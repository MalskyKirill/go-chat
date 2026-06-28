package websocket

import "encoding/json"

type OutgoingEvent struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

type IncomingEvent struct {
	Type    string `json:"type"`
	ChatID  int64  `json:"chat_id,omitempty"`
	Content string `json:"content,omitempty"`
}

type UserMessage struct {
	UserIDs []int64
	Message []byte
}

func EncodeEvent(eventType string, data any) []byte {
	payload, _ := json.Marshal(OutgoingEvent{
		Type: eventType,
		Data: data,
	})

	return payload
}
