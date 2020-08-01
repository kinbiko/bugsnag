# Panics

This package contains a simple application that shows the recommended pattern of how to report panics to Bugsnag.
In particular, panics should be treated as `unhandled`, so that they affect your application stability score.
