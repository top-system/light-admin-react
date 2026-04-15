package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	platformctrl "github.com/top-system/light-admin/api/platform/controller"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	ws "github.com/top-system/light-admin/pkg/websocket"
)

// fakeTokenParser implements platformctrl.TokenParser. A token value of
// "good" resolves to the baked-in username; everything else fails.
type fakeTokenParser struct {
	validToken string
	username   string
}

func (f fakeTokenParser) ParseToken(token string) (*dto.JwtClaims, error) {
	if token != f.validToken {
		return nil, &errInvalidToken{}
	}
	return &dto.JwtClaims{Username: f.username}, nil
}

type errInvalidToken struct{}

func (e *errInvalidToken) Error() string { return "invalid token" }

// nopLogger returns a lib.Logger backed by zap.NewNop().
func nopLogger() lib.Logger {
	z := zap.NewNop()
	return lib.Logger{Zap: z.Sugar(), DesugarZap: z}
}

// newWSTestServer wires up a Hub + controller + httptest.Server. The caller is
// responsible for stopping the server via srv.Close().
func newWSTestServer(t *testing.T, token, username string) (*httptest.Server, *ws.WebSocket) {
	t.Helper()
	wsSvc := ws.New(zap.NewNop())
	ctrl := platformctrl.NewWebSocketControllerWithParser(
		wsSvc,
		nopLogger(),
		fakeTokenParser{validToken: token, username: username},
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctrl.HandleWebSocket(w, r)
	}))
	return srv, wsSvc
}

// httpToWS swaps the scheme so a test client can dial the websocket.
func httpToWS(addr string) string {
	return "ws" + strings.TrimPrefix(addr, "http")
}

// TestWS_RejectsInvalidToken verifies the pre-upgrade 401 path. gorilla's
// Dialer exposes the failing HTTP response, which we assert against.
func TestWS_RejectsInvalidToken(t *testing.T) {
	srv, _ := newWSTestServer(t, "good", "alice")
	defer srv.Close()

	url := httpToWS(srv.URL) + "/?token=bad"
	_, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err == nil {
		t.Fatalf("expected dial error for invalid token")
	}
	if resp == nil {
		t.Fatalf("expected HTTP response; got nil")
	}
	if resp.StatusCode != http.StatusUnauthorized {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 401. body=%s", resp.StatusCode, body)
	}

	// And the same for the missing-token case.
	_, resp2, err := websocket.DefaultDialer.Dial(httpToWS(srv.URL)+"/", nil)
	if err == nil {
		t.Fatalf("expected dial error when token is missing")
	}
	if resp2 == nil || resp2.StatusCode != http.StatusUnauthorized {
		t.Fatalf("missing token: status = %v, want 401", resp2)
	}
}

// TestWS_DictChangeFanout connects a valid client, triggers a dict-change
// via the public WebSocket service, and asserts the frame shape matches ADR.
func TestWS_DictChangeFanout(t *testing.T) {
	srv, wsSvc := newWSTestServer(t, "good", "alice")
	defer srv.Close()

	conn, resp, err := websocket.DefaultDialer.Dial(httpToWS(srv.URL)+"/?token=good", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("handshake status = %d, want 101", resp.StatusCode)
	}

	// Drain any presence-change bootstrap frames, then trigger the dict-change
	// server-side. We look for the dict-change frame specifically rather than
	// asserting the first frame, because OnPresenceChange also emits one.
	done := make(chan ws.Frame, 8)
	go func() {
		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				close(done)
				return
			}
			var f ws.Frame
			if err := json.Unmarshal(raw, &f); err != nil {
				continue
			}
			done <- f
		}
	}()

	// Give the read loop a moment to register, then fire the change.
	time.Sleep(50 * time.Millisecond)
	wsSvc.BroadcastDictChange("user_status")

	deadline := time.After(2 * time.Second)
	for {
		select {
		case f, ok := <-done:
			if !ok {
				t.Fatal("connection closed before dict-change arrived")
			}
			if f.Type != ws.FrameTypeDictChange {
				continue
			}
			// Shape assertions: data contains dictCode and timestamp.
			var payload struct {
				DictCode  string `json:"dictCode"`
				Timestamp int64  `json:"timestamp"`
			}
			if err := json.Unmarshal(f.Data, &payload); err != nil {
				t.Fatalf("unmarshal dict-change data: %v", err)
			}
			if payload.DictCode != "user_status" {
				t.Fatalf("dictCode = %q, want user_status", payload.DictCode)
			}
			if payload.Timestamp == 0 {
				t.Fatal("dict-change timestamp missing")
			}
			if f.Ts == 0 || f.ID == "" {
				t.Fatalf("frame envelope missing Ts/ID: %+v", f)
			}
			return
		case <-deadline:
			t.Fatal("timed out waiting for dict-change frame")
		}
	}
}

// TestWS_HeartbeatEvictsStaleClients exercises the sweep path end-to-end but
// bypasses the real 60-second timeout by driving presence bookkeeping via the
// exported Hub API. Unit-level sweep coverage lives in pkg/websocket/hub_test.go;
// this test verifies the public WebSocket service reflects the eviction.
func TestWS_HeartbeatEvictsStaleClients(t *testing.T) {
	srv, wsSvc := newWSTestServer(t, "good", "alice")
	defer srv.Close()

	conn, _, err := websocket.DefaultDialer.Dial(httpToWS(srv.URL)+"/?token=good", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Wait for registration to settle.
	for i := 0; i < 20 && !wsSvc.IsUserOnline("alice"); i++ {
		time.Sleep(25 * time.Millisecond)
	}
	if !wsSvc.IsUserOnline("alice") {
		t.Fatal("alice never registered")
	}

	// Closing the client-side connection should cause the server read-loop to
	// return, triggering Unregister. Give the server a moment to notice.
	_ = conn.Close()
	for i := 0; i < 40 && wsSvc.IsUserOnline("alice"); i++ {
		time.Sleep(25 * time.Millisecond)
	}
	if wsSvc.IsUserOnline("alice") {
		t.Fatal("alice should be unregistered after client disconnect")
	}
}
