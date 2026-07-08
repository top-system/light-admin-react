package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/top-system/light-admin/db/store"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/pkg/uuid"
)

// LogRepository is the sqlc/pgx-backed persistence layer for operation logs.
type LogRepository struct {
	q      store.Store
	logger lib.Logger
}

// NewLogRepository creates a new log repository bound to the pool-level Queries.
func NewLogRepository(q store.Store, logger lib.Logger) LogRepository {
	return LogRepository{
		q:      q,
		logger: logger,
	}
}

// WithTx returns a copy bound to the given transaction-scoped Queries.
func (a LogRepository) WithTx(q store.Store) LogRepository {
	a.q = q
	return a
}

// Query 查询日志列表. Ordering is fixed to create_time DESC (the previous default).
func (a LogRepository) Query(param *system.LogQueryParam) (*system.LogQueryResult, error) {
	ctx := context.Background()

	var module *string
	if param.Module != "" {
		module = ptr(param.Module)
	}
	var keywords *string
	if param.Keywords != "" {
		keywords = ptr("%" + param.Keywords + "%")
	}
	createFrom := startOfDayFilter(param.CreateTimeFrom)
	createTo := endOfDayFilter(param.CreateTimeTo)

	total, err := a.q.CountLogs(ctx, store.CountLogsParams{
		Module:     module,
		Keywords:   keywords,
		CreateFrom: createFrom,
		CreateTo:   createTo,
	})
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.Logs, 0)
	if total > 0 {
		limit, offset := pageBounds(param.PaginationParam)
		rows, err := a.q.ListLogs(ctx, store.ListLogsParams{
			Module:     module,
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
			list = append(list, toDomainLog(r))
		}
	}

	return &system.LogQueryResult{
		Pagination: &dto.Pagination{
			PageNum:  param.PaginationParam.GetPageNum(),
			PageSize: param.PaginationParam.GetPageSize(),
			Total:    total,
		},
		List: list,
	}, nil
}

// Get 获取日志详情
func (a LogRepository) Get(id string) (*system.Log, error) {
	row, err := a.q.GetLog(context.Background(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.DatabaseRecordNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainLog(row), nil
}

// Create 创建日志, assigning a UUID when the ID is empty.
func (a LogRepository) Create(log *system.Log) error {
	if log.ID == "" {
		log.ID = uuid.NewID()
	}

	err := a.q.CreateLog(context.Background(), store.CreateLogParams{
		ID:              log.ID,
		Module:          log.Module,
		RequestMethod:   log.RequestMethod,
		RequestParams:   log.RequestParams,
		ResponseContent: log.ResponseContent,
		Content:         log.Content,
		RequestUri:      log.RequestURI,
		Method:          log.Method,
		Ip:              log.IP,
		Province:        log.Province,
		City:            log.City,
		ExecutionTime:   log.ExecutionTime,
		Browser:         log.Browser,
		BrowserVersion:  log.BrowserVersion,
		Os:              log.OS,
		CreateBy:        log.CreateBy,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Delete 删除日志 (hard delete; sys_log has no soft-delete column).
func (a LogRepository) Delete(id string) error {
	if err := a.q.DeleteLog(context.Background(), id); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// BatchDelete 批量删除日志
func (a LogRepository) BatchDelete(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	if err := a.q.BatchDeleteLogs(context.Background(), ids); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func toDomainLog(r store.SysLog) *system.Log {
	return &system.Log{
		ID:              r.ID,
		Module:          r.Module,
		RequestMethod:   r.RequestMethod,
		RequestParams:   r.RequestParams,
		ResponseContent: r.ResponseContent,
		Content:         r.Content,
		RequestURI:      r.RequestUri,
		Method:          r.Method,
		IP:              r.Ip,
		Province:        r.Province,
		City:            r.City,
		ExecutionTime:   r.ExecutionTime,
		Browser:         r.Browser,
		BrowserVersion:  r.BrowserVersion,
		OS:              r.Os,
		CreateBy:        r.CreateBy,
		CreateTime:      dto.DateTime(r.CreateTime.Time),
	}
}
