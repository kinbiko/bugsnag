package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const usage = `Bugsnag is a tool for calling Bugsnag's APIs

Usage:

bugsnag <command> --flag1 --flag2 (...)

The commands are:
	release -- Report a build or a release to Bugsnag's build API.

For more details about a command, run:

bugsnag <command> --help`

type application struct {
	releaseCmd   *flag.FlagSet
	releaseFlags *releaseFlags
}

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
	releaseCmd := flag.NewFlagSet("release", flag.ExitOnError)
	app := application{releaseFlags: newReleaseFlags(releaseCmd)}

	switch args[0] {
	case "release":
		releaseCmd.Parse(args[1:])
		return app.runRelease(envvars)
	}
	return fmt.Errorf(usage)

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
