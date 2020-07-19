package bugsnag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Notifier is the key type of this package.
type Notifier struct {
	cfg *Configuration

	// sessions
	sessionCh              chan *session
	sessions               []*session
	sessionPublishInterval time.Duration

	loopOnce sync.Once

	reportCh   chan *report
	shutdownCh chan struct{}
}

type report struct {
	ctx    context.Context
	report *JSONErrorReport
}

// ErrorReportSanitizer allows you to modify the payload being sent to Bugsnag just before it's being sent.
// No further modifications will happen to the payload after this is run.
// You may return a nil Context in order to prevent the payload from being sent at all.
// This context will be attached to the http.Request used for the request to
// Bugsnag, so you are also free to set deadlines etc as you see fit.
// The ctx param provided will be the ctx from the deepest location where
// Wrap is called, falling back to the ctx given to Notify.
// It is recommended to return an unrelated ctx from this method instead of
// a derivative of the input ctx, as the deadline properties etc. of the
// returned ctx is used in the HTTP request to Bugsnag.
type ErrorReportSanitizer func(ctx context.Context, p *JSONErrorReport) context.Context

// New constructs a new Notifier with the given configuration.
func New(cfg Configuration) (*Notifier, error) { //nolint:gocritic // We want to pass by value here as the configuration should be considered immutable
	if cfg.EndpointNotify == "" {
		cfg.EndpointNotify = "https://notify.bugsnag.com"
		cfg.EndpointSessions = "https://sessions.bugsnag.com"
	}
	cfg.runtimeConstants = makeRuntimeConstants()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	const bufChanSize = 16

	return &Notifier{
		cfg: &cfg,

		sessions:               []*session{},
		sessionPublishInterval: time.Minute,

		sessionCh: make(chan *session, bufChanSize),
		reportCh:  make(chan *report, bufChanSize),

		shutdownCh: make(chan struct{}),
	}, nil
}

// Close shuts down the notifier, flushing any unsent reports and sessions.
// Any further calls to StartSession and Notify will panic, so it is your
// responsibility to only call Close at an appropriate time.
func (n *Notifier) Close() {
	// OK, so we have that warning above in the documentation about the panics,
	// but I'd much rather just drop the sessions/reports. I haven't bothered
	// figuring out how to do this yet in a clean (race-condition-free) manner.
	n.shutdownCh <- struct{}{}
}

// Notify reports the given error to Bugsnag.
func (n *Notifier) Notify(ctx context.Context, err error) {
	if err == nil {
		logErr(errors.New("error missing in call to (*bugsnag.Notifier).Notify. no error reported to Bugsnag"))
		return
	}
	n.loopOnce.Do(func() { go n.loop() })
	r := n.makeReport(ctx, err)
	if sanitizer := n.cfg.ErrorReportSanitizer; sanitizer != nil {
		ctx = sanitizer(context.Background(), r.report)
		if ctx == nil {
			// A nil ctx indicates that we should not send the payload.
			// Useful for testing etc.
			return
		}
	}
	n.reportCh <- r
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
	t := time.NewTicker(n.sessionPublishInterval)
	for {
		select {
		case r := <-n.reportCh:
			n.sendErrorReport(r)
		case s := <-n.sessionCh:
			n.sessions = append(n.sessions, s)
		case <-t.C:
			n.flushSessions()
		case <-n.shutdownCh:
			close(n.shutdownCh)

			close(n.reportCh)
			for r := range n.reportCh {
				n.sendErrorReport(r)
			}

			close(n.sessionCh)
			for s := range n.sessionCh {
				n.sessions = append(n.sessions, s)
			}

			t.Stop()
			n.flushSessions()
			return
		}
	}
}

type causer interface {
	Cause() error
}

func (n *Notifier) makeReport(ctx context.Context, err error) *report {
	unhandled := makeUnhandled(err)
	cd, ctx := extractAugmentedContextData(ctx, err, unhandled)
	return &report{
		report: &JSONErrorReport{
			APIKey:   n.cfg.APIKey,
			Notifier: makeNotifier(n.cfg),
			Events: []*JSONEvent{
				{
					PayloadVersion: "5",
					Context:        cd.bContext,
					Unhandled:      unhandled,
					Severity:       makeSeverity(err),
					SeverityReason: &JSONSeverityReason{Type: severityReasonType(err)},
					Exceptions:     makeExceptions(err),
					Breadcrumbs:    cd.breadcrumbs,
					User:           cd.user,
					App:            makeJSONApp(n.cfg),
					Device:         n.cfg.makeJSONDevice(),
					Session:        cd.session,
					Metadata:       cd.metadata,
				},
			},
		},
		ctx: ctx,
	}
}

func (n *Notifier) sendErrorReport(r *report) {
	b, err := json.Marshal(r.report)
	if err != nil {
		logErr(fmt.Errorf("unable to marshal JSON: %w", err))
	}
	req, err := http.NewRequest("POST", n.cfg.EndpointNotify, bytes.NewBuffer(b))
	if err != nil {
		logErr(fmt.Errorf("unable to create new request: %w", err))
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Bugsnag-Api-Key", n.cfg.APIKey)
	req.Header.Add("Bugsnag-Payload-Version", "5")
	req.Header.Add("Bugsnag-Sent-At", time.Now().UTC().Format(time.RFC3339))
	res, err := http.DefaultClient.Do(req.WithContext(r.ctx))
	if err != nil {
		logErr(fmt.Errorf("unable to perform HTTP request: %w", err))
		return
	}
	if err := res.Body.Close(); err != nil {
		logErr(err)
	}
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
	if e := extractLowestBugsnagError(err); e != nil {
		if e.Severity != severityUndetermined {
			return "userSpecifiedSeverity"
		}
		if e.Unhandled {
			prefix = "unhandled"
		}
		if e.Panic {
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
	for i, err := range errs {
		ep := &JSONException{ErrorClass: reflect.TypeOf(err).String(), Message: err.Error()}
		if berr, ok := err.(*Error); ok {
			ep.Stacktrace = berr.stacktrace
		}
		eps[len(errs)-i-1] = ep // reverse the order to match the API
	}
	return eps
}

func makeJSONApp(cfg *Configuration) *JSONApp {
	return &JSONApp{
		Version:      cfg.AppVersion,
		ID:           cfg.runtimeConstants.appID,
		ReleaseStage: cfg.ReleaseStage,
		Duration:     time.Since(cfg.appStartTime).Milliseconds(),
	}
}

func (c *runtimeConstants) makeJSONDevice() *JSONDevice {
	return &JSONDevice{
		Hostname:        c.hostname,
		OSName:          c.osName,
		OSVersion:       c.osVersion,
		MemStats:        memStats(),
		GoroutineCount:  runtime.NumGoroutine(),
		RuntimeVersions: map[string]string{"go": c.goVersion},
	}
}

func memStats() map[string]interface{} {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	b, err := json.Marshal(m)
	if err != nil {
		logErr(fmt.Errorf("unable to marshal memstats: %w", err))
		return nil
	}

	memStats := map[string]interface{}{}
	err = json.Unmarshal(b, &memStats)
	if err != nil {
		logErr(fmt.Errorf("unable to unmarshal memstats into a map: %w", err))
		return nil
	}

	// These are just to long to add and makes is more likely that we'd hit the
	// 1MB limit
	delete(memStats, "PauseNs")
	delete(memStats, "PauseEnd")
	delete(memStats, "BySize")
	return memStats
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

func logErr(err error) {
	fmt.Fprintf(os.Stderr, "ERROR (bugsnag): %s\n", err.Error())
}
