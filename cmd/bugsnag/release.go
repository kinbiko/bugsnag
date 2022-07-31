package main

import (
	"flag"
	"fmt"
	"os"
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
			`Required. The version of your application that you're reporting.`,
		),

		releaseStage: releaseCmd.String(
			"release-stage",
			"",
			`Optional. The environment in which your application is running.`,
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

func (app *application) runRelease(envVars map[string]string) error {
	rf := app.releaseFlags
	if *rf.debug {
		app.printReleaseDebug()
	}

	req := &builds.JSONBuildRequest{
		APIKey:            *rf.apiKey,
		AppVersion:        *rf.appVersion,
		ReleaseStage:      *rf.releaseStage,
		BuilderName:       *rf.builder,
		Metadata:          splitByEquals(strings.Split(*rf.metadata, ",")),
		AppVersionCode:    *rf.appVersionCode,
		AppBundleVersion:  *rf.appBundleVersion,
		AutoAssignRelease: *rf.autoAssignRelease,
	}

	if *rf.repository != "" || *rf.revision != "" {
		req.SourceControl = &builds.JSONSourceControl{
			Provider:   *rf.provider,
			Repository: *rf.repository,
			Revision:   *rf.revision,
		}
	}

	populateReleaseDefaults(req, envVars)

	if err := req.Validate(); err != nil {
		return fmt.Errorf("Invalid build data: %w\nSee 'bugsnag release --help'", err)
	}

	publisher := builds.DefaultPublisher()
	if endpoint := *rf.endpoint; endpoint != "" {
		publisher = builds.NewPublisher(endpoint)
	}
	if err := publisher.Publish(req); err != nil {
		return err
	}

	fmt.Printf("release info published for version %s\n", req.AppVersion)
	return nil
}

func (app *application) printReleaseDebug() {
	rf := app.releaseFlags
	fmt.Printf("--api-key=%s\n", *rf.apiKey)
	fmt.Printf("--app-version=%s\n", *rf.appVersion)
	fmt.Printf("--release-stage=%s\n", *rf.releaseStage)
	fmt.Printf("--provider=%s\n", *rf.provider)
	fmt.Printf("--repository=%s\n", *rf.repository)
	fmt.Printf("--revision=%s\n", *rf.revision)
	fmt.Printf("--builder=%s\n", *rf.builder)
	fmt.Printf("--metadata=%s\n", *rf.metadata)
	fmt.Printf("--auto-assign-release=%v\n", *rf.autoAssignRelease)
	fmt.Printf("--endpoint=%s\n", *rf.endpoint)
	fmt.Printf("--app-version-code=%d\n", *rf.appVersionCode)
	fmt.Printf("--app-bundle-version=%s\n", *rf.appBundleVersion)
	fmt.Printf("--debug=%v\n", *rf.debug)
}

func populateReleaseDefaults(req *builds.JSONBuildRequest, envVars map[string]string) {
	if req.APIKey == "" {
		req.APIKey = os.Getenv("BUGSNAG_API_KEY")
	}
}
