package websocket

type Hub struct {
	clients map[int64]map[*Client]bool

	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int64]map[*Client]bool),

		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte),
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

		}
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
