# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application === x86
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o runtime main.go

# Runtime stage
FROM alpine:latest

WORKDIR /root/

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /build/runtime .

# Expose port 8080
EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["./runtime"]
