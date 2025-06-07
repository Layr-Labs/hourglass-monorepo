# Build stage
FROM golang:1.23.6-alpine AS builder

WORKDIR /app
COPY main.go .

# Build the binary
RUN go build -o test-container main.go

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests (if needed)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/test-container .

# Create non-root user
RUN addgroup -g 1001 -S testuser && \
    adduser -u 1001 -S testuser -G testuser

USER testuser

# Expose port 8080
EXPOSE 8080

# Health check
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./test-container"]