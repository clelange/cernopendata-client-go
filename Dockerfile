# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.buildVersion=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" -o cernopendata-client ./cmd/cernopendata-client

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
# ca-certificates: for TLS connections
RUN apk add --no-cache ca-certificates

# Copy the binary
COPY --from=builder /app/cernopendata-client /usr/local/bin/cernopendata-client

# Set entrypoint
ENTRYPOINT ["cernopendata-client"]
