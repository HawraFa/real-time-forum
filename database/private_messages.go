package database

import (
	"database/sql"
)

// GetUserMessages retrieves the last 10 messages between two users
func GetUserMessages(db *sql.DB, user1ID, user2ID int64, offset int) ([]PrivateMessage, error) {
	query := `
		SELECT id, sender_id, receiver_id, message_text, sent_at, is_read 
		FROM private_messages 
		WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)
		ORDER BY sent_at DESC
		LIMIT 10 OFFSET ?`

	rows, err := db.Query(query, user1ID, user2ID, user2ID, user1ID, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []PrivateMessage
	for rows.Next() {
		var msg PrivateMessage
		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Message, &msg.SentAt, &msg.IsRead)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// SaveMessage saves a new private message and updates the last interaction
func SaveMessage(db *sql.DB, senderID, receiverID int64, message string) error {
	// Insert the message
	result, err := db.Exec(`
		INSERT INTO private_messages (sender_id, receiver_id, message_text)
		VALUES (?, ?, ?)`,
		senderID, receiverID, message)
	if err != nil {
		return err
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Update or insert last interaction
	_, err = db.Exec(`
		INSERT INTO chat_last_interactions (user1_id, user2_id, last_message_id, last_interaction_time)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user1_id, user2_id) DO UPDATE SET
			last_message_id = ?,
			last_interaction_time = CURRENT_TIMESTAMP`,
		senderID, receiverID, messageID, messageID)

	return err
}

// GetUserChats retrieves all chat conversations for a user, sorted by last interaction
func GetUserChats(db *sql.DB, userID int64) ([]ChatLastInteraction, error) {
	query := `
		SELECT 
			CASE 
				WHEN user1_id = ? THEN user2_id 
				ELSE user1_id 
			END as other_user_id,
			last_message_id,
			last_interaction_time
		FROM chat_last_interactions
		WHERE user1_id = ? OR user2_id = ?
		ORDER BY last_interaction_time DESC`

	rows, err := db.Query(query, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []ChatLastInteraction
	for rows.Next() {
		var chat ChatLastInteraction
		err := rows.Scan(&chat.User2ID, &chat.LastMessageID, &chat.LastInteractionTime)
		if err != nil {
			return nil, err
		}
		chat.User1ID = userID
		chats = append(chats, chat)
	}
	return chats, nil
}

// UpdateUserStatus updates a user's online status and last seen time
func UpdateUserStatus(db *sql.DB, userID int64, isOnline bool) error {
	_, err := db.Exec(`
		INSERT INTO user_status (user_id, is_online, last_seen)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id) DO UPDATE SET
		is_online = ?,
		last_seen = CURRENT_TIMESTAMP`,
		userID, isOnline, isOnline)
	return err
}

// GetOnlineUsers retrieves all currently online users
func GetOnlineUsers(db *sql.DB) ([]int64, error) {
	rows, err := db.Query(`
		SELECT user_id 
		FROM user_status 
		WHERE is_online = true 
		AND last_seen > datetime('now', '-5 minutes')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		users = append(users, userID)
	}
	return users, nil
}

// MarkMessagesAsRead marks all messages from a sender to a receiver as read
func MarkMessagesAsRead(db *sql.DB, senderID, receiverID int64) error {
	_, err := db.Exec(`
		UPDATE private_messages 
		SET is_read = true 
		WHERE sender_id = ? AND receiver_id = ? AND is_read = false`,
		senderID, receiverID)
	return err
} 

