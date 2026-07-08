package repository

import (
	"context"
	"errors"
	"time"

	"github.com/top-system/light-admin/db/store"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/tenant"
	"github.com/top-system/light-admin/pkg/uuid"
)

// TenantRepository is the sqlc/pgx-backed persistence layer for tenants.
type TenantRepository struct {
	q      store.Store
	logger lib.Logger
}

// NewTenantRepository creates a new tenant repository bound to the pool-level Queries.
func NewTenantRepository(q store.Store, logger lib.Logger) TenantRepository {
	return TenantRepository{q: q, logger: logger}
}

// WithTx returns a copy bound to the given transaction-scoped Queries.
func (a TenantRepository) WithTx(q store.Store) TenantRepository {
	a.q = q
	return a
}

// Query lists tenants. Ordering is fixed to id DESC (the previous default).
func (a TenantRepository) Query(param *tenant.TenantQueryParam) (*tenant.TenantQueryResult, error) {
	ctx := context.Background()

	var keywords *string
	if param.Keywords != "" {
		keywords = ptr("%" + param.Keywords + "%")
	}
	var status *int32
	if param.Status != nil {
		status = ptr(int32(*param.Status))
	}

	total, err := a.q.CountTenants(ctx, store.CountTenantsParams{Keywords: keywords, Status: status})
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(tenant.Tenants, 0)
	if total > 0 {
		limit, offset := pageBounds(param.PaginationParam)
		rows, err := a.q.ListTenants(ctx, store.ListTenantsParams{Keywords: keywords, Status: status, Limit: limit, Offset: offset})
		if err != nil {
			return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
		}
		for _, r := range rows {
			list = append(list, toDomainTenant(r))
		}
	}

	p := dtoPagination(total, param.GetPageNum(), param.GetPageSize())
	return &tenant.TenantQueryResult{List: list, Pagination: &p}, nil
}

func (a TenantRepository) GetByCode(code string) (*tenant.Tenant, error) {
	row, err := a.q.GetTenantByCode(context.Background(), code)
	if err != nil {
		if errors.Is(err, store.ErrNoRows) {
			return nil, apperrors.TenantNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainTenant(row), nil
}

func (a TenantRepository) Get(id string) (*tenant.Tenant, error) {
	row, err := a.q.GetTenant(context.Background(), id)
	if err != nil {
		if errors.Is(err, store.ErrNoRows) {
			return nil, apperrors.TenantNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainTenant(row), nil
}

// Create inserts a new tenant, assigning a UUID when the ID is empty.
func (a TenantRepository) Create(t *tenant.Tenant) error {
	if t.ID == "" {
		t.ID = uuid.NewID()
	}
	err := a.q.CreateTenant(context.Background(), store.CreateTenantParams{
		ID:        t.ID,
		Code:      t.Code,
		Name:      t.Name,
		Status:    int32(t.Status),
		CreateBy:  t.CreateBy,
		IsDeleted: int32(t.IsDeleted),
		Now:       time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a TenantRepository) Update(id string, t *tenant.Tenant) error {
	err := a.q.UpdateTenant(context.Background(), store.UpdateTenantParams{
		ID:       id,
		Code:     t.Code,
		Name:     t.Name,
		Status:   int32(t.Status),
		UpdateBy: t.UpdateBy,
		Now:      time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a TenantRepository) Delete(id string) error {
	if err := a.q.SoftDeleteTenant(context.Background(), store.SoftDeleteTenantParams{ID: id, Now: time.Now()}); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func toDomainTenant(r store.TTenant) *tenant.Tenant {
	return &tenant.Tenant{
		ID:         r.ID,
		Code:       r.Code,
		Name:       r.Name,
		Status:     int(r.Status),
		CreateTime: dto.DateTime(r.CreateTime),
		CreateBy:   r.CreateBy,
		UpdateTime: dto.DateTime(r.UpdateTime),
		UpdateBy:   r.UpdateBy,
		IsDeleted:  int(r.IsDeleted),
	}
}
