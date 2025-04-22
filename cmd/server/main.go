package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/paper-social/notification-service/api/graphql"
	"github.com/paper-social/notification-service/api/metrics"
	notificationProto "github.com/paper-social/notification-service/api/proto/gen/github.com/paper-social/notification-service/api/proto/notification"
	postProto "github.com/paper-social/notification-service/api/proto/gen/github.com/paper-social/notification-service/api/proto/post"
	"github.com/paper-social/notification-service/internal/models"
	"github.com/paper-social/notification-service/internal/queue"
	"github.com/paper-social/notification-service/internal/service"
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

	// Start gRPC server
	grpcListener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	// Start gRPC server in a goroutine
	go func() {
		log.Println("Starting gRPC server on :50051")
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Create GraphQL resolver and handler
	graphqlResolver := graphql.NewResolver(store)
	graphqlHandler := graphql.Handler(graphqlResolver)

	// Create metrics handler
	metricsHandler := metrics.Handler(notificationQueue)

	// Set up HTTP server
	mux := http.NewServeMux()
	mux.Handle("/graphql", graphqlHandler)
	mux.HandleFunc("/metrics", metricsHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Paper.Social Notification Service\n\nAvailable endpoints:\n- /graphql: GraphQL API\n- /metrics: Metrics API")
	})

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()

	// Set up signal handling for graceful shutdowna ctrl + c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)

	// Wait for shutdown signal
	<-signalChan
	log.Println("Shutting down servers...")

	// Gracefully stop the gRPC server
	grpcServer.GracefulStop()

	// Gracefully stop the HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server shutdown failed: %v", err)
	}

	log.Println("Servers stopped")
}
