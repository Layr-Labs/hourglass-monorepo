# Multi-stage Docker build for Hourglass TypeScript Performer SDK
FROM node:18-alpine AS base

# Install build dependencies
RUN apk add --no-cache \
    python3 \
    make \
    g++ \
    protoc \
    protobuf-dev

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./
COPY tsconfig.json ./

# Install dependencies
RUN npm ci --only=production

# Development stage
FROM base AS development

# Install dev dependencies
RUN npm ci

# Copy source code
COPY . .

# Build the project
RUN npm run build

# Production stage
FROM node:18-alpine AS production

# Install runtime dependencies
RUN apk add --no-cache \
    dumb-init \
    tini

# Create non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S performer -u 1001

# Set working directory
WORKDIR /app

# Copy built application
COPY --from=development --chown=performer:nodejs /app/dist ./dist
COPY --from=development --chown=performer:nodejs /app/node_modules ./node_modules
COPY --from=development --chown=performer:nodejs /app/package*.json ./

# Create required directories
RUN mkdir -p /app/logs /app/data && \
    chown -R performer:nodejs /app/logs /app/data

# Switch to non-root user
USER performer

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD node -e "const http = require('http'); \
        const options = { hostname: 'localhost', port: 8080, path: '/health', timeout: 5000 }; \
        const req = http.request(options, (res) => { \
            process.exit(res.statusCode === 200 ? 0 : 1); \
        }); \
        req.on('error', () => process.exit(1)); \
        req.end();"

# Use tini for proper signal handling
ENTRYPOINT ["/sbin/tini", "--"]

# Default command
CMD ["node", "dist/index.js"]