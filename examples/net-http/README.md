# Simple net/http middleware example

This app runs a HTTP server on port `8080`, which returns a JSON payload for each request.
The HTTP endpoint is wrapped with two Bugsnag related middleware:

- Starts a session for each new HTTP request.
- Attach request information to Bugsnag reports, so that they appear under a 'request' tab in the dashboard.
- Report panics that happen in the app as unhandled panics (reflected in your stability score).
