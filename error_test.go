package bugsnag

import "testing"

func TestSeverityReasonType(t *testing.T) {
	for _, tc := range []struct {
		exp string
		err Error
	}{
		{exp: "handledException", err: Error{}},
		{exp: "unhandledException", err: Error{Unhandled: true}},
		{exp: "handledPanic", err: Error{Panic: true}},
		{exp: "unhandledPanic", err: Error{Unhandled: true, Panic: true}},
		{exp: "userSpecifiedSeverity", err: Error{Severity: SeverityError}},
		{exp: "userSpecifiedSeverity", err: Error{Severity: SeverityError, Unhandled: true, Panic: true}},
	} {
		t.Run(tc.exp, func(t *testing.T) {
			if got := severityReasonType(&tc.err); got != tc.exp {
				t.Errorf("expected severity reason type '%s' but got '%s'", tc.exp, got)
			}
		})
	}
}
