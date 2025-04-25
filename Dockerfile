FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:latest

WORKDIR /app

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binary and static files
COPY --from=builder /server .
COPY --from=builder /app/api/graphql/schema.graphql ./api/graphql/schema.graphql

EXPOSE 
EXPOSE 8080
EXPOSE 50051

# Run the service
CMD ["./server"] 