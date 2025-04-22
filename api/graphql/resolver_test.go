package graphql

import (
	"testing"
	"time"

	"github.com/paper-social/notification-service/internal/models"
)

func TestGetNotifications(t *testing.T) {
	// Create a store with sample data
	store := models.NewStore()

	// Add some sample notifications
	userID := "test-user"
	store.Notifications[userID] = []*models.Notification{
		{
			ID:           "n1",
			UserID:       userID,
			PostID:       "p1",
			PostAuthorID: "author1",
			Content:      "Test notification 1",
			Read:         false,
			CreatedAt:    time.Now().Add(-2 * time.Hour),
		},
		{
			ID:           "n2",
			UserID:       userID,
			PostID:       "p2",
			PostAuthorID: "author2",
			Content:      "Test notification 2",
			Read:         true,
			CreatedAt:    time.Now().Add(-1 * time.Hour),
		},
		{
			ID:           "n3",
			UserID:       userID,
			PostID:       "p3",
			PostAuthorID: "author1",
			Content:      "Test notification 3",
			Read:         false,
			CreatedAt:    time.Now(),
		},
	}

	// Create resolver
	resolver := NewResolver(store)

	// We'll test the fallback method directly since we can't mock the gRPC service easily in this test
	notifications := resolver.getNotificationsFromStore(userID)

	// Verify results
	if len(notifications) != 3 {
		t.Errorf("Expected 3 notifications, got %d", len(notifications))
	}

	// Verify order (most recent first)
	if notifications[0].id != "n3" || notifications[1].id != "n2" || notifications[2].id != "n1" {
		t.Errorf("Notifications not sorted correctly by time")
	}

	// Verify content
	if notifications[0].content != "Test notification 3" {
		t.Errorf("Expected content 'Test notification 3', got '%s'", notifications[0].content)
	}

	// Test for non-existent user
	emptyNotifications := resolver.getNotificationsFromStore("non-existent")

	if len(emptyNotifications) != 0 {
		t.Errorf("Expected 0 notifications for non-existent user, got %d", len(emptyNotifications))
	}
}
