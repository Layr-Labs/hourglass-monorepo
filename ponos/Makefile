.PHONY: build clean test

GO = $(shell which go)
BIN = ./bin

PROTO_OPTS=--proto_path=protos --go_out=paths=source_relative:protos

all: build/cmd/executor build/cmd/aggregator

# -----------------------------------------------------------------------------
# Dependencies
# -----------------------------------------------------------------------------
deps: deps/go deps/dev
	./scripts/installDeps.sh

.PHONY: deps/dev
deps/dev:
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
	go build -o ${BIN}/executor ./cmd/executor

.PHONY: build/cmd/aggregator
build/cmd/aggregator:
	go build -o ${BIN}/aggregator ./cmd/aggregator

# -----------------------------------------------------------------------------
# Run binaries
# -----------------------------------------------------------------------------
.PHONY: run/aggregator
run/aggregator:
	go run ./cmd/aggregator/*.go run

.PHONY: run/executor
run/executor:
	go run ./cmd/executor/*.go run

# -----------------------------------------------------------------------------
# Tests and linting
# -----------------------------------------------------------------------------

.PHONY: test
test:
	$(GO) test -v -p 1 -parallel 1 ./...

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
