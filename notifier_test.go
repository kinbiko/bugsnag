package bugsnag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kinbiko/jsonassert"
)

func TestApp(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  *Configuration
		exp  string
	}{
		{name: "no optional config", cfg: &Configuration{}, exp: `{ "duration": 5000 }`},
		{
			name: "all optional config",
			cfg:  &Configuration{AppVersion: "1.5.2", ReleaseStage: "staging"},
			exp:  `{ "duration": 5000, "releaseStage": "staging", "version": "1.5.2" }`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.cfg.appStartTime = time.Now().Add(-5 * time.Second)
			payload, err := json.Marshal(makeJSONApp(tc.cfg))
			if err != nil {
				t.Fatal(err)
			}
			jsonassert.New(t).Assertf(string(payload), tc.exp)
		})
	}
}

func TestDevice(t *testing.T) {
	n, err := New(Configuration{APIKey: "abcd1234abcd1234abcd1234abcd1234", ReleaseStage: "dev", AppVersion: "1.2.3"})
	if err != nil {
		t.Fatal(err)
	}
	n.cfg.runtimeConstants = runtimeConstants{
		hostname:  "myHost",
		osVersion: "4.1.12",
		goVersion: "1.15",
		osName:    "linux innit",
	}
	payload, err := json.Marshal(n.makeJSONDevice())
	if err != nil {
		t.Fatal(err)
	}

	jsonassert.New(t).Assertf(string(payload), `{
			"memStats": "<<PRESENCE>>",
			"hostname": "myHost",
			"osName": "linux innit",
			"osVersion": "4.1.12",
			"goroutineCount": "<<PRESENCE>>",
			"runtimeVersions": {"go": "1.15"}
		}`)
}

func TestMakeExceptions(t *testing.T) {
	n, err := New(Configuration{APIKey: "abcd1234abcd1234abcd1234abcd1234", ReleaseStage: "dev", AppVersion: "1.2.3"})
	if err != nil {
		t.Fatal(err)
	}

	ctx := WithBugsnagContext(context.Background(), "/api/user/1523")
	ctx = WithMetadatum(ctx, "tab", "one", "1")

	err = errors.New("1st error (errors.New)")
	err = fmt.Errorf("2nd error (github.com/pkg/errors.Wrap): %w", err)
	err = n.Wrap(WithMetadatum(ctx, "tab", "two", "2"), err, "3rd error (%s.(*Notifier).Wrap)", "bugsnag")
	err = fmt.Errorf("4th error (fmt.Errorf('percent-w')): %w", err)

	eps := makeExceptions(err)
	payload, _ := json.Marshal(eps)

	jsonassert.New(t).Assertf(string(payload), `[
		{
			"errorClass": "*fmt.wrapError",
			"message": "4th error (fmt.Errorf('percent-w')): 3rd error (bugsnag.(*Notifier).Wrap): 2nd error (github.com/pkg/errors.Wrap): 1st error (errors.New)",
			"stacktrace": null
		}, {
			"errorClass": "*bugsnag.Error",
			"message": "3rd error (bugsnag.(*Notifier).Wrap): 2nd error (github.com/pkg/errors.Wrap): 1st error (errors.New)",
			"stacktrace": [
				{"file":"<<PRESENCE>>","inProject":false,"lineNumber":"<<PRESENCE>>","method":"testing.tRunner"},
				{"file":"<<PRESENCE>>","inProject":false,"lineNumber":"<<PRESENCE>>","method":"<<PRESENCE>>"}
			]
		}, {
			"errorClass": "*fmt.wrapError",
			"message": "2nd error (github.com/pkg/errors.Wrap): 1st error (errors.New)",
			"stacktrace": null
		}, {
			"errorClass": "*errors.errorString",
			"message": "1st error (errors.New)",
			"stacktrace": null
		}
	]`)
}

func TestNoPanicsWhenShuttingDown(t *testing.T) {
	cfg := Configuration{
		APIKey:       "abcd1234abcd1234abcd1234abcd1234",
		ReleaseStage: "dev",
		AppVersion:   "1.2.3",
		// Dummy endpoints to avoid accidentally sending payloads to Bugsnag.
		// Note that the panics are most likely to occur *after* any execution
		// of the Sanitizer functions, and thus setting different endpoints
		// that will fail are more appropriate.
		EndpointNotify:   "http://0.0.0.0:1234",
		EndpointSessions: "http://0.0.0.0:1234",
	}

	t.Run("doesn't panic if invoking StartSession or Notify", func(t *testing.T) {
		n, err := New(cfg)
		if err != nil {
			t.Fatal(err)
		}
		ctx := n.StartSession(context.Background())
		n.Notify(ctx, fmt.Errorf("oooi"))
		n.Close()
	})

	t.Run("StartSession and Notify are uncalled", func(t *testing.T) {
		n, err := New(cfg)
		if err != nil {
			t.Fatal(err)
		}
		n.Close()
	})

	t.Run("Notify after Close invokes InternalErrorCallback", func(t *testing.T) {
		var got error
		cfg.InternalErrorCallback = func(err error) { got = err }

		n, err := New(cfg)
		if err != nil {
			t.Fatal(err)
		}
		n.Close()
		n.Notify(context.Background(), fmt.Errorf("oops"))

		if got == nil {
			t.Fatal("expected error but got none")
		}

		if exp, got := "did you invoke Notify", got.Error(); !strings.Contains(got, exp) {
			t.Errorf("expected error message containing %s but got %s", exp, got)
		}
	})

	t.Run("StartSession after Close invokes InternalErrorCallback", func(t *testing.T) {
		var got error
		cfg.InternalErrorCallback = func(err error) { got = err }

		n, err := New(cfg)
		if err != nil {
			t.Fatal(err)
		}
		n.Close()
		n.StartSession(context.Background())

		if got == nil {
			t.Fatal("expected error but got none")
		}

		if exp, got := "did you invoke StartSession", got.Error(); !strings.Contains(got, exp) {
			t.Errorf("expected error message containing %s but got %s", exp, got)
		}
	})

	t.Run("Close after Close invokes InternalErrorCallback", func(t *testing.T) {
		var got error
		cfg.InternalErrorCallback = func(err error) { got = err }

		n, err := New(cfg)
		if err != nil {
			t.Fatal(err)
		}
		n.Close()
		n.Close()

		if got == nil {
			t.Fatal("expected error but got none")
		}

		if exp, got := "did you invoke Close", got.Error(); !strings.Contains(got, exp) {
			t.Errorf("expected error message containing %s but got %s", exp, got)
		}
	})
}

func TestInternalErrorCallback(t *testing.T) {
	t.Run("gets invoked when set", func(t *testing.T) {
		var got error
		n, err := New(Configuration{
			APIKey:       "abcd1234abcd1234abcd1234abcd1234",
			ReleaseStage: "dev",
			AppVersion:   "1.2.3",
			InternalErrorCallback: func(err error) {
				got = err
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		n.Notify(nil, nil) //nolint:staticcheck // Need to verify that the notifier doesn't die

		if got == nil {
			t.Error("expected an error in the error callback but got none")
		}
	})

	t.Run("doesn't panic when not set", func(t *testing.T) {
		n, err := New(Configuration{APIKey: "abcd1234abcd1234abcd1234abcd1234", ReleaseStage: "dev", AppVersion: "1.2.3"})
		if err != nil {
			t.Fatal(err)
		}

		n.Notify(nil, nil) //nolint:staticcheck // Need to verify that the notifier doesn't die
	})
}
