package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kinbiko/bugsnag/builds"
)

const usage = `Bugsnag is a tool for calling Bugsnag's APIs

Usage:

bugsnag <command> --flag1 --flag2 (...)

The commands are:
	release -- Report a build or a release to Bugsnag's build API.

For more details about a command, run:

bugsnag <command> --help`

func main() {
	if err := run(os.Args[1:], getEnvVars()); err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}
}

func run(args []string, envvars map[string]string) error {
	if len(args) == 0 {
		return fmt.Errorf(usage)
	}

	var (
		releaseCmd = flag.NewFlagSet("release", flag.ExitOnError)
		flagAPIKey = releaseCmd.String(
			"api-key",
			"",
			`Required. Your Bugsnag project's 32-digit hex string.`,
		)

		flagAppVersion = releaseCmd.String(
			"app-version",
			"",
			`Required. The version of your application that you're reporting.`,
		)

		flagReleaseStage = releaseCmd.String(
			"release-stage",
			"",
			`Optional. The environment in which your application is running.`,
		)

		flagProvider = releaseCmd.String(
			"provider",
			"",
			`Optional. Your version control provider.
Valid values: "github", "github-enterprise", "bitbucket", "bitbucket-server", "gitlab", "gitlab-onpremise"`,
		)

		flagRepository = releaseCmd.String(
			"repository",
			"",
			`Optional, unless revision is also set. URL of your repository.`,
		)

		flagRevision = releaseCmd.String(
			"revision",
			"",
			`Optional, unless repository is also set.
The SHA (or 7-character shorthand) of the git commit associated with the build.`,
		)

		flagBuilder = releaseCmd.String(
			"builder",
			"",
			`Optional. The name of the person or bot performing this release.`,
		)

		flagMetadata = releaseCmd.String(
			"metadata",
			"",
			`Optional. Format is "KEY1=VALUE1,KEY2=VALUE2"`,
		)

		flagAutoAssignRelease = releaseCmd.Bool(
			"auto-assign-release",
			false,
			`Optional. Set to true if the logic in your application is unaware of its own app version.`,
		)

		flagEndpoint = releaseCmd.String(
			"endpoint",
			"",
			`Optional. If you're running Bugsnag on-prem, set this to the build API's URL`,
		)

		flagAppBundleVersion = releaseCmd.String(
			"app-bundle-version",
			"",
			`Optional. Applies to Apple platform builds only.`,
		)

		flagAppVersionCode = releaseCmd.Int(
			"app-version-code",
			0,
			`Optional. Applies to Android builds only.`,
		)

		flagDebug = releaseCmd.Bool("debug", false, "Optional. Turn on for debug logs")
	)

	switch args[0] {
	case "release":
		releaseCmd.Parse(args[1:])
	default:
		return fmt.Errorf(usage)
	}

	if *flagDebug {
		fmt.Printf("--api-key=%s\n", *flagAPIKey)
		fmt.Printf("--app-version=%s\n", *flagAppVersion)
		fmt.Printf("--release-stage=%s\n", *flagReleaseStage)
		fmt.Printf("--provider=%s\n", *flagProvider)
		fmt.Printf("--repository=%s\n", *flagRepository)
		fmt.Printf("--revision=%s\n", *flagRevision)
		fmt.Printf("--builder=%s\n", *flagBuilder)
		fmt.Printf("--metadata=%s\n", *flagMetadata)
		fmt.Printf("--auto-assign-release=%v\n", *flagAutoAssignRelease)
		fmt.Printf("--endpoint=%s\n", *flagEndpoint)
		fmt.Printf("--app-version-code=%d\n", *flagAppVersionCode)
		fmt.Printf("--app-bundle-version=%s\n", *flagAppBundleVersion)
		fmt.Printf("--debug=%v\n", *flagDebug)
	}

	req := builds.JSONBuildRequest{
		APIKey:            *flagAPIKey,
		AppVersion:        *flagAppVersion,
		ReleaseStage:      *flagReleaseStage,
		BuilderName:       *flagBuilder,
		Metadata:          makeMetadata(*flagMetadata),
		AppVersionCode:    *flagAppVersionCode,
		AppBundleVersion:  *flagAppBundleVersion,
		AutoAssignRelease: *flagAutoAssignRelease,
	}

	if *flagRepository != "" || *flagRevision != "" {
		req.SourceControl = &builds.JSONSourceControl{
			Provider:   *flagProvider,
			Repository: *flagRepository,
			Revision:   *flagRevision,
		}
	}

	_ = req
	return nil
}

func makeMetadata(str string) map[string]string {
	m := map[string]string{}
	kvps := strings.Split(str, ",")
	for _, kvp := range kvps {
		if kvp != "" {
			pair := strings.SplitN(kvp, "=", 2)
			m[pair[0]] = pair[1]
		}
	}
	return m
	// TODO: Refactor as it's basically the same logic as getEnvVars

}

func getEnvVars() map[string]string {
	m := map[string]string{}
	for _, v := range os.Environ() {
		pair := strings.SplitN(v, "=", 2)
		m[pair[0]] = pair[1]
	}
	return m
}
