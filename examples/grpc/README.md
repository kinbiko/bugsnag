# gRPC example

This example contains a gRPC server and client that communicate with each other periodically, demonstrating how to:

- Propagate Bugsnag diagnostic data in gRPC requests.
- Set up gRPC middleware that will report warnings on handled errors, and panics as unhandled events.

## Generating the comments package from `.proto` file from scratch:

> Note: At the time of writing, the grpc/protobuf toolchain is in the middle of a migration, and the installation instructions below reflect a proven process at the time.
> This process is likely much simpler now.

Ensure you have `protoc` and `protoc-gen-go` installed:

```console
# Assuming you're on a mac:
brew install protobuf # Includes the 'protoc' binary.

# Install the protoc-gen-go plugin to generate go sourcecode from the compiled protobuf
# Fetch the package
cd
go get github.com/golang/protobuf/protoc-gen-go
cd -

# Install the v1.3.0 release.
cd $GOPATH/src/github.com/golang/protobuf/protoc-gen-go
git checkout v1.3.0
go install .
cd -
```

Then ensure that the `comments/` directory exists:

```console
mkdir comments
```

Now, you should be able to (re-)generate the `comments`, assuming you're in this directory:

```console
protoc -I . comments.proto --go_out=plugins=grpc:comments
```
