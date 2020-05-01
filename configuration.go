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
	APIKey string // Required. The 32 hex-character API Key for your Bugsnag project.

	AppVersion   string // Optional, but highly recommended.
	ReleaseStage string // Optional, but highly recommended.

	// Optional. The endpoint to send error reports to. Configure if you're
	// using an on-premise installation of Bugsnag. Defaults to
	// https://notify.bugsnag.com
	EndpointNotify string
	// Optional. The endpoint to send sessions to. Configure if you're using an
	// on-premise installation of Bugsnag. Defaults to
	// https://sessions.bugsnag.com
	EndpointSessions string

	ErrorReportSanitizer ErrorReportSanitizer

	runtimeConstants
}

func (cfg *Configuration) validate() error {
	if r := regexp.MustCompile("^[0-9a-f]{32}$"); !r.Match([]byte(cfg.APIKey)) {
		return fmt.Errorf(`API key must be 32 hex characters, but got "%s"`, cfg.APIKey)
	}

	validURL := func(cand string) bool {
		if _, err := url.ParseRequestURI(cand); err != nil {
			return false
		}
		u, err := url.Parse(cand)
		return err == nil && u.Scheme != "" && u.Host != ""
	}

	if !validURL(cfg.EndpointNotify) {
		return fmt.Errorf(`notify endpoint be avalid URL, got "%s"`, cfg.EndpointNotify)
	}
	if !validURL(cfg.EndpointSessions) {
		return fmt.Errorf(`sessions endpoint be avalid URL, got "%s"`, cfg.EndpointSessions)
	}
	return nil
}

type runtimeConstants struct {
	hostname, osVersion, goVersion, osName string

	// modulePath defines the root of the project that uses this package.
	// Used to identify if a file is "in-project" or a third party library,
	// which is in turn used by Bugsnag to group errors by the top stackframe
	// that's "in project".
	modulePath   string
	appStartTime time.Time
}

func makeRuntimeConstants() runtimeConstants {
	hostname, _ := os.Hostname()

	modulePath := ""
	if bi, ok := debug.ReadBuildInfo(); ok {
		modulePath = bi.Main.Path
	}
	return runtimeConstants{
		hostname:     hostname,
		osVersion:    osVersion(),
		goVersion:    runtime.Version(),
		osName:       runtime.GOOS,
		modulePath:   modulePath,
		appStartTime: time.Now(),
	}
}
