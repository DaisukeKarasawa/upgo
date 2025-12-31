# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies (including gcc for CGO/SQLite and Node.js for frontend)
RUN apk add --no-cache git make gcc musl-dev nodejs npm

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build frontend (if web directory exists)
# Ensure web/dist directory exists even if build fails
RUN if [ -d "web" ]; then \
      cd web && \
      npm install && \
      npm run build || mkdir -p dist; \
    else \
      mkdir -p web/dist; \
    fi

# Build the application (CGO required for SQLite)
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o upgo ./cmd/server

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates curl sqlite

# Copy binary from builder
COPY --from=builder /build/upgo .

# Copy web assets from builder
# Note: web/dist directory is always created in builder stage (even if build fails)
COPY --from=builder /build/web/dist ./web/dist

# Create directories for data, logs, and backups
RUN mkdir -p data logs backups

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD curl -f http://localhost:8081/health || exit 1

# Run the application
CMD ["./upgo"]
