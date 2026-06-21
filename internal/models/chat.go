package models

import "time"

const (
	ChatTypePrivate = "private"
	ChatTypeGroup   = "group"
)

type Chat struct {
	ID        int64
	Type      string
	Title     string
	CreatedAt time.Time
}

type ChatMember struct {
	ChatID    int64
	UserID    int64
	Username  string
	CreatedAt time.Time
}
