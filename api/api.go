package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/paper-social/notification-service/internal/models"
)

type HttpApi struct {
	engine *gin.Engine
	store  *models.Store
	port   string
}

func NewHttpApi(store *models.Store, port string) *HttpApi {
	gin.SetMode(gin.ReleaseMode)
	server := &HttpApi{
		engine: gin.Default(),
		store:  store,
		port:   port,
	}
	server.setupMiddleware()
	server.setupRoutes()
	return server
}

func (s *HttpApi) setupMiddleware() {
	s.engine.Use(gin.Recovery())
	s.engine.Use(gin.Logger())
}

func (s *HttpApi) setupRoutes() {
	api := s.engine.Group("/api")

	// Register metric routes
	s.RegisterMetricRoutes(api)
}

func (s *HttpApi) Start() error {
	addr := fmt.Sprintf(":%s", s.port)
	return s.engine.Run(addr)
}
