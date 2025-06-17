package database

import (
	"time"
)

type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Avatar    string `json:"avatar"`
	Gender    string `json:"gender"`
	Age       int    `json:"age"`
	IsOnline  bool   `json:"isOnline"`
}
type Post struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
	CategoryID    int
	LikesCount    int       `json:"likes_count"`
	DislikesCount int       `json:"dislikes_count"`
	CommentsCount int       `json:"comments_count"`
	Avatar        string    `json:"avatar"`
	Username      string    `json:"username"`
	Comments      []Comment `json:"comments"`
}

// Category represents a category structure
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Comment represents a comment structure
type Comment struct {
	ID            int    `json:"id"`
	PostID        int    `json:"post_id"`
	UserID        int    `json:"user_id"`
	Content       string `json:"content"`
	CreatedAt     string `json:"created_at"`
	LikesCount    int    `json:"likes_count"`
	DislikesCount int    `json:"dislikes_count"`
	Username      string `json:"username"`
	Avatar        string `json:"avatar"`
}

// Reaction represents a reaction structure
type Reaction struct {
	ID        int
	UserID    int
	PostID    int
	CommentID int
	Type      string
	Timestamp string
}

type UserStatus struct {
	UserID   int64  `json:"userId"`
	IsOnline bool   `json:"isOnline"`
	LastSeen string `json:"lastSeen"`
}

type PrivateMessage struct {
	ID         int64  `json:"id"`
	SenderID   int64  `json:"senderId"`
	ReceiverID int64  `json:"receiverId"`
	Content    string `json:"content"`
	SentAt     string `json:"timestamp"`
	Username   string `json:"username"`
}

type ChatLastInteraction struct {
	User1ID             int64  `json:"user1Id"`
	User2ID             int64  `json:"user2Id"`
	LastMessageID       int64  `json:"lastMessageId"`
	LastInteractionTime string `json:"lastInteractionTime"`
}
