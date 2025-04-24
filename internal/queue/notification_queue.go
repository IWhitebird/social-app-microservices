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
	shutdownChan chan struct{}
}

// NewNotificationQueue creates a new notification queue
func NewNotificationQueue(store *models.Store, workerCount, maxRetries int) *NotificationQueue {
	return &NotificationQueue{
		jobs:         make(chan NotificationJob, 1000), // Buffer size of 1000
		store:        store,
		workerCount:  workerCount,
		maxRetries:   maxRetries,
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

			startTime := time.Now()
			success := q.processNotification(job)

			timeTakenToDeliver := time.Since(startTime)

			if success {
				q.store.Metrics.AverageDeliveryTime = (q.store.Metrics.AverageDeliveryTime*float64(q.store.Metrics.TotalNotificationsSent-1) + float64(timeTakenToDeliver)) / float64(q.store.Metrics.TotalNotificationsSent)
				q.store.Metrics.TotalNotificationsSent++
			} else {
				q.store.Metrics.FailedAttempts++
			}

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
	userID := notification.UserID
	if _, exists := q.store.Notifications[userID]; !exists {
		q.store.Notifications[userID] = []*models.Notification{}
	}
	q.store.Notifications[userID] = append(q.store.Notifications[userID], notification)

	return true
}
