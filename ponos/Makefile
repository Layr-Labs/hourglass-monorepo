.PHONY: build clean test

GO = $(shell which go)
BIN = ./bin

GO_FLAGS=-ldflags "-X 'github.com/Layr-Labs/hourglass-monorepo/ponos/internal/version.Version=$(shell cat ../VERSION)' -X 'github.com/Layr-Labs/hourglass-monorepo/ponos/internal/version.Commit=$(shell cd ../ && git rev-parse --short HEAD)'"

PROTO_OPTS=--proto_path=protos --go_out=paths=source_relative:protos

all: deps/go build/cmd/executor build/cmd/aggregator

# -----------------------------------------------------------------------------
# Dependencies
# -----------------------------------------------------------------------------
deps: deps/go
	./scripts/installDeps.sh
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
	$(GO) install go.uber.org/mock/mockgen@latest
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

.PHONY: build/local-container
build/local-container:
	cd ../ && docker build -t hourglass:latest -f ./ponos/Dockerfile .

.PHONY: build/operator-container
build/operator-container:
	cd ../hourglass-operator && $(MAKE) docker-build IMG=hourglass/operator:test

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

.PHONY: build/sleepy-test-performer-container
build/sleepy-test-performer-container:
	docker build -t sleepy-hello-performer:latest -f internal/testUtils/sleepyPerformer/Dockerfile .

.PHONY: build/test-container
build/test-container:
	docker build -t ponos-test-container:latest ./internal/testUtils/testContainer/

.PHONY: test
test:
	GOFLAGS="-count=1" ./scripts/goTest.sh -v -p 1 -parallel 1 ./...

.PHONY: gen-mocks
gen-mocks:
	@echo "Generating mock for ethereum.Client interface..."
	mockgen -source=pkg/clients/ethereum/client.go -destination=mocks/mock_ethereum_client.go -package=mocks Client
	@echo "Generating mock for transactionLogParser.LogParser interface..."
	mockgen -source=pkg/transactionLogParser/transactionLogParser.go -destination=mocks/mock_log_parser.go -package=mocks LogParser
	@echo "Generating mock for IBlockContextManager interface..."
	mockgen -source=pkg/contextManager/blockContextManager.go -destination=mocks/mock_block_context_manager.go -package=mocks IBlockContextManager
	@echo "Generating mock for IContractStore interface..."
	mockgen -source=pkg/contractStore/contractStore.go -destination=mocks/mock_contract_store.go -package=mocks IContractStore
	@echo "Generating mock for IContractCaller interface..."
	mockgen -source=pkg/contractCaller/contractCaller.go -destination=mocks/mock_contract_caller.go -package=mocks IContractCaller
	@echo "Generating mock for IPeeringDataFetcher interface..."
	mockgen -source=pkg/peering/peering.go -destination=mocks/mock_peering_data_fetcher.go -package=mocks IPeeringDataFetcher

.PHONY: ci-test
ci-test: build/test-performer-container build/test-container build/sleepy-test-performer-container build/operator-container test

.PHONY: lint
lint:
	golangci-lint run --timeout "5m"

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
anvil/start/l1:
	anvil \
		--fork-url https://late-crimson-dew.quiknode.pro/56c000eadf175378343de407c56e0ccd62801fe9 \
		--fork-block-number 23477799 \
		--load-state ./internal/testData/anvil-l1-state.json \
		--chain-id 1 \
		--block-time 2 \
		--host 0.0.0.0

.PHONY: anvil/start
anvil/start/l2:
	anvil \
		--fork-url https://still-attentive-slug.base-mainnet.quiknode.pro/91bfa66d45c9f3ac7ef9e9ca35b2acc8ba41160a/ \
		--fork-block-number 36235532 \
		--load-state ./internal/testData/anvil-l2-state.json \
		--chain-id 8453 \
		--port 9545 \
		--block-time 2 \
		--host 0.0.0.0

