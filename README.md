# Paper.Social Notification Service

A distributed notification delivery service for Paper.Social that handles near real-time notifications when users publish posts.

## Features

- gRPC service for accepting new post events
- Distributed notification delivery to followers using Go routines and queues
- GraphQL API for retrieving user notifications
- Failure handling with retry logic
- Metrics endpoint for monitoring
- Dockerized for easy deployment

## Architecture

The notification service consists of the following components:

1. **gRPC Service**: Accepts new post events and triggers notification creation
2. **Notification Queue**: Processes notifications with a worker pool and handles retries
3. **GraphQL API**: Provides an endpoint to fetch notifications for a specific user
4. **Metrics API**: Exposes statistics about notification delivery

## Setup Instructions

### Prerequisites

- Go 1.16 or later
- Docker (optional)
- Make (optional)

### Using Make (Recommended)

The project includes a Makefile with common commands:

```
# Build the server and client
make build

# Run the server
make run

# Run the test client (as user u1)
make client

# Run tests
make test

# Clean build artifacts
make clean

# Build Docker image
make docker

# Run Docker container
make docker-run
```

### Local Development

1. Clone the repository:
   ```
   git clone https://github.com/paper-social/notification-service.git
   cd notification-service
   ```

2. Install dependencies:
   ```
   go mod download
   ```

3. Run the server:
   ```
   go run cmd/server/main.go
   ```

4. Test the service with the included client:
   ```
   # Run the server in one terminal
   go run cmd/server/main.go
   
   # Run the client in another terminal with a user ID
   go run cmd/client/main.go u1
   ```

### Using Docker

1. Build the Docker image:
   ```
   docker build -t paper-social/notification-service .
   ```

2. Run the container:
   ```
   docker run -p 8080:8080 -p 50051:50051 paper-social/notification-service
   ```

## API Usage

### gRPC: PublishPost

The gRPC service exposes a `PublishPost` method that accepts a new post and triggers notifications to followers.

Example using grpcurl:
```
grpcurl -plaintext -d '{"id": "p123", "user_id": "u1", "content": "Hello world!", "created_at": 1633027200}' localhost:50051 notification.NotificationService/PublishPost
```

Using the test client:
```
go run cmd/client/main.go u1
```
This will create a post as user with ID "u1" and notify all followers.

### GraphQL: getNotifications

The GraphQL API provides a `getNotifications` query to retrieve a user's notifications.

Endpoint: `http://localhost:8080/graphql`

Example query:
```graphql
{
  getNotifications(userId: "u1") {
    id
    userId
    postId
    postAuthorId
    content
    read
    createdAt
  }
}
```

You can test this using curl:
```
curl -X POST -H "Content-Type: application/json" --data '{"query": "{ getNotifications(userId: \"u1\") { id userId postId content read createdAt } }"}' http://localhost:8080/graphql
```

### Metrics API

The metrics endpoint provides statistics about notification delivery.

Endpoint: `http://localhost:8080/metrics`

Example:
```
curl http://localhost:8080/metrics
```

## Assumptions

1. All data is stored in-memory and pre-populated with sample data.
2. In a production environment, you would use a persistent database.
3. User authentication is not implemented and would be handled by another service.
4. The service simulates a 10% failure rate to demonstrate retry logic.

## Future Improvements

1. Add persistent storage for notifications and user data
2. Implement authentication and authorization
3. Add more comprehensive metrics and monitoring
4. Scale the worker pool based on queue size
5. Add proper logging and tracing
6. Implement rate limiting for notification deliveries 