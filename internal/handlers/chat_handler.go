package handlers

import (
	"errors"
	"go-chat/internal/dto"
	"go-chat/internal/middleware"
	"go-chat/internal/service"
	"net/http"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

func (h *ChatHandler) CreatePrivateChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	curentUserID, ok := middleware.GetUserID(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreatePrivateChatRequest

	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.chatService.CreatePrivateChat(r.Context(), curentUserID, req)
	if err != nil {
		writeChatError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

func (h *ChatHandler) CreateGroupChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	currentUserID, ok := middleware.GetUserID(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateGroupChatRequest

	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.chatService.CreatePublicChat(r.Context(), currentUserID, req)
	if err != nil {
		writeChatError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

func (h *ChatHandler) GetMyChats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	currentUserID, ok := middleware.GetUserID(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "inauthorized")
		return
	}

	response, err := h.chatService.GetUserChats(r.Context(), currentUserID)
	if err != nil {
		writeChatError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

func writeChatError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		WriteError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, service.ErrCannotChatSelf):
		WriteError(w, http.StatusBadRequest, "cannot create private chat with yourself")
	case errors.Is(err, service.ErrGroupTitleEmpty):
		WriteError(w, http.StatusBadRequest, "group title is required")
	case errors.Is(err, service.ErrGroupMembersEmpty):
		WriteError(w, http.StatusBadRequest, "group must have at least one another member")
	case errors.Is(err, service.ErrUserNotFound):
		WriteError(w, http.StatusNotFound, "user not found")
	default:
		WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
