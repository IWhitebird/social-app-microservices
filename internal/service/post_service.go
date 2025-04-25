package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/iwhitebird/social-app-microservices/internal/models"
	"github.com/iwhitebird/social-app-microservices/internal/queue"
	postProto "github.com/iwhitebird/social-app-microservices/proto/generated/post/proto"
)

// PostService implements the gRPC post service
type PostService struct {
	postProto.UnimplementedPostServiceServer
	store *models.Store
	queue *queue.NotificationQueue
	mu    sync.Mutex
}

// NewPostService creates a new PostService
func NewPostService(store *models.Store, queue *queue.NotificationQueue) *PostService {
	return &PostService{
		store: store,
		queue: queue,
	}
}

// PublishPost handles a new post and creates notifications for followers
func (s *PostService) PublishPost(ctx context.Context, post *postProto.Post) (*postProto.NotificationResponse, error) {
	log.Printf("Received PublishPost request for user %s", post.UserId)

	// Convert proto post to internal post
	internalPost := &models.Post{
		ID:      uuid.New().String(),
		UserID:  post.UserId,
		Content: post.Content,
	}

	// Store the post
	s.mu.Lock()
	s.store.Posts[internalPost.UserID] = internalPost
	s.mu.Unlock()

	// Get followers of the post author
	var followers []string
	for _, val := range s.store.Users {
		if val.ID == post.UserId {
			followers = val.Followers
		}
	}

	log.Printf("Creating notifications for %d followers of user %s", len(followers), post.UserId)
	// Create and queue notifications for each follower
	for _, followerID := range followers {
		notification := &models.Notification{
			ID:        uuid.New().String(),
			UserID:    followerID,
			PostID:    internalPost.ID,
			Content:   fmt.Sprintf("%s posted: %s", post.UserId, post.Content),
			Read:      false,
			CreatedAt: time.Now(),
		}

		// Queue the notification for delivery
		s.queue.EnqueueNotification(notification)
	}

	return &postProto.NotificationResponse{
		Success:             true,
		Message:             fmt.Sprintf("Post published, %d notifications queued", len(followers)),
		NotificationsQueued: int32(len(followers)),
	}, nil
}
