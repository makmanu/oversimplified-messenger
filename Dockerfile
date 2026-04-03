# Build stage
FROM golang:1.21 AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 go build -o listener .

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies for sqlite3
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/listener .

# Create volume directory for database
RUN mkdir -p /app/data

# Expose port
EXPOSE 50505

# Set the entrypoint
ENTRYPOINT ["./listener"]
