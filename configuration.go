package bugsnag

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"time"
)

// Configuration represents all of the possible configurations for the notifier.
type Configuration struct {

	// Required configuration options:

	// The 32 hex-character API Key for your Bugsnag project.
	APIKey string
	// The version of your application, as semver.
	AppVersion string
	// The stage in your release cycle, e.g. "development", "production", etc.
	// Any non-empty value is valid.
	ReleaseStage string

	// Optional configuration options:

	// The endpoint to send error reports to. Configure if you're
	// using an on-premise installation of Bugsnag. Defaults to
	// https://notify.bugsnag.com
	EndpointNotify string
	// The endpoint to send sessions to. Configure if you're using an
	// on-premise installation of Bugsnag. Defaults to
	// https://sessions.bugsnag.com
	EndpointSessions string

	// If defined its SanitizeErrorReport will be invoked just before each API
	// call to Bugsnag.
	ErrorReportSanitizer ErrorReportSanitizer

	runtimeConstants
}

func validURL(cand string) bool {
	if _, err := url.ParseRequestURI(cand); err != nil {
		return false
	}
	u, err := url.Parse(cand)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func (cfg *Configuration) validate() error {
	if r := regexp.MustCompile("^[0-9a-f]{32}$"); !r.Match([]byte(cfg.APIKey)) {
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
	if r := regexp.MustCompile(semverRegex); !r.Match([]byte(cfg.AppVersion)) {
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
	rc := runtimeConstants{
		osVersion:    osVersion(),
		goVersion:    runtime.Version(),
		osName:       runtime.GOOS,
		appStartTime: time.Now(),
	}
	rc.hostname, _ = os.Hostname()
	if bi, ok := debug.ReadBuildInfo(); ok {
		rc.appID = bi.Path
		for _, dep := range bi.Deps {
			if dep.Path == "github.com/kinbiko/bugsnag" {
				rc.notifierVersion = dep.Version
				break
			}
		}
	}
	return rc
}
