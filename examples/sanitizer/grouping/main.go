package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kinbiko/bugsnag"
)

func main() {
	n, _ := bugsnag.New(bugsnag.Configuration{
		APIKey:       os.Getenv("BUGSNAG_API_KEY"),
		AppVersion:   "1.2.3",
		ReleaseStage: "production",
		ErrorReportSanitizer: func(ctx context.Context, r *bugsnag.JSONErrorReport) context.Context {
			// manually specify which frames are 'in project'. Bugsnag groups
			// by the top in-project stackframe by default.
			for _, ev := range r.Events {
				for _, ex := range ev.Exceptions {
					for _, sf := range ex.Stacktrace {
						if strings.Contains(sf.File, "examples") {
							sf.InProject = true
						}
					}
				}
			}

			return context.Background()
		},
	})
	defer n.Close()

	ctx := context.Background()
	n.Notify(ctx, n.Wrap(ctx, fmt.Errorf("oh ploppers"), "got some err"))
}
