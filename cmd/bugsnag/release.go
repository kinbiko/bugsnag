package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/kinbiko/bugsnag/builds"
)

// These fields represent command line flags and all have to be pointers as
// they will be unset until Parse is called.
type releaseFlags struct {
	apiKey            *string
	appVersion        *string
	releaseStage      *string
	provider          *string
	repository        *string
	revision          *string
	builder           *string
	metadata          *string
	autoAssignRelease *bool
	endpoint          *string
	appVersionCode    *int
	appBundleVersion  *string
	debug             *bool
}

//nolint:funlen // This function is long but it's just a bunch of flag declarations.
func newReleaseFlags(releaseCmd *flag.FlagSet) *releaseFlags {
	return &releaseFlags{
		apiKey: releaseCmd.String(
			"api-key",
			"",
			`Required. Your Bugsnag project's 32-digit hex string.
bugsnag will look for a BUGSNAG_API_KEY environment variable if no value is provided.`,
		),

		appVersion: releaseCmd.String(
			"app-version",
			"",
			`Required. The version of your application that you're reporting.
bugsnag will look for a BUGSNAG_APP_VERSION environment variable if no value is provided.
Failing that, bugsnag will look for an APP_VERSION environment variable.`,
		),

		releaseStage: releaseCmd.String(
			"release-stage",
			"production",
			`Optional. The environment in which your application is running.
Set to "" explicitly to report a build instead of a release.`,
		),

		provider: releaseCmd.String(
			"provider",
			"",
			`Optional. Your version control provider.
Valid values: "github", "github-enterprise", "bitbucket", "bitbucket-server", "gitlab", "gitlab-onpremise"`,
		),

		repository: releaseCmd.String(
			"repository",
			"",
			`Optional, unless revision is also set. URL of your repository.`,
		),

		revision: releaseCmd.String(
			"revision",
			"",
			`Optional, unless repository is also set.
The SHA (or 7-character shorthand) of the git commit associated with the build.`,
		),

		builder: releaseCmd.String(
			"builder",
			"",
			`Optional. The name of the person or bot performing this release.`,
		),

		metadata: releaseCmd.String(
			"metadata",
			"",
			`Optional. Format is "KEY1=VALUE1,KEY2=VALUE2"`,
		),

		autoAssignRelease: releaseCmd.Bool(
			"auto-assign-release",
			false,
			`Optional. Set to true if the logic in your application is unaware of its own app version.`,
		),

		endpoint: releaseCmd.String(
			"endpoint",
			"",
			`Optional. If you're running Bugsnag on-prem, set this to the build API's URL`,
		),

		appBundleVersion: releaseCmd.String(
			"app-bundle-version",
			"",
			`Optional. Applies to Apple platform builds only.`,
		),

		appVersionCode: releaseCmd.Int(
			"app-version-code",
			0,
			`Optional. Applies to Android builds only.`,
		),

		debug: releaseCmd.Bool("debug", false, "Turn on for debug logs"),
	}
}

func makeRelease(flags *releaseFlags) *builds.JSONBuildRequest {
	req := &builds.JSONBuildRequest{
		APIKey:            *flags.apiKey,
		AppVersion:        *flags.appVersion,
		ReleaseStage:      *flags.releaseStage,
		BuilderName:       *flags.builder,
		Metadata:          splitByEquals(strings.Split(*flags.metadata, ",")),
		AppVersionCode:    *flags.appVersionCode,
		AppBundleVersion:  *flags.appBundleVersion,
		AutoAssignRelease: *flags.autoAssignRelease,
		SourceControl:     nil,
	}

	if *flags.repository != "" || *flags.revision != "" {
		req.SourceControl = &builds.JSONSourceControl{
			Provider:   *flags.provider,
			Repository: *flags.repository,
			Revision:   *flags.revision,
		}
	}
	return req
}

func (app *application) runRelease(envVars map[string]string) error {
	flags := app.releaseFlags
	if *flags.debug {
		app.printReleaseDebug()
	}

	req := makeRelease(flags)
	populateReleaseDefaults(req, envVars)

	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid build data: %w\nSee 'bugsnag release --help'", err)
	}

	publisher := builds.DefaultPublisher()
	if endpoint := *flags.endpoint; endpoint != "" {
		publisher = builds.NewPublisher(endpoint)
	}
	if err := publisher.Publish(req); err != nil {
		return fmt.Errorf("unable to publish: %w", err)
	}

	logf("release info published for version %s\n", req.AppVersion)
	return nil
}

func (app *application) printReleaseDebug() {
	flags := app.releaseFlags
	logf("--api-key=%s\n", *flags.apiKey)
	logf("--app-version=%s\n", *flags.appVersion)
	logf("--release-stage=%s\n", *flags.releaseStage)
	logf("--provider=%s\n", *flags.provider)
	logf("--repository=%s\n", *flags.repository)
	logf("--revision=%s\n", *flags.revision)
	logf("--builder=%s\n", *flags.builder)
	logf("--metadata=%s\n", *flags.metadata)
	logf("--auto-assign-release=%v\n", *flags.autoAssignRelease)
	logf("--endpoint=%s\n", *flags.endpoint)
	logf("--app-version-code=%d\n", *flags.appVersionCode)
	logf("--app-bundle-version=%s\n", *flags.appBundleVersion)
	logf("--debug=%v\n", *flags.debug)
}

func populateReleaseDefaults(req *builds.JSONBuildRequest, envVars map[string]string) {
	if req.APIKey == "" {
		req.APIKey = envVars["BUGSNAG_API_KEY"]
	}

	if req.AppVersion == "" {
		req.AppVersion = envVars["BUGSNAG_APP_VERSION"]
		if req.AppVersion == "" {
			req.AppVersion = envVars["APP_VERSION"]
		}
	}
}
