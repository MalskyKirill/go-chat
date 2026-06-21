package repositories

import (
	"context"
	"errors"
	"go-chat/internal/models"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatReposytory struct {
	db *pgxpool.Pool
}

func NewChatReposytory(db *pgxpool.Pool) *ChatReposytory {
	return &ChatReposytory{
		db: db,
	}
}

func (r *ChatReposytory) CreateChat(ctx context.Context, chat *models.Chat, membersIDs []int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var title any = nil
	if strings.TrimSpace(chat.Title) != "" {
		title = strings.TrimSpace(chat.Title)
	}

	query := `
		INSERT into chats(type, title)
		VALUES ($1, $2)
		RETURNING id, COALESCE(title, ''), created_at
	`

	err = tx.QueryRow(ctx, query, chat.Type, title).Scan(&chat.ID, &chat.Title, &chat.CreatedAt)
	if err != nil {
		return err
	}

	for _, memberID := range membersIDs {
		_, err := tx.Exec(
			ctx,
			`
				INSERT INTO chat_members (chat_id, user_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`,
			chat.ID,
			memberID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *ChatReposytory) FindPrivateChatBetweenUsers(ctx context.Context, firstUserID int64, secondUserID int64) (*models.Chat, error) {
	query := `
		SELECT c.id, c.type, COALESCE(c.title, ''), c.created_at
		FROM chats AS c
		JOIN chat_members AS cm ON cm.chat_id = c.id
		WHERE c.type = 'private' AND cm.user_id IN ($1, $2)
		GROUP BY c.id, c.type, c.title, c.created_at
		HAVING COUNT (DISTINCT cm.user_id) = 2 
			AND (
				SELECT COUNT(*)
				FROM chat_members
				WHERE chat_id = c.id
			) = 2
		LIMIT 1;
	`

	var chat models.Chat

	err := r.db.QueryRow(ctx, query, firstUserID, secondUserID).Scan(&chat.ID, &chat.Type, &chat.Title, &chat.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &chat, nil
}

func (r *ChatReposytory) FindByID(ctx context.Context, chatID int64) (*models.Chat, error) {
	query := `
		SELECT c.id, c.type, COALESCE(c.title, ''), c.created_at
		FROM chats AS c
		WHERE c.id = $1
	`

	var chat models.Chat

	err := r.db.QueryRow(ctx, query, chatID).Scan(&chat.ID, &chat.Type, &chat.Title, &chat.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &chat, nil
}

func (r *ChatReposytory) FindChatsByUserId(ctx context.Context, userID int64) ([]models.Chat, error) {
	query := `
		SELECT c.id, c.type, COALESCE(c.title, ''), c.created_at
		FROM chats as c
		JOIN chat_members AS cm ON cm.chat_id = c.id
		WHERE cm.user_id = $1
		ORDER BY c.created_at DESC
		`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var chats []models.Chat

	for rows.Next() {
		var chat models.Chat

		if err := rows.Scan(&chat.ID, &chat.Type, &chat.Title, &chat.CreatedAt); err != nil {
			return nil, err
		}

		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

func (r *ChatReposytory) FindMembersByChatId(ctx context.Context, chatID int64) ([]models.ChatMember, error) {
	query := `
		SELECT cm.chat_id, cm.user_id, u.username, cm.created_at
		FROM chat_members AS cm
		JOIN users AS u ON cm.user_id = u.id
		WHERE cm.chat_id = $1
		ORDER BY u.username ASC
	`

	rows, err := r.db.Query(ctx, query, chatID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var chatMembers []models.ChatMember
	for rows.Next() {
		var chatMember models.ChatMember

		if err := rows.Scan(&chatMember.ChatID, &chatMember.UserID, &chatMember.Username, &chatMember.CreatedAt); err != nil {
			return nil, err
		}

		chatMembers = append(chatMembers, chatMember)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chatMembers, nil
}

func (r *ChatReposytory) IsUserMember(ctx context.Context, chatID int64, userID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM chat_members
			WHERE chat_id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, chatID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
