package repository

import (
	"context"
	"errors"
	"time"

	"github.com/top-system/light-admin/db/store"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
)

// DownloadRepository is the sqlc/pgx-backed persistence layer for download tasks.
// The table uses a BIGSERIAL surrogate key, so IDs are int64 in the data layer and
// converted to/from the domain model's uint64.
type DownloadRepository struct {
	q      store.Store
	logger lib.Logger
}

// NewDownloadRepository creates a new download repository bound to the pool-level Queries.
func NewDownloadRepository(q store.Store, logger lib.Logger) DownloadRepository {
	return DownloadRepository{q: q, logger: logger}
}

// WithTx returns a copy bound to the given transaction-scoped Queries.
func (a DownloadRepository) WithTx(q store.Store) DownloadRepository {
	a.q = q
	return a
}

// Query 查询下载任务列表. Ordering is fixed to created_at DESC.
func (a DownloadRepository) Query(param *system.DownloadTaskQueryParam) (*system.DownloadTaskQueryResult, error) {
	ctx := context.Background()

	var status, downloader, keywords *string
	if param.Status != "" {
		status = ptr(param.Status)
	}
	if param.Downloader != "" {
		downloader = ptr(param.Downloader)
	}
	if param.Keywords != "" {
		keywords = ptr("%" + param.Keywords + "%")
	}
	createFrom := startOfDayFilter(param.CreateTimeFrom)
	createTo := endOfDayFilter(param.CreateTimeTo)

	total, err := a.q.CountDownloadTasks(ctx, store.CountDownloadTasksParams{
		Status:     status,
		Downloader: downloader,
		Keywords:   keywords,
		CreateFrom: createFrom,
		CreateTo:   createTo,
	})
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.DownloadTasks, 0)
	if total > 0 {
		limit, offset := pageBounds(param.PaginationParam)
		rows, err := a.q.ListDownloadTasks(ctx, store.ListDownloadTasksParams{
			Status:     status,
			Downloader: downloader,
			Keywords:   keywords,
			CreateFrom: createFrom,
			CreateTo:   createTo,
			Limit:      limit,
			Offset:     offset,
		})
		if err != nil {
			return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
		}
		for _, r := range rows {
			list = append(list, toDomainDownloadTask(r))
		}
	}

	return &system.DownloadTaskQueryResult{
		Pagination: &dto.Pagination{
			PageNum:  param.PaginationParam.GetPageNum(),
			PageSize: param.PaginationParam.GetPageSize(),
			Total:    total,
		},
		List: list,
	}, nil
}

func (a DownloadRepository) Get(id uint64) (*system.DownloadTask, error) {
	return a.getOne(func() (store.SysDownloadTask, error) {
		return a.q.GetDownloadTask(context.Background(), int64(id))
	})
}

func (a DownloadRepository) GetByTaskID(taskID string) (*system.DownloadTask, error) {
	return a.getOne(func() (store.SysDownloadTask, error) {
		return a.q.GetDownloadTaskByTaskID(context.Background(), taskID)
	})
}

func (a DownloadRepository) GetByQueueTaskID(queueTaskID uint64) (*system.DownloadTask, error) {
	return a.getOne(func() (store.SysDownloadTask, error) {
		return a.q.GetDownloadTaskByQueueTaskID(context.Background(), int64(queueTaskID))
	})
}

func (a DownloadRepository) getOne(fetch func() (store.SysDownloadTask, error)) (*system.DownloadTask, error) {
	row, err := fetch()
	if err != nil {
		if errors.Is(err, store.ErrNoRows) {
			return nil, apperrors.DatabaseRecordNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainDownloadTask(row), nil
}

// Create inserts a new download task and writes the generated ID back onto the model.
func (a DownloadRepository) Create(task *system.DownloadTask) error {
	id, err := a.q.CreateDownloadTask(context.Background(), store.CreateDownloadTaskParams{
		QueueTaskID:   int64(task.QueueTaskID),
		TaskID:        task.TaskID,
		Hash:          task.Hash,
		Name:          task.Name,
		Url:           task.URL,
		Downloader:    task.Downloader,
		Status:        task.Status,
		Total:         task.Total,
		Downloaded:    task.Downloaded,
		DownloadSpeed: task.DownloadSpeed,
		Uploaded:      task.Uploaded,
		UploadSpeed:   task.UploadSpeed,
		SavePath:      task.SavePath,
		ErrorMessage:  task.ErrorMessage,
		OwnerID:       task.OwnerID,
		Now:           time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	task.ID = uint64(id)
	return nil
}

// Update saves the full task row (mirrors the previous GORM Save).
func (a DownloadRepository) Update(task *system.DownloadTask) error {
	err := a.q.UpdateDownloadTask(context.Background(), store.UpdateDownloadTaskParams{
		ID:            int64(task.ID),
		QueueTaskID:   int64(task.QueueTaskID),
		TaskID:        task.TaskID,
		Hash:          task.Hash,
		Name:          task.Name,
		Url:           task.URL,
		Downloader:    task.Downloader,
		Status:        task.Status,
		Total:         task.Total,
		Downloaded:    task.Downloaded,
		DownloadSpeed: task.DownloadSpeed,
		Uploaded:      task.Uploaded,
		UploadSpeed:   task.UploadSpeed,
		SavePath:      task.SavePath,
		ErrorMessage:  task.ErrorMessage,
		OwnerID:       task.OwnerID,
		Now:           time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a DownloadRepository) UpdateStatus(id uint64, status string, downloaded, total, downloadSpeed, uploaded, uploadSpeed int64, errorMessage string) error {
	err := a.q.UpdateDownloadTaskStatus(context.Background(), store.UpdateDownloadTaskStatusParams{
		ID:            int64(id),
		Status:        status,
		Downloaded:    downloaded,
		Total:         total,
		DownloadSpeed: downloadSpeed,
		Uploaded:      uploaded,
		UploadSpeed:   uploadSpeed,
		ErrorMessage:  errorMessage,
		Now:           time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a DownloadRepository) Delete(id uint64) error {
	if err := a.q.DeleteDownloadTask(context.Background(), int64(id)); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a DownloadRepository) BatchDelete(ids []uint64) error {
	if len(ids) == 0 {
		return nil
	}
	int64IDs := make([]int64, len(ids))
	for i, id := range ids {
		int64IDs[i] = int64(id)
	}
	if err := a.q.BatchDeleteDownloadTasks(context.Background(), int64IDs); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// GetStatusCounts 获取各状态的任务数量
func (a DownloadRepository) GetStatusCounts() (*system.DownloadTaskStatsVO, error) {
	rows, err := a.q.ListDownloadStatusCounts(context.Background())
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	stats := &system.DownloadTaskStatsVO{}
	for _, c := range rows {
		stats.TotalCount += c.Count
		switch c.Status {
		case "downloading":
			stats.DownloadingCount = c.Count
		case "seeding":
			stats.SeedingCount = c.Count
		case "completed":
			stats.CompletedCount = c.Count
		case "error":
			stats.ErrorCount = c.Count
		}
	}
	return stats, nil
}

// GetActiveTaskIDs 获取活跃任务ID列表（用于状态同步）
func (a DownloadRepository) GetActiveTaskIDs() ([]system.DownloadTask, error) {
	rows, err := a.q.ListActiveDownloadTasks(context.Background())
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	tasks := make([]system.DownloadTask, 0, len(rows))
	for _, r := range rows {
		tasks = append(tasks, system.DownloadTask{
			ID:          uint64(r.ID),
			QueueTaskID: uint64(r.QueueTaskID),
			TaskID:      r.TaskID,
			Hash:        r.Hash,
			Downloader:  r.Downloader,
		})
	}
	return tasks, nil
}

// UpdateFromDownloader 从下载器同步更新任务完整信息. task_id/hash/name/save_path are
// only overwritten when a non-empty value is supplied (handled in SQL).
func (a DownloadRepository) UpdateFromDownloader(id uint64, taskID, hash, name, savePath, status string, downloaded, total, downloadSpeed, uploaded, uploadSpeed int64, errorMessage string) error {
	err := a.q.UpdateDownloadTaskFromDownloader(context.Background(), store.UpdateDownloadTaskFromDownloaderParams{
		ID:            int64(id),
		Status:        status,
		Downloaded:    downloaded,
		Total:         total,
		DownloadSpeed: downloadSpeed,
		Uploaded:      uploaded,
		UploadSpeed:   uploadSpeed,
		ErrorMessage:  errorMessage,
		TaskID:        taskID,
		Hash:          hash,
		Name:          name,
		SavePath:      savePath,
		Now:           time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func toDomainDownloadTask(r store.SysDownloadTask) *system.DownloadTask {
	return &system.DownloadTask{
		ID:            uint64(r.ID),
		QueueTaskID:   uint64(r.QueueTaskID),
		TaskID:        r.TaskID,
		Hash:          r.Hash,
		Name:          r.Name,
		URL:           r.Url,
		Downloader:    r.Downloader,
		Status:        r.Status,
		Total:         r.Total,
		Downloaded:    r.Downloaded,
		DownloadSpeed: r.DownloadSpeed,
		Uploaded:      r.Uploaded,
		UploadSpeed:   r.UploadSpeed,
		SavePath:      r.SavePath,
		ErrorMessage:  r.ErrorMessage,
		OwnerID:       r.OwnerID,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
		DeletedAt:     dto.DateTime(tsOrZero(r.DeletedAt)),
	}
}
