package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/top-system/light-admin/db/sqlc"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/pkg/queue"
)

// TaskRepository is the sqlc/pgx-backed persistence layer for the admin Job module.
//
// It reads and hard-deletes rows in sys_tasks. This is deliberately distinct from
// the pkg/queue engine's own TaskRepository (which owns task lifecycle writes and
// uses soft-delete): the admin side neither filters nor sets deleted_at, mirroring
// the previous models/system.Task GORM behaviour (that model has no DeletedAt).
type TaskRepository struct {
	q      sqlc.Querier
	logger lib.Logger
}

// NewTaskRepository creates a new task repository bound to the pool-level Queries.
func NewTaskRepository(q *sqlc.Queries, logger lib.Logger) TaskRepository {
	return TaskRepository{q: q, logger: logger}
}

// WithTx returns a copy bound to the given transaction-scoped Queries.
func (a TaskRepository) WithTx(q *sqlc.Queries) TaskRepository {
	a.q = q
	return a
}

// Query 查询任务列表. Ordering is fixed to created_at DESC.
func (a TaskRepository) Query(param *system.TaskQueryParam) (*system.TaskQueryResult, error) {
	ctx := context.Background()

	var taskType, status, correlationID, keywords *string
	if param.Type != "" {
		taskType = ptr(param.Type)
	}
	if param.Status != "" {
		status = ptr(param.Status)
	}
	if param.CorrelationID != "" {
		correlationID = ptr(param.CorrelationID)
	}
	if param.Keywords != "" {
		keywords = ptr("%" + param.Keywords + "%")
	}
	createFrom := startOfDayFilter(param.CreateTimeFrom)
	createTo := endOfDayFilter(param.CreateTimeTo)

	total, err := a.q.CountTasks(ctx, sqlc.CountTasksParams{
		Type:          taskType,
		Status:        status,
		CorrelationID: correlationID,
		Keywords:      keywords,
		CreateFrom:    createFrom,
		CreateTo:      createTo,
	})
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.Tasks, 0)
	if total > 0 {
		limit, offset := pageBounds(param.PaginationParam)
		rows, err := a.q.ListTasks(ctx, sqlc.ListTasksParams{
			Type:          taskType,
			Status:        status,
			CorrelationID: correlationID,
			Keywords:      keywords,
			CreateFrom:    createFrom,
			CreateTo:      createTo,
			Limit:         limit,
			Offset:        offset,
		})
		if err != nil {
			return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
		}
		for _, r := range rows {
			list = append(list, toDomainTask(r))
		}
	}

	return &system.TaskQueryResult{
		Pagination: &dto.Pagination{
			PageNum:  param.PaginationParam.GetPageNum(),
			PageSize: param.PaginationParam.GetPageSize(),
			Total:    total,
		},
		List: list,
	}, nil
}

// Get 获取任务详情
func (a TaskRepository) Get(id uint64) (*system.Task, error) {
	row, err := a.q.GetTask(context.Background(), int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.DatabaseRecordNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainTask(row), nil
}

// Delete 删除任务 (hard delete, matching the previous admin behaviour).
func (a TaskRepository) Delete(id uint64) error {
	if err := a.q.DeleteTask(context.Background(), int64(id)); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// BatchDelete 批量删除任务
func (a TaskRepository) BatchDelete(ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	int64IDs := make([]int64, len(ids))
	for i, id := range ids {
		int64IDs[i] = int64(id)
	}
	if err := a.q.BatchDeleteTasks(context.Background(), int64IDs); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// GetTaskTypes 获取所有任务类型
func (a TaskRepository) GetTaskTypes() ([]system.TaskTypeVO, error) {
	types, err := a.q.ListTaskTypes(context.Background())
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	taskTypes := make([]system.TaskTypeVO, 0, len(types))
	for _, t := range types {
		taskTypes = append(taskTypes, system.TaskTypeVO{Label: t, Value: t})
	}
	return taskTypes, nil
}

// GetStatusCounts 获取各状态的任务数量
func (a TaskRepository) GetStatusCounts() (*system.TaskStatsVO, error) {
	rows, err := a.q.ListTaskStatusCounts(context.Background())
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	stats := &system.TaskStatsVO{}
	for _, c := range rows {
		switch queue.Status(c.Status) {
		case queue.StatusQueued:
			stats.QueuedCount = c.Count
		case queue.StatusProcessing:
			stats.ProcessingCount = c.Count
		case queue.StatusCompleted:
			stats.CompletedCount = c.Count
		case queue.StatusError:
			stats.ErrorCount = c.Count
		case queue.StatusCanceled:
			stats.CanceledCount = c.Count
		}
	}
	return stats, nil
}

func toDomainTask(r sqlc.SysTask) *system.Task {
	return &system.Task{
		ID:               uint64(r.ID),
		Type:             r.Type,
		Status:           queue.Status(r.Status),
		CorrelationID:    r.CorrelationID,
		OwnerID:          r.OwnerID,
		PrivateState:     r.PrivateState,
		RetryCount:       int(r.PublicRetryCount),
		ExecutedDuration: r.PublicExecutedDuration,
		Error:            r.PublicError,
		ErrorHistory:     r.PublicErrorHistory,
		ResumeTime:       r.PublicResumeTime,
		CreatedAt:        r.CreatedAt.Time,
		UpdatedAt:        r.UpdatedAt.Time,
	}
}
