package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterMetricRoutes registers all metric related routes
func (s *ApiServer) RegisterMetricRoutes(v1 *gin.RouterGroup) {
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
func (s *ApiServer) GetMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"store_metrics": s.store.Metrics,
			"system_status": "healthy",
		},
	})
}
