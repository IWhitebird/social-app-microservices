package graphql

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/graph-gophers/graphql-go"
	notificationProto "github.com/paper-social/notification-service/api/proto/gen/github.com/paper-social/notification-service/api/proto/notification"
	"github.com/paper-social/notification-service/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Resolver is the root resolver for the GraphQL schema
type Resolver struct {
	store      *models.Store
	grpcClient notificationProto.NotificationServiceClient
}

// NewResolver creates a new GraphQL resolver
func NewResolver(store *models.Store) *Resolver {
	// Connect to gRPC server
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}

	// Create gRPC client
	grpcClient := notificationProto.NewNotificationServiceClient(conn)

	return &Resolver{
		store:      store,
		grpcClient: grpcClient,
	}
}

// Notification represents a notification in the GraphQL schema
type Notification struct {
	id           string
	userId       string
	postId       string
	postAuthorId string
	content      string
	read         bool
	createdAt    time.Time
}

// GetNotifications returns notifications for a user
func (r *Resolver) GetNotifications(args struct{ UserID string }) []*Notification {
	userID := args.UserID

	// Use the gRPC client to get notifications
	stream, err := r.grpcClient.GetNotifications(context.Background(), &notificationProto.UserId{UserId: userID})
	if err != nil {
		log.Printf("Error getting notifications via gRPC: %v", err)
		// Fall back to store if gRPC fails
		return r.getNotificationsFromStore(userID)
	}

	var notifications []*Notification
	for {
		notification, err := stream.Recv()
		if err != nil {
			break
		}

		// Convert timestamp to time.Time
		createdAt := time.Unix(notification.CreatedAt, 0)

		// Convert from proto model to GraphQL model
		notifications = append(notifications, &Notification{
			id:           notification.Id,
			userId:       notification.UserId,
			postId:       notification.PostId,
			postAuthorId: notification.PostAuthorId,
			content:      notification.Content,
			read:         notification.Read,
			createdAt:    createdAt,
		})
	}

	// Sort notifications by created time (most recent first)
	sort.Slice(notifications, func(i, j int) bool {
		return notifications[i].createdAt.After(notifications[j].createdAt)
	})

	// Limit to 20 most recent notifications
	if len(notifications) > 20 {
		notifications = notifications[:20]
	}

	return notifications
}

// getNotificationsFromStore is a fallback method to get notifications from the local store
func (r *Resolver) getNotificationsFromStore(userID string) []*Notification {
	userNotifications, exists := r.store.Notifications[userID]
	if !exists {
		return []*Notification{}
	}

	// Sort notifications by created time (most recent first)
	sort.Slice(userNotifications, func(i, j int) bool {
		return userNotifications[i].CreatedAt.After(userNotifications[j].CreatedAt)
	})

	// Convert from internal model to GraphQL model
	var result []*Notification
	for i, notification := range userNotifications {
		if i >= 20 {
			break // Limit to 20 most recent notifications
		}

		result = append(result, &Notification{
			id:           notification.ID,
			userId:       notification.UserID,
			postId:       notification.PostID,
			postAuthorId: notification.PostAuthorID,
			content:      notification.Content,
			read:         notification.Read,
			createdAt:    notification.CreatedAt,
		})
	}

	return result
}

// ID returns the notification ID
func (n *Notification) ID() graphql.ID {
	return graphql.ID(n.id)
}

// UserID returns the user ID
func (n *Notification) UserID() string {
	return n.userId
}

// PostID returns the post ID
func (n *Notification) PostID() string {
	return n.postId
}

// PostAuthorID returns the post author ID
func (n *Notification) PostAuthorID() string {
	return n.postAuthorId
}

// Content returns the notification content
func (n *Notification) Content() string {
	return n.content
}

// Read returns whether the notification has been read
func (n *Notification) Read() bool {
	return n.read
}

// CreatedAt returns the notification creation time
func (n *Notification) CreatedAt() string {
	return n.createdAt.Format(time.RFC3339)
}
