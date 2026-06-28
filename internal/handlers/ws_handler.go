package handlers

import (
	"go-chat/internal/auth"
	"go-chat/internal/service"
	"go-chat/internal/websocket"
	"net/http"
	"strings"

	gorilla "github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	hub            *websocket.Hub
	jwtSecret      string
	messageService *service.MessageService
}

func NewWebSocketHandler(hub *websocket.Hub, jswSecret string, messageService *service.MessageService) *WebSocketHandler {
	return &WebSocketHandler{
		hub:            hub,
		jwtSecret:      jswSecret,
		messageService: messageService,
	}
}

var upgrader = gorilla.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *WebSocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not alowed")
	}

	tokenString := getTokenFromRequest(r)
	if tokenString == "" {
		WriteError(w, http.StatusUnauthorized, "missing token")
		return
	}

	claims, err := auth.ParseToken(h.jwtSecret, tokenString)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := websocket.NewClient(h.hub, claims.UserID, conn, h.messageService)

	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

func getTokenFromRequest(r *http.Request) string {
	token := r.URL.Query().Get("token")
	if token != "" {
		return token
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return ""
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return parts[1]
}
