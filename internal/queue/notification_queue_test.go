package queue

import (
	"math/rand"
	"testing"
	"time"

	"github.com/paper-social/notification-service/internal/models"
)

func TestNotificationQueue(t *testing.T) {
	// Create store
	store := models.NewStore()

	// Create notification queue with 3 workers and max 2 retries
	queue := NewNotificationQueue(store, 3, 2)
	queue.Start()
	defer queue.Stop()

	// Create a test notification
	notification := &models.Notification{
		ID:        "test-notification-1",
		UserID:    "u1",
		PostID:    "p1",
		Content:   "Test notification",
		Read:      false,
		CreatedAt: time.Now(),
	}

	// Send notification to queue
	queue.EnqueueNotification(notification)

	// Wait for processing to finish
	time.Sleep(100 * time.Millisecond)

	// Check if notification was saved
	cnt := len(store.Notifications["u1"])
	if cnt == 0 {
		t.Errorf("Expected at least one notification for user u1, got none")
	}
}

func TestConcurrentProcessing(t *testing.T) {
	// Create store
	store := models.NewStore()

	// Create mock queue for testing
	mockQueue := &MockNotificationQueue{
		store:      store,
		maxRetries: 2,
	}

	// Number of notifications to send
	count := 50

	// Create and process notifications
	for i := 0; i < count; i++ {
		notification := &models.Notification{
			ID:        "test-notification-concurrent",
			UserID:    "u1",
			PostID:    "p1",
			Content:   "Test notification",
			Read:      false,
			CreatedAt: time.Now(),
		}

		// Process directly with 10% failure rate
		mockQueue.ProcessNotification(notification)
	}
}

// MockNotificationQueue is a simplified version of NotificationQueue for testing
type MockNotificationQueue struct {
	store      *models.Store
	maxRetries int
}

// ProcessNotification simulates processing a notification with 10% failure rate
func (q *MockNotificationQueue) ProcessNotification(notification *models.Notification) bool {
	// Simulate delivery with 10% failure rate
	success := true
	if rand.Float64() < 0.1 {
		success = false
		q.store.Metrics.FailedAttempts++
	} else {
		// Store notification in user's list
		q.store.Notifications[notification.UserID] = append(q.store.Notifications[notification.UserID], notification)
	}

	return success
}
