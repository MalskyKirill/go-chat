package dto

import "time"

type CreatePrivateChatRequest struct {
	UserID int64 `json:"user_id"`
}

type CreateGroupChatRequest struct {
	Title      string  `json:"title"`
	MembersIDs []int64 `json:"members_ids"`
}

type ChatMemberResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type ChatResponse struct {
	ID          int64                `json:"id"`
	Type        string               `json:"type"`
	Title       string               `json:"title"`
	ChatMembers []ChatMemberResponse `json:"members"`
	CreatedAt   time.Time            `json:"created_at"`
}
