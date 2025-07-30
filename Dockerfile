# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev
ARG BUILD_TIME
ARG TARGETOS
ARG TARGETARCH

# Build binary
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" \
    -o sheetsync ./cmd/sheetsync

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh sheetsync

# Set working directory
WORKDIR /home/sheetsync

# Copy binary from builder stage
COPY --from=builder /app/sheetsync /usr/local/bin/sheetsync

# Copy default configuration
COPY --from=builder /app/.sheetsync.yaml /home/sheetsync/.sheetsync.yaml.example

# Change ownership
RUN chown -R sheetsync:sheetsync /home/sheetsync

# Switch to non-root user
USER sheetsync

# Set default command
ENTRYPOINT ["sheetsync"]
CMD ["--help"]

# Add labels
LABEL org.opencontainers.image.title="SheetSync"
LABEL org.opencontainers.image.description="Version control for Excel files"
LABEL org.opencontainers.image.vendor="Classic Homes"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/Classic-Homes/sheetsync"