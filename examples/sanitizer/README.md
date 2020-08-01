# Sanitizer

The `ErrorReportSanitizer` and `SessionReportSanitizer` types available in this package are perhaps among the most powerful features of this package.
They allow you to modify the request body just before it's being sent to Bugsnag's servers, and even prevent the reporting entirely.
This can be useful for:

-   Adding data in the payloads (Example: `augment`).
-   Removing sensitive data in the payloads (Example: `sanitize`).
-   Verifying calls to Bugsnag in tests (Example: `tests`),
-   Improving inaccurate error grouping (Example: `grouping`),
-   Sampling noisy errors before they even reach Bugsnag's servers (Example: `sampling`),
