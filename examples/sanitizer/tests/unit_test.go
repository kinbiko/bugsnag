package unit_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kinbiko/bugsnag"
	unit "github.com/kinbiko/bugsnag/examples/sanitizer/tests"
)

func TestBugsnagGetsNotified(t *testing.T) {
	for _, tc := range []struct {
		name   string
		in     bool
		expMsg string
	}{
		{name: "dangerous operation succeeds", in: false, expMsg: ""},
		{name: "dangerous operation fails", in: true, expMsg: "something happened in the dangerous operation"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				// Note: We're closing over errMsg instead of doing the
				// assertion inside the sanitizer, in order to verify the
				// positive case as well.
				errMsg = ""

				n, err = bugsnag.New(bugsnag.Configuration{
					APIKey: "abcd1234abcd1234abcd1234abcd1234", AppVersion: "1.2.3", ReleaseStage: "test",
					ErrorReportSanitizer: func(ctx context.Context, r *bugsnag.JSONErrorReport) error {
						errMsg = r.Events[0].Exceptions[0].Message
						return errors.New("prevents sending the payload to Bugsnag")
					},
				})
			)

			if err != nil {
				t.Fatal(err)
			}

			unit.NewUnit(n).DoDangerousOperation(context.Background(), tc.in)
			if errMsg != tc.expMsg {
				t.Errorf("expected error message '%s' but got '%s'", tc.expMsg, errMsg)
			}
		})
	}
}
