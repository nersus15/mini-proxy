## DEVELOPMENT STAGE
FROM golang:1.25-trixie AS development

# Set direktori kerja
WORKDIR /app

# Copy source code
COPY . .

# Download dependencies
RUN go work sync

# Install Watch tool untuk live reload saat development
RUN go install github.com/air-verse/air@latest

# Instal Delve untuk debugging
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Build Producer dengan CGO untuk Kafka support
RUN go build -o /app/main ./webcore/main.go

EXPOSE 2021

# CMD ["air", "--build.cmd", "go build -o /app/main /app/webcore/main.go", "--build.bin", "/app/webcore", "--debug.host", "0.0.0.0", "--debug.port", "2345"]
CMD ["air", "-c", "/app/.air-proxy.toml"]


## BUILD STAGE
FROM golang:1.25-trixie AS builder

# Install packages required for CGO and librdkafka di Debian
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    librdkafka-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy source code
COPY . .

# Download dependencies
RUN go work sync

# Build the application dengan CGO (Tanpa tag musl karena Debian menggunakan glibc)
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/main /app/webcore/main.go

# Build the migrate tool
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/migrate webcore/init/migrate.go


## PRODUCTION STAGE
FROM debian:trixie-slim AS production

# Install runtime dependencies untuk librdkafka dan CA certificates di Debian
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    librdkafka1 \
    libssl3 \
    zlib1g \
    libstdc++6 \
    libgcc-s1 \
    && rm -rf /var/lib/apt/lists/*

# Create user and group dengan limited privileges (non-root) menggunakan sintaks Debian
RUN groupadd -g 2000 webcore && \
    useradd -m -u 2000 -g webcore -s /bin/bash webcore

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main /usr/local/bin/webcore
COPY --from=builder /app/migrate /usr/local/bin/migrate

# Copy configuration files
COPY --from=builder /app/config.yaml.example ./config.yaml
COPY --from=builder /app/access.yaml.example ./access.yaml
COPY --from=builder /app/webcore/init/migrations ./webcore/init/migrations

RUN mkdir -p /var/sqlite/

# Set ownership of working directory to webcore user
RUN chown -R webcore:webcore /app
RUN chown -R webcore:webcore /var/sqlite

# Switch to webcore user
USER webcore

# Expose port
EXPOSE 2021

# Run the application
CMD ["webcore", "proxy"]
