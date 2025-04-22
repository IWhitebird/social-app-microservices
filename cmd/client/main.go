package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	postProto "github.com/paper-social/notification-service/api/proto/gen/github.com/paper-social/notification-service/api/proto/post"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client [user_id]")
		fmt.Println("Example: client u1")
		os.Exit(1)
	}

	userID := os.Args[1]

	// Connect to gRPC server
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create client
	client := postProto.NewPostServiceClient(conn)

	// Create a new post
	post := &postProto.Post{
		Id:      fmt.Sprintf("p%d", time.Now().Unix()),
		UserId:  userID,
		Content: fmt.Sprintf("Test post from user %s at %s", userID, time.Now().Format(time.RFC3339)),
	}

	// Send post to server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := client.PublishPost(ctx, post)
	if err != nil {
		log.Fatalf("PublishPost failed: %v", err)
	}

	fmt.Printf("Response: %v\n", response)
}
