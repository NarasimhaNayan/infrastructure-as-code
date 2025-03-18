# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Install required dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod ./
# COPY go.sum ./ # Uncomment this after running go mod tidy

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o devops-assign ./cmd/main

# Final stage
FROM alpine:3.14

WORKDIR /app

# Install required runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /app/devops-assign .
COPY --from=builder /app/migrations /app/migrations

# Set environment variables
ENV APP_ENV=production \
    POSTGRES_HOST=db \
    POSTGRES_PORT=5432 \
    POSTGRES_DB=vulndb \
    POSTGRES_USER=vulnuser \
    POSTGRES_PASSWORD=vulnpass

# Expose port
EXPOSE 8000

# Run the binary
CMD ["./devops-assign"]