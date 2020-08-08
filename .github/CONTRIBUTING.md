# Contributing

> Be kind and inclusive, both to people and when talking about the Bugsnag product.

## PRs

If you wish to contribute a change to this project, please create an issue first to discuss.
If you do not raise an issue before submitting a (significant) PR then your PR may be dismissed without much consideration.

Changes that aims to align the feature set of this package to the [officially supported `bugsnag-go` package](https://github.com/bugsnag/bugsnag-go) are unlikely to be accepted.
On the other hand, if there's a new feature implemented in the Bugsnag product that affects this notifier, then your chances are good.

PRs only improving the documentation are welcome without raising an issue first.

Ensure that:

1. The tests pass. (Enable GitHub actions on your fork of the repository in order to run CI)
1. The linter has 0 issues against the root package. Issues against the examples are OK within reason.
1. You don't introduce any new dependencies (ask in an issue if you feel strongly that it's necessary).
1. You follow the existing commit message convention.
