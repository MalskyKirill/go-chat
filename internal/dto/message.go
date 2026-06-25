package dto

import "time"

type SendMessageRequest struct {
	Content string `json:"content"`
}

type MessageResponse struct {
	ID             int64     `json:"id"`
	ChatID         int64     `json:"chat_id"`
	SenderID       int64     `json:"sender_id"`
	SenderUsername string    `json:"sender_username"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}
