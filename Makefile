BINARY_NAME=cernopendata-client
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

LDFLAGS = -X main.buildVersion=$(VERSION)

.PHONY: all test test-integration test-short build clean build-all

all: build

test:
	go test ./...

test-integration: build
	go test -tags=integration ./...

test-short:
	go test -short ./...

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/cernopendata-client

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64

build-darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/cernopendata-client

build-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/cernopendata-client

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/cernopendata-client

build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/cernopendata-client
