package service

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/iwhitebird/social-app-microservices/internal/models"
	"github.com/iwhitebird/social-app-microservices/internal/queue"
	notificationProto "github.com/iwhitebird/social-app-microservices/proto/generated/notification/proto"
	"google.golang.org/protobuf/types/known/emptypb"
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

	// Get the 20 most recent notifications
	startIdx := 0
	if len(userNotifications) > 20 {
		startIdx = len(userNotifications) - 20
	}
	recentNotifications := userNotifications[startIdx:]

	// Send each notification to the client
	for _, notification := range recentNotifications {
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

func (s *NotificationService) GetNotificationMetrics(ctx context.Context, in *emptypb.Empty) (*notificationProto.NotificationMetrics, error) {
	notificationMetrics := &notificationProto.NotificationMetrics{
		TotalNotificationsSent: int64(s.store.Metrics.TotalNotificationsSent),
		FailedAttempts:         int64(s.store.Metrics.FailedAttempts),
		AverageDeliveryTime:    float64(s.store.Metrics.AverageDeliveryTime),
	}
	return notificationMetrics, nil
}
