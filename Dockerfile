# Multi-stage build for optimal image size
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version info
ARG VERSION=dev
ARG BUILD_TIME
ARG COMMIT

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X 'main.version=${VERSION}' -X 'main.buildTime=${BUILD_TIME}' -X 'main.commit=${COMMIT}'" \
    -o aid ./cmd/aid/

# Production image
FROM scratch

# Copy certificates for HTTPS support
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /build/aid /usr/local/bin/aid

# Set the binary as entrypoint
ENTRYPOINT ["/usr/local/bin/aid"]

# Default command
CMD ["--help"]

# Labels for better maintainability
LABEL org.opencontainers.image.title="AI Distiller"
LABEL org.opencontainers.image.description="High-performance CLI tool for extracting essential code structure from large codebases"
LABEL org.opencontainers.image.url="https://github.com/janreges/ai-distiller"
LABEL org.opencontainers.image.source="https://github.com/janreges/ai-distiller"
LABEL org.opencontainers.image.vendor="Jan Reges"
LABEL org.opencontainers.image.licenses="MIT"