// Package main is the entrypoint for the bugsnag command-line tool.
package main

import (
	"flag"
	"fmt"
	"os"
)

const usage = `bugsnag is a tool for calling Bugsnag's APIs

Usage:

bugsnag <command> --flag1 --flag2 (...)

The commands are:
	release -- Report a build or a release to Bugsnag's build API.

For more details about a command, run:

bugsnag <command> --help`

type application struct {
	releaseFlags *releaseFlags
}

func main() {
	if err := run(os.Args[1:], splitByEquals(os.Environ())); err != nil {
		logf("%s\n", err.Error())
		os.Exit(1)
	}
}

func run(args []string, envvars map[string]string) error {
	releaseCmd := flag.NewFlagSet("release", flag.ExitOnError)
	app := application{releaseFlags: newReleaseFlags(releaseCmd)}

	if len(args) == 0 {
		return fmt.Errorf(usage)
	}
	if args[0] == "release" {
		err := releaseCmd.Parse(args[1:])
		if err != nil {
			return fmt.Errorf("%w:\n%s", err, usage)
		}
		return app.runRelease(envvars)
	}
	return fmt.Errorf(usage)
}
