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
	EventCounts *SessionEventsPayload
}

// StartSession attaches Bugsnag session data to a copy of the given
// context.Context, and returns the new context.Context.
// Records the newly started session and will at some point flush this session.
func (n *Notifier) StartSession(ctx context.Context) context.Context {
	n.sessionOnce.Do(func() { go n.startSessionTracking() })
	sessionID, _ := uuid.NewV4()
	session := &session{
		StartedAt:   time.Now(),
		ID:          sessionID,
		EventCounts: &SessionEventsPayload{},
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
	payload, err := json.Marshal(makeSessionPayload(cfg, sessions))
	if err != nil {
		return fmt.Errorf("unable to marshal json: %v", err)
	}

	req, err := http.NewRequest("POST", cfg.EndpointSessions, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("unable to create request: %v", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Bugsnag-Api-Key", cfg.APIKey)
	req.Header.Add("Bugsnag-Payload-Version", "1.0")
	req.Header.Add("Bugsnag-Sent-At", time.Now().UTC().Format(time.RFC3339))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		_ = res.Body.Close()
		return fmt.Errorf("unable to deliver session: %v", err)
	}
	if s := res.StatusCode; s != http.StatusAccepted {
		_ = res.Body.Close()
		return fmt.Errorf("expected 202 response status, got HTTP %d", s)
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

func makeReportSessionPayload(ctx context.Context, unhandled bool) *SessionPayload {
	if sess := incrementEventCountAndGetSession(ctx, unhandled); sess != nil {
		return &SessionPayload{
			ID:        sess.ID.String(),
			StartedAt: sess.StartedAt.Format(time.RFC3339),
			Events: &SessionEventsPayload{
				Handled:   sess.EventCounts.Handled,
				Unhandled: sess.EventCounts.Unhandled,
			},
		}
	}
	return nil
}

func makeSessionPayload(cfg *Configuration, sessions []*session) *sessionPayload {
	return &sessionPayload{
		Notifier: makeNotifier(),
		App:      makeAppPayload(cfg),
		Device:   cfg.makeDevicePayload(),
		SessionCounts: []sessionCountsPayload{
			{
				// This timestamp assumes that the sessions happen at more or
				// less the same point in time
				StartedAt:       sessions[0].StartedAt.UTC().Format(time.RFC3339),
				SessionsStarted: len(sessions),
			},
		},
	}
}
