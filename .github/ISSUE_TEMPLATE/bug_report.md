---
name: Bug report
about: Describe an issue you've seen with this package
title: "[BUG]"
labels: bug
assignees: ''

---

<!-- TRY THIS FIRST:

This package comes with an `InternalErrorCallback` configuration option, which you can use to forward any internal errors to a callback of your choosing. You will be expected to provide logs of the errors that are passed to this callback, if this doesn't explain your problem.

-->

### What exactly did you do?

<!-- Code snippets are preferable here. -->

### What did you expect would happen?

<!-- Be explicit, and if possible, explain why. -->

### What actually happened?

<!-- Please paste any output that you saw in your logs.

If you're reporting a bug against the notifier, make sure you set a InternalErrorCallback configuration option in the constructor, and log all errors that appear here.

If you're reporting a bug against the 'bugsnag' command line application, please include the complete command you're executing (include the --debug flag please). You may redact your API key.
-->

```console
any output goes here
```

### Additional info

- Output from `go version`: [e.g. `go version go1.14.5 darwin/amd64`]
- Version of `github.com/kinbiko/bugsnag-go`: [e.g. `v.0.9.0`]

<!-- Please add any other context about the problem here. -->
