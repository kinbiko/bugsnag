module github.com/kinbiko/bugsnag/examples

go 1.14

replace github.com/kinbiko/bugsnag => ../

require (
	github.com/DataDog/datadog-go v3.7.2+incompatible
	github.com/golang/protobuf v1.4.2
	github.com/kinbiko/bugsnag v0.0.0
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/objx v0.3.0 // indirect
	google.golang.org/grpc v1.30.0
)
