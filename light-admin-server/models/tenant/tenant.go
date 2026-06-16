package tenant

import "github.com/top-system/light-admin/models/dto"

// Tenant 租户模型
// Status: 1-正常 0-禁用
type Tenant struct {
	ID         string       `gorm:"primaryKey;type:char(32)" json:"id"`
	Code       string       `gorm:"column:code;size:64;not null;uniqueIndex:uniq_tenant_code" json:"code"`
	Name       string       `gorm:"column:name;size:128;not null" json:"name"`
	Status     int          `gorm:"column:status;default:1" json:"status"`
	CreateTime dto.DateTime `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	CreateBy   string       `gorm:"column:create_by" json:"createBy"`
	UpdateTime dto.DateTime `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
	UpdateBy   string       `gorm:"column:update_by" json:"updateBy"`
	IsDeleted  int          `gorm:"column:is_deleted;default:0" json:"isDeleted"`
}

func (Tenant) TableName() string { return "t_tenant" }

type Tenants []*Tenant

// TenantForm 租户表单
type TenantForm struct {
	ID     string `json:"id"`
	Code   string `json:"code" validate:"required"`
	Name   string `json:"name" validate:"required"`
	Status int    `json:"status"`
}

// TenantQueryParam 租户查询参数
type TenantQueryParam struct {
	dto.PaginationParam
	dto.OrderParam
	Keywords string `query:"keywords"`
	Status   *int   `query:"status"`
}

// TenantQueryResult 租户查询结果
type TenantQueryResult struct {
	List       Tenants         `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}
