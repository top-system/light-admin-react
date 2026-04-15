package websocket

import (
	"encoding/json"
	"time"
)

// Built-in frame types for the plain JSON protocol.
const (
	FrameTypePing         = "ping"
	FrameTypePong         = "pong"
	FrameTypeOnlineCount  = "online-count"
	FrameTypeDictChange   = "dict-change"
	FrameTypeNotice       = "notice"
	FrameTypeMessage      = "message" // point-to-point message
	FrameTypeSystem       = "system"  // system-wide broadcast
)

// Frame is the wire format exchanged over the WebSocket. Keep it minimal: a
// discriminator (Type), an opaque payload (Data), an optional correlation ID,
// and a server-assigned timestamp in milliseconds.
type Frame struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
	ID   string          `json:"id,omitempty"`
	Ts   int64           `json:"ts,omitempty"`
}

// NewFrame builds a frame, JSON-encoding an arbitrary payload into Data. If
// payload is nil, Data is omitted. If payload is already a json.RawMessage or
// []byte, it is used verbatim.
func NewFrame(frameType string, payload any) (Frame, error) {
	f := Frame{Type: frameType, Ts: time.Now().UnixMilli()}
	if payload == nil {
		return f, nil
	}
	switch v := payload.(type) {
	case json.RawMessage:
		f.Data = v
	case []byte:
		f.Data = v
	default:
		raw, err := json.Marshal(payload)
		if err != nil {
			return Frame{}, err
		}
		f.Data = raw
	}
	return f, nil
}

// MustFrame is NewFrame for well-formed payloads; panics on marshal error.
func MustFrame(frameType string, payload any) Frame {
	f, err := NewFrame(frameType, payload)
	if err != nil {
		panic(err)
	}
	return f
}
