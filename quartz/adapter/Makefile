.PHONY: build docker-build docker-push run test clean

# Variables
BINARY_NAME=quartz-adapter
DOCKER_IMAGE=quartz-adapter
DOCKER_TAG?=latest
AWS_REGION?=us-east-1
AWS_ACCOUNT_ID?=$(shell aws sts get-caller-identity --query Account --output text 2>/dev/null)
ECR_REGISTRY=$(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com

# Build Go binary
build:
	go build -ldflags="-w -s" -o $(BINARY_NAME) main.go

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Push to ECR
docker-push: docker-build
	@echo "Logging in to ECR..."
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(ECR_REGISTRY)
	@echo "Tagging image..."
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(ECR_REGISTRY)/quartz-performer:$(DOCKER_TAG)
	@echo "Pushing image..."
	docker push $(ECR_REGISTRY)/quartz-performer:$(DOCKER_TAG)

# Run locally
run: build
	AVS_ADDRESS=0x1234567890123456789012345678901234567890 \
	OPERATOR_SET_ID=1 \
	PORT=8080 \
	./$(BINARY_NAME)

# Run with Docker
docker-run: docker-build
	docker run --rm \
		-e AVS_ADDRESS=0x1234567890123456789012345678901234567890 \
		-e OPERATOR_SET_ID=1 \
		-e AWS_REGION=$(AWS_REGION) \
		-p 8080:8080 \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Test
test:
	go test -v ./...

# Clean
clean:
	rm -f $(BINARY_NAME)
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true

# Download dependencies
deps:
	go mod download
	go mod tidy

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build Go binary"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-push  - Push Docker image to ECR"
	@echo "  run          - Run locally"
	@echo "  docker-run   - Run with Docker"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"