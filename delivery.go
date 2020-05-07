package bugsnag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"time"
)

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

type causer interface {
	Cause() error
}

type ctxData struct {
	bContext    string
	breadcrumbs []*JSONBreadcrumb
	user        *User
	session     *JSONSession
	metadata    map[string]map[string]interface{}
}

func (n *Notifier) makeReport(ctx context.Context, err error) *JSONErrorReport {
	return &JSONErrorReport{
		APIKey:   n.cfg.APIKey,
		Notifier: makeNotifier(),
		Events:   makeEvents(ctx, n.cfg, err),
	}
}

func makeEvents(ctx context.Context, cfg *Configuration, err error) []*JSONEvent {
	unhandled := makeUnhandled(err)
	ctxData := extractInnermostCtx(ctx, err, unhandled)
	return []*JSONEvent{
		{
			PayloadVersion: "5",
			Context:        ctxData.bContext,
			Unhandled:      unhandled,
			Severity:       makeSeverity(err),
			SeverityReason: &JSONSeverityReason{Type: severityReasonType(err)},
			Exceptions:     makeExceptions(err),
			Breadcrumbs:    ctxData.breadcrumbs,
			User:           ctxData.user,
			App:            makeJSONApp(cfg),
			Device:         cfg.makeJSONDevice(),
			Session:        ctxData.session,
			Metadata:       ctxData.metadata,
		},
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

func extractInnermostCtx(ctx context.Context, err error, unhandled bool) *ctxData {
	data := &ctxData{
		bContext:    makeContext(ctx),
		breadcrumbs: makeBreadcrumbs(ctx),
		user:        makeUser(ctx),
		session:     makeJSONSession(ctx, unhandled),
		metadata:    Metadata(ctx),
	}
	var e error = err
	for {
		if berr, ok := e.(*Error); ok {
			ctx = berr.ctx
			if ctx != nil {
				data.updateFromCtx(ctx, unhandled)
			}
		}
		e = errors.Unwrap(e)
		if e == nil {
			break
		}
	}

	if data.bContext == "" {
		data.bContext = err.Error()
	}
	return data
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
		RuntimeVersions: map[string]string{"go": c.goVersion},
	}
}

func makeUser(ctx context.Context) *User {
	u := ctx.Value(userKey)
	if u == nil {
		return nil
	}
	user, ok := u.(*User)
	if !ok {
		return nil
	}
	return user
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

func logErr(err error) {
	fmt.Fprintf(os.Stderr, "ERROR (bugsnag): %s\n", err.Error())
}

func makeNotifier() *JSONNotifier {
	return &JSONNotifier{
		Name:    "Alternative Go Notifier",
		URL:     "https://github.com/kinbiko/bugsnag",
		Version: notifierVersion,
	}
}

func (data *ctxData) updateFromCtx(ctx context.Context, unhandled bool) {
	if dataBContext := makeContext(ctx); dataBContext != "" {
		data.bContext = dataBContext
	}
	if dataBreadcrumbs := makeBreadcrumbs(ctx); dataBreadcrumbs != nil {
		data.breadcrumbs = dataBreadcrumbs
	}
	if dataUser := makeUser(ctx); dataUser != nil {
		data.user = dataUser
	}
	if dataSession := makeJSONSession(ctx, unhandled); dataSession != nil {
		data.session = dataSession
	}

	dataMetadata := Metadata(ctx)
	if dataMetadata == nil {
		return
	}
	if data.metadata == nil {
		data.metadata = map[string]map[string]interface{}{}
	}
	for tab, kvps := range dataMetadata {
		for k, v := range kvps {
			data.metadata[tab][k] = v
		}
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

func makeContext(ctx context.Context) string {
	if v, ok := ctx.Value(contextKey).(string); ok {
		return v
	}
	return ""
}
