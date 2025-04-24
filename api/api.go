package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/paper-social/notification-service/internal/models"
)

type ApiServer struct {
	engine *gin.Engine
	store  *models.Store
	port   string
}

func NewApiServer(store *models.Store, port string) *ApiServer {
	server := &ApiServer{
		engine: gin.Default(),
		store:  store,
		port:   port,
	}

	server.setupMiddleware()
	server.setupRoutes()
	return server
}

func (s *ApiServer) setupMiddleware() {
	s.engine.Use(gin.Recovery())
	s.engine.Use(gin.Logger())
}

func (s *ApiServer) setupRoutes() {
	api := s.engine.Group("/api")

	// Register metric routes
	s.RegisterMetricRoutes(api)
}

func (s *ApiServer) Start() error {
	addr := fmt.Sprintf(":%s", s.port)
	return s.engine.Run(addr)
}
