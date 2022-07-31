package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	if err := run(os.Args[1:], getEnvVars()); err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}
}

func run(args []string, envvars map[string]string) error {
	if len(args) == 0 || args[0] != "release" {
		return fmt.Errorf("See 'bugsnag --help' for usage")
	}

	var (
		flagAPIKey = flag.String(
			"api-key",
			"",
			`Required. Your Bugsnag project's 32-digit hex string.`,
		)

		flagAppVersion = flag.String(
			"app-version",
			"",
			`Required. The version of your application that you're reporting.`,
		)

		flagReleaseStage = flag.String(
			"release-stage",
			"",
			`Optional. The environment in which your application is running.`,
		)

		flagProvider = flag.String(
			"provider",
			"",
			`Optional. Your version control provider.
Valid values: "github", "github-enterprise", "bitbucket", "bitbucket-server", "gitlab", "gitlab-onpremise"`,
		)

		flagRepository = flag.String(
			"repository",
			"",
			`Optional, unless revision is also set. URL of your repository.`,
		)

		flagRevision = flag.String(
			"revision",
			"",
			`Optional, unless repository is also set.
The SHA (or 7-character shorthand) of the git commit associated with the build.`,
		)

		flagBuilder = flag.String(
			"builder",
			"",
			`Optional. The name of the person or bot performing this release.`,
		)

		flagMetadata = flag.String(
			"metadata",
			"",
			`Optional. Format is "KEY1=VALUE1,KEY2=VALUE2"`,
		)

		flagAutoAssignRelease = flag.Bool(
			"auto-assign-release",
			false,
			`Optional. Set to true if the logic in your application is unaware of its own app version.`,
		)

		flagEndpoint = flag.String(
			"endpoint",
			"",
			`Optional. If you're running Bugsnag on-prem, set this to the build API's URL`,
		)

		flagAppBundleVersion = flag.String(
			"app-bundle-version",
			"",
			`Optional. Applies to Apple platform builds only.`,
		)

		flagAppVersionCode = flag.Int(
			"app-version-code",
			0,
			`Optional. Applies to Android builds only.`,
		)
	)
	flag.Parse()

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
	return nil
}

func getEnvVars() map[string]string {
	m := map[string]string{}
	for _, v := range os.Environ() {
		pair := strings.SplitN(v, "=", 2)
		m[pair[0]] = pair[1]
	}
	return m
}
