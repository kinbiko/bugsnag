module github.com/kinbiko/bugsnag/examples

go 1.24.0

replace github.com/kinbiko/bugsnag => ../

require (
	github.com/DataDog/datadog-go v4.8.3+incompatible
	github.com/golang/protobuf v1.5.4
	github.com/kinbiko/bugsnag v0.0.0
	github.com/sirupsen/logrus v1.9.3
	google.golang.org/grpc v1.79.3
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
