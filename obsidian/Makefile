.PHONY: all build test docker-build docker-push deploy clean

# Variables
REGISTRY ?= localhost:5000
VERSION ?= v1.0.0
IMAGE_SERVICE = $(REGISTRY)/obsidian-service:$(VERSION)
IMAGE_OPERATOR = $(REGISTRY)/obsidian-operator:$(VERSION)

# Build
all: build

build:
	go build -o bin/obsidian-service ./cmd/obsidian-service
	go build -o bin/operator ./cmd/operator
	go build -o bin/attestation-agent ./cmd/attestation-agent

test:
	go test -v ./...

# Docker
docker-build:
	docker build -f build/Dockerfile.service -t $(IMAGE_SERVICE) .
	docker build -f build/Dockerfile.operator -t $(IMAGE_OPERATOR) .

docker-push: docker-build
	docker push $(IMAGE_SERVICE)
	docker push $(IMAGE_OPERATOR)

# CRDs
generate:
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

manifests:
	controller-gen crd rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=deployments/kubernetes/crds

# Deploy
deploy: manifests
	kubectl apply -f deployments/kubernetes/crds/
	kubectl apply -f deployments/kubernetes/

# Local development
run-local:
	go run ./cmd/obsidian-service

run-operator:
	go run ./cmd/operator

# Clean
clean:
	rm -rf bin/
	kubectl delete -f deployments/kubernetes/ || true