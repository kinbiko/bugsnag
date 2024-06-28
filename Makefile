LINTER_VERSION := v1.59.1

.PHONY: all
all: clean bin/bugsnag lint test-race

.PHONY: test
test:
	go test -count=1 -coverprofile=profile.cov ./...

.PHONY: test-race
test-race:
	go test -count=1 -race ./...

.PHONY: clean
clean:
	-rm -r ./bin

.PHONY: lint
lint: bin/linter
	./bin/linter run ./...

.PHONY: dependencies
dependencies: go.mod go.sum
	go get -v -t -d ./...

# Note: These next targets should not be PHONY

bin/bugsnag:
	go build -o bin/bugsnag ./cmd/bugsnag

bin/linter: Makefile .golangci.yml
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin $(LINTER_VERSION)
	mv ./bin/golangci-lint ./bin/linter
