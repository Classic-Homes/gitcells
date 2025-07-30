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
    -o gitcells ./cmd/gitcells

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh gitcells

# Set working directory
WORKDIR /home/gitcells

# Copy binary from builder stage
COPY --from=builder /app/gitcells /usr/local/bin/gitcells

# Copy default configuration
COPY --from=builder /app/.gitcells.yaml /home/gitcells/.gitcells.yaml.example

# Change ownership
RUN chown -R gitcells:gitcells /home/gitcells

# Switch to non-root user
USER gitcells

# Set default command
ENTRYPOINT ["gitcells"]
CMD ["--help"]

# Add labels
LABEL org.opencontainers.image.title="GitCells"
LABEL org.opencontainers.image.description="Version control for Excel files"
LABEL org.opencontainers.image.vendor="Classic Homes"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/Classic-Homes/gitcells"