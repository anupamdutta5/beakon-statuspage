# Multi-stage Dockerfile for Enterprise Status Page
# Stage 1: Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o statuspage cmd/api/main.go

# Stage 2: Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S statuspage && \
    adduser -u 1001 -S statuspage -G statuspage

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/statuspage .

# Copy web assets
COPY --from=builder /app/web ./web

# Copy configs directory
COPY --from=builder /app/configs ./configs

# Change ownership to non-root user
RUN chown -R statuspage:statuspage /app

# Switch to non-root user
USER statuspage

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./statuspage"]
