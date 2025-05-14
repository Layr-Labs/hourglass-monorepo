.PHONY: build clean test

GO = $(shell which go)
BIN = ./bin

GO_FLAGS=-ldflags "-X 'github.com/Layr-Labs/hourglass-monorepo/ponos/internal/version.Version=$(shell cat VERSION)' -X 'github.com/Layr-Labs/hourglass-monorepo/ponos/internal/version.Commit=$(shell cd ../ && git rev-parse --short HEAD)'"

PROTO_OPTS=--proto_path=protos --go_out=paths=source_relative:protos

all: deps/go build/cmd/executor build/cmd/aggregator build/cmd/bls-helper

# -----------------------------------------------------------------------------
# Dependencies
# -----------------------------------------------------------------------------
deps: deps/go
	./scripts/installDeps.sh
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
	cd protos && buf dep update


.PHONY: deps/go
deps/go:
	${GO} mod tidy

# -----------------------------------------------------------------------------
# Build protobufs
# -----------------------------------------------------------------------------
.PHONY: proto
proto:
	buf generate protos

# -----------------------------------------------------------------------------
# Build binaries
# -----------------------------------------------------------------------------
.PHONY: cmd/executor
build/cmd/executor:
	go build $(GO_FLAGS) -o ${BIN}/executor ./cmd/executor

.PHONY: build/cmd/executor/linux-arm64
build/cmd/executor/linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(GO_FLAGS) -o ${BIN}/linux/arm64/executor ./cmd/executor

.PHONY: build/cmd/executor/linux-amd64
build/cmd/executor/linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(GO_FLAGS) -o ${BIN}/linux/arm64/executor ./cmd/executor

.PHONY: build/cmd/executor/darwin-arm64
build/cmd/executor/darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build $(GO_FLAGS) -o ${BIN}/darwin/arm64/executor ./cmd/executor

.PHONY: build/cmd/executor/darwin-amd64
build/cmd/executor/darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(GO_FLAGS) -o ${BIN}/darwin/arm64/executor ./cmd/executor


.PHONY: build/cmd/aggregator
build/cmd/aggregator:
	go build $(GO_FLAGS) -o ${BIN}/aggregator ./cmd/aggregator

.PHONY: build/cmd/aggregator/linux-arm64
build/cmd/aggregator/linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(GO_FLAGS) -o ${BIN}/linux/arm64/aggregator ./cmd/aggregator

.PHONY: build/cmd/aggregator/linux-amd64
build/cmd/aggregator/linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(GO_FLAGS) -o ${BIN}/linux/arm64/aggregator ./cmd/aggregator

.PHONY: build/cmd/aggregator/darwin-arm64
build/cmd/aggregator/darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build $(GO_FLAGS) -o ${BIN}/linux/darwin/aggregator ./cmd/aggregator

.PHONY: build/cmd/aggregator/darwin-amd64
build/cmd/aggregator/darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(GO_FLAGS) -o ${BIN}/linux/darwin/aggregator ./cmd/aggregator

# bls-helper
.PHONY: build/cmd/bls-helper
build/cmd/bls-helper:
	go build $(GO_FLAGS) -o ${BIN}/bls-helper ./cmd/blsHelper

.PHONY: build/cmd/bls-helper/linux-arm64
build/cmd/bls-helper/linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(GO_FLAGS) -o ${BIN}/linux/arm64/bls-helper ./cmd/blsHelper

.PHONY: build/cmd/bls-helper/linux-amd64
build/cmd/bls-helper/linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(GO_FLAGS) -o ${BIN}/linux/arm64/bls-helper ./cmd/blsHelper

.PHONY: build/cmd/bls-helper/darwin-arm64
build/cmd/bls-helper/darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build $(GO_FLAGS) -o ${BIN}/linux/darwin/bls-helper ./cmd/blsHelper

.PHONY: build/cmd/bls-helper/darwin-amd64
build/cmd/bls-helper/darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(GO_FLAGS) -o ${BIN}/linux/darwin/bls-helper ./cmd/blsHelper

.PHONY: build/local-container
build/local-container:
	cd ../ && docker build -t hourglass:latest -f ./ponos/Dockerfile .

# -----------------------------------------------------------------------------
# Release builds
# -----------------------------------------------------------------------------
.PHONY: release
release:
	$(MAKE) build/cmd/aggregator/darwin-arm64
	$(MAKE) build/cmd/aggregator/darwin-amd64
	$(MAKE) build/cmd/aggregator/linux-arm64
	$(MAKE) build/cmd/aggregator/linux-amd64
	$(MAKE) build/cmd/executor/darwin-arm64
	$(MAKE) build/cmd/executor/darwin-amd64
	$(MAKE) build/cmd/executor/linux-arm64
	$(MAKE) build/cmd/executor/linux-amd64

release-with-docker:
	cd ../ && docker run --rm -v `pwd`:/build golang:1.23.6-bookworm /bin/bash -c "git config --global --add safe.directory /build && cd /build/ponos && make release"

.PHONY: build/cmd/keygen
build/cmd/keygen:
	go build -o ${BIN}/keygen ./cmd/keygen

# -----------------------------------------------------------------------------
# Run binaries
# -----------------------------------------------------------------------------
.PHONY: run/aggregator
run/aggregator:
	go run ./cmd/aggregator/*.go run $(ARGS)

.PHONY: run/executor
run/executor:
	go run ./cmd/executor/*.go run

# -----------------------------------------------------------------------------
# Tests and linting
# -----------------------------------------------------------------------------

.PHONY: build/test-performer
build/test-performer:
	cd ../demo && make build

.PHONY: build/test-performer-container
build/test-performer-container:
	cd ../demo && make build-container

.PHONY: test
test:
	GOFLAGS="-count=1" $(GO) test -v -p 1 -parallel 1 ./...

.PHONY: ci-test
ci-test: build/test-performer-container test

.PHONY: lint
lint:
	golangci-lint run

.PHONY: fmt
fmt:
	gofmt -w .

.PHONY: fmtcheck
fmtcheck:
	@unformatted_files=$$(gofmt -l .); \
	if [ -n "$$unformatted_files" ]; then \
		echo "The following files are not properly formatted:"; \
		echo "$$unformatted_files"; \
		echo "Please run 'gofmt -w .' to format them."; \
		exit 1; \
	fi

.PHONY: anvil/start
anvil/start:
	anvil \
		--fork-url https://tame-fabled-liquid.quiknode.pro/f27d4be93b4d7de3679f5c5ae881233f857407a0/ \
		--fork-block-number 22396947 \
		--load-state ./internal/testData/anvil-state.json \
		--chain-id 31337 \
		--block-time 5
