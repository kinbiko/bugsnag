package bugsnag

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// Error allows you to specify certain properties of an error in the Bugsnag dashboard.
// Setting Unhandled to true indicates that the application was not able to
// gracefully handle an error or panic that occurred in the system. This will
// make the reported error affect your app's stability score.
// Setting panic to true indicates that the application experienced a (caught)
// panic, as opposed to just reporting an error.
// You may specify what severity your error should be reported with. Values can
// be one of SeverityInfo, SeverityWarning, and SeverityError. The default
// severity for unhandled or panicking Errors is "error", and "warning"
// otherwise.
type Error struct {
	Err       error
	Unhandled bool
	Panic     bool
	Severity  severity

	ctx        context.Context
	stacktrace []*JSONStackframe
	msg        string
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.msg == "" {
		return e.Err.Error()
	}
	return e.msg
}

// Unwrap is the conventional method for getting the underlying error of a
// bugsnag.Error.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Wrap attaches ctx data and wraps the given error with message, and
// associates a stacktrace to the error based on the frame at which Wrap was
// called.
// Any attached diagnostic data from this ctx will be preserved should you
// return the returned error further up the stack.
func (n *Notifier) Wrap(ctx context.Context, err error, msgAndFmtArgs ...interface{}) *Error {
	return Wrap(ctx, err, msgAndFmtArgs...)
}

// Wrap attaches ctx data and wraps the given error with message, and
// associates a stacktrace to the error based on the frame at which Wrap was
// called.
// Any attached diagnostic data from this ctx will be preserved should you
// return the returned error further up the stack.
func Wrap(ctx context.Context, err error, msgAndFmtArgs ...interface{}) *Error {
	if err == nil {
		return nil
	}
	berr := &Error{
		Err:        err,
		stacktrace: makeStacktrace(makeModulePath()),
		ctx:        ctx,
	}
	if len(msgAndFmtArgs) >= 1 {
		msg, ok := msgAndFmtArgs[0].(string)
		if ok {
			msg = fmt.Sprintf(msg, msgAndFmtArgs[1:]...)
			berr.msg = fmt.Sprintf("%s: %s", msg, err.Error())
		}
	}
	return berr
}

func makeStacktrace(module string) []*JSONStackframe {
	ptrs := [50]uintptr{}
	// Skip 0 frames as we strip this manually later by ignoring any frames
	// including github.com/kinbiko/bugsnag (or below).
	pcs := ptrs[0:runtime.Callers(0, ptrs[:])]

	stacktrace := make([]*JSONStackframe, len(pcs))
	for i, pc := range pcs {
		pc-- // pc - 1 is the *real* program counter, for reasons beyond me.

		file, lineNumber, method := "unknown", 0, "unknown"
		if fn := runtime.FuncForPC(pc); fn != nil {
			file, lineNumber = fn.FileLine(pc)
			method = fn.Name()
		}
		inProject := module != "" && strings.Contains(method, module) || strings.Contains(method, "main.main")

		stacktrace[i] = &JSONStackframe{File: file, LineNumber: lineNumber, Method: method, InProject: inProject}
	}

	// Drop any frames from this package, and further down, for example Go
	// stdlib packages. Rather than trying to guess how many frames to skip,
	// this approach will work better on multiple platforms
	lastBugsnagIndex := 0
	for i, sf := range stacktrace {
		if strings.Contains(sf.Method, "github.com/kinbiko/bugsnag.") {
			lastBugsnagIndex = i
		}
	}
	return stacktrace[lastBugsnagIndex+1:]
}

// makeModulePath defines the root of the project that uses this package.
// Used to identify if a file is "in-project" or a third party library,
// which is in turn used by Bugsnag to group errors by the top stackframe
// that's "in project".
func makeModulePath() string {
	if bi, ok := debug.ReadBuildInfo(); ok {
		return bi.Main.Path
	}
	return ""
}
