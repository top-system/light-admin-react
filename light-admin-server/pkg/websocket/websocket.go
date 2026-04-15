package websocket

import (
	"time"

	"go.uber.org/zap"
)

// WebSocket is the application-facing service around Hub. Handlers and
// business code depend on this type via FX; the Hub itself is exposed for
// the WebSocket controller which owns the upgrade / read-pump.
type WebSocket struct {
	Hub    *Hub
	logger *zap.Logger
}

// New constructs a WebSocket service backed by a fresh Hub. It wires the
// presence callback so the hub broadcasts online counts on connect/disconnect.
func New(logger *zap.Logger) *WebSocket {
	hub := NewHub(logger)
	ws := &WebSocket{
		Hub:    hub,
		logger: logger.With(zap.String("module", "websocket")),
	}
	hub.OnPresenceChange = ws.broadcastOnlineCount
	return ws
}

// broadcastOnlineCount emits a fresh online-count frame to every connected
// client.  Kept deliberately cheap — it is fired on every presence change.
func (ws *WebSocket) broadcastOnlineCount() {
	ws.Hub.Broadcast(MustFrame(FrameTypeOnlineCount, ws.Hub.SessionCount()))
}

// BroadcastDictChange notifies every client that a dict has been edited so
// their caches can invalidate.
func (ws *WebSocket) BroadcastDictChange(dictCode string) {
	if dictCode == "" {
		return
	}
	ws.Hub.Broadcast(MustFrame(FrameTypeDictChange, map[string]any{
		"dictCode":  dictCode,
		"timestamp": time.Now().UnixMilli(),
	}))
}

// SendNotification pushes a notice to a specific user's connections.
func (ws *WebSocket) SendNotification(username string, payload any) {
	if username == "" || payload == nil {
		return
	}
	ws.Hub.SendToUser(username, MustFrame(FrameTypeNotice, payload))
}

// BroadcastSystemMessage fans out a system announcement to everyone.
func (ws *WebSocket) BroadcastSystemMessage(message string) {
	if message == "" {
		return
	}
	ws.Hub.Broadcast(MustFrame(FrameTypeSystem, map[string]any{
		"sender":    "System",
		"content":   message,
		"timestamp": time.Now().UnixMilli(),
	}))
}

// SendToUser sends a direct message from sender to receiver.
func (ws *WebSocket) SendToUser(sender, receiver, message string) {
	if receiver == "" {
		return
	}
	ws.Hub.SendToUser(receiver, MustFrame(FrameTypeMessage, map[string]any{
		"sender":    sender,
		"content":   message,
		"timestamp": time.Now().UnixMilli(),
	}))
}

// BroadcastNotice broadcasts a textual notice to all clients.
func (ws *WebSocket) BroadcastNotice(message string) {
	ws.Hub.Broadcast(MustFrame(FrameTypeNotice, map[string]any{
		"content":   message,
		"timestamp": time.Now().UnixMilli(),
	}))
}

// GetOnlineUserCount returns distinct online usernames.
func (ws *WebSocket) GetOnlineUserCount() int {
	return ws.Hub.OnlineCount()
}

// GetOnlineUsers returns the current presence snapshot.
func (ws *WebSocket) GetOnlineUsers() []OnlineUser {
	return ws.Hub.OnlineUsers()
}

// IsUserOnline reports whether the given user has any live connections.
func (ws *WebSocket) IsUserOnline(username string) bool {
	return ws.Hub.IsOnline(username)
}
