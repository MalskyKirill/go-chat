package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"go-chat/internal/dto"
	"go-chat/internal/service"
	"log"
	"time"

	gorilla "github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 54 * time.Second
	maxMessageSize = 4096
)

type Client struct {
	UserID int64
	Hub    *Hub
	Conn   *gorilla.Conn
	Send   chan []byte

	messageService *service.MessageService
}

func NewClient(hub *Hub, userID int64, conn *gorilla.Conn, messageService *service.MessageService) *Client {
	return &Client{
		UserID:         userID,
		Hub:            hub,
		Conn:           conn,
		Send:           make(chan []byte, 256),
		messageService: messageService,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		_ = c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)

	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))

	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, rawMessage, err := c.Conn.ReadMessage()
		if err != nil {
			if gorilla.IsUnexpectedCloseError(
				err,
				gorilla.CloseGoingAway,
				gorilla.CloseAbnormalClosure,
			) {
				log.Printf("websocket read error: %v", err)
			}

			break
		}

		c.handleIncomingMessage(rawMessage)
	}

}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				_ = c.Conn.WriteMessage(gorilla.CloseMessage, []byte{})
				return
			}

			writer, err := c.Conn.NextWriter(gorilla.TextMessage)
			if err != nil {
				return
			}

			if _, err := writer.Write(message); err != nil {
				writer.Close()
				return
			}

			if err := writer.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.Conn.WriteMessage(gorilla.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleIncomingMessage(rawMessage []byte) {
	var event IncomingEvent

	if err := json.Unmarshal(rawMessage, &event); err != nil {
		c.sendError("invalid_json", "invalid json message")
		return
	}

	switch event.Type {
	case "message.send":
		c.handleMessageSend(event)
	default:
		c.sendError("unknown_event", "unknown event type")
	}
}

func (c *Client) handleMessageSend(event IncomingEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message, meberIDs, err := c.messageService.SendMessageAndGetChatMemberIDs(ctx, c.UserID, event.ChatID, dto.SendMessageRequest{Content: event.Content})
	if err != nil {
		c.sendServiceError(err)
		return
	}

	c.Hub.SendToUsers <- UserMessage{
		UserIDs: meberIDs,
		Message: EncodeEvent("message.new", message),
	}
}

func (c *Client) sendServiceError(err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		c.sendError("invalid_input", "invalid input")
	case errors.Is(err, service.ErrChatNotFound):
		c.sendError("chat_not_found", "chat not found")
	case errors.Is(err, service.ErrNotChatMember):
		c.sendError("forbidden", "you are not a member of this chat")
	case errors.Is(err, service.ErrMessageEmpty):
		c.sendError("message_empty", "message content is empty")
	case errors.Is(err, service.ErrMessageTooLong):
		c.sendError("message_too_long", "message content is too long")
	default:
		c.sendError("internal_error", "internal server error")
	}
}

func (c *Client) sendError(code string, message string) {
	c.Send <- EncodeEvent("error", map[string]any{
		"code":    code,
		"message": message,
	})
}
