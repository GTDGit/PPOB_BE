# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=Asia/Jakarta

# Copy binary from builder
COPY --from=builder /app/main .

# Create logs directory
RUN mkdir -p /app/logs

# Expose port
EXPOSE 8080

# Run
CMD ["./main"]
