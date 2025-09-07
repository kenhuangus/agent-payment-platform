# Use official Go image with CGO support
FROM golang:1.23-alpine AS builder

# Install git and build dependencies for CGO
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled for SQLite support
ENV CGO_ENABLED=1
RUN go build -o orchestration services/orchestration/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/orchestration .

# Expose port (if needed)
EXPOSE 8080

# Set environment variables
ENV USE_SQLITE=true
ENV DB_NAME=agent_payments

# Run the binary
CMD ["./orchestration"]
