package bugsnag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Notifier is the key type of this package.
type Notifier struct {
	cfg *Configuration

	// sessions
	sessionChannel         chan *session
	sessions               []*session
	sessionMutex           sync.Mutex
	sessionPublishInterval time.Duration
	sessionOnce            sync.Once
}

// ErrorReportSanitizer allows you to modify the payload being sent to Bugsnag just before it's being sent.
// No further modifications will happen to the payload after this is run.
// You may return a nil Context in order to prevent the payload from being sent at all.
// This context will be attached to the http.Request used for the request to
// Bugsnag, so you are also free to set deadlines etc as you see fit.
type ErrorReportSanitizer interface {
	SanitizeErrorReport(ctx context.Context, p *JSONErrorReport) context.Context
}

// New constructs a new Notifier with the given configuration
func New(cfg Configuration) (*Notifier, error) { //nolint:gocritic // We want to pass by value here as the configuration should be considered immutable
	if cfg.EndpointNotify == "" {
		cfg.EndpointNotify = "https://notify.bugsnag.com"
		cfg.EndpointSessions = "https://sessions.bugsnag.com"
	}
	cfg.runtimeConstants = makeRuntimeConstants()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &Notifier{
		cfg: &cfg,

		sessionChannel:         make(chan *session, 1),
		sessions:               []*session{},
		sessionMutex:           sync.Mutex{},
		sessionPublishInterval: time.Minute,
	}, nil
}

// Notify reports the given error to Bugsnag.
func (n *Notifier) Notify(ctx context.Context, err error) {
	if err == nil {
		logErr(errors.New("error missing in call to (*bugsnag.Notifier).Notify. no error reported to Bugsnag"))
		return
	}
	report := n.makeReport(ctx, err)
	if sanitizer := n.cfg.ErrorReportSanitizer; sanitizer != nil {
		ctx = sanitizer.SanitizeErrorReport(context.Background(), report)
		if ctx == nil {
			// A nil ctx indicates that we should not send the payload.
			// Useful for testing etc.
			return
		}
	}
	b, err := json.Marshal(report)
	if err != nil {
		logErr(fmt.Errorf("unable to marshal JSON: %w", err))
	}
	req, err := http.NewRequest("POST", n.cfg.EndpointNotify, bytes.NewBuffer(b))
	if err != nil {
		logErr(fmt.Errorf("unable to create new request: %w", err))
	}
	res, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		logErr(fmt.Errorf("unable to perform HTTP request: %w", err))
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			logErr(err)
		}
	}()
}
