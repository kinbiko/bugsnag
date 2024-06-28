package bugsnag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kinbiko/jsonassert"
)

func TestApp(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
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
			"runtimeMetrics": "<<PRESENCE>>",
			"hostname": "myHost",
			"osName": "linux innit",
			"osVersion": "4.1.12",
			"goroutineCount": "<<PRESENCE>>",
			"runtimeVersions": {"go": "1.15"}
		}`)
}

func TestMakeExceptions(t *testing.T) {
	t.Parallel()
	n, err := New(Configuration{APIKey: "abcd1234abcd1234abcd1234abcd1234", ReleaseStage: "dev", AppVersion: "1.2.3"})
	if err != nil {
		t.Fatal(err)
	}

	ctx := n.WithBugsnagContext(context.Background(), "/api/user/1523")
	ctx = n.WithMetadatum(ctx, "tab", "one", "1")

	err = errors.New("1st error (errors.New)")
	err = fmt.Errorf("2nd error (github.com/pkg/errors.Wrap): %w", err)
	err = n.Wrap(n.WithMetadatum(ctx, "tab", "two", "2"), err, "3rd error (%s.(*Notifier).Wrap)", "bugsnag")
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

func TestInternalErrorCallback(t *testing.T) {
	t.Parallel()
	t.Run("gets invoked when set", func(t *testing.T) {
		t.Parallel()
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
		n.Notify(nil, nil)

		if got == nil {
			t.Error("expected an error in the error callback but got none")
		}
	})

	t.Run("doesn't panic when not set", func(t *testing.T) {
		t.Parallel()
		n, err := New(Configuration{APIKey: "abcd1234abcd1234abcd1234abcd1234", ReleaseStage: "dev", AppVersion: "1.2.3"})
		if err != nil {
			t.Fatal(err)
		}

		n.Notify(nil, nil)
	})
}
