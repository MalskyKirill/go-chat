package repositories

import (
	"context"
	"errors"
	"go-chat/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}
func (r *MessageRepository) Create(ctx context.Context, message *models.Message) error {
	query := `
		INSERT INTO message(chat_id, sender_id, content)
		VALUES($1, $2, $3)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query, message.ChatID, message.SenderID, message.Content).Scan(&message.ID, &message.CreatedAt)

	return err
}

func (r *MessageRepository) FindById(ctx context.Context, messageID int64) (*models.Message, error) {
	query := `
		SELECT m.id, m.chat_id, m.sender_id, u.username, m.content, m.created_at
		FROM messages AS m
		JOIN users AS u ON u.id = m.sender_id
		WHERE m.id = $1
	`

	var message models.Message

	err := r.db.QueryRow(ctx, query, messageID).Scan(
		&message.ID,
		&message.ChatID,
		&message.SenderID,
		&message.SenderUsername,
		&message.Content,
		&message.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (r *MessageRepository) FindByChatID(ctx context.Context, chatID int64, limit int, offset int) ([]models.Message, error) {
	query := `
		SELECT m.id, m.chat_id, m.sender_id, u.username, m.content, m.created_at
		FROM messages as m
		JOIN users AS u ON u.id = m.sender_id
		WHERE chat_id = $1
		ORDER BY m.created_at ASC
		LIMIL $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	messages := make([]models.Message, 0)

	for rows.Next() {
		var message models.Message

		if err := rows.Scan(
			&message.ID,
			&message.ChatID,
			&message.SenderID,
			&message.SenderUsername,
			&message.Content,
			&message.CreatedAt,
		); err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
