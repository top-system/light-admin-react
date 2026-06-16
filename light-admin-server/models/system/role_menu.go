package system

import (
	"github.com/top-system/light-admin/models/dto"
)

// RoleMenu 角色菜单关联模型
type RoleMenu struct {
	RoleID string `gorm:"column:role_id;type:char(32);not null;uniqueIndex:uk_roleid_menuid" json:"roleId"`
	MenuID string `gorm:"column:menu_id;type:char(32);not null;uniqueIndex:uk_roleid_menuid" json:"menuId"`
}

// TableName 指定表名
func (RoleMenu) TableName() string {
	return "t_role_menu"
}

type RoleMenus []*RoleMenu

type RoleMenuQueryParam struct {
	dto.PaginationParam
	dto.OrderParam

	RoleID  string
	RoleIDs []string
}

type RoleMenuQueryResult struct {
	List       RoleMenus       `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

func (a RoleMenus) ToMap() map[string]*RoleMenu {
	m := make(map[string]*RoleMenu)
	for _, item := range a {
		m[item.MenuID] = item
	}
	return m
}

func (a RoleMenus) ToRoleIDMap() map[string]RoleMenus {
	m := make(map[string]RoleMenus)
	for _, item := range a {
		m[item.RoleID] = append(m[item.RoleID], item)
	}
	return m
}

func (a RoleMenus) ToMenuIDs() []string {
	var idList []string
	m := make(map[string]struct{})

	for _, item := range a {
		if _, ok := m[item.MenuID]; ok {
			continue
		}
		idList = append(idList, item.MenuID)
		m[item.MenuID] = struct{}{}
	}

	return idList
}
