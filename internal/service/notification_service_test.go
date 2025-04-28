package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/iwhitebird/social-app-microservices/internal/models"
	"github.com/iwhitebird/social-app-microservices/internal/queue"
	"github.com/iwhitebird/social-app-microservices/internal/service"
	notificationProto "github.com/iwhitebird/social-app-microservices/proto/generated/notification/proto"
	postProto "github.com/iwhitebird/social-app-microservices/proto/generated/post/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MockNotificationStream implements the NotificationService_GetNotificationsServer interface
type MockNotificationStream struct {
	Ctx          context.Context
	ReceivedMsgs []*notificationProto.Notification
}

func (s *MockNotificationStream) Send(notification *notificationProto.Notification) error {
	s.ReceivedMsgs = append(s.ReceivedMsgs, notification)
	return nil
}

func (s *MockNotificationStream) SetHeader(md metadata.MD) error {
	return nil
}

func (s *MockNotificationStream) SendHeader(md metadata.MD) error {
	return nil
}

func (s *MockNotificationStream) SetTrailer(md metadata.MD) {}

func (s *MockNotificationStream) Context() context.Context {
	return s.Ctx
}

func (s *MockNotificationStream) SendMsg(m interface{}) error {
	notification, ok := m.(*notificationProto.Notification)
	if !ok {
		return nil
	}
	return s.Send(notification)
}

func (s *MockNotificationStream) RecvMsg(m interface{}) error {
	return nil
}

func TestGetNotifications(t *testing.T) {
	// Create real store
	store := models.NewStore()

	// Create real queue
	notificationQueue := queue.NewNotificationQueue(store, 3, 2)
	notificationQueue.Start()
	defer notificationQueue.Stop()

	// Create notification service
	notificationService := service.NewNotificationService(store, notificationQueue)

	// Initialize test data
	initTestData(store)

	// Test cases
	tests := []struct {
		name          string
		userID        string
		expectedCount int
		shouldBeEmpty bool
	}{
		{
			name:          "existing user with notifications",
			userID:        "test-user-1",
			expectedCount: 1,
			shouldBeEmpty: false,
		},
		{
			name:          "existing user with notifications",
			userID:        "test-user-2",
			expectedCount: 1,
			shouldBeEmpty: false,
		},
		{
			name:          "user with no notifications",
			userID:        "test-user-3",
			expectedCount: 0,
			shouldBeEmpty: true,
		},
		{
			name:          "non-existing user",
			userID:        "nonexistent",
			expectedCount: 0,
			shouldBeEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create user ID proto message
			userID := &notificationProto.UserId{
				UserId: tt.userID,
			}

			// Create mock stream
			mockStream := &MockNotificationStream{
				Ctx: context.Background(),
			}

			// Call GetNotifications
			err := notificationService.GetNotifications(userID, mockStream)
			assert.NoError(t, err)

			// Check results
			receivedNotifications := mockStream.ReceivedMsgs

			if tt.shouldBeEmpty {
				assert.Empty(t, receivedNotifications)
			} else {
				assert.Len(t, receivedNotifications, tt.expectedCount)

				if tt.expectedCount > 0 {
					// Verify notification fields
					for _, notification := range receivedNotifications {
						assert.Equal(t, tt.userID, notification.UserId)
						assert.NotEmpty(t, notification.Id)
						assert.NotEmpty(t, notification.Content)
						assert.NotZero(t, notification.CreatedAt)
					}
				}
			}
		})
	}
}

func TestGetNotificationMetrics(t *testing.T) {
	// Create real store
	store := models.NewStore()

	// Create real queue
	notificationQueue := queue.NewNotificationQueue(store, 3, 2)
	notificationQueue.Start()
	defer notificationQueue.Stop()

	// Create notification service
	notificationService := service.NewNotificationService(store, notificationQueue)

	// Initialize test data with metrics
	initTestData(store)
	store.Metrics.TotalNotificationsSent = 10
	store.Metrics.FailedAttempts = 2
	store.Metrics.AverageDeliveryTime = 150.5

	// Call GetNotificationMetrics
	metrics, err := notificationService.GetNotificationMetrics(context.Background(), &emptypb.Empty{})

	// Assert no error
	assert.NoError(t, err)

	// Assert metrics values match the data
	assert.Equal(t, int64(10), metrics.TotalNotificationsSent)
	assert.Equal(t, int64(2), metrics.FailedAttempts)
	assert.Equal(t, 150.5, metrics.AverageDeliveryTime)
}

func TestGetNotificationsStreamError(t *testing.T) {
	// Create real store
	store := models.NewStore()

	// Create real queue
	notificationQueue := queue.NewNotificationQueue(store, 3, 2)
	notificationQueue.Start()
	defer notificationQueue.Stop()

	// Create notification service
	notificationService := service.NewNotificationService(store, notificationQueue)

	// Add test data
	initTestData(store)

	// Create a user ID proto message
	userID := &notificationProto.UserId{
		UserId: "test-user-1",
	}

	// Create mock stream
	mockStream := &MockNotificationStream{
		Ctx: context.Background(),
	}

	// Test the service behavior
	err := notificationService.GetNotifications(userID, mockStream)

	// Since there are notifications in the test data for test-user-1,
	// they should be sent to the stream without errors
	assert.NoError(t, err)
}

func TestNotificationServiceIntegration(t *testing.T) {
	// Create real store
	store := models.NewStore()

	// Create real queue
	notificationQueue := queue.NewNotificationQueue(store, 3, 2)
	notificationQueue.Start()
	defer notificationQueue.Stop()

	// Create services
	notificationService := service.NewNotificationService(store, notificationQueue)
	postService := service.NewPostService(store, notificationQueue)

	// Add some test users with followers
	store.Users["user1"] = &models.User{
		ID:        "user1",
		Username:  "user1",
		Followers: []string{"user2", "user3"},
	}

	store.Users["user2"] = &models.User{
		ID:        "user2",
		Username:  "user2",
		Followers: []string{},
	}

	store.Users["user3"] = &models.User{
		ID:        "user3",
		Username:  "user3",
		Followers: []string{},
	}

	// Create a test post through the post service
	ctx := context.Background()
	postResp, err := postService.PublishPost(ctx, &postProto.Post{
		UserId:  "user1",
		Content: "Test post content",
	})

	// Assert post creation was successful
	assert.NoError(t, err)
	assert.True(t, postResp.Success)
	assert.Equal(t, int32(2), postResp.NotificationsQueued)

	// Allow time for notifications to be processed
	time.Sleep(100 * time.Millisecond)

	// Now get notifications for one of the followers
	userID := &notificationProto.UserId{
		UserId: "user2",
	}

	// Create mock stream
	mockStream := &MockNotificationStream{
		Ctx: context.Background(),
	}

	// Get notifications
	err = notificationService.GetNotifications(userID, mockStream)
	assert.NoError(t, err)

	// Check if the notification was received
	assert.Len(t, mockStream.ReceivedMsgs, 1, "Expected one notification for user2")

	// Verify notification content if it exists
	if len(mockStream.ReceivedMsgs) > 0 {
		notification := mockStream.ReceivedMsgs[0]
		assert.Equal(t, "user2", notification.UserId)
		assert.Contains(t, notification.Content, "Test post content")
	}
}

// Helper function to initialize test data
func initTestData(store *models.Store) {
	// Test users
	users := []*models.User{
		{ID: "test-user-1", Username: "testuser1", Followers: []string{"test-user-2", "test-user-3"}},
		{ID: "test-user-2", Username: "testuser2", Followers: []string{"test-user-1", "test-user-3"}},
		{ID: "test-user-3", Username: "testuser3", Followers: []string{"test-user-1", "test-user-2"}},
	}

	for _, u := range users {
		store.Users[u.ID] = u
	}

	// Test posts
	posts := []*models.Post{
		{ID: "test-post-1", UserID: "test-user-1", Content: "Test post 1"},
		{ID: "test-post-2", UserID: "test-user-2", Content: "Test post 2"},
	}

	for _, p := range posts {
		store.Posts[p.UserID] = p
	}

	// Test notifications
	now := time.Now()
	notifications := []*models.Notification{
		{
			ID:        "test-notification-1",
			UserID:    "test-user-1",
			PostID:    "test-post-2",
			Content:   "Test notification 1",
			Read:      false,
			CreatedAt: now,
			Status:    models.NotificationStatusDelivered,
		},
		{
			ID:        "test-notification-2",
			UserID:    "test-user-2",
			PostID:    "test-post-1",
			Content:   "Test notification 2",
			Read:      true,
			CreatedAt: now.Add(-1 * time.Hour),
			Status:    models.NotificationStatusDelivered,
		},
	}

	for _, n := range notifications {
		userID := n.UserID
		if _, exists := store.Notifications[userID]; !exists {
			store.Notifications[userID] = []*models.Notification{}
		}
		store.Notifications[userID] = append(store.Notifications[userID], n)
	}
}
