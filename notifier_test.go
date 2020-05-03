package bugsnag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kinbiko/jsonassert"
	pkgerrors "github.com/pkg/errors"
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
			payload, err := json.Marshal(makeAppPayload(tc.cfg))
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
	payload, err := json.Marshal(cfg.makeDevicePayload())
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

func TestUserAndContext(t *testing.T) {
	exp := User{ID: "id", Name: "name", Email: "email"}
	if got := makeUser(WithUser(context.Background(), exp)); *got != exp {
		t.Errorf("expected that when I add '%+v' to the context what I get back ('%+v') should be equal", exp, got)
	}
}

func TestMetadata(t *testing.T) {
	ctx := WithMetadatum(context.Background(), "app", "id", "15011-2")
	ctx = WithMetadata(ctx, "device", map[string]interface{}{"model": "15023-2"})
	md := metadata(ctx)
	if appID, exp := md["app"]["id"], "15011-2"; appID != exp {
		t.Errorf("expected app.id to be '%s' but was '%s'", exp, appID)
	}
	if deviceModel, exp := md["device"]["model"], "15023-2"; deviceModel != exp {
		t.Errorf("expected device.model to be '%s' but was '%s'", exp, deviceModel)
	}
}

func TestMakeExceptions(t *testing.T) {
	n, err := New(Configuration{APIKey: "abcd1234abcd1234abcd1234abcd1234"})
	if err != nil {
		t.Fatal(err)
	}
	// Need to set this, as tests don't count as 'main' packages, which means
	// that the 'github.com/kinbiko/bugsnag' package won't be picked up
	// automatically
	n.cfg.runtimeConstants.modulePath = "github.com/kinbiko/bugsnag"

	ctx := WithBugsnagContext(context.Background(), "/api/user/1523")
	ctx = WithMetadatum(ctx, "tab", "one", "1")

	err = errors.New("1st error (errors.New)")
	err = pkgerrors.Wrap(err, "2nd error (github.com/pkg/errors.Wrap)")
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
				{"file":"<<PRESENCE>>","inProject":false,"lineNumber":909,"method":"testing.tRunner"},
				{"file":"<<PRESENCE>>","inProject":false,"lineNumber":"<<PRESENCE>>","method":"<<PRESENCE>>"}
			]
		}, {
			"errorClass": "*errors.withStack",
			"message": "2nd error (github.com/pkg/errors.Wrap): 1st error (errors.New)",
			"stacktrace": null
		}, {
			"errorClass": "*errors.errorString",
			"message": "1st error (errors.New)",
			"stacktrace": null
		}
	]`)
}
