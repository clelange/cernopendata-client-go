.PHONY: all test build clean

all: build

test:
	go test ./...

build:
	go build -o cernopendata-client ./cmd/cernopendata-client

clean:
	rm -f cernopendata-client
