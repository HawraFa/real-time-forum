package database
import (
	"time"
)
type User struct {
	ID       int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Gender   string `json:"gender"`
	Age      int    `json:"age"`
	LastSeen time.Time `json:"last_seen"`
	IsOnline bool  `json:"is_online"`
}
type Post struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Image         string    `json:"image,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	CategoryID    int  `json:"category_id"`
	LikesCount    int `json:"likes_count"`    
	DislikesCount int `json:"dislikes_count"` 
	CommentsCount int `json:"comments_count"` 
	Avatar 		  string `json:"avatar"`
	Username  	  string `json:"username"`
	Comments 	  []Comment `json:"comments"`
}
// Category represents a category structure
type Category struct {
	ID   int `json:"id"`
	Name string `json:"name"`
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
type PrivateMessage struct {
	ID         int
	SenderID   int
	ReceiverID int
	Content    string
	CreatedAt  time.Time
	IsRead     bool
	Sender     User
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
	IsOnline  bool      `json:"isOnline"`
	LastSeen  time.Time `json:"lastSeen"`
	JoinDate  time.Time `json:"joinDate"`
	PostCount int       `json:"postCount"`
} 