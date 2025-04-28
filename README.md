# Go Notification Service

A modern, scalable microservice architecture with golang. for highly scalable systems with realtime noitfication.

## 🌟 Features

- Business Layer with GRPC Server and scalaable via - mciroservices.
- GraphQL and REST Api Layyer for communationbn between business layera and user

## 🏗️ Basic Architecture

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

## 📁 Project Structure

```
├── api/                  # Rest Api
├── cmd/                  # Application entry points
│   └── server/           # Main server application
├── graph/                # GraphQL implementation (using gqlgen)
│   ├── generated/        # Auto-generated GraphQL code (Auto Generated)
│   ├── model/            # GraphQL data models (Auto Generated)
│   ├── gql/              # Graphql schema files which will be used for generating other graphql definations. 
│   └── /                 # GraphQL resolvers
├── internal/             # Private application code
│   ├── models/           # Data model / Store
│   ├── config/           # Env & Config
│   ├── queue/            # Notification queue implementation
│   └── service/          # gRPC service implementations
├── proto/                # Protocol Buffer definitions
│   └── generated/        # Generated gRPC code
│   └── /                 # Proto Definations
```

## 🚀 Getting Started

### Prerequisites

- Go 1.24
- Protocol Buffers compiler (for development)

### For running the project

1. Clone the repository / Extract from zip:

2. Copy .env.example file to .env

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run the server:
   ```bash
   make run
   ```

### Docker Deployment

To build and run using Docker:

```bash
make docker
make docker-run
```

## 🔌 API Endpoints (With provided env file)

### REST API
- `GET http://localhost:3000/api/metrics` - Get notification metrics

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
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests only 
make test-integration

# Run benchmarks
make test-bench

# Run all test types
make test-all

# Generate test coverage report
make test-coverage
```

For more details about testing, see the [test documentation](test/README.md).

## 🔄 Design Decisions

### In-memory Storage
The current implementation uses in-memory storage for simplicity, which means data isn't persistent across restarts. In a production environment, you'd want to replace this with a proper database.

### Notification Queue
The notification queue uses a worker pool pattern with configurable concurrency. This provides a good balance between performance and resource usage, but might need tuning for high-load scenarios.

### Protocol Support
Supporting multiple protocols (REST, GraphQL, gRPC) increases complexity but provides flexibility for different client requirements. gRPC is used internally for efficiency, while GraphQL and REST are offered for client convenience.

### Error Handling
The notification queue includes retry logic for failed deliveries, balancing reliability with performance. The exact retry strategy can be customized as needed.
