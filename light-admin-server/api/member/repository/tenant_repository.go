package repository

import (
	"gorm.io/gorm"

	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/tenant"
)

type TenantRepository struct {
	db     lib.Database
	logger lib.Logger
}

func NewTenantRepository(db lib.Database, logger lib.Logger) TenantRepository {
	return TenantRepository{db: db, logger: logger}
}

func (a TenantRepository) Query(param *tenant.TenantQueryParam) (*tenant.TenantQueryResult, error) {
	db := a.db.ORM.Model(&tenant.Tenant{}).Where("is_deleted = ?", 0)
	if v := param.Keywords; v != "" {
		db = db.Where("code LIKE ? OR name LIKE ?", "%"+v+"%", "%"+v+"%")
	}
	if v := param.Status; v != nil {
		db = db.Where("status = ?", *v)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	var list tenant.Tenants
	err := db.Order(param.ParseOrder()).
		Offset((param.GetPageNum() - 1) * param.GetPageSize()).
		Limit(param.GetPageSize()).
		Find(&list).Error
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	p := dtoPagination(total, param.GetPageNum(), param.GetPageSize())
	return &tenant.TenantQueryResult{List: list, Pagination: &p}, nil
}

func (a TenantRepository) GetByCode(code string) (*tenant.Tenant, error) {
	var t tenant.Tenant
	result := a.db.ORM.Where("code = ? AND is_deleted = ?", code, 0).First(&t)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.TenantNotFound
		}
		return nil, errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}
	return &t, nil
}

func (a TenantRepository) Get(id string) (*tenant.Tenant, error) {
	var t tenant.Tenant
	result := a.db.ORM.Where("id = ? AND is_deleted = ?", id, 0).First(&t)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.TenantNotFound
		}
		return nil, errors.Wrap(errors.DatabaseInternalError, result.Error.Error())
	}
	return &t, nil
}

func (a TenantRepository) Create(t *tenant.Tenant) error {
	if err := a.db.ORM.Create(t).Error; err != nil {
		return errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a TenantRepository) Update(id string, t *tenant.Tenant) error {
	err := a.db.ORM.Model(&tenant.Tenant{}).Where("id = ?", id).
		Select("code", "name", "status", "update_by").Updates(t).Error
	if err != nil {
		return errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a TenantRepository) Delete(id string) error {
	err := a.db.ORM.Model(&tenant.Tenant{}).Where("id = ?", id).
		Update("is_deleted", 1).Error
	if err != nil {
		return errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return nil
}
