FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files and download dependencies
COPY go.mod go.sum* ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# Create a minimal production image
FROM alpine:latest

WORKDIR /app

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binary and static files
COPY --from=builder /server .
COPY --from=builder /app/api/graphql/schema.graphql ./api/graphql/schema.graphql

# Expose ports
EXPOSE 8080
EXPOSE 50051

# Run the service
CMD ["./server"] 