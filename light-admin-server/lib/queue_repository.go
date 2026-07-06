package lib

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/top-system/light-admin/db/sqlc"
	"github.com/top-system/light-admin/pkg/queue"
)

// queueTaskRepository is the sqlc/pgx-backed implementation of
// queue.TaskRepository. It lives in the application layer so pkg/queue stays
// framework-agnostic (it depends only on the interface); the implementation is
// injected into the queue engine via ExtrasModule.
//
// It maps queue.TaskModel <-> the sys_tasks row and preserves the previous GORM
// behaviour: reads and writes ignore soft-deleted rows and Delete is a soft delete.
type queueTaskRepository struct {
	q *sqlc.Queries
}

// NewQueueTaskRepository builds the sqlc-backed queue.TaskRepository from the
// pool-bound sqlc Queries. It is registered as an fx provider and consumed by
// both the queue engine (ExtrasModule) and services that read task state.
func NewQueueTaskRepository(q *sqlc.Queries) queue.TaskRepository {
	return &queueTaskRepository{q: q}
}

// Create inserts a new task and back-fills the generated ID and timestamps onto
// the model so the engine can transition it to the persisted state.
func (r *queueTaskRepository) Create(ctx context.Context, task *queue.TaskModel) error {
	row, err := r.q.CreateQueueTask(ctx, sqlc.CreateQueueTaskParams{
		Type:                   task.Type,
		Status:                 string(task.Status),
		CorrelationID:          task.CorrelationID.String(),
		OwnerID:                task.OwnerID,
		PrivateState:           task.PrivateState,
		PublicRetryCount:       int32(task.PublicState.RetryCount),
		PublicExecutedDuration: int64(task.PublicState.ExecutedDuration),
		PublicError:            task.PublicState.Error,
		PublicErrorHistory:     marshalErrorHistory(task.PublicState.ErrorHistory),
		PublicResumeTime:       task.PublicState.ResumeTime,
	})
	if err != nil {
		return err
	}
	task.ID = uint64(row.ID)
	task.CreatedAt = row.CreatedAt.Time
	task.UpdatedAt = row.UpdatedAt.Time
	return nil
}

// Update writes the current model state back to the (live) row.
func (r *queueTaskRepository) Update(ctx context.Context, task *queue.TaskModel) error {
	row, err := r.q.UpdateQueueTask(ctx, sqlc.UpdateQueueTaskParams{
		ID:                     int64(task.ID),
		Type:                   task.Type,
		Status:                 string(task.Status),
		CorrelationID:          task.CorrelationID.String(),
		OwnerID:                task.OwnerID,
		PrivateState:           task.PrivateState,
		PublicRetryCount:       int32(task.PublicState.RetryCount),
		PublicExecutedDuration: int64(task.PublicState.ExecutedDuration),
		PublicError:            task.PublicState.Error,
		PublicErrorHistory:     marshalErrorHistory(task.PublicState.ErrorHistory),
		PublicResumeTime:       task.PublicState.ResumeTime,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return queue.ErrTaskNotFound
		}
		return err
	}
	task.UpdatedAt = row.UpdatedAt.Time
	return nil
}

// GetByID returns the live task with the given ID, or queue.ErrTaskNotFound.
func (r *queueTaskRepository) GetByID(ctx context.Context, id uint64) (*queue.TaskModel, error) {
	row, err := r.q.GetQueueTask(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, queue.ErrTaskNotFound
		}
		return nil, err
	}
	return toQueueModel(row), nil
}

// GetPendingTasks returns all live tasks in a pending status, optionally filtered
// by type.
func (r *queueTaskRepository) GetPendingTasks(ctx context.Context, types ...string) ([]*queue.TaskModel, error) {
	var filter []string
	if len(types) > 0 {
		filter = types
	}
	rows, err := r.q.ListPendingQueueTasks(ctx, filter)
	if err != nil {
		return nil, err
	}
	tasks := make([]*queue.TaskModel, 0, len(rows))
	for _, row := range rows {
		tasks = append(tasks, toQueueModel(row))
	}
	return tasks, nil
}

// Delete soft-deletes the task (sets deleted_at), matching the previous engine
// behaviour.
func (r *queueTaskRepository) Delete(ctx context.Context, id uint64) error {
	return r.q.SoftDeleteQueueTask(ctx, int64(id))
}

// toQueueModel maps a sys_tasks row to the framework-agnostic queue.TaskModel.
func toQueueModel(row sqlc.SysTask) *queue.TaskModel {
	var history queue.StringSlice
	_ = history.Scan(row.PublicErrorHistory)

	return &queue.TaskModel{
		ID:            uint64(row.ID),
		Type:          row.Type,
		Status:        queue.Status(row.Status),
		CorrelationID: uuid.FromStringOrNil(row.CorrelationID),
		OwnerID:       row.OwnerID,
		PrivateState:  row.PrivateState,
		PublicState: queue.TaskPublicState{
			RetryCount:       int(row.PublicRetryCount),
			ExecutedDuration: time.Duration(row.PublicExecutedDuration),
			Error:            row.PublicError,
			ErrorHistory:     history,
			ResumeTime:       row.PublicResumeTime,
		},
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

// marshalErrorHistory renders the StringSlice to its stored string form, reusing
// the type's driver.Valuer so the on-disk representation is unchanged.
func marshalErrorHistory(s queue.StringSlice) string {
	v, err := s.Value()
	if err != nil {
		return "[]"
	}
	if str, ok := v.(string); ok {
		return str
	}
	return "[]"
}
