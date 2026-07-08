package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"github.com/top-system/light-admin/api/system/service"
	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/lib"
	dto "github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/pkg/echox"
	"github.com/top-system/light-admin/pkg/uuid"
	ws "github.com/top-system/light-admin/pkg/websocket"
)

// TokenParser is the narrow surface the WebSocket controller needs from the
// auth service. Defined here so tests can plug in an in-memory verifier
// without dragging in a Cache / DB.
type TokenParser interface {
	ParseToken(token string) (*dto.JwtClaims, error)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(_ *http.Request) bool {
		return true // production should restrict origins
	},
	// No STOMP subprotocols — plain JSON over the default subprotocol.
}

// WebSocketController handles the /ws upgrade and the HTTP fan-out endpoints.
type WebSocketController struct {
	ws          *ws.WebSocket
	logger      lib.Logger
	authService TokenParser
}

// NewWebSocketController creates the controller and starts the heartbeat loop.
func NewWebSocketController(
	websocket *ws.WebSocket,
	logger lib.Logger,
	authService service.AuthService,
) WebSocketController {
	return NewWebSocketControllerWithParser(websocket, logger, authService)
}

// NewWebSocketControllerWithParser is the test-friendly constructor; it accepts the
// narrower TokenParser interface. Production code should continue to use
// NewWebSocketController so FX wiring resolves.
func NewWebSocketControllerWithParser(
	websocket *ws.WebSocket,
	logger lib.Logger,
	authService TokenParser,
) WebSocketController {
	ctrl := WebSocketController{
		ws:          websocket,
		logger:      logger,
		authService: authService,
	}
	// Heartbeat runs for the lifetime of the process. Not tied to lifecycle
	// hooks because the hub has no persistent resources to flush on shutdown.
	go ctrl.ws.Hub.RunHeartbeat(make(chan struct{}))
	return ctrl
}

// HandleWebSocket is the raw http.Handler installed ahead of the Echo engine.
// Authentication happens pre-upgrade via the `token` query parameter; any
// invalid token produces a 401 without touching the upgrade machinery.
func (c WebSocketController) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	c.logger.Info(fmt.Sprintf("WebSocket upgrade request from: %s", r.RemoteAddr))

	token := r.URL.Query().Get("token")
	if token == "" {
		if h := r.Header.Get("Authorization"); len(h) > 7 && h[:7] == "Bearer " {
			token = h[7:]
		}
	}
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	claims, err := c.authService.ParseToken(token)
	if err != nil {
		c.logger.Warn(fmt.Sprintf("WebSocket auth failed: %v", err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.logger.Error(fmt.Sprintf("Failed to upgrade to websocket: %v", err))
		return
	}

	clientID := uuid.MustString()
	client := c.ws.Hub.Register(clientID, claims.Username, conn)
	_ = client // client handle not needed outside the read loop

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.ws.Hub.MarkPong(clientID)
		return nil
	})

	c.readLoop(clientID, claims.Username, conn)
}

// readLoop consumes frames from a single client until the connection dies.
// The plain protocol currently only accepts `pong` frames from the client;
// everything else is ignored but logged. The HTTP `sendToAll` / `sendToUser`
// endpoints remain the canonical way for authenticated users to publish.
func (c WebSocketController) readLoop(clientID, username string, conn *websocket.Conn) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(fmt.Sprintf("panic in ws readLoop: %v", r))
		}
		c.ws.Hub.Unregister(clientID)
		c.logger.Info(fmt.Sprintf("WebSocket disconnected: user=%s, client=%s", username, clientID))
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			c.logger.Debug(fmt.Sprintf("ws ReadMessage error: %v", err))
			return
		}
		if len(message) == 0 {
			continue
		}
		var frame ws.Frame
		if err := json.Unmarshal(message, &frame); err != nil {
			c.logger.Warn(fmt.Sprintf("ws malformed frame from %s: %v", username, err))
			continue
		}
		switch frame.Type {
		case ws.FrameTypePong:
			c.ws.Hub.MarkPong(clientID)
		case ws.FrameTypePing:
			// Client-initiated ping: reply with a pong so the keep-alive
			// works even behind aggressive proxies that eat WS control frames.
			_ = conn.WriteMessage(websocket.TextMessage, mustJSON(ws.MustFrame(ws.FrameTypePong, nil)))
		default:
			c.logger.Debug(fmt.Sprintf("ws ignoring frame type=%s from %s", frame.Type, username))
		}
	}
}

func mustJSON(v any) []byte {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return raw
}

// Connect is the Echo-bound variant preserved for route wiring. In practice
// requests to /ws are intercepted before Echo in bootstrap.go; this exists so
// the route file still compiles and appears in Swagger.
//
//	@tags			WebSocket
//	@summary		WebSocket Connect
//	@description	Upgrade to a plain JSON WebSocket. Authenticate via ?token=<jwt>.
//	@produce		json
//	@success		101	"Switching Protocols"
//	@failure		401	{object}	echox.Response	"unauthorized"
//	@failure		400	{object}	echox.Response	"bad request"
//	@router			/ws [get]
func (c WebSocketController) Connect(ctx echo.Context) error {
	c.HandleWebSocket(ctx.Response().Writer, ctx.Request())
	return nil
}

// === HTTP fan-out endpoints =================================================

// SendToAllRequest is the body for POST /websocket/sendToAll.
type SendToAllRequest struct {
	Message string `json:"message" validate:"required"`
}

// SendToAll broadcasts a notice to every connected client.
//
//	@tags			WebSocket
//	@summary		Send message to all users
//	@accept			json
//	@produce		json
//	@param			body	body		SendToAllRequest	true	"Message"
//	@success		200		{object}	echox.Response		"ok"
//	@failure		400		{object}	echox.Response		"bad request"
//	@router			/api/v1/websocket/sendToAll [post]
func (c WebSocketController) SendToAll(ctx echo.Context) error {
	var req SendToAllRequest
	if err := ctx.Bind(&req); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	c.ws.BroadcastNotice(req.Message)
	return echox.Response{Code: http.StatusOK, Message: "Message sent to all users"}.JSON(ctx)
}

// SendToUserRequest is the body for POST /websocket/sendToUser.
type SendToUserRequest struct {
	Username string `json:"username" validate:"required"`
	Message  string `json:"message"  validate:"required"`
}

// SendToUser delivers a direct message to a specific user's connections.
//
//	@tags			WebSocket
//	@summary		Send message to specific user
//	@accept			json
//	@produce		json
//	@param			body	body		SendToUserRequest	true	"Message"
//	@success		200		{object}	echox.Response		"ok"
//	@failure		400		{object}	echox.Response		"bad request"
//	@router			/api/v1/websocket/sendToUser [post]
func (c WebSocketController) SendToUser(ctx echo.Context) error {
	var req SendToUserRequest
	if err := ctx.Bind(&req); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	senderName := "System"
	if sender := ctx.Get(constants.CurrentUser); sender != nil {
		if name, ok := sender.(string); ok {
			senderName = name
		}
	}
	c.ws.SendToUser(senderName, req.Username, req.Message)
	return echox.Response{Code: http.StatusOK, Message: "Message sent to user"}.JSON(ctx)
}

// GetOnlineUsers returns the presence snapshot.
//
//	@tags			WebSocket
//	@summary		Get online users
//	@produce		json
//	@success		200	{object}	echox.Response{data=[]ws.OnlineUser}	"ok"
//	@router			/api/v1/websocket/online-users [get]
func (c WebSocketController) GetOnlineUsers(ctx echo.Context) error {
	return echox.Response{Code: http.StatusOK, Data: c.ws.GetOnlineUsers()}.JSON(ctx)
}

// GetOnlineCount returns the distinct-user count.
//
//	@tags			WebSocket
//	@summary		Get online user count
//	@produce		json
//	@success		200	{object}	echox.Response{data=int}	"ok"
//	@router			/api/v1/websocket/online-count [get]
func (c WebSocketController) GetOnlineCount(ctx echo.Context) error {
	return echox.Response{Code: http.StatusOK, Data: c.ws.GetOnlineUserCount()}.JSON(ctx)
}

// BroadcastDictChangeRequest is the body for POST /websocket/dict-change.
type BroadcastDictChangeRequest struct {
	DictCode string `json:"dictCode" validate:"required"`
}

// BroadcastDictChange is an HTTP entry point used by the dict controller to
// fan out invalidation notifications.
//
//	@tags			WebSocket
//	@summary		Broadcast dict change
//	@accept			json
//	@produce		json
//	@param			body	body		BroadcastDictChangeRequest	true	"Dict code"
//	@success		200		{object}	echox.Response				"ok"
//	@router			/api/v1/websocket/dict-change [post]
func (c WebSocketController) BroadcastDictChange(ctx echo.Context) error {
	var req BroadcastDictChangeRequest
	if err := ctx.Bind(&req); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	c.ws.BroadcastDictChange(req.DictCode)
	return echox.Response{Code: http.StatusOK, Message: "Dict change notification sent"}.JSON(ctx)
}
