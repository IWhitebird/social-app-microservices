# Go Notification Service

A modern, scalable microservice architecture using Golang, featuring a queue system with a notification prototype.

## 🌟 Features

- Business Layer implemented with a gRPC Server, scalable via microservices.
- GraphQL and REST API Layer for communication between the business layer and the user.

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
├── api/                  # REST API
├── cmd/                  # Application entry points
│   └── server/           # Main server application
├── graph/                # GraphQL implementation (using gqlgen) and resolvers
│   ├── generated/        # Auto-generated GraphQL code (Auto Generated)
│   ├── model/            # GraphQL data models (Auto Generated)
│   ├── gql/              # GraphQL schema files used for generating other GraphQL files
├── internal/             # Private application code
│   ├── models/           # Data model / Store
│   ├── config/           # Environment Variables & Config
│   ├── queue/            # Notification queue implementation
│   └── service/          # gRPC service implementations
├── proto/                # Protocol Buffer definitions
│   └── generated/        # Generated gRPC code
```

## 🚀 Getting Started

### Prerequisites

- Go 1.24
- Protocol Buffers compiler (for development)

### For running the project

1. Clone the repository:
   ```
   git clone https://github.com/iwhitebird/social-app-microservices
   ```
   or if you downloaded the repository as a ZIP file from GitHub:
   ```
   unzip social-app-microservices-main.zip
   ```

2. Copy the `.env.example` file to `.env`:
   ```
   cp .env.example .env
   ```


3. Install dependencies:
   ```bash
   go mod download
   ```

4. Run the server:
   ```bash
   make run all
   ```
   or you can specify a single service to run instead of all. Available options are [all, http, graphql, grpc].

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

You can run this queries in on graphql playground 

```
query GetNotifications($userID: String!) {
  getNotifications(userID: $userID) {
    id
    userID
    postID
    content
    read
  }
}
```

```
mutation publishPost {
  publishPost(input: {
    userID: "u1",
    content: "myfirst post"
  }) {
    success
    message
    notificationsQueued
  }
}
```

```
query GetNotificationMetrics {
  getNotificationMetrics {
    totalNotificationsSent
    failedAttempts
    averageDeliveryTime
  }
}
```


### gRPC
- Service running on port 50051
- `PublishPost` - Publish a post and send corresponding notifications
- `GetNotifications` - Get notifications for a user
- `GetNotificationMetrics` - Get metrics about notification delivery

## 💻 Development

### Generating Protocol Buffers

   Generates Protocol Buffer files, automatically rewriting existing ones or creating new ones.

```bash
make protogen notification
make protogen post
```

### Generating GraphQL Code

   Generates GraphQL files using `gqlgen`, overwrites old ones, and creates new ones.

```bash
make gqlgen
```

### Running Tests
This will test our notification queue, notification service, and post service.

```bash
# Run all tests
make test
```


## Design Decisions


### Running the Servers
We are using the `.env` file for reading ports and command-line arguments for specifying which servers to run. This allows running individual servers. Currently, only the GraphQL and HTTP servers can be run individually, but the backend needs to run in the same environment due to in-memory storage.

### Backend Layer
For our backend layer, we are using gRPC for inter-service communication. gRPC is a binary-based TCP protocol for remote procedure calls. Our services can work independently and call procedures on other services. However, this introduces networking latency costs, but we have a good trade-off for scaling individual systems. We are using the official protogen compiler for compiling our .protofiles.

### Notification Queue
The notification queue is implemented using a worker queue pattern. This approach offers excellent control over concurrency and resource usage. By making the number of workers configurable, it allows for auto-scaling based on traffic. The notification queue is helpful for background tasks as it doesn't block the user from receiving a response and can retry on failure until successful.


### API Layer
For the API layer, we have implemented both HTTP (using Gin) and GraphQL (using `gqlgen`). `gqlgen` helps in automatically generating boilerplate code from schemas, making the process fast and maintainable, leaving the resolver implementation to the developer. These API layers also act as gRPC clients that communicate with the gRPC backend services.


## Future Upgrades & Current Flaws

- **Separate Notification Queue:** Refactor the notification queue into a more general-purpose, generic queue system that can accept dynamic functions, not limited to the notification service. Add options to spin up multiple worker servers via command-line arguments or environment variables. Implement a central datastore like Redis for communication and task management between different workers running in parallel.
- **Use a Real Database:** Introduce an actual database (e.g., PostgreSQL, MongoDB) to enable proper segregation of microservices, which are currently coupled due to shared in-memory storage.
- **Improve Logging:** Enhance logging beyond the current basic `log`. Integrate structured logging and metrics collection with tools like Prometheus and Grafana or the ELK stack for better observability.
- **Streamline Model Handling:** Create scripts to automate the generation or synchronization of models across different layers (datastore, proto, GraphQL). Currently, creating a model requires manual updates in potentially three places. Automating parts of this process would improve code scalability and reduce errors.
- **Enhance Error Handling:** Improve error handling and reporting. As mentioned in the logging point, integrate with monitoring tools like Datadog or Sentry for production-level error tracking and alerting.
- **Implement Security Measures:** Add rate limiting, authentication, and a firewall for the public-facing APIs.
