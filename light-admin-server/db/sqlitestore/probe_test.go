package sqlitestore

// Opt-in smoke probes against a real SQLite database file produced by
// `migrate` + `setup` (multi-database-plan §7 端到端冒烟). Skipped unless
// SMOKE_DB points at the database file:
//
//	SMOKE_DB=path/to/app.db go test ./db/sqlitestore -run TestProbe -v

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/top-system/light-admin/db/store"
)

func TestProbeKeywords(t *testing.T) {
	path := os.Getenv("SMOKE_DB")
	if path == "" {
		t.Skip("SMOKE_DB not set")
	}
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := New(db)
	ctx := context.Background()

	all, err := s.ListUsers(ctx, store.ListUsersParams{})
	if err != nil {
		t.Fatalf("ListUsers(nil filters): %v", err)
	}
	t.Logf("all users: %d", len(all))
	for _, u := range all {
		t.Logf("  user=%s create_time=%v", u.Username, u.CreateTime)
	}

	kw := "%admin%"
	filtered, err := s.ListUsers(ctx, store.ListUsersParams{Keywords: &kw})
	if err != nil {
		t.Fatalf("ListUsers(keywords): %v", err)
	}
	t.Logf("keyword-filtered users: %d", len(filtered))

	n, err := s.CountUsers(ctx, store.CountUsersParams{Keywords: &kw})
	if err != nil {
		t.Fatalf("CountUsers(keywords): %v", err)
	}
	t.Logf("keyword count: %d", n)
}

func TestProbeQueueTaskLifecycle(t *testing.T) {
	path := os.Getenv("SMOKE_DB")
	if path == "" {
		t.Skip("SMOKE_DB not set")
	}
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := New(db)
	ctx := context.Background()
	now := time.Now()

	created, err := s.CreateQueueTask(ctx, store.CreateQueueTaskParams{
		Type: "probe_task", Status: "queued", CorrelationID: "probe-1",
		Now: now,
	})
	if err != nil {
		t.Fatalf("CreateQueueTask: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("RETURNING did not yield an id")
	}
	t.Logf("created id=%d created_at=%v", created.ID, created.CreatedAt)

	updated, err := s.UpdateQueueTask(ctx, store.UpdateQueueTaskParams{
		ID: created.ID, Type: "probe_task", Status: "processing", CorrelationID: "probe-1",
		PublicRetryCount: 1, Now: now.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("UpdateQueueTask: %v", err)
	}
	if updated.Status != "processing" || updated.PublicRetryCount != 1 {
		t.Fatalf("update not reflected: %+v", updated)
	}

	pending, err := s.ListPendingQueueTasks(ctx, []string{"probe_task"})
	if err != nil {
		t.Fatalf("ListPendingQueueTasks: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending probe_task, got %d", len(pending))
	}
	none, err := s.ListPendingQueueTasks(ctx, []string{"other_type"})
	if err != nil || len(none) != 0 {
		t.Fatalf("type filter leak: %d, %v", len(none), err)
	}
	all, err := s.ListPendingQueueTasks(ctx, nil)
	if err != nil || len(all) < 1 {
		t.Fatalf("nil filter should list all pending: %d, %v", len(all), err)
	}

	if err := s.SoftDeleteQueueTask(ctx, store.SoftDeleteQueueTaskParams{ID: created.ID, Now: now.Add(2 * time.Second)}); err != nil {
		t.Fatalf("SoftDeleteQueueTask: %v", err)
	}
	if _, err := s.GetQueueTask(ctx, created.ID); !errors.Is(err, store.ErrNoRows) {
		t.Fatalf("expected store.ErrNoRows after soft delete, got %v", err)
	}
	// The admin view ignores deleted_at and hard-deletes.
	if _, err := s.GetTask(ctx, created.ID); err != nil {
		t.Fatalf("admin GetTask should still see the row: %v", err)
	}
	if err := s.DeleteTask(ctx, created.ID); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	t.Log("queue task lifecycle OK")
}
