# Bugsnag Examples

This directory contains various code examples of how take advantage of the more powerful features of the notifier.

## Runnable examples

**You'll need to set the `BUGSNAG_API_KEY` environment variable** (only required for these examples, not for the `github.com/kinbiko/bugsnag` package) to the API key of one of your projects to see the events in your dashboard.

All of the examples can be run with:

```console
$ go run --trimpath ./cmd/$EXAMPLE_NAME
```

where `$EXAMPLE_NAME` is listed in each of the below sections.

For example:

```console
$ go run --trimpath ./cmd/cli
application done, closing down immediately, but errors/sessions are still reported.
```

Use `--trimpath` to get stackframes agnostic of which system the binary was built on. This is recommended for your production application as well.

### Command line application

> `$EXAMPLE_NAME=cli`

The main point to keep in mind when introducing bugsnag in command line applications is remembering to call `notifier.Close` when shutting down.
This is because command line applications tend to return very quickly, compared to longer-run processes such as servers.
Because the application shuts down very quickly, the `main` function will close which effectively kills all other goroutines running as well -- including the `bugsnag` looping goroutine that regularly fires off HTTP requests.
`notifier.Close` ensures sessions/errors are sent before the application closes.

### gRPC server and client

> `$EXAMPLE_NAME=grpc/client` and `$EXAMPLE_NAME=grpc/server`

This example contains a gRPC server and client that communicate with each other periodically, demonstrating how to:

- Propagate Bugsnag diagnostic data in gRPC requests.
- Set up gRPC middleware that will report warnings on handled errors, and panics as unhandled events.

### net/http HTTP server example

> `$EXAMPLE_NAME=nethttp`

This app runs a HTTP server on port `8080`, which returns a JSON payload for each request.
The HTTP endpoint is wrapped with two Bugsnag related middleware:

- Starts a session for each new HTTP request.
- Attach request information to Bugsnag reports, so that they appear under a 'request' tab in the dashboard.
- Report panics that happen in the app as unhandled panics (reflected in your stability score).

### Panicking applications

> `$EXAMPLE_NAME=panics`

This package contains a simple application that shows the recommended pattern of how to report panics to Bugsnag.
In particular, panics should be treated as `unhandled`, so that they affect your application stability score.

### Using the Sanitizer types

The `ErrorReportSanitizer` and `SessionReportSanitizer` types available in this package are perhaps the most powerful features of this package.
They allow you to modify the request body just before it's being sent to Bugsnag's servers, and even prevent the reporting entirely.
This can be useful for:

- Adding data in the payloads (`$EXAMPLE_NAME=sanitizer/augment`).
- Removing sensitive data in the payloads (`$EXAMPLE_NAME=sanitizer/sanitize`).
- Improving inaccurate error grouping (`$EXAMPLE_NAME=sanitizer/grouping`),
- etc.

## Non-runnable examples

Unlike the above examples these aren't intended to be run as stand-alone applications. Instead they show development patterns using this package.

### Observability structural pattern

The code in `o11y` is a structural pattern that combines observability (o11y) tools like logging, metrics, and error reporting into one struct for easy use in the application logic.

### Testing

The example in `sanitizer/tests` shows how to configure a notifier for testing purposes, so that you can make assertions against errors being reported to Bugsnag without actually reporting them and adding cost.
