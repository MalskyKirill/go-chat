package websocket

import (
	"context"
	"log"
	"time"
)

type StatusAudienceProvider interface {
	GetStatusAudience(ctx context.Context, userID int64) ([]int64, error)
}

type OnlineUsersRequest struct {
	UserIDs  []int64
	Response chan []int64
}

type Hub struct {
	clients map[int64]map[*Client]bool

	Register    chan *Client
	Unregister  chan *Client
	Broadcast   chan []byte
	SendToUsers chan UserMessage
	OnlineUsers chan OnlineUsersRequest

	audienceProvider StatusAudienceProvider
}

func NewHub(audienceProvider StatusAudienceProvider) *Hub {
	return &Hub{
		clients: make(map[int64]map[*Client]bool),

		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Broadcast:   make(chan []byte),
		SendToUsers: make(chan UserMessage),
		OnlineUsers: make(chan OnlineUsersRequest),

		audienceProvider: audienceProvider,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.broadcastToAll(message)

		case userMessage := <-h.SendToUsers:
			h.sendToUsers(userMessage.UserIDs, userMessage.Message)

		case request := <-h.OnlineUsers:
			request.Response <- h.filterOnlineUserIDs(request.UserIDs)
		}
	}
}

func (h *Hub) FilterOnlineUserIDs(userIDs []int64) []int64 {
	response := make(chan []int64, 1)

	request := OnlineUsersRequest{
		UserIDs:  userIDs,
		Response: response,
	}

	select {
	case h.OnlineUsers <- request:
	case <-time.After(2 * time.Second):
		return []int64{}
	}

	select {
	case result := <-response:
		return result
	case <-time.After(2 * time.Second):
		return []int64{}
	}
}

func (h *Hub) registerClient(client *Client) {
	wasOffline := false

	if _, ok := h.clients[client.UserID]; !ok {
		h.clients[client.UserID] = make(map[*Client]bool)
		wasOffline = true
	}

	h.clients[client.UserID][client] = true

	client.Send <- EncodeEvent("connection.ready", map[string]any{
		"user_id": client.UserID,
	})

	if wasOffline {
		h.broadcastToAll(EncodeEvent("user.online", map[string]any{
			"user_id": client.UserID,
		}))
	}
}

func (h *Hub) unregisterClient(client *Client) {
	removed := h.removeClient(client)
	if !removed {
		return
	}

	if !h.isUserOnline(client.UserID) {
		h.broadcastToAll(EncodeEvent("user.offline", map[string]any{
			"user_id": client.UserID,
		}))
	}

}

func (h *Hub) removeClient(client *Client) bool {
	userClients, ok := h.clients[client.UserID]
	if !ok {
		return false
	}

	if _, ok := userClients[client]; !ok {
		return false
	}

	delete(userClients, client)
	close(client.Send)

	if len(userClients) == 0 {
		delete(h.clients, client.UserID)
	}

	return true
}

func (h *Hub) isUserOnline(userID int64) bool {
	userClients, ok := h.clients[userID]
	return ok && len(userClients) > 0
}

func (h *Hub) broadcastToAll(message []byte) {
	for _, userClients := range h.clients {
		for client := range userClients {
			select {
			case client.Send <- message:
			default:
				h.removeClient(client)
			}
		}
	}
}

func (h *Hub) sendToUsers(userIDs []int64, message []byte) {
	for _, userID := range userIDs {
		userClient, ok := h.clients[userID]
		if !ok {
			continue
		}

		for client := range userClient {
			select {
			case client.Send <- message:
			default:
				h.registerClient(client)
			}
		}
	}
}

func (h *Hub) filterOnlineUserIDs(userIDs []int64) []int64 {
	seen := make(map[int64]bool)
	result := make([]int64, 0)

	for _, userId := range userIDs {
		if seen[userId] {
			continue
		}

		seen[userId] = true

		if h.isUserOnline(userId) {
			result = append(result, userId)
		}
	}

	return result
}

func (h *Hub) sendStatusEventToAudience(userID int64, eventType string) {
	audienceUserIDs := h.getStatusAudience(userID)
	if len(audienceUserIDs) == 0 {
		return
	}

	h.sendToUsers(audienceUserIDs, EncodeEvent(eventType, map[string]any{
		"user_id": userID,
	}))
}

func (h *Hub) getStatusAudience(userId int64) []int64 {
	if h.audienceProvider != nil {
		return []int64{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	userIDs, err := h.audienceProvider.GetStatusAudience(ctx, userId)
	if err != nil {
		log.Printf("failed to get status audience: %v", err)
		return []int64{}
	}

	return userIDs
}
