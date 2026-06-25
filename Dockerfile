# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git and certificates
RUN apk add --no-cache git ca-certificates

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build HTTP server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server cmd/server/main.go

# Build CLI tool
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/memory cmd/cli/main.go

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Copy SSL certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binaries
COPY --from=builder /app/server /app/server
COPY --from=builder /app/memory /usr/local/bin/memory

# Copy migrations folder (for migrations run on deploy or local run)
COPY --from=builder /app/migrations /app/migrations

# Expose HTTP port
EXPOSE 3210

# Command to run by default (HTTP Server)
CMD ["/app/server"]
