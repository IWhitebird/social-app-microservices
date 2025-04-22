package metrics

import (
	"encoding/json"
	"net/http"

	"github.com/paper-social/notification-service/internal/queue"
)

// MetricsResponse defines the structure for metrics API response
type MetricsResponse struct {
	QueueSize     int `json:"queue_size"`
	WorkersActive int `json:"workers_active"`
	TotalHandled  int `json:"total_handled"`
	TotalFailed   int `json:"total_failed"`
}

// Handler creates an HTTP handler for the metrics endpoint
func Handler(notificationQueue *queue.NotificationQueue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get metrics from the notification queue
		metrics := MetricsResponse{
			QueueSize:     notificationQueue.Size(),
			WorkersActive: notificationQueue.ActiveWorkers(),
			TotalHandled:  notificationQueue.TotalHandled(),
			TotalFailed:   notificationQueue.TotalFailed(),
		}

		// Return as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
	}
}
