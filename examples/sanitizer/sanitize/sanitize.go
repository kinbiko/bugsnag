package sanitize

import (
	"context"
	"fmt"
	"os"

	"github.com/kinbiko/bugsnag"
)

func Run() {
	n, _ := bugsnag.New(bugsnag.Configuration{
		APIKey:       os.Getenv("BUGSNAG_API_KEY"),
		AppVersion:   "1.2.3",
		ReleaseStage: "production",
		ErrorReportSanitizer: func(ctx context.Context, r *bugsnag.JSONErrorReport) error {
			ev := r.Events[0]
			rs := ev.App.ReleaseStage
			if rs == "production" {
				ev.User.Name = ""
				ev.User.Email = ""
			}
			// Don't send any reports in dev
			if rs == "development" {
				return nil
			}
			return nil
		},
	})
	defer n.Close()

	ctx := n.WithUser(context.Background(), bugsnag.User{
		ID: "123",
		// The next two fields will be sanitized in production.
		Name:  "River Tam",
		Email: "river@serenity.space",
	})

	n.Notify(ctx, fmt.Errorf("oh ploppers"))
}
