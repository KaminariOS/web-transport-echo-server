
FROM golang:1.25-alpine AS builder

# Enable Go modules and configure working directory
WORKDIR /app

# Copy go.mod and go.sum first (for caching dependencies)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build statically linked binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o server .

# ---- Runtime stage ----
FROM scratch

# Copy binary from builder
COPY --from=builder /app/server /server

# Expose HTTP port
EXPOSE 4443

EXPOSE 8000 

# Use non-root user (optional but recommended)
USER 1000

# Run the binary
ENTRYPOINT ["/server"]
