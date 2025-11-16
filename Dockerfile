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

RUN apk --no-cache add ca-certificates curl tini

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/meshtastic-mqtt-relay .

ENV HEALTHCHECK_PORT=8099
HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 CMD [ "curl", "--fail", "http://localhost:${HEALTHCHECK_PORT}/" ]
# # Run as non-root user
# RUN adduser -D -u 1000 appuser
# USER appuser

# Expose any necessary ports (none required for MQTT client)

ENTRYPOINT ["./meshtastic-mqtt-relay"]
