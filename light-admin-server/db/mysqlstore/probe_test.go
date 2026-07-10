package mysqlstore

// Opt-in smoke probes against a real MySQL database prepared by `migrate` +
// `setup` (multi-database-plan §7 端到端冒烟). Skipped unless SMOKE_MYSQL_DSN
// points at the database:
//
//	SMOKE_MYSQL_DSN='root:pass@tcp(host:port)/dbname?parseTime=true&loc=Asia%2FShanghai' \
//	  go test ./db/mysqlstore -run TestProbe -v

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/top-system/light-admin/db/store"
)

func TestProbeQueueTaskLifecycle(t *testing.T) {
	dsn := os.Getenv("SMOKE_MYSQL_DSN")
	if dsn == "" {
		t.Skip("SMOKE_MYSQL_DSN not set")
	}
	db, err := sql.Open("mysql", dsn)
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
		t.Fatal("execlastid did not yield an id")
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
