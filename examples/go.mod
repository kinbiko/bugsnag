module github.com/kinbiko/bugsnag/examples

go 1.18

replace github.com/kinbiko/bugsnag => ../

require (
	github.com/DataDog/datadog-go v4.8.3+incompatible
	github.com/golang/protobuf v1.5.2
	github.com/kinbiko/bugsnag v0.0.0
	github.com/sirupsen/logrus v1.8.1
	google.golang.org/grpc v1.45.0
)

require (
	github.com/Microsoft/go-winio v0.5.2 // indirect
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
	golang.org/x/text v0.3.0 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
)
