package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iwhitebird/social-app-microservices/api"
	"github.com/iwhitebird/social-app-microservices/internal/config"
	"github.com/iwhitebird/social-app-microservices/internal/models"
	"github.com/iwhitebird/social-app-microservices/internal/queue"
	"github.com/iwhitebird/social-app-microservices/internal/service"
	notificationProto "github.com/iwhitebird/social-app-microservices/proto/generated/notification/proto"
	postProto "github.com/iwhitebird/social-app-microservices/proto/generated/post/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	resolver "github.com/iwhitebird/social-app-microservices/graph"
	graph "github.com/iwhitebird/social-app-microservices/graph/generated"
	"github.com/vektah/gqlparser/v2/ast"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	store  *models.Store
	logger *slog.Logger
)

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	store = models.NewStore()
	store.InitSampleData()

	logger.Info("starting servers", "config", cfg)

	if cfg.IsServerEnabled("http") {
		go RunHTTPServer(cfg)
	}
	if cfg.IsServerEnabled("grpc") {
		go RunGRPCServer(cfg)
	}
	if cfg.IsServerEnabled("gql") {
		go RunGQlServer(cfg)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)

	<-signalChan
	logger.Info("shutting down servers...")
	logger.Info("servers stopped")
}

func RunHTTPServer(cfg *config.Config) {
	grpcAddr := fmt.Sprintf("%s:%s", cfg.GRPCHost, cfg.GRPCPort)
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to connect to notification service", "error", err)
		return
	}
	defer conn.Close()

	notificationClient := notificationProto.NewNotificationServiceClient(conn)
	postClient := postProto.NewPostServiceClient(conn)

	server := api.NewHttpApi(notificationClient, postClient, cfg.HTTPPort)

	logger.Info("starting HTTP server", "port", cfg.HTTPPort)

	if err := server.Start(); err != nil {
		logger.Error("failed to start HTTP server", "error", err)
	}
}

func RunGQlServer(cfg *config.Config) {
	grpcAddr := fmt.Sprintf("%s:%s", cfg.GRPCHost, cfg.GRPCPort)
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to connect to services", "error", err)
		return
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

	logger.Info("starting GraphQL server", "port", cfg.GQLPort, "playground", fmt.Sprintf("http://localhost:%s/", cfg.GQLPort))

	if err := http.ListenAndServe(":"+cfg.GQLPort, nil); err != nil {
		logger.Error("failed to start GraphQL server", "error", err)
	}
}

func RunGRPCServer(cfg *config.Config) {
	notificationQueue := queue.NewNotificationQueue(store, 5, 3)
	notificationQueue.Start()
	defer notificationQueue.Stop()

	notificationService := service.NewNotificationService(store, notificationQueue)
	postService := service.NewPostService(store, notificationQueue)

	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()

	notificationProto.RegisterNotificationServiceServer(grpcServer, notificationService)
	postProto.RegisterPostServiceServer(grpcServer, postService)
	reflection.Register(grpcServer)

	grpcListener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Error("failed to listen for gRPC", "error", err)
		return
	}

	logger.Info("starting gRPC server", "port", cfg.GRPCPort)

	if err := grpcServer.Serve(grpcListener); err != nil {
		logger.Error("failed to serve gRPC", "error", err)
	}
}
