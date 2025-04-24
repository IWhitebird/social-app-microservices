package main

import (
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/paper-social/notification-service/api"
	"github.com/paper-social/notification-service/internal/models"
	"github.com/paper-social/notification-service/internal/queue"
	"github.com/paper-social/notification-service/internal/service"
	notificationProto "github.com/paper-social/notification-service/proto/generated/notification/proto"
	postProto "github.com/paper-social/notification-service/proto/generated/post/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	resolver "github.com/paper-social/notification-service/graph"
	graph "github.com/paper-social/notification-service/graph/generated"
	"github.com/vektah/gqlparser/v2/ast"
	"google.golang.org/grpc/credentials/insecure"
)

var store *models.Store

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create store and notification queue first since we need it for all servers
	store = models.NewStore()
	store.InitSampleData()

	// Start all servers
	go RunHTTPServer()
	go RunGRPCServer()
	go RunGQlServer()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)

	<-signalChan
	log.Println("Shutting down servers...")
	log.Println("Servers stopped")
}

func RunHTTPServer() {
	const defaultPort = "3000"
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Create and start the HTTP server
	server := api.NewApiServer(store, port)
	log.Println("Starting HTTP server on http://localhost:" + port)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}

}

func RunGQlServer() {
	const defaultPort = "8080"
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Set up gRPC connections
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to notification service: %v", err)
	}
	defer conn.Close()

	notificationClient := notificationProto.NewNotificationServiceClient(conn)
	postClient := postProto.NewPostServiceClient(conn)

	srv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: resolver.NewResolver(notificationClient, postClient),
	}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func RunGRPCServer() {
	notificationQueue := queue.NewNotificationQueue(store, 5, 3)
	notificationQueue.Start()
	defer notificationQueue.Stop()

	// Create the gRPC services
	notificationService := service.NewNotificationService(store, notificationQueue)
	postService := service.NewPostService(store, notificationQueue)

	// Set up the gRPC server
	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()

	notificationProto.RegisterNotificationServiceServer(grpcServer, notificationService)
	postProto.RegisterPostServiceServer(grpcServer, postService)
	reflection.Register(grpcServer)

	grpcListener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	log.Println("Starting gRPC server on :50051")

	if err := grpcServer.Serve(grpcListener); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}

}
