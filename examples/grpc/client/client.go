package client

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kinbiko/bugsnag"
	pb "github.com/kinbiko/bugsnag/examples/grpc/comments"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type application struct {
	ntf *bugsnag.Notifier
}

func Run() {
	ctx := context.Background()
	notifier, err := bugsnag.New(bugsnag.Configuration{APIKey: os.Getenv("BUGSNAG_API_KEY"), AppVersion: "v1.2.3", ReleaseStage: "dev"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't create Bugsnag notifier: %s", err.Error())
	}
	defer notifier.Close()

	app := &application{ntf: notifier}

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		panic(err)
	}
	defer func() { _ = conn.Close() }()

	client := pb.NewCommentServiceClient(conn)

	for {
		time.Sleep(1 * time.Second)
		app.callServer(ctx, client)
	}
}

func (a *application) callServer(ctx context.Context, client pb.CommentServiceClient) {
	ctx = a.ntf.WithBreadcrumb(ctx, bugsnag.Breadcrumb{
		Name: "gRPC call",
		Metadata: map[string]interface{}{
			"invoked at": time.Now().Format(time.RFC3339),
			"app name":   "client",
		},
	})

	ctx = a.ntf.WithBugsnagContext(ctx, "users/123/comments") // Pretend that this is a HTTP endpoint that initiated the gRPC call
	ctx = a.ntf.WithMetadata(ctx, "gRPC", map[string]interface{}{"client": "comments"})
	ctx = a.ntf.WithUser(ctx, bugsnag.User{ID: "123", Name: "River Tam", Email: "river@serentiy.space"})
	ctx = metadata.AppendToOutgoingContext(ctx, "bugsnag-diagnostics", string(a.ntf.Serialize(ctx)))

	fmt.Println("invoking GetComment")
	_, _ = client.GetComment(ctx, &pb.GetCommentReq{Id: "123"})
}
