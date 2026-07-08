package service

import (
	"context"

	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/db/store"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
)

// RoleService service layer
type RoleService struct {
	logger             lib.Logger
	txManager          lib.TxManager
	userRepository     repository.UserRepository
	roleRepository     repository.RoleRepository
	roleMenuRepository repository.RoleMenuRepository
	menuRepository     repository.MenuRepository
	permissionCache    PermissionCache
}

// NewRoleService creates a new role service
func NewRoleService(
	logger lib.Logger,
	txManager lib.TxManager,
	userRepository repository.UserRepository,
	roleRepository repository.RoleRepository,
	roleMenuRepository repository.RoleMenuRepository,
	menuRepository repository.MenuRepository,
	permissionCache PermissionCache,
) RoleService {
	return RoleService{
		logger:             logger,
		txManager:          txManager,
		userRepository:     userRepository,
		roleRepository:     roleRepository,
		roleMenuRepository: roleMenuRepository,
		menuRepository:     menuRepository,
		permissionCache:    permissionCache,
	}
}

func (a RoleService) Query(param *system.RoleQueryParam) (roleQR *system.RoleQueryResult, err error) {
	return a.roleRepository.Query(param)
}

func (a RoleService) Get(id string) (*system.Role, error) {
	role, err := a.roleRepository.Get(id)
	if err != nil {
		return nil, err
	}

	// Get role menu IDs
	menuIDs, err := a.roleMenuRepository.GetMenuIDsByRoleID(id)
	if err != nil {
		return nil, err
	}
	role.MenuIds = menuIDs

	return role, nil
}

func (a RoleService) GetByCode(code string) (*system.Role, error) {
	return a.roleRepository.GetByCode(code)
}

func (a RoleService) CheckName(item *system.Role) error {
	qr, err := a.roleRepository.Query(&system.RoleQueryParam{Name: item.Name})
	if err != nil {
		return err
	}

	for _, role := range qr.List {
		if role.ID != item.ID {
			return errors.RoleAlreadyExists
		}
	}

	return nil
}

func (a RoleService) CheckCode(item *system.Role) error {
	qr, err := a.roleRepository.Query(&system.RoleQueryParam{Code: item.Code})
	if err != nil {
		return err
	}

	for _, role := range qr.List {
		if role.ID != item.ID {
			return errors.RoleCodeAlreadyExists
		}
	}

	return nil
}

func (a RoleService) Create(role *system.Role) (string, error) {
	if err := a.CheckName(role); err != nil {
		return "", err
	}

	if err := a.CheckCode(role); err != nil {
		return "", err
	}

	// Create the role and its menu associations atomically.
	err := a.txManager.RunInTx(context.Background(), func(q store.Store) error {
		roleRepo := a.roleRepository.WithTx(q)
		if err := roleRepo.Create(role); err != nil {
			return err
		}

		if len(role.MenuIds) > 0 {
			roleMenuRepo := a.roleMenuRepository.WithTx(q)
			if err := a.assignMenusToRole(roleMenuRepo, role.ID, role.MenuIds); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return role.ID, nil
}

func (a RoleService) Update(id string, role *system.Role) error {
	oRole, err := a.Get(id)
	if err != nil {
		return err
	}

	role.ID = id

	if role.Name != oRole.Name {
		if err = a.CheckName(role); err != nil {
			return err
		}
	}

	if role.Code != oRole.Code {
		if err = a.CheckCode(role); err != nil {
			return err
		}
	}

	if err := a.roleRepository.Update(id, role); err != nil {
		return err
	}

	// 清除该角色相关用户的权限缓存
	a.permissionCache.InvalidateRoleCache(id)

	return nil
}

func (a RoleService) Delete(id string) error {
	_, err := a.roleRepository.Get(id)
	if err != nil {
		return err
	}

	userQR, err := a.userRepository.Query(&system.UserQueryParam{
		RoleIDs: []string{id},
	})

	if err != nil {
		return err
	} else if userQR.Pagination.Total > 0 {
		return errors.RoleNotAllowDeleteWithUser
	}

	// 先清除该角色相关用户的权限缓存
	a.permissionCache.InvalidateRoleCache(id)

	// Remove menu associations and soft-delete the role atomically.
	return a.txManager.RunInTx(context.Background(), func(q store.Store) error {
		if err := a.roleMenuRepository.WithTx(q).DeleteByRoleID(id); err != nil {
			return err
		}
		return a.roleRepository.WithTx(q).Delete(id)
	})
}

func (a RoleService) UpdateStatus(id string, status int) error {
	_, err := a.roleRepository.Get(id)
	if err != nil {
		return err
	}

	return a.roleRepository.UpdateStatus(id, status)
}

// GetRoleMenuIds 获取角色的菜单ID列表
func (a RoleService) GetRoleMenuIds(roleID string) ([]string, error) {
	return a.roleMenuRepository.GetMenuIDsByRoleID(roleID)
}

// AssignMenusToRole 为角色分配菜单
func (a RoleService) AssignMenusToRole(roleID string, menuIDs []string) error {
	_, err := a.roleRepository.Get(roleID)
	if err != nil {
		return err
	}

	// Replace the role's menu associations atomically.
	err = a.txManager.RunInTx(context.Background(), func(q store.Store) error {
		roleMenuRepo := a.roleMenuRepository.WithTx(q)

		// Delete existing associations
		if err := roleMenuRepo.DeleteByRoleID(roleID); err != nil {
			return err
		}

		return a.assignMenusToRole(roleMenuRepo, roleID, menuIDs)
	})
	if err != nil {
		return err
	}

	// 清除该角色相关用户的权限缓存
	a.permissionCache.InvalidateRoleCache(roleID)

	return nil
}

// assignMenusToRole inserts the given menu associations using the supplied
// (possibly transaction-bound) role-menu repository.
func (a RoleService) assignMenusToRole(roleMenuRepo repository.RoleMenuRepository, roleID string, menuIDs []string) error {
	if len(menuIDs) == 0 {
		return nil
	}

	roleMenus := make([]*system.RoleMenu, 0, len(menuIDs))
	for _, menuID := range menuIDs {
		roleMenus = append(roleMenus, &system.RoleMenu{
			RoleID: roleID,
			MenuID: menuID,
		})
	}

	return roleMenuRepo.BatchCreate(roleMenus)
}

// ListRoleOptions 获取角色下拉选项
func (a RoleService) ListRoleOptions() ([]system.RoleOption, error) {
	qr, err := a.roleRepository.Query(&system.RoleQueryParam{
		Status:          1,
		PaginationParam: dto.PaginationParam{PageSize: 1000, PageNum: 1},
	})

	if err != nil {
		return nil, err
	}

	return qr.List.ToOptions(), nil
}
