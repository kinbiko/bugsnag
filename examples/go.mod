module github.com/kinbiko/bugsnag/examples

go 1.18

replace github.com/kinbiko/bugsnag => ../

require (
	github.com/DataDog/datadog-go v4.8.3+incompatible
	github.com/golang/protobuf v1.5.2
	github.com/kinbiko/bugsnag v0.0.0
	github.com/sirupsen/logrus v1.8.1
	google.golang.org/grpc v1.53.0
)

require (
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/stretchr/testify v1.8.0 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)
