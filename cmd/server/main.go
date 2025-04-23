package main

import (
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/paper-social/notification-service/internal/models"
	"github.com/paper-social/notification-service/internal/queue"
	"github.com/paper-social/notification-service/internal/service"
	notificationProto "github.com/paper-social/notification-service/proto/generated/notification/proto"
	postProto "github.com/paper-social/notification-service/proto/generated/post/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Set seed for random generators
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create data store and initialize with sample data
	store := models.NewStore()
	store.InitSampleData()

	// Create notification queue with 5 workers and max 3 retries
	notificationQueue := queue.NewNotificationQueue(store, 5, 3)
	notificationQueue.Start()
	defer notificationQueue.Stop()

	// Create the gRPC services
	notificationService := service.NewNotificationService(store, notificationQueue)
	postService := service.NewPostService(store, notificationQueue)

	// Set up the gRPC server
	grpcServer := grpc.NewServer()
	notificationProto.RegisterNotificationServiceServer(grpcServer, notificationService)
	postProto.RegisterPostServiceServer(grpcServer, postService)
	reflection.Register(grpcServer)

	grpcListener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	go func() {
		log.Println("Starting gRPC server on :50051")
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)

	<-signalChan
	log.Println("Shutting down servers...")
	grpcServer.GracefulStop()

	log.Println("Servers stopped")
}
