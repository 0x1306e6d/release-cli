BINARY := release-cli
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X github.com/0x1306e6d/release-cli/internal/cli.Version=$(VERSION)"

.PHONY: build install test lint fmt vet clean

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/release-cli

install:
	go install $(LDFLAGS) ./cmd/release-cli

test:
	go test ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -rf bin/
