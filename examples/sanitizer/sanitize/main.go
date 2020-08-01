package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kinbiko/bugsnag"
)

func main() {
	n, _ := bugsnag.New(bugsnag.Configuration{
		APIKey:       os.Getenv("BUGSNAG_API_KEY"),
		AppVersion:   "1.2.3",
		ReleaseStage: "production",
		ErrorReportSanitizer: func(ctx context.Context, r *bugsnag.JSONErrorReport) context.Context {
			if ev := r.Events[0]; ev.App.ReleaseStage == "production" {
				ev.User.Name = ""
				ev.User.Email = ""
			}
			return context.Background()
		},
	})
	defer n.Close()

	ctx := bugsnag.WithUser(context.Background(), bugsnag.User{
		ID: "123",
		// The next two fields will be sanitized in production.
		Name:  "River Tam",
		Email: "river@serenity.space",
	})

	n.Notify(ctx, fmt.Errorf("oh ploppers"))
}
