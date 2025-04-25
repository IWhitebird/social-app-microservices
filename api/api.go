package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	notificationProto "github.com/iwhitebird/social-app-microservices/proto/generated/notification/proto"
	postProto "github.com/iwhitebird/social-app-microservices/proto/generated/post/proto"
)

type HttpApi struct {
	engine             *gin.Engine
	port               string
	notificationClient notificationProto.NotificationServiceClient
	postClient         postProto.PostServiceClient
}

func NewHttpApi(notificationClient notificationProto.NotificationServiceClient, postClient postProto.PostServiceClient, port string) *HttpApi {
	gin.SetMode(gin.ReleaseMode)
	server := &HttpApi{
		engine:             gin.Default(),
		port:               port,
		notificationClient: notificationClient,
		postClient:         postClient,
	}
	server.engine.Use(gin.Recovery())
	server.engine.Use(gin.Logger())
	server.setupRoutes()

	return server
}

func (s *HttpApi) setupRoutes() {
	api := s.engine.Group("/api")
	s.RegisterMetricRoutes(api)
}

func (s *HttpApi) Start() error {
	addr := fmt.Sprintf(":%s", s.port)
	return s.engine.Run(addr)
}
