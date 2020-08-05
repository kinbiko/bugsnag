# Bugsnag Go notifier (unofficial rewrite)

[![Build Status](https://github.com/kinbiko/bugsnag/workflows/Go/badge.svg)](https://github.com/kinbiko/bugsnag/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/kinbiko/bugsnag)](https://goreportcard.com/report/github.com/kinbiko/bugsnag)
[![Latest version](https://img.shields.io/github/tag/kinbiko/bugsnag.svg?label=latest%20version&style=flat)](https://github.com/kinbiko/bugsnag/releases)
[![Go Documentation](http://img.shields.io/badge/godoc-documentation-blue.svg?style=flat)](https://pkg.go.dev/github.com/kinbiko/bugsnag?tab=doc)
[![License](https://img.shields.io/github/license/kinbiko/bugsnag.svg?style=flat)](https://github.com/kinbiko/mokku/blob/master/.github/LICENSE)

Well-documented, maintainable, idiomatic, opinionated, and **unofficial** rewrite of the [Bugsnag Go notifier](https://github.com/bugsnag/bugsnag-go).

## Usage

Set up your `*bugsnag.Notifier`, the type that exposes the main API of this package, based on a `bugsnag.Configuration`.

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
    err := notifier.Wrap(ctx, fmt.Errorf("unhandled panic when calling FooBar: %v", r))
    err.Unhandled = true
    err.Panic = true
    notifier.Notify(ctx, err)
}
```

### Enabling session tracking to establish a stability score

For each session, usually synonymous with 'request' (HTTP/gRPC/AMQP/PubSub/etc.), you should call `ctx = notifier.StartSession(ctx)`, usually performed in a middleware function.
Any **unhandled** errors that are reported along with this `ctx` will count negatively towards your stability score.

## Examples

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
