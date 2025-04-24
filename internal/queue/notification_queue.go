package queue

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/paper-social/notification-service/internal/models"
)

// NotificationJob represents a job to send a notification
type NotificationJob struct {
	Notification *models.Notification
	Attempt      int
}

// NotificationQueue handles the queue of notifications to be sent
type NotificationQueue struct {
	jobs         chan NotificationJob
	store        *models.Store
	workerCount  int
	maxRetries   int
	wg           sync.WaitGroup
	metrics      *Metrics
	shutdownChan chan struct{}
}

// Metrics tracks notification delivery statistics
type Metrics struct {
	mu                 sync.Mutex
	totalSent          int
	failedAttempts     int
	totalDeliveryTime  time.Duration
	notificationsCount int
	activeWorkers      int
}

// GetMetrics returns the current metrics
func (m *Metrics) GetMetrics() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	avgDeliveryTime := float64(0)
	if m.notificationsCount > 0 {
		avgDeliveryTime = float64(m.totalDeliveryTime.Milliseconds()) / float64(m.notificationsCount)
	}

	return map[string]interface{}{
		"total_notifications_sent": m.totalSent,
		"failed_attempts":          m.failedAttempts,
		"average_delivery_time_ms": avgDeliveryTime,
	}
}

// NewNotificationQueue creates a new notification queue
func NewNotificationQueue(store *models.Store, workerCount, maxRetries int) *NotificationQueue {
	return &NotificationQueue{
		jobs:         make(chan NotificationJob, 1000), // Buffer size of 1000
		store:        store,
		workerCount:  workerCount,
		maxRetries:   maxRetries,
		metrics:      &Metrics{},
		shutdownChan: make(chan struct{}),
	}
}

// Start starts the notification queue workers
func (q *NotificationQueue) Start() {
	log.Printf("Starting %d notification workers", q.workerCount)
	for i := 0; i < q.workerCount; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
}

// Stop stops the notification queue
func (q *NotificationQueue) Stop() {
	close(q.shutdownChan)
	q.wg.Wait()
	log.Println("Notification queue stopped")
}

// EnqueueNotification adds a notification to the queue
func (q *NotificationQueue) EnqueueNotification(notification *models.Notification) {
	q.jobs <- NotificationJob{
		Notification: notification,
		Attempt:      1,
	}
}

// worker processes notifications from the queue
func (q *NotificationQueue) worker(id int) {
	defer q.wg.Done()
	log.Printf("Worker %d started", id)

	for {
		select {
		case job := <-q.jobs:
			q.metrics.mu.Lock()
			q.metrics.activeWorkers++
			q.metrics.mu.Unlock()

			startTime := time.Now()
			success := q.processNotification(job)
			duration := time.Since(startTime)

			q.metrics.mu.Lock()
			q.metrics.activeWorkers--
			if success {
				q.metrics.totalSent++
				q.metrics.totalDeliveryTime += duration
				q.metrics.notificationsCount++
			} else {
				q.metrics.failedAttempts++
			}
			q.metrics.mu.Unlock()

		case <-q.shutdownChan:
			log.Printf("Worker %d shutting down", id)
			return
		}
	}
}

// processNotification attempts to send a notification with retry logic
func (q *NotificationQueue) processNotification(job NotificationJob) bool {
	notification := job.Notification
	attempt := job.Attempt

	// Simulate delivery with 10% failure rate
	if rand.Float64() < 0.1 {
		log.Printf("Failed to send notification to user %s for post %s (attempt %d)",
			notification.UserID, notification.PostID, attempt)

		if attempt < q.maxRetries {
			// Exponential backoff: 1s, 2s, 4s...
			backoff := time.Duration(1<<(attempt-1)) * time.Second
			log.Printf("Retrying in %v...", backoff)

			time.Sleep(backoff)

			q.jobs <- NotificationJob{
				Notification: notification,
				Attempt:      attempt + 1,
			}
		} else {
			log.Printf("Max retries exceeded for notification to user %s for post %s",
				notification.UserID, notification.PostID)
		}
		return false
	}

	// Simulation of successful delivery
	log.Printf("Notification sent to user %s for post %s",
		notification.UserID, notification.PostID)

	// Store notification in user's list
	q.metrics.mu.Lock()
	userID := notification.UserID
	if _, exists := q.store.Notifications[userID]; !exists {
		q.store.Notifications[userID] = []*models.Notification{}
	}
	q.store.Notifications[userID] = append(q.store.Notifications[userID], notification)
	q.metrics.mu.Unlock()

	return true
}

// GetMetrics returns the current queue metrics
func (q *NotificationQueue) GetMetrics() map[string]interface{} {
	return q.metrics.GetMetrics()
}

// Size returns the current number of jobs in the queue
func (q *NotificationQueue) Size() int {
	return len(q.jobs)
}

// ActiveWorkers returns the number of active workers
func (q *NotificationQueue) ActiveWorkers() int {
	q.metrics.mu.Lock()
	defer q.metrics.mu.Unlock()
	return q.metrics.activeWorkers
}

// TotalHandled returns the total number of notifications processed
func (q *NotificationQueue) TotalHandled() int {
	q.metrics.mu.Lock()
	defer q.metrics.mu.Unlock()
	return q.metrics.totalSent
}

// TotalFailed returns the total number of failed notification attempts
func (q *NotificationQueue) TotalFailed() int {
	q.metrics.mu.Lock()
	defer q.metrics.mu.Unlock()
	return q.metrics.failedAttempts
}
