package augment

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
			app := r.Events[0].App
			app.ID = "kinbiko-some-app-worker"
			app.Type = "worker" // Note: this field is indexed for 'free', i.e. no custom filter required.
			return nil
		},
	})
	defer n.Close()

	ctx := n.WithMetadatum(context.Background(), "app", "someCustomProperty", "someCustomValue")
	n.Notify(ctx, fmt.Errorf("oh ploppers"))
}
