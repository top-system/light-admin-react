package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/top-system/light-admin/db/sqlc"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/member"
	"github.com/top-system/light-admin/pkg/uuid"
)

// MemberRepository is the sqlc/pgx-backed persistence layer for members. Every
// query is tenant-scoped (filtered by tenant_id) to enforce row-level isolation.
type MemberRepository struct {
	q      sqlc.Querier
	logger lib.Logger
}

// NewMemberRepository creates a new member repository bound to the pool-level Queries.
func NewMemberRepository(q *sqlc.Queries, logger lib.Logger) MemberRepository {
	return MemberRepository{q: q, logger: logger}
}

// NewMemberRepositoryWithQuerier builds a repository over any sqlc.Querier. It
// exists so tests can inject an in-memory Querier; production wiring uses
// NewMemberRepository with the pool-bound *sqlc.Queries.
func NewMemberRepositoryWithQuerier(q sqlc.Querier, logger lib.Logger) MemberRepository {
	return MemberRepository{q: q, logger: logger}
}

// WithTx returns a copy bound to the given transaction-scoped Queries.
func (a MemberRepository) WithTx(q *sqlc.Queries) MemberRepository {
	a.q = q
	return a
}

func (a MemberRepository) GetByUsername(tenantID, username string) (*member.Member, error) {
	row, err := a.q.GetMemberByUsername(context.Background(), sqlc.GetMemberByUsernameParams{
		TenantID: tenantID,
		Username: username,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.MemberRecordNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainMember(row), nil
}

func (a MemberRepository) GetByID(tenantID, id string) (*member.Member, error) {
	row, err := a.q.GetMemberByID(context.Background(), sqlc.GetMemberByIDParams{
		TenantID: tenantID,
		ID:       id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.MemberRecordNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainMember(row), nil
}

func (a MemberRepository) ExistsUsername(tenantID, username string) (bool, error) {
	count, err := a.q.CountMembersByUsername(context.Background(), sqlc.CountMembersByUsernameParams{
		TenantID: tenantID,
		Username: username,
	})
	if err != nil {
		return false, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return count > 0, nil
}

// Create inserts a new member, assigning a UUID when the ID is empty.
func (a MemberRepository) Create(m *member.Member) error {
	if m.ID == "" {
		m.ID = uuid.NewID()
	}
	err := a.q.CreateMember(context.Background(), sqlc.CreateMemberParams{
		ID:        m.ID,
		TenantID:  m.TenantID,
		Username:  m.Username,
		Email:     m.Email,
		Mobile:    m.Mobile,
		Password:  m.Password,
		Nickname:  m.Nickname,
		Avatar:    m.Avatar,
		Gender:    int32(m.Gender),
		Status:    int32(m.Status),
		IsDeleted: int32(m.IsDeleted),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a MemberRepository) UpdateProfile(tenantID, id string, form *member.MemberProfileForm) error {
	err := a.q.UpdateMemberProfile(context.Background(), sqlc.UpdateMemberProfileParams{
		TenantID: tenantID,
		ID:       id,
		Nickname: form.Nickname,
		Avatar:   form.Avatar,
		Gender:   int32(form.Gender),
		Mobile:   form.Mobile,
		Email:    form.Email,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a MemberRepository) UpdateStatus(tenantID, id string, status int) error {
	err := a.q.UpdateMemberStatus(context.Background(), sqlc.UpdateMemberStatusParams{
		TenantID: tenantID,
		ID:       id,
		Status:   int32(status),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a MemberRepository) UpdatePassword(tenantID, id, hashed string) error {
	err := a.q.UpdateMemberPassword(context.Background(), sqlc.UpdateMemberPasswordParams{
		TenantID: tenantID,
		ID:       id,
		Password: hashed,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a MemberRepository) UpdateLoginInfo(tenantID, id, ip string) error {
	err := a.q.UpdateMemberLoginInfo(context.Background(), sqlc.UpdateMemberLoginInfoParams{
		TenantID:    tenantID,
		ID:          id,
		LastLoginIp: ip,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Query 后台分页查询（可跨租户，按 tenantId 可选过滤）. Ordering is fixed to id DESC.
func (a MemberRepository) Query(param *member.MemberQueryParam) (*member.MemberQueryResult, error) {
	ctx := context.Background()

	var tenantID *string
	if param.TenantID != "" {
		tenantID = ptr(param.TenantID)
	}
	var keywords *string
	if param.Keywords != "" {
		keywords = ptr("%" + param.Keywords + "%")
	}
	var status *int32
	if param.Status != nil {
		status = ptr(int32(*param.Status))
	}

	total, err := a.q.CountMembers(ctx, sqlc.CountMembersParams{TenantID: tenantID, Keywords: keywords, Status: status})
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(member.Members, 0)
	if total > 0 {
		limit, offset := pageBounds(param.PaginationParam)
		rows, err := a.q.ListMembers(ctx, sqlc.ListMembersParams{
			TenantID: tenantID,
			Keywords: keywords,
			Status:   status,
			Limit:    limit,
			Offset:   offset,
		})
		if err != nil {
			return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
		}
		for _, r := range rows {
			m := toDomainMember(r)
			m.CleanSecure() // never expose password hashes in list responses
			list = append(list, m)
		}
	}

	p := dtoPagination(total, param.GetPageNum(), param.GetPageSize())
	return &member.MemberQueryResult{List: list, Pagination: &p}, nil
}

func toDomainMember(r sqlc.TMember) *member.Member {
	return &member.Member{
		ID:            r.ID,
		TenantID:      r.TenantID,
		Username:      r.Username,
		Email:         r.Email,
		Mobile:        r.Mobile,
		Password:      r.Password,
		Nickname:      r.Nickname,
		Avatar:        r.Avatar,
		Gender:        int(r.Gender),
		Status:        int(r.Status),
		LastLoginTime: dto.DateTime(r.LastLoginTime.Time),
		LastLoginIP:   r.LastLoginIp,
		CreateTime:    dto.DateTime(r.CreateTime.Time),
		UpdateTime:    dto.DateTime(r.UpdateTime.Time),
		IsDeleted:     int(r.IsDeleted),
	}
}
