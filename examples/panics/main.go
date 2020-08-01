package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kinbiko/bugsnag"
)

func main() {
	ctx := context.Background()
	n, _ := bugsnag.New(bugsnag.Configuration{APIKey: os.Getenv("BUGSNAG_API_KEY"), AppVersion: "1.2.3", ReleaseStage: "dev"})
	defer n.Close()

	defer func() {
		r := recover()
		if r == nil {
			return
		}
		err := fmt.Errorf("%v", r)
		if recErr, ok := r.(error); ok {
			err = recErr
		}
		bErr := n.Wrap(ctx, err)
		bErr.Unhandled = true
		bErr.Panic = true
		n.Notify(ctx, bErr)
	}()

	panic("oh ploppers")
}
