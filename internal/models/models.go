package models

import (
	"time"
)

type User struct {
	ID        string   `json:"id"`
	Username  string   `json:"username"`
	Followers []string `json:"followers"`
}

type Post struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	Content string `json:"content"`
}

type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusFailed    NotificationStatus = "failed"
)

type Notification struct {
	ID          string             `json:"id"`
	UserID      string             `json:"user_id"`
	PostID      string             `json:"post_id"`
	Content     string             `json:"content"`
	Read        bool               `json:"read"`
	CreatedAt   time.Time          `json:"created_at"`
	Status      NotificationStatus `json:"status"`
	RetryCount  int                `json:"retry_count"`
	LastRetry   *time.Time         `json:"last_retry,omitempty"`
	DeliveredAt *time.Time         `json:"delivered_at,omitempty"`
}

// Metrics related structs
type NotificationMetrics struct {
	TotalNotificationsSent int           `json:"total_notifications_sent"`
	FailedAttempts         int           `json:"failed_attempts"`
	AverageDeliveryTime    time.Duration `json:"average_delivery_time"`
}

type Store struct {
	//UUID -> User
	Users map[string]*User
	//UserId -> Post
	Posts map[string]*Post
	//UserId -> []Notification
	Notifications map[string][]*Notification

	//Metrics Singleton
	Metrics *NotificationMetrics
}

func NewStore() *Store {
	return &Store{
		Users:         make(map[string]*User),
		Posts:         make(map[string]*Post),
		Notifications: make(map[string][]*Notification),
		Metrics:       &NotificationMetrics{},
	}
}

func (s *Store) InitSampleData() {
	users := []*User{
		{ID: "u1", Username: "alice", Followers: []string{"u2", "u3", "u4", "u5"}},
		{ID: "u2", Username: "bob", Followers: []string{"u1", "u3", "u4", "u5"}},
		{ID: "u3", Username: "charlie", Followers: []string{"u1", "u2", "u4", "u5"}},
		{ID: "u4", Username: "david", Followers: []string{"u1", "u2", "u3", "u5"}},
		{ID: "u5", Username: "eve", Followers: []string{"u1", "u2", "u3", "u4"}},
	}

	for _, u := range users {
		s.Users[u.ID] = u
	}

	posts := []*Post{
		{UserID: "u1", Content: "Hello from Alice!"},
		{UserID: "u2", Content: "Bob's first post"},
		{UserID: "u3", Content: "Charlie shares news"},
		{UserID: "u4", Content: "David's photo post"},
		{UserID: "u5", Content: "Eve's thoughts"},
	}

	for _, p := range posts {
		s.Posts[p.UserID] = p
	}
}
