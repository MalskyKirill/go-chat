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
	ErrUserNotFound      = errors.New("user not found")
	ErrCannotChatSelf    = errors.New("cannot create private chat with yourself")
	ErrGroupTitleEmpty   = errors.New("group title is required")
	ErrGroupMembersEmpty = errors.New("group must have at least one another member")
)

type ChatService struct {
	chatRepo *repositories.ChatReposytory
	userRepo *repositories.UserRepository
}

func NewChatService(chatRepo *repositories.ChatReposytory, userRepo *repositories.UserRepository) *ChatService {
	return &ChatService{
		chatRepo: chatRepo,
		userRepo: userRepo,
	}
}

func (s *ChatService) CreatePrivateChat(ctx context.Context, currentUserID int64, req dto.CreatePrivateChatRequest) (*dto.ChatResponse, error) {
	if req.UserID <= 0 {
		return nil, ErrInvalidInput
	}

	if req.UserID == currentUserID {
		return nil, ErrCannotChatSelf
	}

	_, err := s.userRepo.FindByID(ctx, req.UserID)
	if errors.Is(err, repositories.ErrNotFound) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	existingChat, err := s.chatRepo.FindPrivateChatBetweenUsers(ctx, currentUserID, req.UserID)
	if err == nil {
		return s.buildChatResponse(ctx, existingChat)
	}

	if !errors.Is(err, repositories.ErrNotFound) {
		return nil, err
	}

	chat := &models.Chat{
		Type: models.ChatTypePrivate,
	}

	membersIDs := []int64{
		currentUserID,
		req.UserID,
	}

	if err := s.chatRepo.CreateChat(ctx, chat, membersIDs); err != nil {
		return nil, err
	}

	return s.buildChatResponse(ctx, chat)
}

func (s *ChatService) CreatePublicChat(ctx context.Context, currentUserID int64, req dto.CreateGroupChatRequest) (*dto.ChatResponse, error) {
	req.Title = strings.TrimSpace(req.Title)

	if req.Title == "" {
		return nil, ErrGroupTitleEmpty
	}

	if len(req.MembersIDs) == 0 {
		return nil, ErrGroupMembersEmpty
	}

	membersIDs := uniqueIDs(append(req.MembersIDs, currentUserID))

	if len(membersIDs) < 2 {
		return nil, ErrGroupMembersEmpty
	}

	for _, memberID := range membersIDs {
		if memberID <= 0 {
			return nil, ErrInvalidInput
		}

		_, err := s.userRepo.FindByID(ctx, memberID)
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrUserNotFound
		}

		if err != nil {
			return nil, err
		}
	}

	chat := &models.Chat{
		Type:  models.ChatTypeGroup,
		Title: req.Title,
	}

	if err := s.chatRepo.CreateChat(ctx, chat, membersIDs); err != nil {
		return nil, err
	}

	return s.buildChatResponse(ctx, chat)
}

func (s *ChatService) GetUserChats(ctx context.Context, userID int64) ([]dto.ChatResponse, error) {
	chats, err := s.chatRepo.FindChatsByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := make([]dto.ChatResponse, 0, len(chats))

	for _, chat := range chats {
		chatCopy := chat

		chatResponse, err := s.buildChatResponse(ctx, &chatCopy)
		if err != nil {
			return nil, err
		}

		response = append(response, *chatResponse)
	}

	return response, nil
}

func (s *ChatService) GetStatusAudience(ctx context.Context, userID int64) ([]int64, error) {
	return s.chatRepo.FindRelatedUserIDsByUserID(ctx, userID)
}

func (s *ChatService) GetRelatedUserIDs(ctx context.Context, currentUserID int64) ([]int64, error) {
	return s.chatRepo.FindRelatedUserIDsByUserID(ctx, currentUserID)
}

func (s *ChatService) GetChatMemberIDs(ctx context.Context, currentUserID int64, chatID int64) ([]int64, error) {
	if chatID <= 0 {
		return nil, ErrInvalidInput
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

	members, err := s.chatRepo.FindMembersByChatId(ctx, chatID)
	if err != nil {
		return nil, err
	}

	membersIDs := make([]int64, 0, len(members))

	for _, member := range members {
		membersIDs = append(membersIDs, member.UserID)
	}

	return membersIDs, nil
}

func (s *ChatService) GetUsersByIDs(ctx context.Context, userIDs []int64) ([]dto.OnlineUserResponse, error) {
	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	response := make([]dto.OnlineUserResponse, 0, len(users))

	for _, user := range users {
		response = append(response, dto.OnlineUserResponse{
			ID:       user.ID,
			Username: user.Username,
		})
	}

	return response, nil
}

func (s *ChatService) buildChatResponse(ctx context.Context, chat *models.Chat) (*dto.ChatResponse, error) {
	members, err := s.chatRepo.FindMembersByChatId(ctx, chat.ID)
	if err != nil {
		return nil, err
	}

	memberResponses := make([]dto.ChatMemberResponse, 0, len(members))
	for _, member := range members {
		memberResponses = append(memberResponses, dto.ChatMemberResponse{
			ID:       member.UserID,
			Username: member.Username,
		})
	}

	return &dto.ChatResponse{
		ID:          chat.ID,
		Type:        chat.Type,
		Title:       chat.Title,
		ChatMembers: memberResponses,
		CreatedAt:   chat.CreatedAt,
	}, nil
}

func uniqueIDs(ids []int64) []int64 {
	seen := make(map[int64]bool)
	result := make([]int64, 0, len(ids))

	for _, id := range ids {
		if seen[id] {
			continue
		}

		seen[id] = true
		result = append(result, id)
	}

	return result
}
