package repository

import (
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/member"
	"github.com/top-system/light-admin/pkg/scopes"
)

type MemberRepository struct {
	db     lib.Database
	logger lib.Logger
}

func NewMemberRepository(db lib.Database, logger lib.Logger) MemberRepository {
	return MemberRepository{db: db, logger: logger}
}

func (a MemberRepository) GetByUsername(tenantID, username string) (*member.Member, error) {
	var m member.Member
	err := a.db.ORM.Model(&member.Member{}).
		Scopes(scopes.Tenant(tenantID)).
		Where("username = ? AND is_deleted = ?", username, 0).
		First(&m).Error
	if err != nil {
		return nil, errors.MemberRecordNotFound
	}
	return &m, nil
}

func (a MemberRepository) GetByID(tenantID, id string) (*member.Member, error) {
	var m member.Member
	err := a.db.ORM.Model(&member.Member{}).
		Scopes(scopes.Tenant(tenantID)).
		Where("id = ? AND is_deleted = ?", id, 0).
		First(&m).Error
	if err != nil {
		return nil, errors.MemberRecordNotFound
	}
	return &m, nil
}

func (a MemberRepository) ExistsUsername(tenantID, username string) (bool, error) {
	var count int64
	err := a.db.ORM.Model(&member.Member{}).
		Scopes(scopes.Tenant(tenantID)).
		Where("username = ? AND is_deleted = ?", username, 0).
		Count(&count).Error
	if err != nil {
		return false, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return count > 0, nil
}

func (a MemberRepository) Create(m *member.Member) error {
	if err := a.db.ORM.Create(m).Error; err != nil {
		return errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a MemberRepository) UpdateProfile(tenantID, id string, m *member.Member) error {
	err := a.db.ORM.Model(&member.Member{}).
		Scopes(scopes.Tenant(tenantID)).
		Where("id = ?", id).
		Select("nickname", "avatar", "gender", "mobile", "email").
		Updates(m).Error
	if err != nil {
		return errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a MemberRepository) UpdateStatus(tenantID, id string, status int) error {
	err := a.db.ORM.Model(&member.Member{}).
		Scopes(scopes.Tenant(tenantID)).
		Where("id = ?", id).Update("status", status).Error
	if err != nil {
		return errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a MemberRepository) UpdatePassword(tenantID, id, hashed string) error {
	err := a.db.ORM.Model(&member.Member{}).
		Scopes(scopes.Tenant(tenantID)).
		Where("id = ?", id).Update("password", hashed).Error
	if err != nil {
		return errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a MemberRepository) UpdateLoginInfo(tenantID, id, ip string) error {
	err := a.db.ORM.Model(&member.Member{}).
		Scopes(scopes.Tenant(tenantID)).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_ip":   ip,
			"last_login_time": nowDateTime(),
		}).Error
	if err != nil {
		return errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Query 后台分页查询（可跨租户，按 tenantId 可选过滤）
func (a MemberRepository) Query(param *member.MemberQueryParam) (*member.MemberQueryResult, error) {
	db := a.db.ORM.Model(&member.Member{}).Where("is_deleted = ?", 0)
	if v := param.TenantID; v != "" {
		db = db.Where("tenant_id = ?", v)
	}
	if v := param.Keywords; v != "" {
		db = db.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?", "%"+v+"%", "%"+v+"%", "%"+v+"%")
	}
	if v := param.Status; v != nil {
		db = db.Where("status = ?", *v)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	var list member.Members
	err := db.Order(param.ParseOrder()).
		Offset((param.GetPageNum() - 1) * param.GetPageSize()).
		Limit(param.GetPageSize()).
		Find(&list).Error
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	for _, m := range list {
		m.CleanSecure()
	}

	p := dtoPagination(total, param.GetPageNum(), param.GetPageSize())
	return &member.MemberQueryResult{List: list, Pagination: &p}, nil
}
