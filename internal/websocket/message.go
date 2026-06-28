package websocket

import "encoding/json"

type OutgoingEvent struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

func EncodeEvent(eventType string, data any) []byte {
	payload, _ := json.Marshal(OutgoingEvent{
		Type: eventType,
		Data: data,
	})

	return payload
}
