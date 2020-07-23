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
	cfg := Configuration{
		runtimeConstants: runtimeConstants{
			hostname:  "myHost",
			osVersion: "4.1.12",
			goVersion: "1.15",
			osName:    "linux innit",
		},
	}
	payload, err := json.Marshal(cfg.makeJSONDevice())
	if err != nil {
		t.Fatal(err)
	}

	jsonassert.New(t).Assertf(string(payload), `{
			"memStats": "<<PRESENCE>>",
			"hostname": "myHost",
			"osName": "linux innit",
			"osVersion": "4.1.12",
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

func TestShuttingDown(t *testing.T) {
	t.Run("doesn't panic if invoking StartSession or Notify", func(t *testing.T) {
		n, err := New(Configuration{APIKey: "abcd1234abcd1234abcd1234abcd1234", ReleaseStage: "dev", AppVersion: "1.2.3"})
		if err != nil {
			t.Fatal(err)
		}
		ctx := n.StartSession(context.Background())
		n.Notify(ctx, fmt.Errorf("oooi"))
		n.Close()
	})

	t.Run("doens't panic even if StartSession and Notify are uncalled", func(t *testing.T) {
		n, err := New(Configuration{APIKey: "abcd1234abcd1234abcd1234abcd1234", ReleaseStage: "dev", AppVersion: "1.2.3"})
		if err != nil {
			t.Fatal(err)
		}
		n.Close()
	})
}
