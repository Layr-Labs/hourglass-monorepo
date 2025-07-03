.PHONY: all build test clean clean-proto proto deps lint docker

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=obsidian

# Directories
PROTO_DIR=api/proto
GEN_DIR=api/proto

all: proto deps build

deps:
	$(GOMOD) download
	$(GOMOD) tidy

proto: clean-proto
	@echo "Generating Go code from proto files..."
	@mkdir -p $(GEN_DIR)/orchestrator
	@mkdir -p $(GEN_DIR)/proxy
	@mkdir -p $(GEN_DIR)/registry
	protoc --go_out=. --go_opt=module=github.com/hourglass/obsidian \
		--go-grpc_out=. --go-grpc_opt=module=github.com/hourglass/obsidian \
		-I. \
		$(PROTO_DIR)/*.proto

clean-proto:
	@echo "Cleaning generated proto files..."
	rm -rf $(GEN_DIR)/orchestrator
	rm -rf $(GEN_DIR)/proxy
	rm -rf $(GEN_DIR)/registry
	rm -f $(GEN_DIR)/*.pb.go

build:
	$(GOBUILD) -o bin/$(BINARY_NAME) -v ./cmd/obsidian

test:
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...

clean: clean-proto
	$(GOCMD) clean
	rm -rf bin/
	rm -f coverage.txt

lint:
	golangci-lint run

docker:
	docker build -t obsidian:latest .

# Development helpers
run: build
	./bin/$(BINARY_NAME) --config=config/development.yaml

run-gateway:
	$(GOCMD) run ./cmd/obsidian --config=config/development.yaml

# Install required tools
install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# CDK commands
cdk-install:
	cd cdk/obsidian-cdk && npm install

cdk-build: cdk-install
	cd cdk/obsidian-cdk && npm run build

cdk-deploy-dev: cdk-build
	cd cdk/obsidian-cdk && npx cdk deploy ObsidianEC2Stack --context env=dev

cdk-deploy-prod: cdk-build
	cd cdk/obsidian-cdk && npx cdk deploy ObsidianHybridStack --context env=prod