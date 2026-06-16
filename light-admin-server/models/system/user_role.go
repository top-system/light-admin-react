package system

import (
	"github.com/top-system/light-admin/models/dto"
)

// UserRole 用户角色关联模型
type UserRole struct {
	UserID string `gorm:"column:user_id;type:char(32);primaryKey" json:"userId"`
	RoleID string `gorm:"column:role_id;type:char(32);primaryKey" json:"roleId"`
}

// TableName 指定表名
func (UserRole) TableName() string {
	return "t_user_role"
}

type UserRoles []*UserRole

type UserRoleQueryParam struct {
	dto.PaginationParam
	dto.OrderParam

	UserID  string
	UserIDs []string
}

type UserRoleQueryResult struct {
	List       UserRoles       `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

func (a UserRoles) ToMap() map[string]*UserRole {
	m := make(map[string]*UserRole)
	for _, item := range a {
		m[item.RoleID] = item
	}
	return m
}

func (a UserRoles) ToRoleIDs() []string {
	list := make([]string, len(a))
	for i, item := range a {
		list[i] = item.RoleID
	}
	return list
}

func (a UserRoles) ToUserIDMap() map[string]UserRoles {
	m := make(map[string]UserRoles)
	for _, item := range a {
		m[item.UserID] = append(m[item.UserID], item)
	}
	return m
}
