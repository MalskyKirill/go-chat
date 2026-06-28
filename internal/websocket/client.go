package websocket

import (
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
}

func NewClient(hub *Hub, userID int64, conn *gorilla.Conn) *Client {
	return &Client{
		UserID: userID,
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
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
		_, _, err := c.Conn.ReadMessage()
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
