package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/kinbiko/bugsnag"
)

func Run() {
	ctx := context.Background()
	n, err := bugsnag.New(bugsnag.Configuration{APIKey: os.Getenv("BUGSNAG_API_KEY"), AppVersion: "1.2.3", ReleaseStage: "dev"})
	if err != nil {
		panic(err)
	}
	defer n.Close()

	n.Notify(n.StartSession(ctx), fmt.Errorf("ooi"))

	// There *may* be a few seconds delay however before it shows up in your dashboard.
	fmt.Println("application done, closing down immediately, but errors/sessions are still reported.")
}
