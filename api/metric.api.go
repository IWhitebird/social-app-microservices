package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterMetricRoutes registers all metric related routes
func (s *HttpApi) RegisterMetricRoutes(v1 *gin.RouterGroup) {
	metrics := v1.Group("/metrics")
	{
		metrics.GET("", s.GetMetrics)
	}
}

// GetMetrics godoc
// @Summary Get system metrics
// @Description Get detailed metrics about the system
// @Tags metrics
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /metrics [get]
func (s *HttpApi) GetMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"store_metrics": gin.H{
				"total_notifications_sent": s.store.Metrics.TotalNotificationsSent,
				"failed_attempts":          s.store.Metrics.FailedAttempts,
				"average_delivery_time":    fmt.Sprintf("%dms", int(s.store.Metrics.AverageDeliveryTime*1000)),
			},
			"system_status": "healthy",
		},
	})
}
