BINARY := release-cli
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X github.com/0x1306e6d/release-cli/internal/cli.Version=$(VERSION)"

PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64

.PHONY: build install test lint fmt vet clean release-artifacts

build:
	go build $(LDFLAGS) -o dist/$(BINARY) ./cmd/release-cli

install:
	go install $(LDFLAGS) ./cmd/release-cli

release-artifacts:
	@for platform in $(PLATFORMS); do \
		os=$${platform%%-*}; \
		arch=$${platform##*-}; \
		echo "Building $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o dist/$(BINARY)-$$os-$$arch/$(BINARY) ./cmd/release-cli && \
		tar -czf dist/$(BINARY)-$$os-$$arch.tar.gz -C dist/$(BINARY)-$$os-$$arch $(BINARY) && \
		rm -rf dist/$(BINARY)-$$os-$$arch; \
	done

test:
	go test ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -rf dist/
