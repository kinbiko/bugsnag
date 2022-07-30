package server

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/kinbiko/bugsnag"
	pb "github.com/kinbiko/bugsnag/examples/grpc/comments"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func Run() {
	notifier, err := bugsnag.New(bugsnag.Configuration{APIKey: os.Getenv("BUGSNAG_API_KEY"), AppVersion: "1.2.3", ReleaseStage: "dev"})
	if err != nil {
		panic(err)
	}
	defer notifier.Close()
	s := &server{Notifier: notifier}

	fmt.Println("starting gRPC server")
	s.gRPCStart()
}

func (s *server) bugsnagMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
	ctx = s.WithMetadatum(ctx, "app", "id", "comments-server")
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		ctx = s.Deserialize(ctx, []byte(md["bugsnag-diagnostics"][0]))
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			if recErr, ok := r.(error); ok {
				err = recErr
			}
			bErr := bugsnag.Wrap(ctx, err)
			bErr.Unhandled = true
			bErr.Panic = true
			// Be careful not to pass in the ctx from the gRPC call to Notify,
			// as this context is likely cancelled by the time the notifier
			// makes the HTTP request to Bugsnag's servers.
			// All the values in the ctx are attached as part of s.Wrap above.
			s.Notify(context.Background(), bErr)
			err = bErr
		}
	}()

	res, err = handler(ctx, req)
	if err != nil {
		s.Notify(ctx, s.Wrap(ctx, err))
	}
	return res, err
}

func (s *server) gRPCStart() {
	port := ":50051"
	lis, err := net.Listen("tcp", port) //nolint:gosec // This is just an example. Fix your own security issues
	if err != nil {
		panic("unable to open port " + port)
	}
	srv := grpc.NewServer(grpc.UnaryInterceptor(s.bugsnagMiddleware))
	pb.RegisterCommentServiceServer(srv, s)
	if err := srv.Serve(lis); err != nil {
		panic("failed to serve gRPC: " + err.Error())
	}
}

type server struct {
	pb.CommentServiceServer
	*bugsnag.Notifier
}

func (s *server) GetComment(ctx context.Context, in *pb.GetCommentReq) (*pb.GetCommentRes, error) {
	fmt.Println("GetComment invoked")
	if true {
		return nil, bugsnag.Wrap(ctx, fmt.Errorf("oh ploppers"))
	}
	return &pb.GetCommentRes{Id: in.GetId(), Msg: "I aim to misbehave"}, nil
}
