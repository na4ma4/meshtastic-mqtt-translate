# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o meshtastic-mqtt-relay ./cmd/meshtastic-mqtt-relay

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/meshtastic-mqtt-relay .

# # Run as non-root user
# RUN adduser -D -u 1000 appuser
# USER appuser

# Expose any necessary ports (none required for MQTT client)

ENTRYPOINT ["./meshtastic-mqtt-relay"]
