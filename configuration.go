package bugsnag

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"time"
)

// Configuration represents all of the possible configurations for the notifier.
// Only APIKey, AppVersion, and ReleaseStage is required.
type Configuration struct {
	// Required configuration options:

	// The 32 hex-character API Key that identifies your Bugsnag project.
	APIKey string
	// The version of your application, as semver.
	// **Strictly** speaking, Bugsnag doesn't require you to use semver, but in
	// Go this is **highly** recommended (esp. with go mod). If you absolutely
	// *have* to use a non-semver AppVersion, set this configuration option to
	// any valid semver, and change the AppVersion as part of of your
	// ErrorReportSanitizer.
	AppVersion string
	// The stage in your release cycle, e.g. "development", "production", etc.
	// Any non-empty value is valid.
	// **Strictly** speaking, Bugsnag doesn't require you to set this field.
	// However, **a lot** of the features in Bugsnag are
	// significantly improved by setting this field, so it is required in this
	// package.
	ReleaseStage string

	// Optional configuration options:

	// The endpoint to send error reports to. Configure if you're
	// using an on-premise installation of Bugsnag. Defaults to
	// https://notify.bugsnag.com
	// If this field is set you must also set EndpointSessions.
	EndpointNotify string
	// The endpoint to send sessions to. Configure if you're using an
	// on-premise installation of Bugsnag. Defaults to
	// https://sessions.bugsnag.com
	// If this field is set you must also set EndpointNotify.
	EndpointSessions string

	// If defined it will be invoked just before each error report API call to
	// Bugsnag. See the GoDoc on the ErrorReportSanitizer type for more details.
	ErrorReportSanitizer ErrorReportSanitizer

	// If defined it will be invoked just before each session report API call
	// to Bugsnag. See the GoDoc on the SessionReportSanitizer type for more details.
	SessionReportSanitizer SessionReportSanitizer

	// InternalErrorCallback gets invoked with a descriptive error for any
	// internal issues in the notifier that's preventing normal operation.
	// Configuring this function may be useful if you need to debug missing
	// reports or sessions.
	InternalErrorCallback func(err error)

	runtimeConstants
}

func validURL(cand string) bool {
	if _, err := url.ParseRequestURI(cand); err != nil {
		return false
	}
	u, err := url.Parse(cand)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func (cfg *Configuration) populateDefaults() {
	if cfg.EndpointNotify == "" {
		cfg.EndpointNotify = "https://notify.bugsnag.com"
		cfg.EndpointSessions = "https://sessions.bugsnag.com"
	}
	// Default to NOOP callbacks.
	if cfg.ErrorReportSanitizer == nil {
		cfg.ErrorReportSanitizer = func(_ context.Context, _ *JSONErrorReport) error { return nil }
	}
	if cfg.SessionReportSanitizer == nil {
		cfg.SessionReportSanitizer = func(_ *JSONSessionReport) error { return nil }
	}
	if cfg.InternalErrorCallback == nil {
		cfg.InternalErrorCallback = func(_ error) {}
	}
}

func (cfg *Configuration) validate() error {
	if r := regexp.MustCompile("^[0-9a-f]{32}$"); !r.MatchString(cfg.APIKey) {
		return fmt.Errorf(`API key must be 32 hex characters, but got "%s"`, cfg.APIKey)
	}

	if !validURL(cfg.EndpointNotify) {
		return fmt.Errorf(`notify endpoint be a valid URL, got "%s"`, cfg.EndpointNotify)
	}
	if !validURL(cfg.EndpointSessions) {
		return fmt.Errorf(`sessions endpoint be a valid URL, got "%s"`, cfg.EndpointSessions)
	}
	if cfg.ReleaseStage == "" {
		return fmt.Errorf("release stage must be set")
	}
	semverRegex := `v?([0-9]+)(\.[0-9]+)?(\.[0-9]+)?(-([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?(\+([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?`
	if r := regexp.MustCompile(semverRegex); !r.MatchString(cfg.AppVersion) {
		return fmt.Errorf("app version must be valid semver")
	}
	return nil
}

type runtimeConstants struct {
	hostname        string
	osVersion       string
	goVersion       string
	osName          string
	notifierVersion string
	appID           string

	appStartTime time.Time
}

func makeRuntimeConstants() runtimeConstants {
	var (
		appID = ""
		// This next line is for developer sanity:
		// Bugsnag will drop payloads that don't include the notifier payload.
		// However, it is not set when doing development on the notifier locally --
		// only when it's imported in another application.
		// Therefore, set a constant if it is missing from the calculation above.
		notifierVersion = "SNAPSHOT"
	)

	if bi, ok := debug.ReadBuildInfo(); ok {
		appID = bi.Path
		for _, dep := range bi.Deps {
			if dep.Path == "github.com/kinbiko/bugsnag" {
				notifierVersion = dep.Version
				break
			}
		}
	}

	return runtimeConstants{
		osVersion:       osVersion(),
		goVersion:       runtime.Version(),
		osName:          runtime.GOOS,
		appStartTime:    time.Now(),
		hostname:        func() string { h, _ := os.Hostname(); return h }(),
		notifierVersion: notifierVersion,
		appID:           appID,
	}
}
