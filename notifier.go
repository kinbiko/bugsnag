package bugsnag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"reflect"
	"runtime"
	"runtime/metrics"
	"strings"
	"sync"
	"time"
)

// Notifier is the key type of this package, and exposes methods for reporting
// errors and tracking sessions.
type Notifier struct {
	cfg *Configuration

	sessions               []*session
	sessionPublishInterval time.Duration

	reportCh       chan *JSONErrorReport
	sessionCh      chan *session
	shutdownCh     chan struct{}
	shutdownDoneCh chan struct{}
	loopOnce       sync.Once
}

// ErrorReportSanitizer allows you to modify the payload being sent to Bugsnag just before it's being sent.
// The ctx param provided will be the ctx from the deepest location where
// Wrap is called, falling back to the ctx given to Notify.
// You may return a non-nil error in order to prevent the payload from being sent at all.
// This error is then forwarded to the InternalErrorCallback.
// No further modifications to the payload will happen to the payload after this is run.
type ErrorReportSanitizer func(ctx context.Context, p *JSONErrorReport) error

// New constructs a new Notifier with the given configuration.
// You should call Close before shutting down your app in order to ensure that
// all sessions and reports have been sent.
func New(config Configuration) (*Notifier, error) { //nolint:gocritic // We want to pass by value here as the configuration should be considered immutable
	cfg := &config
	cfg.populateDefaults()
	cfg.runtimeConstants = makeRuntimeConstants()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	const bufChanSize = 16

	return &Notifier{
		cfg: cfg,

		sessions:               []*session{},
		sessionPublishInterval: time.Minute,

		sessionCh: make(chan *session, bufChanSize),
		reportCh:  make(chan *JSONErrorReport, bufChanSize),

		shutdownCh:     make(chan struct{}),
		shutdownDoneCh: make(chan struct{}),

		loopOnce: sync.Once{},
	}, nil
}

// Close shuts down the notifier, flushing any unsent reports and sessions.
// Any further calls to StartSession and Notify will call the
// InternalErrorCallback, if provided, with an error.
func (n *Notifier) Close() {
	// Ideally we wouldn't need this guard, but it's the best way I can see to
	// prevent this package from ever panicking.
	defer n.guard("Close")

	// Need to ensure that the loop is running in the first place to not block
	// if Close is called before StartSession/Notify.
	n.loopOnce.Do(func() { go n.loop() })

	// OK, so we have that warning above in the documentation about the panics,
	// but I'd much rather just drop the sessions/reports. I haven't bothered
	// figuring out how to do this yet in a clean (race-condition-free) manner.
	n.shutdownCh <- struct{}{}

	<-n.shutdownDoneCh
}

// Notify reports the given error to Bugsnag.
// Extracts diagnostic data from the context and any *bugsnag.Error errors,
// including wrapped errors.
// Invokes the ErrorReportSanitizer, if set, before sending the error report.
func (n *Notifier) Notify(ctx context.Context, err error) {
	// Ideally we wouldn't need this guard, but it's the best way I can see to
	// prevent this package from ever panicking.
	defer n.guard("Notify")

	if err == nil {
		n.cfg.InternalErrorCallback(errors.New("error missing in call to (*bugsnag.Notifier).Notify. no error reported to Bugsnag"))
		return
	}
	n.loopOnce.Do(func() { go n.loop() })

	var report *JSONErrorReport
	report, ctx = n.makeReport(ctx, err)
	if sErr := n.cfg.ErrorReportSanitizer(ctx, report); sErr != nil {
		n.cfg.InternalErrorCallback(sErr)
		return
	}
	n.reportCh <- report
}

type severity int

const (
	severityUndetermined severity = iota
	// SeverityInfo indicates that the severity of the Error is "info"
	SeverityInfo
	// SeverityWarning indicates that the severity of the Error is "warning"
	SeverityWarning
	// SeverityError indicates that the severity of the Error is "error"
	SeverityError
)

// loop is intended to be an infinitely running goroutine that periodically (as
// defined by sessionPublishInterval) sends sessions, and sends reports as they
// come in. This loop ensures that a spike in errors doesn't consume the upload
// bandwidth for highly concurrent applications.
func (n *Notifier) loop() {
	ticker := time.NewTicker(n.sessionPublishInterval)
	for {
		select {
		case r := <-n.reportCh:
			if err := n.sendErrorReport(r); err != nil {
				n.cfg.InternalErrorCallback(fmt.Errorf("unable to send error report: %w", err))
			}
		case s := <-n.sessionCh:
			n.sessions = append(n.sessions, s)
		case <-ticker.C:
			n.flushSessions()
		case <-n.shutdownCh:
			n.shutdown(ticker)
			close(n.shutdownDoneCh)
			return
		}
	}
}

func (n *Notifier) shutdown(ticker *time.Ticker) {
	close(n.shutdownCh)

	close(n.reportCh)
	for r := range n.reportCh {
		if err := n.sendErrorReport(r); err != nil {
			n.cfg.InternalErrorCallback(fmt.Errorf("unable to send error report when closing Notifier: %w", err))
		}
	}

	close(n.sessionCh)
	for s := range n.sessionCh {
		n.sessions = append(n.sessions, s)
	}

	ticker.Stop()
	n.flushSessions()
}

type causer interface {
	Cause() error
}

func (n *Notifier) makeReport(ctx context.Context, err error) (*JSONErrorReport, context.Context) {
	unhandled := makeUnhandled(err)
	contextData, augmentedCtx := extractAugmentedContextData(ctx, err, unhandled)
	return &JSONErrorReport{
		APIKey:   n.cfg.APIKey,
		Notifier: makeNotifier(n.cfg),
		Events: []*JSONEvent{
			{
				PayloadVersion: "5",
				Context:        contextData.bContext,
				Unhandled:      unhandled,
				Severity:       makeSeverity(err),
				SeverityReason: &JSONSeverityReason{Type: severityReasonType(err)},
				Exceptions:     makeExceptions(err),
				Breadcrumbs:    contextData.breadcrumbs,
				User:           contextData.user,
				App:            makeJSONApp(n.cfg),
				Device:         n.makeJSONDevice(),
				Session:        contextData.session,
				Metadata:       contextData.metadata,
			},
		},
	}, augmentedCtx
}

func (n *Notifier) sendErrorReport(r *JSONErrorReport) error {
	b, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("unable to marshal JSON: %w", err)
	}
	req, err := http.NewRequest("POST", n.cfg.EndpointNotify, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("unable to create new request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Bugsnag-Api-Key", n.cfg.APIKey)
	req.Header.Add("Bugsnag-Payload-Version", "5")
	req.Header.Add("Bugsnag-Sent-At", time.Now().UTC().Format(time.RFC3339))
	// Note we're using a background context here to avoid confusing bugs ala
	// "my errors aren't being sent to Bugsnag" due to users not realizing that
	// the context they provided (which usually is derived from a request) has
	// already been canceled by the time that this request is being made.
	res, err := http.DefaultClient.Do(req.WithContext(context.Background()))
	if err != nil {
		return fmt.Errorf("unable to perform HTTP request: %w", err)
	}
	if err := res.Body.Close(); err != nil {
		return fmt.Errorf("unable to close the response body: %w", err)
	}
	return nil
}

func makeUnhandled(err error) bool {
	for {
		if berr, ok := err.(*Error); ok && berr.Unhandled {
			return true
		}
		err = errors.Unwrap(err)
		if err == nil {
			break
		}
	}
	return false
}

func makeSeverity(err error) string {
	if berr := extractLowestBugsnagError(err); berr != nil {
		if s := berr.Severity; s != severityUndetermined {
			return []string{"undetermined", "info", "warning", "error"}[s]
		}
		if berr.Unhandled || berr.Panic {
			return "error"
		}
	}
	return "warning"
}

func severityReasonType(err error) string {
	var (
		prefix = "handled"
		suffix = "Exception"
	)
	if lowestBugsnagErr := extractLowestBugsnagError(err); lowestBugsnagErr != nil {
		if lowestBugsnagErr.Severity != severityUndetermined {
			return "userSpecifiedSeverity"
		}
		if lowestBugsnagErr.Unhandled {
			prefix = "unhandled"
		}
		if lowestBugsnagErr.Panic {
			suffix = "Panic"
		}
	}
	return prefix + suffix
}

func makeExceptions(err error) []*JSONException {
	var errs []error
	for {
		if err == nil {
			break
		}
		errs = append([]error{err}, errs...)

		switch e := err.(type) {
		case causer:
			// the github.com/pkg/errors package nests its own internal errors,
			// which makes it look like its wrapped twice
			err = e.Cause()
			if e, ok := err.(causer); ok {
				err = e.Cause()
			}
		default:
			err = errors.Unwrap(err)
		}
	}

	eps := make([]*JSONException, len(errs))
	for i, err := range errs { // nolint:varnamelen // indexes are conventionally i
		var stacktrace []*JSONStackframe
		if berr, ok := err.(*Error); ok {
			stacktrace = berr.stacktrace
		}
		// reverse the order to match the API
		eps[len(errs)-i-1] = &JSONException{
			ErrorClass: reflect.TypeOf(err).String(),
			Message:    err.Error(),
			Stacktrace: stacktrace,
		}
	}
	return eps
}

func makeJSONApp(cfg *Configuration) *JSONApp {
	return &JSONApp{
		ID:           cfg.runtimeConstants.appID,
		Version:      cfg.AppVersion,
		ReleaseStage: cfg.ReleaseStage,
		Type:         "",
		Duration:     time.Since(cfg.appStartTime).Milliseconds(),
	}
}

func (n *Notifier) makeJSONDevice() *JSONDevice {
	return &JSONDevice{
		Hostname:        n.cfg.hostname,
		OSName:          n.cfg.osName,
		OSVersion:       n.cfg.osVersion,
		RuntimeMetrics:  runtimeMetrics(),
		GoroutineCount:  runtime.NumGoroutine(),
		RuntimeVersions: map[string]string{"go": n.cfg.goVersion},
	}
}

func runtimeMetrics() map[string]interface{} {
	descs := metrics.All()
	samples := make([]metrics.Sample, len(descs))
	for i := range samples {
		samples[i].Name = descs[i].Name
	}

	metrics.Read(samples)

	runtimeMetrics := map[string]interface{}{}
	for _, sample := range samples {
		switch sample.Value.Kind() {
		case metrics.KindUint64:
			runtimeMetrics[sample.Name] = sample.Value.Uint64()
		case metrics.KindFloat64:
			runtimeMetrics[sample.Name] = sample.Value.Float64()
		case metrics.KindBad:
			// This should never happen because all metrics are supported by construction.
		case metrics.KindFloat64Histogram:
			// Ignore histograms as they contain too much data and we're likely
			// to hit the 1MB limit, and wouldn't want to try and do any
			// analysis on it at runtime for performance reasons.
		default:
			// This block may get invoked if there's a new metric kind being
			// added in future Go versions. Worst case scenario, we miss out
			// on a new metric.
		}
	}

	return runtimeMetrics
}

// osVersion is only available on unix-like systems as it depends on the
// 'uname' command.
func osVersion() string {
	if b, err := exec.Command("uname", "-r").Output(); err == nil {
		return strings.TrimSpace(string(b))
	}
	return ""
}

func makeNotifier(cfg *Configuration) *JSONNotifier {
	return &JSONNotifier{
		Name:    "Alternative Go Notifier",
		URL:     "https://github.com/kinbiko/bugsnag",
		Version: cfg.runtimeConstants.notifierVersion,
	}
}

func extractLowestBugsnagError(err error) *Error {
	var berr *Error
	for {
		if b, ok := err.(*Error); ok {
			berr = b
		}
		err = errors.Unwrap(err)
		if err == nil {
			break
		}
	}
	return berr
}

func (n *Notifier) guard(method string) {
	if p := recover(); p != nil {
		n.cfg.InternalErrorCallback(fmt.Errorf("panic when calling %s (did you invoke %s after calling Close?): %v", method, method, p))
	}
}
