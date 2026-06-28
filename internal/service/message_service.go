package service

import (
	"context"
	"errors"
	"go-chat/internal/dto"
	"go-chat/internal/models"
	"go-chat/internal/repositories"
	"strings"
)

var (
	ErrChatNotFound   = errors.New("chat not found")
	ErrNotChatMember  = errors.New("user is not chat members")
	ErrMessageEmpty   = errors.New("message content is empty")
	ErrMessageTooLong = errors.New("message content is too long")
)

type MessageService struct {
	messageRepo *repositories.MessageRepository
	chatRepo    *repositories.ChatReposytory
}

func NewMessageService(messageRepo *repositories.MessageRepository, chatRepo *repositories.ChatReposytory) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		chatRepo:    chatRepo,
	}
}

func (s *MessageService) SendMessage(ctx context.Context, currentUserID int64, chatID int64, req dto.SendMessageRequest) (*dto.MessageResponse, error) {
	if chatID <= 0 {
		return nil, ErrInvalidInput
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, ErrMessageEmpty
	}

	if len(content) > 4000 {
		return nil, ErrMessageTooLong
	}

	_, err := s.chatRepo.FindByID(ctx, chatID)
	if errors.Is(err, repositories.ErrNotFound) {
		return nil, ErrChatNotFound
	}

	if err != nil {
		return nil, err
	}

	isMember, err := s.chatRepo.IsUserMember(ctx, chatID, currentUserID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, ErrNotChatMember
	}

	message := &models.Message{
		ChatID:   chatID,
		SenderID: currentUserID,
		Content:  content,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	createdMessage, err := s.messageRepo.FindById(ctx, message.ID)
	if err != nil {
		return nil, err
	}

	response := toMessageResponse(createdMessage)

	return &response, nil
}

func (s *MessageService) GetChatMessages(ctx context.Context, currentUserId int64, chatID int64, limit int, offset int) ([]dto.MessageResponse, error) {
	if chatID <= 0 {
		return nil, ErrChatNotFound
	}

	if limit <= 0 {
		limit = 0
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	_, err := s.chatRepo.FindByID(ctx, chatID)
	if errors.Is(err, repositories.ErrNotFound) {
		return nil, ErrChatNotFound
	}

	if err != nil {
		return nil, err
	}

	isMember, err := s.chatRepo.IsUserMember(ctx, chatID, currentUserId)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, ErrNotChatMember
	}

	messages, err := s.messageRepo.FindByChatID(ctx, chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	response := make([]dto.MessageResponse, 0, len(messages))

	for _, message := range messages {
		messageCopy := message
		response = append(response, toMessageResponse((*models.Message)(&messageCopy)))
	}

	return response, nil
}

func toMessageResponse(message *models.Message) dto.MessageResponse {
	return dto.MessageResponse{
		ID:             message.ID,
		ChatID:         message.ChatID,
		SenderID:       message.SenderID,
		SenderUsername: message.SenderUsername,
		Content:        message.Content,
		CreatedAt:      message.CreatedAt,
	}
}
