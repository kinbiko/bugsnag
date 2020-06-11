# Bugsnag

This package is an unofficial rewrite of the Bugsnag Go notifier.

## Differences to the official notifier

Please carefully consider the following table before deciding to use this package.
The official notifier may still be the right choice for you.

| Feature                         | [kinbiko/bugsnag][this-repo]                                                                                  | [bugsnag/bugsnag-go][official-repo]                                                                      |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| License                         | MIT                                                                                                           | MIT                                                                                                      |
| Dependency management           | Go mod                                                                                                        | Go get                                                                                                   |
| Version support                 | last 2 versions of Go                                                                                         | v1.7 and up                                                                                              |
| Send errors reports to Bugsnag  | Yes                                                                                                           | Yes                                                                                                      |
| Send session reports to Bugsnag | Yes                                                                                                           | Yes                                                                                                      |
| Cross platform                  | Yes                                                                                                           | Yes (reduced functionality)                                                                              |
| Officially supported by Bugsnag |                                                                                                               | Yes                                                                                                      |
| Battle tested in production     |                                                                                                               | Yes                                                                                                      |
| Built-in framework support      |                                                                                                               | `net/http`, Gin, Revel, Martini, and Negroni                                                             |
| Panic handling                  | Write your own middleware/`defer recover` call                                                                | `Recover`/`AutoNotify` **and automatic app crash detection** by spinning up another system process       |
| Stability Score judgement       | User defines whether an error is `Unhandled` on a per-`*bugsnag.Error`, which affects the stability score     | Only panics affect score                                                                                 |
| Control over transmitted data   | A more powerful but more involved solution through `Sanitizer`s gives full control over what's sent.          | `OnBeforeNotify` and config options is enough to prevent sending sensitive data.                         |
| Configuration                   | Minimal, relies on educated guesses that can be overridden in a `Sanitizer`.                                  | Lots of config options for tweaking behaviour.                                                           |
| Diagnostic data                 | Attach to `ctx` through `WithFoo` methods or `Error` through `Wrap`. One `Notify` per goroutine suffices.     | Attach on `Notify` -- to get the most diagnostics you should `Notify` when you first discover an `error` |
| Test friendliness               | Error and session reports are easily observed and discarded through `Sanitizer` functions.                    | Error reports are easily ignorable through `NotifyReleaseStages` configuration                           |
| Delivery                        | Async only, with a `*Flush()` method to be deferred in `main` to ensure all payloads are sent before exiting. | Sync or Async                                                                                            |
| Breadcrumbs                     | Yes                                                                                                           | May be implemented in the future.                                                                        |
| Serializable diagnostics        | Pass diagnostics across network boundaries to other services with `Serialize/Deserialize`                     | No, as it's a new feature only in `github.com/kinbiko/bugsnag`, and not officially supported by Bugsnag  |

Perhaps the biggest difference between the two packages is that this package is **significantly easier to maintain**, largely due to:

- not having to support old versions of Go,
- having a small exposed interface,
- not being v1.x yet, so backwards incompatibility is OK,
- not having to adhere to a consistent interface and feature set relative to notifiers of other platforms like Ruby/Android/JavaScript etc.

This has lead to more idiomatic code, at the cost of delegating some of the responsibility to the user.
Common use-cases such as creating custom middleware and attaching HTTP request data to payloads are well-documented.

[official-repo]: https://github.com/bugsnag/bugsnag-go
[this-repo]: https://github.com/kinbiko/bugsnag
