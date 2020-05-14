package bugsnag

// === SHARED ===
// The following payloads are shared between session and report payloads

// JSONNotifier describes the package that triggers the sending of this
// report (the 'notifier').  These properties are used within Bugsnag to track
// error rates from a notifier.
type JSONNotifier struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	URL     string `json:"url"`
}

// JSONApp is information about the app where the error occurred. These
// fields are optional but highly recommended. To display custom app data
// alongside these standard fields on the Bugsnag website, the custom data
// should be included in the metaData object in an app object.
type JSONApp struct {
	// ID is a unique ID for the application
	ID string `json:"id,omitempty"`

	// Version number of the application which generated the error.
	Version string `json:"version,omitempty"`

	// ReleaseStage is the release stage (e.g. "staging" or "production")
	// Should be set to "production" by default.
	ReleaseStage string `json:"releaseStage,omitempty"`

	// Type is the specialized type of the application, such as the worker
	// queue or web framework used.
	Type string `json:"type,omitempty"`

	// Duration is how long the app has been running for in milliseconds.
	Duration int64 `json:"duration"`
}

// JSONDevice is information about the computer/device running the app.
// These fields are optional but highly recommended. To display custom device
// data alongside these standard fields on the Bugsnag website, the custom data
// should be included in the metaData object in a device object.
type JSONDevice struct {
	// Rather than faff about with trying to find something that matches the
	// API exactly, just fit all easily fetchable memory data and abuse
	// the fact you can just add to the device struct whatever you want.
	// Obviously, this may well stop working in the future.
	MemStats map[string]interface{} `json:"memStats,omitempty"`

	// Hostname of the server running your code.
	Hostname string `json:"hostname,omitempty"`

	// OSName is the name of the operating system.
	OSName string `json:"osName,omitempty"`

	// OSVersion is the version of the operating system, only applies to *NIX
	// systems, as defined by "uname -r".
	OSVersion string `json:"osVersion,omitempty"`

	RuntimeVersions map[string]string `json:"runtimeVersions"`
}

// === SESSION ===
// The following payloads only apply to sessions

// JSONSessionReport defines the top level payload object
type JSONSessionReport struct {
	Notifier      *JSONNotifier       `json:"notifier"`
	App           *JSONApp            `json:"app"`
	Device        *JSONDevice         `json:"device"`
	SessionCounts []JSONSessionCounts `json:"sessionCounts"`
}

// JSONSessionCounts defines the .sessionCounts subobject of the payload
type JSONSessionCounts struct {
	StartedAt       string `json:"startedAt"`
	SessionsStarted int    `json:"sessionsStarted"`
}

// === REPORT ===
// The following payloads only apply to reports

// The types written in this file is written against the API defined here:
// https://bugsnagerrorreportingapi.docs.apiary.io/#reference/0/notify/send-error-reports

// JSONErrorReport is the top level struct that encompasses the entire payload that's
// being sent to the Bugsnag's servers upon reporting an error.
type JSONErrorReport struct {
	// The API Key associated with the project. Informs Bugsnag which project has generated this error.
	// This is provided for legacy notifiers. It is preferable to use the Bugsnag-Api-Key header instead.
	APIKey string `json:"apiKey"`

	Notifier *JSONNotifier `json:"notifier"`
	Events   []*JSONEvent  `json:"events"`
}

// JSONEvent is the event that Bugsnag should be notified of.
type JSONEvent struct {
	// The version number of the payload. This is provided for legacy notifiers.
	// It is preferable to use the Bugsnag-Payload-Version header instead.
	PayloadVersion string `json:"payloadVersion,omitempty"`

	// Context is a string representing what was happening in the application
	// at the time of the error. This string could be used for grouping
	// purposes, depending on the event. Usually this would represent the
	// controller and action in a server based project.
	Context string `json:"context,omitempty"`

	// Unhandled indicates whether the error was unhandled. If true, the error
	// was detected by the notifier because it was not handled by the
	// application. If false, the error was handled and reported using Notify.
	Unhandled bool `json:"unhandled"`

	// Severity can take values in ["error", "warning", "info"].
	// "error" should be the default for panics.
	// "warning" should be the default for Notify() calls.
	// "info" can be specified by the user.
	// Severity does not affect the stability score in your dashboard.
	Severity string `json:"severity"`

	SeverityReason *JSONSeverityReason `json:"severityReason,omitempty"`

	// Most of the time there will only be one exception, but multiple
	// (nested/caused-by) errors may be added individually.
	// The innermost error should be first in this array.
	Exceptions []*JSONException `json:"exceptions,omitempty"`

	// This list is sequential and ordered newest to oldest.
	Breadcrumbs []*JSONBreadcrumb `json:"breadcrumbs,omitempty"`

	Request *JSONRequest `json:"request,omitempty"`

	User *JSONUser `json:"user,omitempty"`

	App *JSONApp `json:"app,omitempty"`

	Device *JSONDevice `json:"device,omitempty"`

	Session *JSONSession `json:"session,omitempty"`

	// An object containing any further data you wish to attach to this error
	// event. The key of the outermost map indicates the tab under which to display this information in Bugsnag.
	// The key of the innermost map indicates the property name, and the value is it's value
	Metadata map[string]map[string]interface{} `json:"metaData,omitempty"`

	// GroupingHash is a unique value that can be set in order to override the
	// grouping on the Bugsnag dashboard.
	//  **Warning: Do not set unless you're 100% sure of what you're doing.**
	GroupingHash string `json:"groupingHash,omitempty"`
}

// JSONException is the error or panic that occurred during this the surrounding event.
type JSONException struct {
	// The name of the type of error/panic which occurred. This is used to
	// group errors together so should not ocntain any contextual information
	// that would prevent correct grouping.
	ErrorClass string `json:"errorClass"`

	// The error message associated with the error. Usually this will contain
	// some information about this specific instance of the error and is not
	// used to group the errors.
	Message string `json:"message,omitempty"`

	Stacktrace []*JSONStackframe `json:"stacktrace"`
}

// JSONStackframe represents one line in the Exception's stacktrace.
// Bugsnag uses this information to help with error grouping, as well as
// displaying it to the user.
type JSONStackframe struct {
	// File identifies the name of the file that this frame of the stack was in.
	// This name should be stripped of any unnecessary from the beginning of the path.
	// In addition to error grouping, Bugsnag is able to go to the correct file
	// in GitHub etc. if the path is relative to the root of your repository.
	File string `json:"file"`

	// LineNumber is the line in the file, that this frame of the stack was in.
	// In addition to error grouping, this will be used to navigate to the
	// correct line in a file when source control is properly integrated.
	LineNumber int `json:"lineNumber"`

	// Method identifies the method or function that this particular stackframe was in.
	Method string `json:"method"`

	// InProject identifies if the stackframe is part of the application
	// written by the user, or if it was part of a third party dependency.
	InProject bool `json:"inProject"`
}

// JSONBreadcrumb represents user- and system-initiated events which led up
// to an error, providing additional context.
type JSONBreadcrumb struct {
	// The time at which the event breadcrumb occurred, in ISO8601
	Timestamp string `json:"timestamp,omitempty"`

	// A short summary describing the breadcrumb, such as entering a new application state
	Name string `json:"name"`

	// Type is a category which describes the breadcrumb, from the list of allowed values.
	// Accepted values are: ["navigation", "request", "process", "log", "user", "state", "error", "manual"]
	Type string `json:"type"`

	// Metadata contains any additional information about the breadcrumb, as key/value pairs.
	Metadata map[string]interface{} `json:"metaData,omitempty"`
}

// JSONRequest contains details about the web request from the client that
// experienced the error, if relevant. To display custom request data alongside
// these standard fields on the Bugsnag website, the custom data should be
// included in the metaData object in a request object.
type JSONRequest struct {
	// ClientIP identifies the IP address of the client that sent the request that caused the error.
	ClientIP string `json:"clientIp,omitempty"`

	Headers map[string]string `json:"headers,omitempty"`

	HTTPMethod string `json:"httpMethod,omitempty"`

	URL string `json:"url,omitempty"`

	Referer string `json:"referer,omitempty"`
}

// JSONSeverityReason contains information about why the severity was
// picked.
type JSONSeverityReason struct {
	// "handledPanic" should be used when a panic has been recovered.
	// "unhandledPanic" should be used when a panic has happened without being recovered.
	// "handledError" should be used when reporting an error
	// "errorClass" should be used when the severity is defined because of the type of error that was reported.
	// "log" should be used when a notification is sent as part of a log call and the severity is set based on the log level
	// "signal" should be used when the application has received a signal from the operating system
	// "userCallbackSetSeverity" should be used when the user sets it as part of a callback.
	Type string `json:"type,omitempty"`
}

// JSONUser is user information about the user affected by the error. These fields are
// optional but highly recommended. To display custom user data alongside these
// standard fields on the Bugsnag website, the custom data should be included
// in the metaData object in a user object.
type JSONUser = User

// JSONSession is information associated with the event. This can be used
// alongside the Bugsnag Session Tracking API to associate the event with a
// session so that a release's crash rate can be determined.
type JSONSession struct {
	// ID is a unique identifier of this session.
	ID string `json:"id"`

	// The time (in ISO8601 format) at which the session started.
	StartedAt string `json:"startedAt"`

	Events *JSONSessionEvents `json:"events"`
}

// JSONSessionEvents contain details of the number of handled and unhandled events
// that happened as part of this session (including this event).
type JSONSessionEvents struct {
	// The number of handled events that have occurred in this session (including this event).
	Handled int `json:"handled,omitempty"`
	// The number of unhandled events that have occurred in this session
	// (including this event). Unlikely to be more than 1, as in Go, unhandled
	// events typically indicate that the app is shutting down.
	Unhandled int `json:"unhandled,omitempty"`
}
