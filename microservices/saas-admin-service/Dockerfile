# Build stage
FROM golang:1.25 AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && \
    apt-get install -y git

# Copy the entire project to the container
COPY . /app/

# Set the working directory to the saas-admin-service
WORKDIR /app/microservices/saas-admin-service

# Download all dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o saas-admin ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/saas-admin .
COPY --from=builder /app/.env .

# Install CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Expose port
EXPOSE 8089

# Command to run the executable
CMD ["./saas-admin"]
