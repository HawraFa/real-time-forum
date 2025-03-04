package database

import (
	"time"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Gender   string `json:"gender"`
	Age      int    `json:"age"`
}
type Post struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
	CategoryID    int
	LikesCount    int `json:"likes_count"`    
	DislikesCount int `json:"dislikes_count"` 
	CommentsCount int `json:"comments_count"` 
	Avatar 		  string `json:"avatar"`
	Username  	  string `json:"username"`
	Comments 	  []Comment `json:"comments"`

}

// Category represents a category structure
type Category struct {
	ID   int
	Name string
}

// Comment represents a comment structure
type Comment struct {
	ID            int
	PostID        int
	UserID        int
	Content       string
	CreatedAt     string //could comment this but will mess up queries
	LikesCount    int
	DislikesCount int
	Username      string
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
	UserID    int64     `json:"userId"`
	IsOnline  bool      `json:"isOnline"`
	LastSeen  string    `json:"lastSeen"`
}

type PrivateMessage struct {
	ID         int64     `json:"id"`
	SenderID   int64     `json:"senderId"`
	ReceiverID int64     `json:"receiverId"`
	Message    string    `json:"message"`
	SentAt     string    `json:"sentAt"`
	IsRead     bool      `json:"isRead"`
}

type ChatLastInteraction struct {
	User1ID            int64     `json:"user1Id"`
	User2ID            int64     `json:"user2Id"`
	LastMessageID      int64     `json:"lastMessageId"`
	LastInteractionTime string    `json:"lastInteractionTime"`
}
