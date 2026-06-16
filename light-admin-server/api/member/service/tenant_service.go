package service

import (
	"github.com/top-system/light-admin/api/member/repository"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/tenant"
)

type TenantService struct {
	logger     lib.Logger
	tenantRepo repository.TenantRepository
}

func NewTenantService(logger lib.Logger, tenantRepo repository.TenantRepository) TenantService {
	return TenantService{logger: logger, tenantRepo: tenantRepo}
}

func (a TenantService) Query(param *tenant.TenantQueryParam) (*tenant.TenantQueryResult, error) {
	return a.tenantRepo.Query(param)
}

func (a TenantService) Get(id string) (*tenant.Tenant, error) {
	return a.tenantRepo.Get(id)
}

// ResolveByCode 校验租户码并返回启用中的租户
func (a TenantService) ResolveByCode(code string) (*tenant.Tenant, error) {
	if code == "" {
		return nil, errors.TenantCodeRequired
	}
	t, err := a.tenantRepo.GetByCode(code)
	if err != nil {
		return nil, errors.TenantNotFound
	}
	if t.Status != 1 {
		return nil, errors.TenantDisabled
	}
	return t, nil
}

func (a TenantService) Create(form *tenant.TenantForm, operator string) (string, error) {
	if _, err := a.tenantRepo.GetByCode(form.Code); err == nil {
		return "", errors.TenantCodeExists
	}
	t := &tenant.Tenant{
		Code:     form.Code,
		Name:     form.Name,
		Status:   form.Status,
		CreateBy: operator,
	}
	if err := a.tenantRepo.Create(t); err != nil {
		return "", err
	}
	return t.ID, nil
}

func (a TenantService) Update(id string, form *tenant.TenantForm, operator string) error {
	if _, err := a.tenantRepo.Get(id); err != nil {
		return err
	}
	return a.tenantRepo.Update(id, &tenant.Tenant{
		Code:     form.Code,
		Name:     form.Name,
		Status:   form.Status,
		UpdateBy: operator,
	})
}

func (a TenantService) Delete(id string) error {
	if _, err := a.tenantRepo.Get(id); err != nil {
		return err
	}
	return a.tenantRepo.Delete(id)
}
