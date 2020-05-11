package bugsnag

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kinbiko/jsonassert"
)

func TestSessions(t *testing.T) {
	var (
		apiKey   = "abcd1234abcd1234abcd1234abcd1234"
		payloads = make(chan string, 1)
	)

	checkHeaders := func(t *testing.T, headers http.Header) {
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
		body, _ := ioutil.ReadAll(r.Body)
		payloads <- string(body)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n := Notifier{
		cfg: &Configuration{
			EndpointSessions: ts.URL,
			APIKey:           apiKey,
			AppVersion:       "3.5.1",
			ReleaseStage:     "staging",
			runtimeConstants: runtimeConstants{
				hostname:        "myHost",
				osVersion:       "4.1.12",
				goVersion:       "1.15",
				osName:          "linux innit",
				notifierVersion: "0.1.0",
			},
		},
		sessionChannel:         make(chan *session, 1),
		sessionPublishInterval: time.Microsecond, // Just to make things go a bit faster,
	}
	go n.startSessionTracking()
	n.StartSession(context.Background())

	jsonassert.New(t).Assertf(<-payloads, `{
		"notifier":      { "name": "Alternative Go Notifier", "url": "https://github.com/kinbiko/bugsnag", "version": "0.1.0" },
		"app":           { "releaseStage": "staging", "version": "3.5.1", "duration": "<<PRESENCE>>" },
		"device":        { "osName": "linux innit", "osVersion": "4.1.12", "hostname": "myHost", "runtimeVersions": { "go": "1.15"}, "memStats": "<<PRESENCE>>" },
		"sessionCounts": [{ "startedAt": "<<PRESENCE>>", "sessionsStarted": 1 }]
	}`)
}

func TestSessionAndContext(t *testing.T) {
	for _, tc := range []struct {
		expUnhandledCount, expHandledCount int
		name                               string
		unhandled                          bool
	}{
		{expUnhandledCount: 1, expHandledCount: 0, name: "unhandled", unhandled: true},
		{expUnhandledCount: 0, expHandledCount: 1, name: "handled", unhandled: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			n, err := New(Configuration{APIKey: "abcd1234abcd1234abcd1234abcd1234"})
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
