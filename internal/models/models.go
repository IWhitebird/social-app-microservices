package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// if user A follows B , then A is follower and B is followee
type UserFollow struct {
	FollowerID string `json:"follower_id"`
	FolloweeID string `json:"followee_id"`
}

// Post represents a post by a user
type Post struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Notification represents a notification sent to a user
type Notification struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`        // Recipient
	PostID       string    `json:"post_id"`        // Related post
	PostAuthorID string    `json:"post_author_id"` // Author of the post
	Content      string    `json:"content"`        // Notification text
	Read         bool      `json:"read"`           // Whether the notification has been read
	CreatedAt    time.Time `json:"created_at"`
}

// Store represents our in-memory data store
type Store struct {
	Users         map[string]*User           // UserID -> User
	Posts         map[string]*Post           // PostID -> Post
	Follows       map[string]*UserFollow     // UserID -> []FollowerID
	Notifications map[string][]*Notification // UserID -> []Notification
}

// type filters struct{

// }

// type[T] Model struct{
// 	model T
// 	getAll(filters)
// }

// NewStore creates a new in-memory data store
func NewStore() *Store {
	return &Store{
		Users:         make(map[string]*User),
		Posts:         make(map[string]*Post),
		Follows:       make(map[string]*UserFollow),
		Notifications: make(map[string][]*Notification),
	}
}

// InitSampleData populates the store with sample data
func (s *Store) InitSampleData() {
	// Create sample users
	users := []*User{
		{ID: "u1", Username: "alice"},
		{ID: "u2", Username: "bob"},
		{ID: "u3", Username: "charlie"},
		{ID: "u4", Username: "david"},
		{ID: "u5", Username: "eve"},
	}

	for _, u := range users {
		s.Users[u.ID] = u
	}

	// Set up follower relationships
	follows := []struct {
		followerID string
		followeeID string
	}{
		{"u2", "u1"}, // Bob follows Alice
		{"u1", "u2"}, // Alice follows Bob
		{"u1", "u3"}, // Alice follows Charlie
		{"u2", "u4"}, // Bob follows David
		{"u1", "u5"}, // Alice follows Eve
		{"u2", "u4"}, // Bob follows David (duplicate)
		{"u2", "u5"}, // Bob follows Eve
		{"u1", "u3"}, // Alice follows Charlie (duplicate)
	}

	for _, f := range follows {
		s.Follows[f.followeeID] = &UserFollow{
			FollowerID: f.followerID,
			FolloweeID: f.followeeID,
		}
	}

	// Add some sample posts
	posts := []*Post{
		{ID: "p1", UserID: "u1", Content: "Hello from Alice!", CreatedAt: time.Now().Add(-24 * time.Hour)},
		{ID: "p2", UserID: "u2", Content: "Bob's first post", CreatedAt: time.Now().Add(-12 * time.Hour)},
		{ID: "p3", UserID: "u3", Content: "Charlie shares news", CreatedAt: time.Now().Add(-6 * time.Hour)},
		{ID: "p4", UserID: "u4", Content: "David's photo post", CreatedAt: time.Now().Add(-3 * time.Hour)},
		{ID: "p5", UserID: "u5", Content: "Eve's thoughts", CreatedAt: time.Now().Add(-1 * time.Hour)},
	}

	for _, p := range posts {
		s.Posts[p.ID] = p
	}
}

// func (s *Store) () {

// }
