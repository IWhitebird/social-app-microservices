package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *HttpApi) RegisterMetricRoutes(v1 *gin.RouterGroup) {
	metrics := v1.Group("/metrics")
	{
		metrics.GET("", s.GetMetrics)
	}
}

func (s *HttpApi) GetMetrics(c *gin.Context) {
	notificationMetrics, err := s.notificationClient.GetNotificationMetrics(c, &emptypb.Empty{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get notification metrics",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"store_metrics": gin.H{
				"total_notifications_sent": notificationMetrics.TotalNotificationsSent,
				"failed_attempts":          notificationMetrics.FailedAttempts,
				"average_delivery_time":    notificationMetrics.AverageDeliveryTime,
			},
			"system_status": "healthy",
		},
	})
}
