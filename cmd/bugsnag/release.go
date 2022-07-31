package main

import (
	"flag"
	"fmt"

	"github.com/kinbiko/bugsnag/builds"
)

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
			`Required. Your Bugsnag project's 32-digit hex string.`,
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

		debug: releaseCmd.Bool("debug", false, "Optional. Turn on for debug logs"),
	}
}

func (app *application) runRelease(envVars map[string]string) error {
	if *app.releaseFlags.debug {
		fmt.Printf("--api-key=%s\n", *app.releaseFlags.apiKey)
		fmt.Printf("--app-version=%s\n", *app.releaseFlags.appVersion)
		fmt.Printf("--release-stage=%s\n", *app.releaseFlags.releaseStage)
		fmt.Printf("--provider=%s\n", *app.releaseFlags.provider)
		fmt.Printf("--repository=%s\n", *app.releaseFlags.repository)
		fmt.Printf("--revision=%s\n", *app.releaseFlags.revision)
		fmt.Printf("--builder=%s\n", *app.releaseFlags.builder)
		fmt.Printf("--metadata=%s\n", *app.releaseFlags.metadata)
		fmt.Printf("--auto-assign-release=%v\n", *app.releaseFlags.autoAssignRelease)
		fmt.Printf("--endpoint=%s\n", *app.releaseFlags.endpoint)
		fmt.Printf("--app-version-code=%d\n", *app.releaseFlags.appVersionCode)
		fmt.Printf("--app-bundle-version=%s\n", *app.releaseFlags.appBundleVersion)
		fmt.Printf("--debug=%v\n", *app.releaseFlags.debug)
	}

	req := builds.JSONBuildRequest{
		APIKey:            *app.releaseFlags.apiKey,
		AppVersion:        *app.releaseFlags.appVersion,
		ReleaseStage:      *app.releaseFlags.releaseStage,
		BuilderName:       *app.releaseFlags.builder,
		Metadata:          makeMetadata(*app.releaseFlags.metadata),
		AppVersionCode:    *app.releaseFlags.appVersionCode,
		AppBundleVersion:  *app.releaseFlags.appBundleVersion,
		AutoAssignRelease: *app.releaseFlags.autoAssignRelease,
	}

	if *app.releaseFlags.repository != "" || *app.releaseFlags.revision != "" {
		req.SourceControl = &builds.JSONSourceControl{
			Provider:   *app.releaseFlags.provider,
			Repository: *app.releaseFlags.repository,
			Revision:   *app.releaseFlags.revision,
		}
	}

	if err := req.Validate(); err != nil {
		return fmt.Errorf("Invalid build data: %w\nSee 'bugsnag release --help'", err)
	}

	return nil
}
