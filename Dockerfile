# Build stage
FROM golang:1.26 AS builder

WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build the binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /stickerbook ./cmd/stickerbook

# Runtime stage
FROM debian:stable-slim

# Install CA certificates for HTTPS requests
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /stickerbook /usr/local/bin/stickerbook

# Data directory for collection and packs
VOLUME /data

ENV STICKERBOOK_DATA_DIR=/data

ENTRYPOINT ["/usr/local/bin/stickerbook"]
CMD ["bot"]
