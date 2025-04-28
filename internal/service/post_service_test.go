package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/iwhitebird/social-app-microservices/internal/models"
	"github.com/iwhitebird/social-app-microservices/internal/queue"
	"github.com/iwhitebird/social-app-microservices/internal/service"
	postProto "github.com/iwhitebird/social-app-microservices/proto/generated/post/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

// MockPostStream implements the grpc.ServerStream interface for testing
type MockPostStream struct {
	Ctx          context.Context
	ReceivedMsgs []*postProto.Post
}

func (s *MockPostStream) SetHeader(md metadata.MD) error {
	return nil
}

func (s *MockPostStream) SendHeader(md metadata.MD) error {
	return nil
}

func (s *MockPostStream) SetTrailer(md metadata.MD) {}

func (s *MockPostStream) Context() context.Context {
	return s.Ctx
}

func (s *MockPostStream) SendMsg(m interface{}) error {
	if msg, ok := m.(*postProto.Post); ok {
		s.ReceivedMsgs = append(s.ReceivedMsgs, msg)
	}
	return nil
}

func (s *MockPostStream) RecvMsg(m interface{}) error {
	return nil
}

// NewMockPostStream creates a new MockPostStream instance
func NewMockPostStream() *MockPostStream {
	return &MockPostStream{
		Ctx:          context.Background(),
		ReceivedMsgs: make([]*postProto.Post, 0),
	}
}

// GetReceivedMsgs returns all messages received by the stream
func (s *MockPostStream) GetReceivedMsgs() []*postProto.Post {
	return s.ReceivedMsgs
}

func TestPublishPost(t *testing.T) {
	// Create real store
	store := models.NewStore()

	// Create real queue
	notificationQueue := queue.NewNotificationQueue(store, 3, 2)
	notificationQueue.Start()
	defer notificationQueue.Stop()

	// Create post service
	postService := service.NewPostService(store, notificationQueue)

	// Add test users with followers
	store.Users["user1"] = &models.User{
		ID:        "user1",
		Username:  "user1",
		Followers: []string{"follower1", "follower2", "follower3"},
	}

	// Test cases
	tests := []struct {
		name                  string
		userID                string
		content               string
		expectedSuccess       bool
		expectedNotifications int32
	}{
		{
			name:                  "user with followers",
			userID:                "user1",
			content:               "Test post content",
			expectedSuccess:       true,
			expectedNotifications: 3,
		},
		{
			name:                  "user without followers",
			userID:                "user2",
			content:               "Post with no notifications",
			expectedSuccess:       true,
			expectedNotifications: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create post request
			post := &postProto.Post{
				UserId:  tt.userID,
				Content: tt.content,
			}

			// Call PublishPost
			resp, err := postService.PublishPost(context.Background(), post)

			// Assert no error
			assert.NoError(t, err)

			// Assert response fields
			assert.Equal(t, tt.expectedSuccess, resp.Success)
			assert.Equal(t, tt.expectedNotifications, resp.NotificationsQueued)

			// Check if post was stored
			storedPost, exists := store.Posts[tt.userID]
			assert.True(t, exists)
			assert.Equal(t, tt.content, storedPost.Content)
			assert.Equal(t, tt.userID, storedPost.UserID)

			// Check if notifications were enqueued
			if tt.expectedNotifications > 0 {
				// Allow some time for notifications to be processed
				// because we're using the real queue now
				for attempt := 0; attempt < 5; attempt++ {
					store.Mu.Lock()
					allNotificationsDelivered := true
					for _, followerID := range store.Users[tt.userID].Followers {
						notifications, exists := store.Notifications[followerID]
						if !exists || len(notifications) == 0 {
							allNotificationsDelivered = false
							break
						}
					}
					store.Mu.Unlock()

					if allNotificationsDelivered {
						break
					}

					// Wait a bit for notifications to be processed
					// This is needed because the queue processes asynchronously
					t.Logf("Waiting for notifications to be processed (attempt %d)", attempt+1)
					// Sleep for a short duration to allow queue processing
					time.Sleep(1 * time.Second)
				}

				for _, followerID := range store.Users[tt.userID].Followers {
					// Check if notifications were created for followers
					store.Mu.Lock()
					notifications, exists := store.Notifications[followerID]
					store.Mu.Unlock()

					assert.True(t, exists, "Notifications should exist for follower %s", followerID)
					assert.NotEmpty(t, notifications, "At least one notification should be created for follower %s", followerID)

					// Verify notification content for at least one notification
					found := false
					for _, notification := range notifications {
						if notification.UserID == followerID && notification.Content != "" {
							found = true
							assert.Contains(t, notification.Content, tt.content)
							break
						}
					}
					assert.True(t, found, "Notification with correct content not found for follower %s", followerID)
				}
			}
		})
	}
}

func TestPublishPostWithEmptyContent(t *testing.T) {
	// Create real store
	store := models.NewStore()

	// Create real queue
	notificationQueue := queue.NewNotificationQueue(store, 3, 2)
	notificationQueue.Start()
	defer notificationQueue.Stop()

	// Create post service
	postService := service.NewPostService(store, notificationQueue)

	// Add test users with followers
	store.Users["user1"] = &models.User{
		ID:        "user1",
		Username:  "user1",
		Followers: []string{"follower1", "follower2"},
	}

	// Create post with empty content
	post := &postProto.Post{
		UserId:  "user1",
		Content: "",
	}

	// Call PublishPost
	resp, err := postService.PublishPost(context.Background(), post)

	// Assert response
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, int32(2), resp.NotificationsQueued)

	// Allow some time for processing
	time.Sleep(1 * time.Second)

	// Check if notifications were created with appropriate content
	for _, followerID := range []string{"follower1", "follower2"} {
		store.Mu.Lock()
		notifications, exists := store.Notifications[followerID]
		store.Mu.Unlock()

		assert.True(t, exists, "Notifications should exist for follower %s", followerID)
		assert.NotEmpty(t, notifications, "At least one notification should be created for follower %s", followerID)
	}
}

func TestPostServiceWithNonExistentUser(t *testing.T) {
	// Create real store
	store := models.NewStore()

	// Create real queue
	notificationQueue := queue.NewNotificationQueue(store, 3, 2)
	notificationQueue.Start()
	defer notificationQueue.Stop()

	// Create post service
	postService := service.NewPostService(store, notificationQueue)

	// Create post from non-existent user
	post := &postProto.Post{
		UserId:  "nonexistent",
		Content: "Post from non-existent user",
	}

	// Call PublishPost
	resp, err := postService.PublishPost(context.Background(), post)

	// Assert response
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, int32(0), resp.NotificationsQueued) // No followers = no notifications

	// Verify post was still stored
	storedPost, exists := store.Posts["nonexistent"]
	assert.True(t, exists)
	assert.Equal(t, post.Content, storedPost.Content)
}
