.PHONY: help build test fmt lint install clean test-integration

APP_NAME=hgctl

VERSION_PKG=github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/version
TELEMETRY_PKG=github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/telemetry
CONFIG_PKG=github.com/Layr-Labs/hourglass-monorepo/hgctl-go/internal/config

LD_FLAGS=\
  -X '$(VERSION_PKG).Version=$(shell cat VERSION || echo "dev")' \
  -X '$(VERSION_PKG).Commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")' \
  -X '$(TELEMETRY_PKG).embeddedTelemetryApiKey=$${HGCTL_POSTHOG_API_KEY}'

GO_PACKAGES=./internal/... ./cmd/...
ALL_FLAGS=
GO_FLAGS=-ldflags "$(LD_FLAGS)"
GO=$(shell which go)
BIN=./bin

help: ## Show available commands
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@mkdir -p $(BIN)
	@go build $(GO_FLAGS) -o $(BIN)/$(APP_NAME) cmd/$(APP_NAME)/main.go

test: ## Run tests
	$(GO) test -v ./... -p 1

test-integration: ## Run integration tests
	$(GO) test -v ./... -p 1 -tags=integration

fmt: ## Format code
	@go fmt $(GO_PACKAGES)

lint: ## Run linter
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run $(GO_PACKAGES)

install: build ## Install binary
	@mkdir -p ~/bin
	@cp $(BIN)/$(APP_NAME) ~/bin/
	@echo "hgctl installed to ~/bin/$(APP_NAME)"
	@echo "Make sure ~/bin is in your PATH"

clean: ## Remove binary
	@rm -rf $(BIN) ~/bin/$(APP_NAME)

build/darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(ALL_FLAGS) $(GO) build $(GO_FLAGS) -o release/darwin-arm64/$(APP_NAME) cmd/$(APP_NAME)/main.go

build/darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(ALL_FLAGS) $(GO) build $(GO_FLAGS) -o release/darwin-amd64/$(APP_NAME) cmd/$(APP_NAME)/main.go

build/linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(ALL_FLAGS) $(GO) build $(GO_FLAGS) -o release/linux-arm64/$(APP_NAME) cmd/$(APP_NAME)/main.go

build/linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(ALL_FLAGS) $(GO) build $(GO_FLAGS) -o release/linux-amd64/$(APP_NAME) cmd/$(APP_NAME)/main.go

.PHONY: release
release: ## Build for all platforms
	@mkdir -p release/{darwin-arm64,darwin-amd64,linux-arm64,linux-amd64}
	$(MAKE) build/darwin-arm64
	$(MAKE) build/darwin-amd64
	$(MAKE) build/linux-arm64
	$(MAKE) build/linux-amd64

.PHONY: dev
dev: ## Run in development mode
	go run cmd/$(APP_NAME)/main.go
