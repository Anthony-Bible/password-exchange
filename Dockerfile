# Build stage
FROM ubuntu:24.04 AS builder

# Install dependencies
RUN apt-get update && \
    apt-get install -y git gcc musl-dev libmariadb-dev protobuf-compiler wget unzip && \
    rm -rf /var/lib/apt/lists/*

# Install pinned Go toolchain matching app/go.mod (Go 1.25.0)
RUN wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz && \
    rm go1.25.0.linux-amd64.tar.gz

# Ensure Go toolchain and Go-installed binaries are on PATH
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"

# Install protoc-gen-go and protoc-gen-go-grpc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY app/go.mod app/go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code and protos
COPY protos/ /protos/
COPY app/ ./

# Generate protobuf files
RUN protoc --proto_path=/protos \
       --go_out=. --go_opt=module=github.com/Anthony-Bible/password-exchange/app \
       --go-grpc_out=. --go-grpc_opt=module=github.com/Anthony-Bible/password-exchange/app \
       /protos/database.proto /protos/encryption.proto /protos/message.proto

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o app

# Final stage
FROM ubuntu:24.04

# Install runtime dependencies
RUN apt-get update && \
    apt-get install -y libmariadb3 ca-certificates curl && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary from the builder stage
COPY --from=builder /app/app /app/app

# Copy templates and assets
COPY app/templates /templates
COPY app/assets /app/assets

# Copy API documentation
COPY app/api /app/api

# Set working directory
WORKDIR /app

# Health check for API endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/api/v1/health || exit 1

# Expose ports for web interface and API
EXPOSE 8080

# Command to run
ENTRYPOINT ["./app"]
