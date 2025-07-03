# Build stage
FROM golang:1.24-bookworm AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /owui-metrics-exporter

# Runtime stage
FROM gcr.io/distroless/static:nonroot

# Copy binary from builder stage
COPY --from=builder /owui-metrics-exporter /owui-metrics-exporter

# Use non-root user
USER nonroot:nonroot

# Expose metrics port
EXPOSE 8080

# Command to run the executable
ENTRYPOINT ["/owui-metrics-exporter"]
