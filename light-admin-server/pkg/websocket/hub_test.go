package websocket

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
)

// fakeConn implements the Conn interface for tests.
type fakeConn struct {
	mu        sync.Mutex
	written   [][]byte
	writeErr  error
	closed    atomic.Bool
	onWrite   func([]byte)
	deadline  time.Time
}

func (f *fakeConn) WriteMessage(_ int, data []byte) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.mu.Lock()
	dup := make([]byte, len(data))
	copy(dup, data)
	f.written = append(f.written, dup)
	f.mu.Unlock()
	if f.onWrite != nil {
		f.onWrite(dup)
	}
	return nil
}

func (f *fakeConn) Close() error {
	f.closed.Store(true)
	return nil
}

func (f *fakeConn) SetWriteDeadline(t time.Time) error {
	f.deadline = t
	return nil
}

func (f *fakeConn) frames(t *testing.T) []Frame {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]Frame, 0, len(f.written))
	for _, raw := range f.written {
		var fr Frame
		if err := json.Unmarshal(raw, &fr); err != nil {
			t.Fatalf("unmarshal frame: %v (raw=%s)", err, raw)
		}
		out = append(out, fr)
	}
	return out
}

func newTestHub() *Hub {
	return NewHub(zap.NewNop())
}

func TestHub_RegisterUnregister(t *testing.T) {
	h := newTestHub()
	var presence atomic.Int32
	h.OnPresenceChange = func() { presence.Add(1) }

	c1 := &fakeConn{}
	h.Register("c1", "alice", c1)
	if got := h.OnlineCount(); got != 1 {
		t.Fatalf("OnlineCount = %d, want 1", got)
	}
	if !h.IsOnline("alice") {
		t.Fatalf("alice should be online")
	}
	if presence.Load() != 1 {
		t.Fatalf("presence callback fired %d times, want 1", presence.Load())
	}

	// Same user, second session.
	c2 := &fakeConn{}
	h.Register("c2", "alice", c2)
	if got := h.SessionCount(); got != 2 {
		t.Fatalf("SessionCount = %d, want 2", got)
	}
	if got := h.OnlineCount(); got != 1 {
		t.Fatalf("OnlineCount = %d, want 1 (dedup by username)", got)
	}

	h.Unregister("c1")
	if !c1.closed.Load() {
		t.Fatalf("c1 conn should be closed after Unregister")
	}
	if got := h.SessionCount(); got != 1 {
		t.Fatalf("SessionCount after 1 unregister = %d, want 1", got)
	}
	if !h.IsOnline("alice") {
		t.Fatalf("alice should still be online via c2")
	}

	h.Unregister("c2")
	if h.IsOnline("alice") {
		t.Fatalf("alice should be offline after all sessions closed")
	}
}

func TestHub_SendToUser(t *testing.T) {
	h := newTestHub()
	a1, a2, b1 := &fakeConn{}, &fakeConn{}, &fakeConn{}
	h.Register("a1", "alice", a1)
	h.Register("a2", "alice", a2)
	h.Register("b1", "bob", b1)

	frame := MustFrame(FrameTypeNotice, map[string]string{"body": "hi"})
	n := h.SendToUser("alice", frame)
	if n != 2 {
		t.Fatalf("SendToUser delivered %d, want 2", n)
	}

	for _, c := range []*fakeConn{a1, a2} {
		frames := c.frames(t)
		if len(frames) != 1 {
			t.Fatalf("alice conn: got %d frames, want 1", len(frames))
		}
		if frames[0].Type != FrameTypeNotice {
			t.Fatalf("alice frame type = %q, want %q", frames[0].Type, FrameTypeNotice)
		}
		if frames[0].Ts == 0 {
			t.Fatalf("alice frame missing Ts")
		}
		if frames[0].ID == "" {
			t.Fatalf("alice frame missing ID")
		}
	}
	if len(b1.frames(t)) != 0 {
		t.Fatalf("bob should not receive alice-only message")
	}

	// Sending to an unknown user is a no-op.
	if n := h.SendToUser("nobody", frame); n != 0 {
		t.Fatalf("SendToUser(unknown) = %d, want 0", n)
	}
}

func TestHub_Broadcast(t *testing.T) {
	h := newTestHub()
	conns := []*fakeConn{{}, {}, {}}
	for i, c := range conns {
		h.Register(string(rune('a'+i)), "u"+string(rune('a'+i)), c)
	}

	frame := MustFrame(FrameTypeSystem, "announcement")
	n := h.Broadcast(frame)
	if n != 3 {
		t.Fatalf("Broadcast delivered %d, want 3", n)
	}
	for i, c := range conns {
		if len(c.frames(t)) != 1 {
			t.Fatalf("conn %d: got %d frames, want 1", i, len(c.frames(t)))
		}
	}
}

func TestHub_HeartbeatDisconnects(t *testing.T) {
	h := newTestHub()
	// Freeze time so lastPong is ancient.
	fixed := time.Unix(1_700_000_000, 0)
	h.nowFn = func() time.Time { return fixed }

	c := &fakeConn{}
	client := h.Register("c1", "alice", c)
	// Force last pong far in the past.
	client.lastPong.Store(fixed.Add(-2 * readDeadline).UnixMilli())

	h.sweepAndPing(MustFrame(FrameTypePing, nil))

	if h.IsOnline("alice") {
		t.Fatalf("alice should have been dropped by heartbeat sweep")
	}
	if !c.closed.Load() {
		t.Fatalf("conn should be closed after sweep eviction")
	}
}

func TestHub_HeartbeatKeepsLiveClients(t *testing.T) {
	h := newTestHub()
	fixed := time.Unix(1_700_000_000, 0)
	h.nowFn = func() time.Time { return fixed }

	c := &fakeConn{}
	h.Register("c1", "alice", c)

	h.sweepAndPing(MustFrame(FrameTypePing, nil))

	if !h.IsOnline("alice") {
		t.Fatalf("alice should stay online (pong was fresh)")
	}
	frames := c.frames(t)
	if len(frames) != 1 || frames[0].Type != FrameTypePing {
		t.Fatalf("expected one ping frame, got %+v", frames)
	}
}

func TestHub_SendSerializesPerConn(t *testing.T) {
	h := newTestHub()
	c := &fakeConn{}
	h.Register("c1", "alice", c)

	var wg sync.WaitGroup
	const N = 50
	wg.Add(N)
	for i := range N {
		go func(i int) {
			defer wg.Done()
			h.SendToUser("alice", MustFrame(FrameTypeMessage, i))
		}(i)
	}
	wg.Wait()

	if len(c.frames(t)) != N {
		t.Fatalf("got %d frames, want %d", len(c.frames(t)), N)
	}
}
