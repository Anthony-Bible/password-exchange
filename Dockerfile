# Build stage
FROM ubuntu:24.04 AS builder

# Install dependencies
RUN apt-get update && \
    apt-get install -y golang-go git gcc musl-dev libmariadb-dev && \
    rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY app/go.mod app/go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY app/ ./

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o app

# Final stage
FROM ubuntu:24.04

# Install runtime dependencies
RUN apt-get update && \
    apt-get install -y libmariadb3 ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary from the builder stage
COPY --from=builder /app/app /app/app

# Copy templates and assets
COPY app/templates /templates
COPY app/assets /app/assets

# Set working directory
WORKDIR /app

# Command to run
ENTRYPOINT ["./app"]
