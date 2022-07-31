package builds_test

import (
	"strings"
	"testing"

	"github.com/kinbiko/bugsnag/builds"
)

func makeBigValidReq() *builds.JSONBuildRequest {
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

func makeSmallValidReq() *builds.JSONBuildRequest {
	return &builds.JSONBuildRequest{APIKey: "1234abcd1234abcd1234abcd1234abcd", AppVersion: "1.5.2"}
}

func mustContain(t *testing.T, err error, strs ...string) {
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
