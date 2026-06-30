package handlers

import (
	"errors"
	"net/http"

	"go-chat/internal/middleware"
	"go-chat/internal/service"
	chatws "go-chat/internal/websocket"
)

type OnlineHandler struct {
	chatService *service.ChatService
	hub         *chatws.Hub
}

func NewOnlineHandler(
	chatService *service.ChatService,
	hub *chatws.Hub,
) *OnlineHandler {
	return &OnlineHandler{
		chatService: chatService,
		hub:         hub,
	}
}

func (h *OnlineHandler) GetOnlineRelatedUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	currentUserID, ok := middleware.GetUserID(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	relatedUserIDs, err := h.chatService.GetRelatedUserIDs(r.Context(), currentUserID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	onlineUserIDs := h.hub.FilterOnlineUserIDs(relatedUserIDs)

	response, err := h.chatService.GetUsersByIDs(r.Context(), onlineUserIDs)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

func (h *OnlineHandler) GetOnlineChatMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	currentUserID, ok := middleware.GetUserID(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	chatID, err := getChatIDFromPath(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid chat id")
		return
	}

	chatMemberIDs, err := h.chatService.GetChatMemberIDs(
		r.Context(),
		currentUserID,
		chatID,
	)
	if err != nil {
		writeOnlineError(w, err)
		return
	}

	onlineUserIDs := h.hub.FilterOnlineUserIDs(chatMemberIDs)

	response, err := h.chatService.GetUsersByIDs(r.Context(), onlineUserIDs)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

func writeOnlineError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		WriteError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, service.ErrChatNotFound):
		WriteError(w, http.StatusNotFound, "chat not found")
	case errors.Is(err, service.ErrNotChatMember):
		WriteError(w, http.StatusForbidden, "you are not a member of this chat")
	default:
		WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
