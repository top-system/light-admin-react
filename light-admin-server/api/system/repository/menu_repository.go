package repository

import (
	"context"
	"errors"
	"time"

	"github.com/top-system/light-admin/db/store"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/pkg/uuid"
)

// MenuRepository is the sqlc/pgx-backed persistence layer for menus.
type MenuRepository struct {
	q      store.Store
	logger lib.Logger
}

// NewMenuRepository creates a new menu repository bound to the pool-level Queries.
func NewMenuRepository(q store.Store, logger lib.Logger) MenuRepository {
	return MenuRepository{
		q:      q,
		logger: logger,
	}
}

// WithTx returns a copy bound to the given transaction-scoped Queries.
func (a MenuRepository) WithTx(q store.Store) MenuRepository {
	a.q = q
	return a
}

// menuOrderMode maps an OrderParam to the whitelisted ordering understood by the
// ListMenus query: 1 = sort ASC (the only non-default ordering the app uses),
// 0 = the default id DESC.
func menuOrderMode(o dto.OrderParam) int32 {
	if o.Key == "sort" {
		return 1
	}
	return 0
}

func (a MenuRepository) Query(param *system.MenuQueryParam) (*system.MenuQueryResult, error) {
	ctx := context.Background()

	filter := newMenuFilter(param)

	total, err := a.q.CountMenus(ctx, filter.count())
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.Menus, 0)
	if total > 0 {
		limit, offset := pageBounds(param.PaginationParam)
		rows, err := a.q.ListMenus(ctx, filter.list(menuOrderMode(param.OrderParam), limit, offset))
		if err != nil {
			return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
		}
		for _, r := range rows {
			list = append(list, toDomainMenu(r))
		}
	}

	return &system.MenuQueryResult{
		Pagination: &dto.Pagination{
			PageNum:  param.PaginationParam.GetPageNum(),
			PageSize: param.PaginationParam.GetPageSize(),
			Total:    total,
		},
		List: list,
	}, nil
}

func (a MenuRepository) Get(id string) (*system.Menu, error) {
	row, err := a.q.GetMenu(context.Background(), id)
	if err != nil {
		if errors.Is(err, store.ErrNoRows) {
			return nil, apperrors.DatabaseRecordNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainMenu(row), nil
}

// Create inserts a new menu, assigning a UUID when the ID is empty.
func (a MenuRepository) Create(menu *system.Menu) error {
	if menu.ID == "" {
		menu.ID = uuid.NewID()
	}

	err := a.q.CreateMenu(context.Background(), store.CreateMenuParams{
		ID:         menu.ID,
		ParentID:   menu.ParentID,
		TreePath:   menu.TreePath,
		Name:       menu.Name,
		Type:       int32(menu.Type),
		RouteName:  menu.RouteName,
		RoutePath:  menu.RoutePath,
		Component:  menu.Component,
		Perm:       menu.Perm,
		AlwaysShow: int32(menu.AlwaysShow),
		KeepAlive:  int32(menu.KeepAlive),
		Visible:    int32(menu.Visible),
		Sort:       int32(menu.Sort),
		Icon:       menu.Icon,
		Redirect:   menu.Redirect,
		Params:     menu.Params,
		Now:        time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Update updates the mutable fields of a menu (mirrors the previous Select-scoped
// GORM update).
func (a MenuRepository) Update(id string, menu *system.Menu) error {
	err := a.q.UpdateMenu(context.Background(), store.UpdateMenuParams{
		ID:         id,
		ParentID:   menu.ParentID,
		TreePath:   menu.TreePath,
		Name:       menu.Name,
		Type:       int32(menu.Type),
		RouteName:  menu.RouteName,
		RoutePath:  menu.RoutePath,
		Component:  menu.Component,
		Perm:       menu.Perm,
		AlwaysShow: int32(menu.AlwaysShow),
		KeepAlive:  int32(menu.KeepAlive),
		Visible:    int32(menu.Visible),
		Sort:       int32(menu.Sort),
		Icon:       menu.Icon,
		Redirect:   menu.Redirect,
		Params:     menu.Params,
		Now:        time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Delete hard-deletes a menu (t_menu has no soft-delete column).
func (a MenuRepository) Delete(id string) error {
	if err := a.q.DeleteMenu(context.Background(), id); err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a MenuRepository) UpdateVisible(id string, visible int) error {
	err := a.q.UpdateMenuVisible(context.Background(), store.UpdateMenuVisibleParams{
		ID:      id,
		Visible: int32(visible),
		Now:     time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a MenuRepository) UpdateTreePath(id string, treePath string) error {
	err := a.q.UpdateMenuTreePath(context.Background(), store.UpdateMenuTreePathParams{
		ID:       id,
		TreePath: treePath,
		Now:      time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// GetMenusByRoleIDs 根据角色ID列表获取菜单
func (a MenuRepository) GetMenusByRoleIDs(roleIDs []string) (system.Menus, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}

	rows, err := a.q.ListMenusByRoleIDs(context.Background(), roleIDs)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.Menus, 0, len(rows))
	for _, r := range rows {
		list = append(list, toDomainMenu(r))
	}
	return list, nil
}

// GetButtonPermsByRoleIDs 获取角色关联的按钮权限标识列表
func (a MenuRepository) GetButtonPermsByRoleIDs(roleIDs []string) ([]string, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}

	perms, err := a.q.ListButtonPermsByRoleIDs(context.Background(), roleIDs)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return perms, nil
}

// menuFilter translates a MenuQueryParam into sqlc filter arguments once, so the
// list and count queries stay in sync.
type menuFilter struct {
	ids            []string
	name           *string
	parentID       *string
	prefixTreePath *string
	menuType       *int32
	visible        *int32
	keywords       *string
}

func newMenuFilter(param *system.MenuQueryParam) menuFilter {
	f := menuFilter{}
	if len(param.IDs) > 0 {
		f.ids = param.IDs
	}
	if param.Name != "" {
		f.name = ptr(param.Name)
	}
	// ParentID is a *string: a non-nil value (including "") filters parent_id,
	// while nil disables the predicate — preserving the previous behaviour.
	if param.ParentID != nil {
		f.parentID = ptr(*param.ParentID)
	}
	if param.PrefixTreePath != "" {
		f.prefixTreePath = ptr(param.PrefixTreePath + "%")
	}
	if param.Type != 0 {
		f.menuType = ptr(int32(param.Type))
	}
	if param.Visible != 0 {
		f.visible = ptr(int32(param.Visible))
	}
	if param.Keywords != "" {
		f.keywords = ptr("%" + param.Keywords + "%")
	}
	return f
}

func (f menuFilter) list(orderBy int32, limit, offset *int32) store.ListMenusParams {
	return store.ListMenusParams{
		Ids:            f.ids,
		Name:           f.name,
		ParentID:       f.parentID,
		PrefixTreePath: f.prefixTreePath,
		Type:           f.menuType,
		Visible:        f.visible,
		Keywords:       f.keywords,
		OrderBy:        orderBy,
		Limit:          limit,
		Offset:         offset,
	}
}

func (f menuFilter) count() store.CountMenusParams {
	return store.CountMenusParams{
		Ids:            f.ids,
		Name:           f.name,
		ParentID:       f.parentID,
		PrefixTreePath: f.prefixTreePath,
		Type:           f.menuType,
		Visible:        f.visible,
		Keywords:       f.keywords,
	}
}

func toDomainMenu(r store.TMenu) *system.Menu {
	return &system.Menu{
		ID:         r.ID,
		ParentID:   r.ParentID,
		TreePath:   r.TreePath,
		Name:       r.Name,
		Type:       int(r.Type),
		RouteName:  r.RouteName,
		RoutePath:  r.RoutePath,
		Component:  r.Component,
		Perm:       r.Perm,
		AlwaysShow: int(r.AlwaysShow),
		KeepAlive:  int(r.KeepAlive),
		Visible:    int(r.Visible),
		Sort:       int(r.Sort),
		Icon:       r.Icon,
		Redirect:   r.Redirect,
		Params:     r.Params,
		CreateTime: dto.DateTime(r.CreateTime),
		UpdateTime: dto.DateTime(r.UpdateTime),
	}
}
