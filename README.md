# Bugsnag Go notifier (unofficial rewrite)

[![Build Status](https://github.com/kinbiko/bugsnag/workflows/Go/badge.svg)](https://github.com/kinbiko/bugsnag/actions)
[![Coverage Status](https://coveralls.io/repos/github/kinbiko/bugsnag/badge.svg?branch=main)](https://coveralls.io/github/kinbiko/bugsnag?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/kinbiko/bugsnag)](https://goreportcard.com/report/github.com/kinbiko/bugsnag)
[![Latest version](https://img.shields.io/github/tag/kinbiko/bugsnag.svg?label=latest%20version&style=flat)](https://github.com/kinbiko/bugsnag/releases)
[![Go Documentation](http://img.shields.io/badge/godoc-documentation-blue.svg?style=flat)](https://pkg.go.dev/github.com/kinbiko/bugsnag?tab=doc)
[![License](https://img.shields.io/github/license/kinbiko/bugsnag.svg?style=flat)](https://github.com/kinbiko/bugsnag/blob/main/.github/LICENSE)

Well-documented, maintainable, idiomatic, opinionated, and **unofficial** rewrite of the [Bugsnag Go notifier](https://github.com/bugsnag/bugsnag-go).
See [this document](./.github/official-notifier-difference.md) for an overview of the differences between this package and the official notifier.

In addition to the notifier library in `github.com/kinbiko/bugsnag`, there's a `bugsnag` command-line application in `github.com/kinbiko/bugsnag/cmd/bugsnag` that can easily report builds and releases to Bugsnag for an even better debugging experience.

## Notifier Usage

Make sure you're importing this package, and not the official notifier:

```go
import "github.com/kinbiko/bugsnag"
```

Then set up your `*bugsnag.Notifier`, the type that exposes the main API of this package, based on a `bugsnag.Configuration`.

```go
notifier, err := bugsnag.New(bugsnag.Configuration{
	APIKey:       "<<YOUR API KEY HERE>>",
	AppVersion:   "1.3.5", // Some semver
	ReleaseStage: "production",
})
if err != nil {
	panic(err) // TODO: Handle error in an appropriate manner
}
defer notifier.Close() // Close once you know there are no further calls to notifier.Notify or notifier.StartSession.
```

### Reporting errors

The Notifier is now ready for use.
Call `notifier.Notify()` to notify Bugsnag about an error in your application.

```go
if err != nil {
	notifier.Notify(ctx, err)
}
```

### Attaching diagnostic data

See the [`With*` methods in the docs](https://pkg.go.dev/github.com/kinbiko/bugsnag) to learn how to attach additional information to your error reports:

- User information,
- Breadcrumbs,
- **any** custom "metadata",
- etc.

In Go, errors don't include a stacktrace, so it can be difficult to track where an error originates, if the location that it is being reported is different to where it is first created.
Similarly, any `context.Context` data may get lost if reporting at a location higher up the stack than where the error occurred.
To prevent this loss of information, this package exposes an `notifier.Wrap` method that can wrap an existing error with its stacktrace along with the context data at this location.
You can then safely call `notifier.Notify()` at a single location if you so wish.

```go
if err != nil {
	return notifier.Wrap(ctx, err, "unable to foo the bar")
}
```

You can safely re-wrap this error again should you so wish.

> IMPORTANT: In order to get the most out of this package, it is recommended to wrap your errors as far down the stack as possible.

If you would like to mark your error as unhandled, e.g. in the case of a panic, you should set these flags explicitly on the error returned from Wrap.

```go
if r := recover(); r != nil {
	// You can use notifier.Wrap as well, but this requires a type cast.
	err := bugsnag.Wrap(ctx, fmt.Errorf("unhandled panic when calling FooBar: %v", r))
	err.Unhandled = true
	err.Panic = true
	notifier.Notify(ctx, err)
}
```

### Enabling session tracking to establish a stability score

For each session, usually synonymous with 'request' (HTTP/gRPC/AMQP/PubSub/etc.), you should call `ctx = notifier.StartSession(ctx)`, usually performed in a middleware function.
Any **unhandled** errors that are reported along with this `ctx` will count negatively towards your stability score.

### Examples

Check out the `examples/` directory for more advanced blueprints:

- HTTP middleware.
- gRPC middleware.
- Reporting panics.
- `o11y` structural pattern.
- Advanced feature: Sanitizers
  - Intercepting calls to Bugsnag in tests.
  - Modify data sent to Bugsnag for sanitization/augmentation.
  - etc.

In particular, the examples highlight the additional the features that are different from, or not found, in the [official notifier](https://github.com/bugsnag/bugsnag-go).

## `bugsnag` command-line application usage

Install the binary with:

```console
$ go install github.com/kinbiko/bugsnag/cmd/bugsnag@latest
```

Report a new release to Bugsnag with

```console
$ bugsnag release \
    --api-key=$SOME_API_KEY \
    --app-version=5.1.2 \
    --release-stage=staging \
    --repository=https://github.com/kinbiko/bugsnag \
    --revision=64414f621b33680419b1cb3e5c622b510207ae1e \
    --builder-name=kinbiko \
    --metadata="KEY1=VALUE1,KEY2=VALUE2"
release info published for version 5.1.2
```

If you have environment variables `BUGSNAG_API_KEY` and `BUGSNAG_APP_VERSION` (or `APP_VERSION`) set, then you can skip the first two (required) flags.

See `bugsnag release --help` for more info and additional configuration parameters.
