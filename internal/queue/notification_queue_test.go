package queue

// import (
// 	"math/rand"
// 	"testing"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/paper-social/notification-service/internal/models"
// )

// func TestNotificationQueue(t *testing.T) {
// 	// Create store
// 	store := models.NewStore()

// 	// Create notification queue with 3 workers and max 2 retries
// 	queue := NewNotificationQueue(store, 3, 2)
// 	queue.Start()
// 	defer queue.Stop()

// 	// Create a test notification
// 	notification := &models.Notification{
// 		ID:           "test-notification-1",
// 		UserID:       "u1",
// 		PostID:       "p1",
// 		PostAuthorID: "u2",
// 		Content:      "Test notification",
// 		Read:         false,
// 		CreatedAt:    time.Now(),
// 	}

// 	// Send notification to queue
// 	queue.EnqueueNotification(notification)

// 	// Wait for processing to finish
// 	time.Sleep(100 * time.Millisecond)

// 	// Check if notification was saved
// 	cnt := 0
// 	for _, notification := range store.Notifications {
// 		if notification.UserID == "u1" {
// 			cnt++
// 			break
// 		}
// 	}
// 	if cnt == 0 {
// 		t.Errorf("Expected at least one notification for user u1, got none")
// 	}
// }

// func TestConcurrentProcessing(t *testing.T) {
// 	// Create store
// 	store := models.NewStore()

// 	// Create mock queue for testing
// 	mockQueue := &MockNotificationQueue{
// 		store:      store,
// 		maxRetries: 2,
// 		metrics:    &Metrics{},
// 	}

// 	// Number of notifications to send
// 	count := 50

// 	// Create and process notifications
// 	for i := 0; i < count; i++ {
// 		notification := &models.Notification{
// 			ID:           "test-notification-concurrent",
// 			UserID:       "u1",
// 			PostID:       "p1",
// 			PostAuthorID: "u2",
// 			Content:      "Test notification",
// 			Read:         false,
// 			CreatedAt:    time.Now(),
// 		}

// 		// Process directly with 10% failure rate
// 		mockQueue.ProcessNotification(notification)
// 	}

// 	// Check metrics
// 	metrics := mockQueue.GetMetrics()
// 	totalSent := metrics["total_notifications_sent"].(int)
// 	failedAttempts := metrics["failed_attempts"].(int)

// 	// We expect roughly 90% success rate (10% failure)
// 	expectedSuccessMin := int(float64(count) * 0.7) // Allow wider margin for randomness

// 	if totalSent < expectedSuccessMin {
// 		t.Errorf("Expected at least %d successful deliveries, got %d",
// 			expectedSuccessMin, totalSent)
// 	}

// 	if totalSent+failedAttempts != count {
// 		t.Errorf("Expected metrics to add up to %d, got %d (sent) + %d (failed) = %d",
// 			count, totalSent, failedAttempts, totalSent+failedAttempts)
// 	}
// }

// // MockNotificationQueue is a simplified version of NotificationQueue for testing
// type MockNotificationQueue struct {
// 	store      *models.Store
// 	maxRetries int
// 	metrics    *Metrics
// }

// // ProcessNotification simulates processing a notification with 10% failure rate
// func (q *MockNotificationQueue) ProcessNotification(notification *models.Notification) bool {
// 	// Simulate delivery with 10% failure rate
// 	success := true
// 	if rand.Float64() < 0.1 {
// 		success = false
// 		q.metrics.mu.Lock()
// 		q.metrics.failedAttempts++
// 		q.metrics.mu.Unlock()
// 	} else {
// 		// Store notification in user's list
// 		q.metrics.mu.Lock()
// 		q.metrics.totalSent++
// 		q.store.Notifications[uuid.New().String()] = notification
// 		q.metrics.mu.Unlock()
// 	}

// 	return success
// }

// // GetMetrics returns the current queue metrics
// func (q *MockNotificationQueue) GetMetrics() map[string]interface{} {
// 	q.metrics.mu.Lock()
// 	defer q.metrics.mu.Unlock()

// 	return map[string]interface{}{
// 		"total_notifications_sent": q.metrics.totalSent,
// 		"failed_attempts":          q.metrics.failedAttempts,
// 	}
// }
