package bugsnag

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type session struct {
	ID          string
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
	n.loopOnce.Do(func() { go n.loop() })
	session := &session{
		StartedAt:   time.Now(),
		ID:          uuidv4(),
		EventCounts: &JSONSessionEvents{},
	}
	n.sessionCh <- session
	return context.WithValue(ctx, sessionKey, session)
}

func (n *Notifier) flushSessions() {
	sessions := n.sessions
	n.sessions = nil
	if len(sessions) == 0 {
		return
	}

	if err := n.publishSessions(n.cfg, sessions); err != nil {
		n.cfg.Fallback(fmt.Errorf("unable to publish sessions: %w", err))
	}
}

func (n *Notifier) publishSessions(cfg *Configuration, sessions []*session) error {
	report := n.makeJSONSessionReport(cfg, sessions)
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
			ID:        sess.ID,
			StartedAt: sess.StartedAt.Format(time.RFC3339),
			Events: &JSONSessionEvents{
				Handled:   sess.EventCounts.Handled,
				Unhandled: sess.EventCounts.Unhandled,
			},
		}
	}
	return nil
}

func (n *Notifier) makeJSONSessionReport(cfg *Configuration, sessions []*session) *JSONSessionReport {
	return &JSONSessionReport{
		Notifier: makeNotifier(cfg),
		App:      makeJSONApp(cfg),
		Device:   n.makeJSONDevice(),
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

// uuidv4 returns a randomly generated UUID v4.
// Returns a canonical RFC-4122 string representation of the UUID:
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func uuidv4() string {
	var u [16]byte
	_, _ = io.ReadFull(rand.Reader, u[:])
	u[6] = (u[6] & 0x0f) | (0x04 << 4)  // Version 4
	u[8] = u[8]&(0xff>>2) | (0x02 << 6) // Variation RFC-4122

	buf := make([]byte, 36)
	hex.Encode(buf[0:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], u[10:])

	return string(buf)
}
