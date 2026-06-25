package handlers

import (
	"errors"
	"go-chat/internal/dto"
	"go-chat/internal/middleware"
	"go-chat/internal/service"
	"net/http"
	"strconv"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.GetUserID(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	chatId, err := getChatIDFromPath(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid chat id")
	}

	var req dto.SendMessageRequest

	if err := ReadJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := h.messageService.SendMessage(r.Context(), currentUserID, chatId, req)
	if err != nil {
		writeMessageError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

func (h *MessageHandler) GetMessanges(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.GetUserID(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	chatId, err := getChatIDFromPath(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid chat id")
		return
	}

	limit := getIntQueryParam(r, "limit", 50)
	offset := getIntQueryParam(r, "offset", 0)

	response, err := h.messageService.GetChatMessages(r.Context(), currentUserID, chatId, limit, offset)
	if err != nil {
		writeMessageError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, response)
}

func getChatIDFromPath(r *http.Request) (int64, error) {
	chatIDParam := r.PathValue("chatID")

	chatID, err := strconv.ParseInt(chatIDParam, 10, 64)
	if err != nil {
		return 0, err
	}

	return chatID, nil
}

func getIntQueryParam(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return result
}

func writeMessageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		WriteError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, service.ErrChatNotFound):
		WriteError(w, http.StatusNotFound, "chat not found")
	case errors.Is(err, service.ErrNotChatMember):
		WriteError(w, http.StatusForbidden, "you are not a member of this chat")
	case errors.Is(err, service.ErrMessageEmpty):
		WriteError(w, http.StatusBadRequest, "message content is empty")
	case errors.Is(err, service.ErrMessageTooLong):
		WriteError(w, http.StatusBadRequest, "message content is to long")
	default:
		WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
