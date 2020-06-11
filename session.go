package bugsnag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	uuid "github.com/gofrs/uuid"
)

type session struct {
	ID          uuid.UUID
	StartedAt   time.Time
	EventCounts *JSONSessionEvents
}

// SessionReportSanitizer allows you to modify the payload being sent to Bugsnag just before it's being sent.
// No further modifications will happen to the payload after this is run.
// You may return a nil Context in order to prevent the payload from being sent at all.
// This context will be attached to the http.Request used for the request to
// Bugsnag, so you are also free to set deadlines etc as you see fit.
type SessionReportSanitizer func(p *JSONSessionReport) context.Context

// StartSession attaches Bugsnag session data to a copy of the given
// context.Context, and returns the new context.Context.
// Records the newly started session and will at some point flush this session.
func (n *Notifier) StartSession(ctx context.Context) context.Context {
	n.sessionOnce.Do(func() { go n.startSessionTracking() })
	sessionID, _ := uuid.NewV4()
	session := &session{
		StartedAt:   time.Now(),
		ID:          sessionID,
		EventCounts: &JSONSessionEvents{},
	}
	n.sessionChannel <- session
	return context.WithValue(ctx, sessionKey, session)
}

func (n *Notifier) startSessionTracking() {
	t := time.NewTicker(n.sessionPublishInterval)
	for {
		select {
		case session := <-n.sessionChannel:
			n.sessionMutex.Lock()
			n.sessions = append(n.sessions, session)
			n.sessionMutex.Unlock()
		case <-t.C:
			go n.flushSessions()
		}
	}
}

func (n *Notifier) flushSessions() {
	n.sessionMutex.Lock()
	defer n.sessionMutex.Unlock()

	sessions := n.sessions
	n.sessions = nil
	if len(sessions) == 0 {
		return
	}

	go func() {
		if err := n.publishSessions(n.cfg, sessions); err != nil {
			logErr(err)
		}
	}()
}

func (n *Notifier) publishSessions(cfg *Configuration, sessions []*session) error {
	report := makeJSONSessionReport(cfg, sessions)
	ctx := context.Background()
	if sanitizer := n.cfg.SessionReportSanitizer; sanitizer != nil {
		ctx = sanitizer(report)
		if ctx == nil {
			// A nil ctx indicates that we should not send the payload.
			// Useful for testing etc.
			return nil
		}
	}
	payload, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("unable to marshal json: %w", err)
	}

	req, err := http.NewRequest("POST", cfg.EndpointSessions, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Bugsnag-Api-Key", cfg.APIKey)
	req.Header.Add("Bugsnag-Payload-Version", "1.0")
	req.Header.Add("Bugsnag-Sent-At", time.Now().UTC().Format(time.RFC3339))

	res, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("unable to deliver session: %w", err)
	}
	return res.Body.Close()
}

func incrementEventCountAndGetSession(ctx context.Context, unhandled bool) *session {
	s := ctx.Value(sessionKey)
	if s == nil {
		return nil
	}
	session, ok := s.(*session)
	if !ok {
		return nil
	}

	if unhandled {
		session.EventCounts.Unhandled++
	} else {
		session.EventCounts.Handled++
	}
	return session
}

func makeJSONSession(ctx context.Context, unhandled bool) *JSONSession {
	if sess := incrementEventCountAndGetSession(ctx, unhandled); sess != nil {
		return &JSONSession{
			ID:        sess.ID.String(),
			StartedAt: sess.StartedAt.Format(time.RFC3339),
			Events: &JSONSessionEvents{
				Handled:   sess.EventCounts.Handled,
				Unhandled: sess.EventCounts.Unhandled,
			},
		}
	}
	return nil
}

func makeJSONSessionReport(cfg *Configuration, sessions []*session) *JSONSessionReport {
	return &JSONSessionReport{
		Notifier: makeNotifier(cfg),
		App:      makeJSONApp(cfg),
		Device:   cfg.makeJSONDevice(),
		SessionCounts: []JSONSessionCounts{
			{
				// This timestamp assumes that the sessions happen at more or
				// less the same point in time
				StartedAt:       sessions[0].StartedAt.UTC().Format(time.RFC3339),
				SessionsStarted: len(sessions),
			},
		},
	}
}
