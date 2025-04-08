# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum files first for better caching
COPY go.mod ./
COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o whatsapp-service ./cmd/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/whatsapp-service .

# Environment variables
ENV APP_ENV=production
ENV PORT=3000
ENV LOG_LEVEL=info

# Expose the application port
EXPOSE 3000

# Run the application
CMD ["./whatsapp-service"]