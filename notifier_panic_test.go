//go:build !race
// +build !race

package bugsnag

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// This test intentionally checks for the absence of user-facing panics when
// attempting to write to a closed channel.
// Therefore, the test is only run when the race detector is not running.
func TestNoPanicsWhenShuttingDown(t *testing.T) {
	t.Parallel()
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
		t.Parallel()
		n, err := New(cfg)
		if err != nil {
			t.Fatal(err)
		}
		ctx := n.StartSession(context.Background())
		n.Notify(ctx, errors.New("oooi"))
		n.Close()
	})

	t.Run("StartSession and Notify are uncalled", func(t *testing.T) {
		t.Parallel()
		n, err := New(cfg)
		if err != nil {
			t.Fatal(err)
		}
		n.Close()
	})

	t.Run("Notify after Close invokes InternalErrorCallback", func(t *testing.T) {
		t.Parallel()
		var got error
		cfg.InternalErrorCallback = func(err error) { got = err }

		n, err := New(cfg)
		if err != nil {
			t.Fatal(err)
		}
		n.Close()
		n.Notify(context.Background(), errors.New("oops"))

		if got == nil {
			t.Fatal("expected error but got none")
		}

		if exp, got := "did you invoke Notify", got.Error(); !strings.Contains(got, exp) {
			t.Errorf("expected error message containing %s but got %s", exp, got)
		}
	})

	t.Run("StartSession after Close invokes InternalErrorCallback", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
		var got error
		cfg := cfg
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
