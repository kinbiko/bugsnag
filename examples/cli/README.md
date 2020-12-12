# Command-line application example

The main point to keep in mind when introducing bugsnag in command line applications is remembering to call `notifier.Close` when shutting down.
This is because command line applications tend to return very quickly, compared to longer-run processes such as servers.
Because the application shuts down very quickly, the `main` function will close which effectively kills all other goroutines running as well -- including the bugsnag looping goroutine that regularly fires off HTTP requests.
