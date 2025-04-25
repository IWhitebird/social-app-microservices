# Go Notification Service

A modern, scalable notification system built with Go that provides real-time notifications for social applications through a multi-protocol architecture.

## 🌟 Features

- **Real-time notifications** via gRPC streaming
- **Multi-protocol support**: RESTful API, GraphQL, and gRPC
- **Efficient notification queue** with configurable workers and retry mechanism
- **Metrics collection** for performance monitoring
- **Dockerized** for easy deployment
- **Easily extendable** for various notification types

## 🏗️ Architecture

This project follows a microservice architecture with a focus on performance and scalability:

```
┌─────────────┐         ┌────────────┐
│  HTTP API   │         │   GraphQL  │
└──────┬──────┘         └─────┬──────┘
       │                      │
       ▼                      ▼
┌─────────────────────────────────────┐
│              gRPC Server            │
├─────────────────────────────────────┤
│   Post Service  │ Notification Svc  │
└────────┬────────┴────────┬──────────┘
         │                 │
         ▼                 ▼
┌────────────────┐ ┌──────────────────┐
│   Data Store   │ │Notification Queue│
└────────────────┘ └──────────────────┘
```

The architecture uses:
- **gRPC** for efficient internal service communication
- **Protocol Buffers** for compact, type-safe data serialization
- **GraphQL** for flexible data fetching
- **REST API** for traditional HTTP endpoints
- **In-memory queue** for reliable notification delivery

## 📁 Project Structure

```
├── api/                  # HTTP API implementation
├── cmd/                  # Application entry points
│   └── server/           # Main server application
├── graph/                # GraphQL implementation (using gqlgen)
│   ├── generated/        # Auto-generated GraphQL code
│   ├── model/            # GraphQL data models
│   └── resolvers/        # GraphQL resolvers
├── internal/             # Private application code
│   ├── models/           # Data models
│   ├── queue/            # Notification queue implementation
│   └── service/          # gRPC service implementations
├── proto/                # Protocol Buffer definitions
│   └── generated/        # Generated gRPC code
├── build/                # Compiled application
├── Dockerfile            # Docker container definition
├── Makefile              # Build automation
└── go.mod, go.sum        # Go module dependencies
```

## 🚀 Getting Started

### Prerequisites

- Go 1.20 or higher
- Docker (optional, for containerized deployment)
- Protocol Buffers compiler (for development)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/go-notification.git
   cd go-notification
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the application:
   ```bash
   make build
   ```

4. Run the server:
   ```bash
   make run
   ```

### Docker Deployment

To build and run using Docker:

```bash
make docker
make docker-run
```

## 🔌 API Endpoints

### REST API
- `GET /api/notifications/:userId` - Get notifications for a user
- `GET /api/metrics` - Get notification metrics

### GraphQL
- Playground: http://localhost:8080/
- Endpoint: http://localhost:8080/query

### gRPC
- Service running on port 50051
- `GetNotifications` - Stream notifications for a user
- `GetNotificationMetrics` - Get metrics about notification delivery

## 💻 Development

### Generating Protocol Buffers

```bash
make protogen notification
make protogen post
```

### Generating GraphQL Code

```bash
make gqlgen
```

### Running Tests

```bash
make test
```

## 🔄 Trade-offs and Design Decisions

### In-memory Storage
The current implementation uses in-memory storage for simplicity, which means data isn't persistent across restarts. In a production environment, you'd want to replace this with a proper database.

### Notification Queue
The notification queue uses a worker pool pattern with configurable concurrency. This provides a good balance between performance and resource usage, but might need tuning for high-load scenarios.

### Protocol Support
Supporting multiple protocols (REST, GraphQL, gRPC) increases complexity but provides flexibility for different client requirements. gRPC is used internally for efficiency, while GraphQL and REST are offered for client convenience.

### Error Handling
The notification queue includes retry logic for failed deliveries, balancing reliability with performance. The exact retry strategy can be customized as needed.

## 📊 Performance Considerations

- The gRPC streaming approach minimizes network overhead for real-time notifications
- The worker pool in the notification queue prevents system overload during spikes
- Connection pooling is used for efficient resource management
