package websocket

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Heartbeat parameters.
const (
	pingInterval = 30 * time.Second
	readDeadline = 60 * time.Second
	writeTimeout = 10 * time.Second
)

// Conn is a minimal abstraction over *websocket.Conn so the Hub can be tested
// with fakes. Only the two calls the Hub actually makes are exposed.
type Conn interface {
	WriteMessage(messageType int, data []byte) error
	Close() error
}

// client is an authenticated connection attached to a user.
type client struct {
	id        string
	username  string
	conn      Conn
	writeMu   sync.Mutex
	joinedAt  int64
	lastPong  atomic.Int64 // unix millis
	closed    atomic.Bool
	closeOnce sync.Once
}

func (c *client) send(frame Frame) error {
	data, err := json.Marshal(frame)
	if err != nil {
		return err
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.closed.Load() {
		return nil
	}
	if deadliner, ok := c.conn.(interface{ SetWriteDeadline(time.Time) error }); ok {
		_ = deadliner.SetWriteDeadline(time.Now().Add(writeTimeout))
	}
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *client) close() {
	c.closeOnce.Do(func() {
		c.closed.Store(true)
		_ = c.conn.Close()
	})
}

// OnlineUser is the public view of a user's presence.
type OnlineUser struct {
	Username     string `json:"username"`
	SessionCount int    `json:"sessionCount"`
	LoginTime    int64  `json:"loginTime"`
}

// Hub owns the set of active connections and multiplexes server → client frames.
type Hub struct {
	mu      sync.RWMutex
	byID    map[string]*client            // clientID -> client
	byUser  map[string]map[string]*client // username -> clientID -> client
	logger  *zap.Logger
	nowFn   func() time.Time
	msgSeq  atomic.Uint64

	// OnPresenceChange is fired after Register / Unregister so callers can
	// broadcast online counts, etc. Called with the hub lock NOT held.
	OnPresenceChange func()
}

// NewHub constructs an empty Hub.
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		byID:   make(map[string]*client),
		byUser: make(map[string]map[string]*client),
		logger: logger.With(zap.String("module", "ws-hub")),
		nowFn:  time.Now,
	}
}

// Register attaches conn to username with the given client ID. Returns the
// client handle so the caller can drive the read-pump.
func (h *Hub) Register(clientID, username string, conn Conn) *client {
	c := &client{
		id:       clientID,
		username: username,
		conn:     conn,
		joinedAt: h.nowFn().UnixMilli(),
	}
	c.lastPong.Store(h.nowFn().UnixMilli())

	h.mu.Lock()
	h.byID[clientID] = c
	bucket, ok := h.byUser[username]
	if !ok {
		bucket = make(map[string]*client)
		h.byUser[username] = bucket
	}
	bucket[clientID] = c
	h.mu.Unlock()

	h.logger.Info("client registered",
		zap.String("clientID", clientID),
		zap.String("username", username))

	if h.OnPresenceChange != nil {
		h.OnPresenceChange()
	}
	return c
}

// Unregister removes a client and closes its underlying connection.
func (h *Hub) Unregister(clientID string) {
	h.mu.Lock()
	c, ok := h.byID[clientID]
	if !ok {
		h.mu.Unlock()
		return
	}
	delete(h.byID, clientID)
	if bucket, ok := h.byUser[c.username]; ok {
		delete(bucket, clientID)
		if len(bucket) == 0 {
			delete(h.byUser, c.username)
		}
	}
	h.mu.Unlock()

	c.close()

	h.logger.Info("client unregistered",
		zap.String("clientID", clientID),
		zap.String("username", c.username))

	if h.OnPresenceChange != nil {
		h.OnPresenceChange()
	}
}

// SendToUser delivers a frame to every connection of the given username.
// Returns the number of successful deliveries.
func (h *Hub) SendToUser(username string, frame Frame) int {
	h.mu.RLock()
	bucket := h.byUser[username]
	clients := make([]*client, 0, len(bucket))
	for _, c := range bucket {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	frame = h.stampFrame(frame)
	delivered := 0
	for _, c := range clients {
		if err := c.send(frame); err != nil {
			h.logger.Warn("send-to-user failed",
				zap.String("username", username),
				zap.String("clientID", c.id),
				zap.Error(err))
			continue
		}
		delivered++
	}
	return delivered
}

// Broadcast delivers a frame to every connected client.
func (h *Hub) Broadcast(frame Frame) int {
	h.mu.RLock()
	clients := make([]*client, 0, len(h.byID))
	for _, c := range h.byID {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	frame = h.stampFrame(frame)
	delivered := 0
	for _, c := range clients {
		if err := c.send(frame); err != nil {
			h.logger.Warn("broadcast failed",
				zap.String("clientID", c.id),
				zap.Error(err))
			continue
		}
		delivered++
	}
	return delivered
}

func (h *Hub) stampFrame(f Frame) Frame {
	if f.Ts == 0 {
		f.Ts = h.nowFn().UnixMilli()
	}
	if f.ID == "" {
		f.ID = h.nextMessageID()
	}
	return f
}

func (h *Hub) nextMessageID() string {
	n := h.msgSeq.Add(1)
	// cheap, URL-safe.
	return "m-" + strconvFormatUint(n)
}

// OnlineUsers returns a snapshot of who is connected.
func (h *Hub) OnlineUsers() []OnlineUser {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]OnlineUser, 0, len(h.byUser))
	for name, bucket := range h.byUser {
		earliest := int64(0)
		for _, c := range bucket {
			if earliest == 0 || c.joinedAt < earliest {
				earliest = c.joinedAt
			}
		}
		out = append(out, OnlineUser{
			Username:     name,
			SessionCount: len(bucket),
			LoginTime:    earliest,
		})
	}
	return out
}

// OnlineCount returns distinct online usernames.
func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.byUser)
}

// SessionCount returns total active connections (may exceed user count).
func (h *Hub) SessionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.byID)
}

// IsOnline reports whether any connection exists for username.
func (h *Hub) IsOnline(username string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.byUser[username]) > 0
}

// MarkPong is called by the read-pump whenever the client acknowledges a ping
// (either via a WS-level pong or an application-level `{"type":"pong"}` frame).
func (h *Hub) MarkPong(clientID string) {
	h.mu.RLock()
	c := h.byID[clientID]
	h.mu.RUnlock()
	if c == nil {
		return
	}
	c.lastPong.Store(h.nowFn().UnixMilli())
}

// RunHeartbeat drives the server-side ping loop. Each tick it sends a ping
// frame; any client whose last pong is older than readDeadline is forcibly
// unregistered. Intended to be called in its own goroutine; exits when done
// is closed.
func (h *Hub) RunHeartbeat(done <-chan struct{}) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	ping := MustFrame(FrameTypePing, nil)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			h.sweepAndPing(ping)
		}
	}
}

func (h *Hub) sweepAndPing(ping Frame) {
	now := h.nowFn().UnixMilli()

	h.mu.RLock()
	clients := make([]*client, 0, len(h.byID))
	for _, c := range h.byID {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, c := range clients {
		if now-c.lastPong.Load() > readDeadline.Milliseconds() {
			h.logger.Info("client timed out, dropping",
				zap.String("clientID", c.id),
				zap.String("username", c.username))
			h.Unregister(c.id)
			continue
		}
		if err := c.send(ping); err != nil {
			h.logger.Warn("ping failed, dropping",
				zap.String("clientID", c.id),
				zap.Error(err))
			h.Unregister(c.id)
		}
	}
}

// strconvFormatUint is inlined to avoid pulling in fmt for a hot path.
func strconvFormatUint(u uint64) string {
	if u == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for u > 0 {
		i--
		buf[i] = byte('0' + u%10)
		u /= 10
	}
	return string(buf[i:])
}
