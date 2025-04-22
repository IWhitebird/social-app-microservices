.PHONY: all build run test clean docker docker-run

all: clean build

build:
	@echo "Building server and client..."
	go build -o cmd/server/server cmd/server/main.go
	go build -o cmd/client/client cmd/client/main.go

run:
	@echo "Running server..."
	go run cmd/server/main.go

client:
	@echo "Running client with user u1..."
	go run cmd/client/main.go u1

test:
	@echo "Running tests..."
	go test ./... -v

clean:
	@echo "Cleaning..."
	rm -f cmd/server/server cmd/client/client

docker:
	@echo "Building Docker image..."
	docker build -t paper-social/notification-service .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -p 50051:50051 paper-social/notification-service 

protogen:
	@echo "Generating proto files..."
	@rm -rf api/proto/gen
	@mkdir -p api/proto/gen
	protoc --go_out=api/proto/gen --go-grpc_out=api/proto/gen api/proto/*.proto

