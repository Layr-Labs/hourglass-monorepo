# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git make protoc protobuf-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate protobuf code
RUN make proto

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o obsidian ./cmd/obsidian

# Final stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/obsidian .

# Copy default config
COPY --from=builder /app/config/development.yaml /etc/obsidian/config.yaml

# Expose ports
EXPOSE 8080 9090 9091

# Run the binary
CMD ["./obsidian", "--config=/etc/obsidian/config.yaml"]