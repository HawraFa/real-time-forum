package database

// import (
// 	"time"
// )

type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"-"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	Gender    string `json:"gender"`
	Age       int    `json:"age"`
	FirstName string `json:"firstName"`
	IsOnline  string  `json:"isOnline"`
	LastSeen  string  `json:"lastSeen"`
	LastName  string `json:"lastName"`
	JoinDate  string `json:"joinDate"`
	PostCount int    `json:"postCount"`
}

type Post struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Image         string    `json:"image,omitempty"`
	CreatedAt     string    `json:"created_at"`
	CategoryID    int       `json:"category_id"`
	LikesCount    int       `json:"likes_count"`    
	DislikesCount int       `json:"dislikes_count"` 
	CommentsCount int       `json:"comments_count"` 
	Avatar        string    `json:"avatar"`
	Username      string    `json:"username"`
	Comments      []Comment `json:"comments"`
}

// Category represents a category structure
type Category struct {
	ID   int `json:"id"`
	Name string `json:"name"`
}

// Comment represents a comment structure
type Comment struct {
	ID            int       `json:"id"`
	PostID        int       `json:"postId"`
	UserID        int       `json:"userId"`
	Content       string    `json:"content"`
	CreatedAt     string    `json:"createdAt"`
	LikesCount    int       `json:"likesCount"`
	DislikesCount int       `json:"dislikesCount"`
	Username      string    `json:"username"`
	Avatar        string    `json:"avatar"`
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

type PrivateMessage struct {
	ID         int       `json:"id"`
	SenderID   int       `json:"senderId"`
	ReceiverID int       `json:"receiverId"`
	Content    string    `json:"content"`
	CreatedAt  string    `json:"createdAt"`
	IsRead     bool      `json:"isRead"`
	Sender     User      `json:"sender"`
}

type UserProfile struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Age       int       `json:"age"`
	Gender    string    `json:"gender"`
	Avatar    string    `json:"avatar"`
	IsOnline  string    `json:"isOnline"`
	LastSeen  string    `json:"lastSeen"`
	JoinDate  string    `json:"joinDate"`
	PostCount int       `json:"postCount"`
} 