package queue

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/iwhitebird/social-app-microservices/internal/models"
)

type NotificationJob struct {
	Notification *models.Notification
	Attempt      int
}

type NotificationQueue struct {
	jobs         chan NotificationJob
	store        *models.Store
	workerCount  int
	maxRetries   int
	shutdownChan chan struct{}
	wg           sync.WaitGroup
	mu           sync.Mutex
}

func NewNotificationQueue(store *models.Store, workerCount, maxRetries int) *NotificationQueue {
	return &NotificationQueue{
		jobs:         make(chan NotificationJob, 1000), // Buffer size of 1000
		store:        store,
		workerCount:  workerCount,
		maxRetries:   maxRetries,
		shutdownChan: make(chan struct{}),
		mu:           sync.Mutex{},
	}
}

func (q *NotificationQueue) Start() {
	log.Printf("Starting %d notification workers", q.workerCount)
	for i := range make([]struct{}, q.workerCount) {
		q.wg.Add(1)
		go q.worker(i)
	}
}

func (q *NotificationQueue) Stop() {
	close(q.shutdownChan)
	q.wg.Wait()
	log.Println("Notification queue stopped")
}

func (q *NotificationQueue) EnqueueNotification(notification *models.Notification) {
	q.jobs <- NotificationJob{
		Notification: notification,
		Attempt:      1,
	}
}

func (q *NotificationQueue) worker(id int) {
	defer q.wg.Done()
	log.Printf("Worker %d started", id)

	for {
		select {
		case job := <-q.jobs:

			startTime := time.Now()
			success := q.processNotification(job)

			timeTakenToDeliver := time.Since(startTime)

			q.store.Mu.Lock()
			if success {
				if q.store.Metrics.TotalNotificationsSent == 0 {
					// First successful notification
					q.store.Metrics.AverageDeliveryTime = float64(timeTakenToDeliver)
				} else {
					// Update running average
					q.store.Metrics.AverageDeliveryTime =
						(q.store.Metrics.AverageDeliveryTime*float64(q.store.Metrics.TotalNotificationsSent) +
							float64(timeTakenToDeliver)) / float64(q.store.Metrics.TotalNotificationsSent+1)
				}
				q.store.Metrics.TotalNotificationsSent++
			} else {
				q.store.Metrics.FailedAttempts++
			}
			q.store.Mu.Unlock()

		case <-q.shutdownChan:
			log.Printf("Worker %d shutting down", id)
			return
		}
	}
}

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

	q.store.Mu.Lock()
	userID := notification.UserID
	if _, exists := q.store.Notifications[userID]; !exists {
		q.store.Notifications[userID] = []*models.Notification{}
	}
	q.store.Notifications[userID] = append(q.store.Notifications[userID], notification)
	q.store.Mu.Unlock()

	return true
}
