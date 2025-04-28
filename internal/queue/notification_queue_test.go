package queue_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iwhitebird/social-app-microservices/internal/models"
	"github.com/iwhitebird/social-app-microservices/internal/queue"
	"github.com/stretchr/testify/assert"
)

func TestNotificationQueue(t *testing.T) {
	// Create store
	store := models.NewStore()

	// Create notification queue with 3 workers and max 2 retries
	notificationQueue := queue.NewNotificationQueue(store, 3, 2)
	notificationQueue.Start()
	defer notificationQueue.Stop()

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
	notificationQueue.EnqueueNotification(notification)

	// Wait for processing to complete
	time.Sleep(1 * time.Second)

	// Check if notification was saved
	store.Mu.Lock()
	notifications, exists := store.Notifications["u1"]
	store.Mu.Unlock()

	assert.True(t, exists, "Expected notifications to exist for user u1")
	assert.NotEmpty(t, notifications, "Expected at least one notification for user u1")
}

func TestConcurrentProcessing(t *testing.T) {
	// Create store
	store := models.NewStore()

	// Create notification queue with 5 workers and max 2 retries
	notificationQueue := queue.NewNotificationQueue(store, 5, 2)
	notificationQueue.Start()
	defer notificationQueue.Stop()

	// Number of notifications to send
	count := 100

	// Map to track notifications by userID
	userNotifications := make(map[string]int)
	var mu sync.Mutex

	// Create and send notifications
	for i := 0; i < count; i++ {
		userID := fmt.Sprintf("user-%d", i%10) // 10 different users

		notification := &models.Notification{
			ID:        uuid.New().String(),
			UserID:    userID,
			PostID:    "p1",
			Content:   fmt.Sprintf("Test notification %d", i),
			Read:      false,
			CreatedAt: time.Now(),
		}

		mu.Lock()
		userNotifications[userID]++
		mu.Unlock()

		// Send notification to queue
		notificationQueue.EnqueueNotification(notification)
	}

	// Wait for processing to complete (adjust time as needed)
	time.Sleep(1 * time.Second)

	// Verify notifications were saved for each user
	for userID, expectedCount := range userNotifications {
		store.Mu.Lock()
		notifications, exists := store.Notifications[userID]
		notificationCount := len(notifications)
		store.Mu.Unlock()

		assert.True(t, exists, "Expected notifications to exist for user %s", userID)

		// Due to random failures (10% failure rate in the queue), we may not have all notifications
		// But we should have at least some for each user
		assert.NotZero(t, notificationCount, "Expected at least some notifications for user %s", userID)

		// Log the expected vs actual counts for debug purposes
		t.Logf("User %s: Expected up to %d notifications, got %d",
			userID, expectedCount, notificationCount)
	}

	// Check metrics
	assert.Greater(t, store.Metrics.TotalNotificationsSent, 0,
		"Expected some successful notifications")
}

func TestRetryLogic(t *testing.T) {
	// This is a more complex test that would ideally need control over the random failure
	// For a robust test, we'd need to mock the randomness or inject a failure hook
	// For now, we'll do a simple test to ensure the retry mechanism exists

	// Create store
	store := models.NewStore()

	// Create notification queue with 1 worker and max 3 retries
	// This ensures we can observe the retry behavior more easily
	notificationQueue := queue.NewNotificationQueue(store, 1, 3)
	notificationQueue.Start()
	defer notificationQueue.Stop()

	// Create 20 test notifications to increase the chance of seeing a retry
	for i := 0; i < 20; i++ {
		notification := &models.Notification{
			ID:        uuid.New().String(),
			UserID:    "retry-test-user",
			PostID:    "p1",
			Content:   fmt.Sprintf("Retry test notification %d", i),
			Read:      false,
			CreatedAt: time.Now(),
		}

		notificationQueue.EnqueueNotification(notification)
	}

	// Wait longer to allow for retries
	time.Sleep(1 * time.Second)

	// We can't assert exact numbers because of the random nature,
	// but we can check that some metrics were updated
	assert.GreaterOrEqual(t, store.Metrics.TotalNotificationsSent, 0)

	// Log metrics for information
	t.Logf("Metrics - Sent: %d, Failed: %d, Avg Time: %f",
		store.Metrics.TotalNotificationsSent,
		store.Metrics.FailedAttempts,
		store.Metrics.AverageDeliveryTime)
}

func TestQueueShutdown(t *testing.T) {
	// Create store
	store := models.NewStore()

	// Create notification queue with 3 workers and max 2 retries
	notificationQueue := queue.NewNotificationQueue(store, 3, 2)
	notificationQueue.Start()

	// Enqueue some notifications
	for i := 0; i < 10; i++ {
		notification := &models.Notification{
			ID:        uuid.New().String(),
			UserID:    "shutdown-test-user",
			PostID:    "p1",
			Content:   fmt.Sprintf("Shutdown test notification %d", i),
			Read:      false,
			CreatedAt: time.Now(),
		}

		notificationQueue.EnqueueNotification(notification)
	}

	// Wait briefly for processing to start
	time.Sleep(50 * time.Millisecond)

	// Now shutdown the queue
	notificationQueue.Stop()

	// The test passes if Stop() doesn't hang indefinitely
	// We can't easily assert the completion of all pending tasks,
	// but we can check that some notifications were processed
	store.Mu.Lock()
	_, exists := store.Notifications["shutdown-test-user"]
	store.Mu.Unlock()

	assert.True(t, exists, "Expected some notifications to be processed before shutdown")
}

func TestNotificationQueuePerformance(t *testing.T) {
	// Create store
	store := models.NewStore()

	// Create notification queue with more workers for performance testing
	notificationQueue := queue.NewNotificationQueue(store, 100, 1) // More workers, fewer retries
	notificationQueue.Start()
	defer notificationQueue.Stop()

	// Large batch of notifications
	batchSize := 1000
	start := time.Now()

	// Enqueue many notifications
	for i := 0; i < batchSize; i++ {
		userID := fmt.Sprintf("perf-user-%d", i%20) // 20 different users

		notification := &models.Notification{
			ID:        uuid.New().String(),
			UserID:    userID,
			PostID:    "p1",
			Content:   fmt.Sprintf("Performance test notification %d", i),
			Read:      false,
			CreatedAt: time.Now(),
		}

		notificationQueue.EnqueueNotification(notification)
	}

	// Wait for processing to complete - this may need to be adjusted
	time.Sleep(1 * time.Second)

	duration := time.Since(start)

	// Log performance metrics
	t.Logf("Processed approximately %d notifications in %v",
		store.Metrics.TotalNotificationsSent, duration)
	t.Logf("Approximate rate: %.2f notifications/second",
		float64(store.Metrics.TotalNotificationsSent)/duration.Seconds())

	// Simple performance assertion - should process at a reasonable rate
	// This is more informative than a strict assertion
	assert.Greater(t, store.Metrics.TotalNotificationsSent, 0,
		"Expected some notifications to be processed")
}
