.PHONY: all build run test clean docker docker-run gqlgen update-deps protogen test

all: clean build
build:
	@echo "Building server..."
	CGO_ENABLED=1 go build -o build/server cmd/server/main.go

run:
	@echo "Running server..."
	go run cmd/server/main.go

test:
	@echo "Running tests..."
	go test ./internal/... -v

docker:
	@echo "Building Docker image..."
	docker build -t paper-social/notification-service .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -p 50051:50051 paper-social/notification-service 

# Usage: make protogen ${package_name}
protogen:
	@echo "Generating proto files..."
	@PACKAGE_NAME=$(wordlist 2,2,$(MAKECMDGOALS)); \
	if [ -z "$$PACKAGE_NAME" ]; then \
		echo "Error: PACKAGE name is required. Usage: make protogen packagename"; \
		exit 1; \
	fi; \
	rm -rf proto/generated/$$PACKAGE_NAME; \
	mkdir -p proto/generated/$$PACKAGE_NAME; \
	protoc --go_out=./proto/generated/$$PACKAGE_NAME --go-grpc_out=./proto/generated/$$PACKAGE_NAME \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		proto/$$PACKAGE_NAME.proto

gqlgen:
	go run github.com/99designs/gqlgen generate

update-deps:
	go mod tidy

# Allow passing arguments to protogen
%:
	@:

