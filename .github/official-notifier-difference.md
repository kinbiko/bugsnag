# Difference to official notifier

The following highlights the main differences between [this package](https://github.com/kinbiko/bugsnag), which is not officially supported by Bugsnag, and the [official notifier](https://github.com/bugsnag/bugsnag-go).
In particular, where the official notifier has evolved along with the best practices in the Go ecosystem (and to its credit, never broken backwards compatibility), this package has had the advantage of learning from issues in the official notifier and designing for a simpler, and more powerful set of features.
As a result, this package is significantly easier to maintain, largely due to:

- Only supporting the last 2 versions of Go.
- Minimal public interface.
- Staying on a `v0.x.y` release until the public interface has been sufficiently battle tested.
- By being an unofficial rewrite, this package values idiomatic Go code over consistency with other platforms in the Bugsnag ecosystem.

Use these points to decide which package to use:

- **Go versions supported**: This package supports the last 2 releases of Go. The official notifier supports all versions since `v1.7`.
- **Capturing panics that crash your app**: The official notifier uses a package called `panicwrap` that spins up a monitoring process alongside your app, that listens for stacktraces printed to stderr coming from your app. It uses this to report even fatal crashes of your app. This package assumes no such responsibility, and prioritizes simplicity and maintainability instead, and expects your app to adequately protect against fatal crashes and report as necessary.
- **Cross platform**: Both notifiers are cross-platform, but the `panicwrap` functionality mentioned above does not work on all platforms.
- **Support**: Only the official notifier is supported by Bugsnag. Do not contact their support team with requests relating to this package.
- **Battle tested**: This package is used by a relatively small number of projects, whereas the official notifier has been in use by lots of teams over many years.
- **Framework support**: This package teaches you (via examples) how to set up middleware for a HTTP server and a gRPC server, and trusts that you can apply these ideas to your framework as necessary. The official notifier officially supports several HTTP frameworks such as Gin and Revel.
- **Stability score**: The stability score is a measure of the ratio of unhandled events to total number of sessions. Both packages support a similar solution for tracking sessions (`StartSession()`), but this package allows you to declare any `bugsnag.Error` as `unhandled`, meaning that you get to decide whether your app handled the event. The official notifier only counts panics that crash your app, or panics that are caught by `AutoNotify` as `unhandled`.
- **Control over transmitted data**: Both this package and the official notifier let you manipulate the data before sending it to Bugsnag via the `ErrorReportSanitizer` & `SessionReportSanitizer`, and `OnBeforeNotify` respectively, but this package gives you 100% control over the payload being sent, whereas the official notifier only lets you modify the data that may contain sensitive information.
- **Configuration options**: This package only lets you set a small number of configuration options compared to the official notifier, and relies on assumptions around best practices and opinions instead of extensive configurability. For example:

  - `ReleaseStage` and `AppVersion` are required in this package, whereas they are optional in the official notifier.
  - `Hostname` is not a configuration option in this package, as we set this automatically.
  - `AutoCaptureSessions` is not a configuration option in this package, as you are expected to write your own middleware.
  - `NotifyReleaseStages` and `ParamsFilters` are not configuration options in this package, as you have control of whether or not (and what) to send your payload with the `ErrorReportSanitizer`.
  - etc.

- **Augmenting diagnostic data**: This package keeps track of additional data (e.g. user info, request context, custom metadata) in a `context.Context` type via `With*` methods such as `WithUser`. The official notifier allows you to set more or less the same data, but users have to keep track of this data until the call to `Notify`, when it can be attached to the reported error. Additionally, in order to report as much diagnostic data as possible when using the official notifier, it's recommended to report the error where it first occurs. This package exposes a `Wrap` method that users may call to keep track of both diagnostics data in the `context.Context` and the stacktrace of the error at the point it is wrapped. This error can then be returned (and wrapped) further up the stack, without loss of data.
- **Delivery**: The official notifier lets you configure whether to send reports synchronously or asynchronously. This package will always send reports asynchronously, and provides a `Close` method to ensure that all reports are sent before the app quits.
- **Breadcrumbs**: This official Bugsnag feature is not yet available in the official notifier. This package lets you attach breadcrumbs via `WithBreadcrumb`.
- **Serializable diagnostics**: This **unofficial** feature is only available in this package. Since Go is often used for microservices it is often useful to transmit diagnostic data that relate to the same "request" across services. This package exposes `Serialize` and `Deserialize` for this purpose.
