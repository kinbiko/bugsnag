package builds_test

import (
	"strings"
	"testing"

	"github.com/kinbiko/bugsnag/builds"
)

func TestValidate(t *testing.T) {
	makeValidReq := func() *builds.JSONBuildRequest {
		return &builds.JSONBuildRequest{
			APIKey:       "1234abcd1234abcd1234abcd1234abcd",
			AppVersion:   "1.5.2",
			ReleaseStage: "staging",
			BuilderName:  "River Tam",
			Metadata: map[string]string{
				"Tickets": "JIRA-1234, JIRA-4321",
			},
			SourceControl: &builds.JSONSourceControl{
				Provider:   "github",
				Repository: "https://github.com/kinbiko/bugsnag",
				Revision:   "9fc0b224985fc09d1ced97e51a0e8f166f1d190a",
			},
			AppVersionCode:    33,
			AppBundleVersion:  "42.3",
			AutoAssignRelease: true,
		}
	}

	mustContain := func(t *testing.T, err error, strs ...string) {
		t.Helper()
		if err == nil {
			t.Fatalf("expected error but got nil")
		}
		errMsg := err.Error()
		for _, str := range strs {
			if !strings.Contains(errMsg, str) {
				t.Errorf("expected error message to contain '%s' but was:\n%s", str, errMsg)
			}
		}
	}

	t.Run("API key", func(t *testing.T) {
		r := makeValidReq()
		r.APIKey = ""
		mustContain(t, r.Validate(), "APIKey", "32 hex")

		r = makeValidReq()
		r.APIKey = "1234567890"
		mustContain(t, r.Validate(), "APIKey", "32 hex")

		r = makeValidReq()
		r.APIKey = "123456789012345678901234567890ZZ"
		mustContain(t, r.Validate(), "APIKey", "32 hex")
	})

	t.Run("App version", func(t *testing.T) {
		r := makeValidReq()
		r.AppVersion = ""
		mustContain(t, r.Validate(), "AppVersion", "present")
	})

	t.Run("Source control", func(t *testing.T) {
		r := makeValidReq()
		r.SourceControl.Repository = ""
		mustContain(t, r.Validate(), "SourceControl.Repository", "present when")

		r = makeValidReq()
		r.SourceControl.Revision = ""
		mustContain(t, r.Validate(), "SourceControl.Revision", "present when")

		t.Run("Provider", func(t *testing.T) {
			r := makeValidReq()
			r.SourceControl.Provider = "sourceforge"
			mustContain(t, r.Validate(), "SourceControl.Provider", "unset", "automatic", "one of", "server")
		})
	})

	t.Run("perfectly valid", func(t *testing.T) {
		t.Run("all values populated", func(t *testing.T) {
			r := makeValidReq()
			if err := r.Validate(); err != nil {
				t.Errorf("unexpected error: %v", err)
			}

		})
		t.Run("only bare-minimum populated", func(t *testing.T) {
			r := &builds.JSONBuildRequest{APIKey: "1234abcd1234abcd1234abcd1234abcd", AppVersion: "1.5.2"}
			if err := r.Validate(); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	})
}
