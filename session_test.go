package bugsnag

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kinbiko/jsonassert"
)

func TestSessions(t *testing.T) {
	t.Parallel()
	var (
		apiKey   = "abcd1234abcd1234abcd1234abcd1234"
		payloads = make(chan string, 1)
	)

	checkHeaders := func(t *testing.T, headers http.Header) {
		t.Helper()
		for _, tc := range []struct{ name, expected string }{
			{name: "Bugsnag-Payload-Version", expected: "1.0"},
			{name: "Content-Type", expected: "application/json"},
			{name: "Bugsnag-Api-Key", expected: apiKey},
		} {
			if got := headers[tc.name][0]; tc.expected != got {
				t.Errorf("Expected header '%s' to be '%s' but was '%s'", tc.name, tc.expected, got)
			}
		}
		if headers["Bugsnag-Sent-At"][0] == "" {
			t.Error("Expected header 'Bugsnag-Sent-At' to be non-empty but was empty")
		}
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checkHeaders(t, r.Header)
		body, _ := io.ReadAll(r.Body)
		payloads <- string(body)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n, err := New(Configuration{
		EndpointSessions: ts.URL,
		EndpointNotify:   ts.URL,
		APIKey:           apiKey,
		AppVersion:       "3.5.1",
		ReleaseStage:     "staging",
	})
	if err != nil {
		t.Fatal(err)
	}
	n.cfg.runtimeConstants = runtimeConstants{
		hostname:        "myHost",
		osVersion:       "4.1.12",
		goVersion:       "1.15",
		osName:          "linux innit",
		notifierVersion: "0.1.0",
	}
	n.sessionPublishInterval = time.Microsecond // Just to make things go a bit faster,
	n.StartSession(context.Background())

	jsonassert.New(t).Assertf(<-payloads, `{
		"notifier":      { "name": "Alternative Go Notifier", "url": "https://github.com/kinbiko/bugsnag", "version": "0.1.0" },
		"app":           { "releaseStage": "staging", "version": "3.5.1", "duration": "<<PRESENCE>>" },
		"device":        { "osName": "linux innit", "osVersion": "4.1.12", "hostname": "myHost", "runtimeVersions": { "go": "1.15"}, "goroutineCount": "<<PRESENCE>>", "runtimeMetrics": "<<PRESENCE>>" },
		"sessionCounts": [{ "startedAt": "<<PRESENCE>>", "sessionsStarted": 1 }]
	}`)
}

func TestSessionAndContext(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		expUnhandledCount, expHandledCount int
		name                               string
		unhandled                          bool
	}{
		{expUnhandledCount: 1, expHandledCount: 0, name: "unhandled", unhandled: true},
		{expUnhandledCount: 0, expHandledCount: 1, name: "handled", unhandled: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			n, err := New(Configuration{APIKey: "abcd1234abcd1234abcd1234abcd1234", ReleaseStage: "dev", AppVersion: "1.2.3"})
			if err != nil {
				t.Fatal(err)
			}
			ctx := n.StartSession(context.Background())
			s := incrementEventCountAndGetSession(ctx, tc.unhandled)
			if got := len(s.ID); got == 0 {
				t.Error("expected session ID to be set but was empty")
			}
			if got := s.StartedAt; got.IsZero() {
				t.Error("expected session StartedAt to be set but was empty")
			}
			got := s.EventCounts
			if got == nil {
				t.Fatal("expected EventCounts to be set but wasn't")
			}
			if got := s.EventCounts.Unhandled; got != tc.expUnhandledCount {
				t.Errorf("expected EventCounts.Unhandled to be %d but %d", tc.expUnhandledCount, got)
			}
			if got := s.EventCounts.Handled; got != tc.expHandledCount {
				t.Errorf("expected EventCounts.Handled to be %d but %d", tc.expHandledCount, got)
			}
		})
	}
}
