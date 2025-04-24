package service

import (
	"fmt"
	"log"
	"sync"

	"github.com/paper-social/notification-service/internal/models"
	"github.com/paper-social/notification-service/internal/queue"
	notificationProto "github.com/paper-social/notification-service/proto/generated/notification/proto"
)

// NotificationService implements the gRPC notification service
type NotificationService struct {
	notificationProto.UnimplementedNotificationServiceServer
	store *models.Store
	queue *queue.NotificationQueue
	mu    sync.Mutex
}

// NewNotificationService creates a new NotificationService
func NewNotificationService(store *models.Store, queue *queue.NotificationQueue) *NotificationService {
	return &NotificationService{
		store: store,
		queue: queue,
	}
}

// GetNotifications streams notifications for a user
func (s *NotificationService) GetNotifications(userId *notificationProto.UserId, stream notificationProto.NotificationService_GetNotificationsServer) error {
	log.Printf("Received GetNotifications request for user %s", userId.UserId)

	userID := userId.UserId

	// Get notifications for the user
	s.mu.Lock()
	userNotifications, exists := s.store.Notifications[userID]
	s.mu.Unlock()

	if !exists {
		return nil
	}

	// Send each notification to the client
	for _, notification := range userNotifications {
		// Convert internal notification to proto notification
		protoNotification := &notificationProto.Notification{
			Id:        notification.ID,
			UserId:    notification.UserID,
			PostId:    notification.PostID,
			Content:   notification.Content,
			Read:      notification.Read,
			CreatedAt: notification.CreatedAt.Unix(),
		}
		fmt.Println(protoNotification)
		// Send the notification
		if err := stream.Send(protoNotification); err != nil {
			log.Printf("Error sending notification: %v", err)
			return err
		}
	}

	return nil
}
