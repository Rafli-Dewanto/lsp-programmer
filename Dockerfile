# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
RUN go build -o main cmd/main.go

# Runtime stage
FROM alpine:3.18

# Add necessary certificates
RUN apk add --no-cache ca-certificates

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs
COPY .env .

# Expose application port
EXPOSE 8080

# Run the application
CMD ["./main"]