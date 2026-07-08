package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/top-system/light-admin/db/store"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/pkg/uuid"
)

// NoticeRepository is the sqlc/pgx-backed persistence layer for notices.
type NoticeRepository struct {
	q      store.Store
	logger lib.Logger
}

// NewNoticeRepository creates a new notice repository bound to the pool-level Queries.
func NewNoticeRepository(q store.Store, logger lib.Logger) NoticeRepository {
	return NoticeRepository{
		q:      q,
		logger: logger,
	}
}

// WithTx returns a copy bound to the given transaction-scoped Queries.
func (a NoticeRepository) WithTx(q store.Store) NoticeRepository {
	a.q = q
	return a
}

// Query 分页查询通知公告. Ordering is fixed to create_time DESC.
func (a NoticeRepository) Query(param *system.NoticeQueryParam) (*system.NoticeQueryResult, error) {
	ctx := context.Background()

	var title *string
	if param.Title != "" {
		title = ptr("%" + param.Title + "%")
	}
	var noticeType *int32
	if param.Type != 0 {
		noticeType = ptr(int32(param.Type))
	}
	var publishStatus *int32
	if param.PublishStatus != nil {
		publishStatus = ptr(int32(*param.PublishStatus))
	}

	total, err := a.q.CountNotices(ctx, store.CountNoticesParams{
		Title:         title,
		Type:          noticeType,
		PublishStatus: publishStatus,
	})
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.Notices, 0)
	if total > 0 {
		limit, offset := pageBounds(param.PaginationParam)
		rows, err := a.q.ListNotices(ctx, store.ListNoticesParams{
			Title:         title,
			Type:          noticeType,
			PublishStatus: publishStatus,
			Limit:         limit,
			Offset:        offset,
		})
		if err != nil {
			return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
		}
		for _, r := range rows {
			list = append(list, toDomainNotice(r))
		}
	}

	return &system.NoticeQueryResult{
		Pagination: &dto.Pagination{
			PageNum:  param.PaginationParam.GetPageNum(),
			PageSize: param.PaginationParam.GetPageSize(),
			Total:    total,
		},
		List: list,
	}, nil
}

func (a NoticeRepository) Get(id string) (*system.Notice, error) {
	row, err := a.q.GetNotice(context.Background(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.DatabaseRecordNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainNotice(row), nil
}

// Create inserts a new notice, assigning a UUID when the ID is empty.
func (a NoticeRepository) Create(notice *system.Notice) error {
	if notice.ID == "" {
		notice.ID = uuid.NewID()
	}

	err := a.q.CreateNotice(context.Background(), store.CreateNoticeParams{
		ID:            notice.ID,
		Title:         notice.Title,
		Content:       notice.Content,
		Type:          int32(notice.Type),
		Level:         notice.Level,
		TargetType:    int32(notice.TargetType),
		TargetUserIds: notice.TargetUserIds,
		PublishStatus: int32(notice.PublishStatus),
		CreateBy:      notice.CreateBy,
		IsDeleted:     int32(notice.IsDeleted),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Update updates the mutable content fields of a notice.
func (a NoticeRepository) Update(id string, notice *system.Notice) error {
	err := a.q.UpdateNotice(context.Background(), store.UpdateNoticeParams{
		ID:            id,
		Title:         notice.Title,
		Content:       notice.Content,
		Type:          int32(notice.Type),
		Level:         notice.Level,
		TargetType:    int32(notice.TargetType),
		TargetUserIds: notice.TargetUserIds,
		UpdateBy:      notice.UpdateBy,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// UpdateStatus updates the publish status, stamping publish_time when publishing
// (status == 1) or revoke_time when revoking (status == -1).
func (a NoticeRepository) UpdateStatus(id string, status int, publisherId string) error {
	params := store.UpdateNoticeStatusParams{
		ID:            id,
		PublishStatus: int32(status),
		PublisherID:   publisherId,
	}
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	if status == 1 {
		params.PublishTime = now
	} else if status == -1 {
		params.RevokeTime = now
	}

	if err := a.q.UpdateNoticeStatus(context.Background(), params); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Delete soft-deletes a single notice.
func (a NoticeRepository) Delete(id string, deletedBy string) error {
	return a.BatchDelete([]string{id}, deletedBy)
}

// BatchDelete soft-deletes multiple notices.
func (a NoticeRepository) BatchDelete(ids []string, deletedBy string) error {
	if len(ids) == 0 {
		return nil
	}
	err := a.q.SoftDeleteNoticesByIDs(context.Background(), store.SoftDeleteNoticesByIDsParams{
		Ids:      ids,
		UpdateBy: deletedBy,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func toDomainNotice(r store.TNotice) *system.Notice {
	return &system.Notice{
		ID:            r.ID,
		Title:         r.Title,
		Content:       r.Content,
		Type:          int(r.Type),
		Level:         r.Level,
		TargetType:    int(r.TargetType),
		TargetUserIds: r.TargetUserIds,
		PublisherId:   r.PublisherID,
		PublishStatus: int(r.PublishStatus),
		PublishTime:   dto.NullDateTime{Time: r.PublishTime.Time, Valid: r.PublishTime.Valid},
		RevokeTime:    dto.NullDateTime{Time: r.RevokeTime.Time, Valid: r.RevokeTime.Valid},
		CreateBy:      r.CreateBy,
		CreateTime:    dto.DateTime(r.CreateTime.Time),
		UpdateBy:      r.UpdateBy,
		UpdateTime:    dto.DateTime(r.UpdateTime.Time),
		IsDeleted:     int(r.IsDeleted),
	}
}
