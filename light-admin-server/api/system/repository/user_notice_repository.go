package repository

import (
	"context"

	"github.com/top-system/light-admin/db/sqlc"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/pkg/uuid"
)

// UserNoticeRepository is the sqlc/pgx-backed persistence layer for per-user
// notice state (t_user_notice).
type UserNoticeRepository struct {
	q      sqlc.Querier
	logger lib.Logger
}

// NewUserNoticeRepository creates a new user notice repository bound to the
// pool-level Queries.
func NewUserNoticeRepository(q *sqlc.Queries, logger lib.Logger) UserNoticeRepository {
	return UserNoticeRepository{
		q:      q,
		logger: logger,
	}
}

// WithTx returns a copy bound to the given transaction-scoped Queries.
func (a UserNoticeRepository) WithTx(q *sqlc.Queries) UserNoticeRepository {
	a.q = q
	return a
}

// GetMyNoticePage 获取我的通知公告分页列表. Ordering is fixed to publish_time DESC.
func (a UserNoticeRepository) GetMyNoticePage(param *system.NoticeQueryParam) ([]system.UserNoticePageVO, int64, error) {
	ctx := context.Background()

	var title *string
	if param.Title != "" {
		title = ptr("%" + param.Title + "%")
	}
	var noticeType *int32
	if param.Type != 0 {
		noticeType = ptr(int32(param.Type))
	}

	total, err := a.q.CountMyNotices(ctx, sqlc.CountMyNoticesParams{
		UserID: param.UserID,
		Title:  title,
		Type:   noticeType,
	})
	if err != nil {
		return nil, 0, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make([]system.UserNoticePageVO, 0)
	if total > 0 {
		limit, offset := pageBounds(param.PaginationParam)
		rows, err := a.q.ListMyNotices(ctx, sqlc.ListMyNoticesParams{
			UserID: param.UserID,
			Title:  title,
			Type:   noticeType,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return nil, 0, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
		}
		for _, r := range rows {
			list = append(list, system.UserNoticePageVO{
				ID:          r.ID,
				NoticeID:    r.NoticeID,
				Title:       derefString(r.Title),
				Type:        int(derefInt32(r.Type)),
				Level:       derefString(r.Level),
				PublishTime: dto.NullDateTime{Time: r.PublishTime.Time, Valid: r.PublishTime.Valid},
				IsRead:      int(r.IsRead),
			})
		}
	}

	return list, total, nil
}

// Create inserts a single user notice, assigning a UUID when the ID is empty.
func (a UserNoticeRepository) Create(userNotice *system.UserNotice) error {
	if userNotice.ID == "" {
		userNotice.ID = uuid.NewID()
	}
	err := a.q.CreateUserNotice(context.Background(), sqlc.CreateUserNoticeParams{
		ID:       userNotice.ID,
		NoticeID: userNotice.NoticeID,
		UserID:   userNotice.UserID,
		IsRead:   int32(userNotice.IsRead),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// BatchCreate inserts multiple user notices in a single statement, assigning a
// UUID to each row that lacks one.
func (a UserNoticeRepository) BatchCreate(userNotices []*system.UserNotice) error {
	if len(userNotices) == 0 {
		return nil
	}

	ids := make([]string, len(userNotices))
	noticeIDs := make([]string, len(userNotices))
	userIDs := make([]string, len(userNotices))
	for i, un := range userNotices {
		if un.ID == "" {
			un.ID = uuid.NewID()
		}
		ids[i] = un.ID
		noticeIDs[i] = un.NoticeID
		userIDs[i] = un.UserID
	}

	err := a.q.BatchCreateUserNotices(context.Background(), sqlc.BatchCreateUserNoticesParams{
		Ids:       ids,
		NoticeIds: noticeIDs,
		UserIds:   userIDs,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// MarkAsRead marks a single notice read for a user.
func (a UserNoticeRepository) MarkAsRead(noticeID, userID string) error {
	err := a.q.MarkUserNoticeRead(context.Background(), sqlc.MarkUserNoticeReadParams{
		NoticeID: noticeID,
		UserID:   userID,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// MarkAllAsRead marks every unread notice read for a user.
func (a UserNoticeRepository) MarkAllAsRead(userID string) error {
	if err := a.q.MarkAllUserNoticesRead(context.Background(), userID); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// DeleteByNoticeID removes all per-user state for a notice.
func (a UserNoticeRepository) DeleteByNoticeID(noticeID string) error {
	if err := a.q.DeleteUserNoticesByNoticeID(context.Background(), noticeID); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// DeleteByNoticeIDs removes all per-user state for multiple notices.
func (a UserNoticeRepository) DeleteByNoticeIDs(noticeIDs []string) error {
	if len(noticeIDs) == 0 {
		return nil
	}
	if err := a.q.DeleteUserNoticesByNoticeIDs(context.Background(), noticeIDs); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}
