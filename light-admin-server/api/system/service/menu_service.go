package service

import (
	"context"
	"sort"

	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/db/store"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
)

// MenuService service layer
type MenuService struct {
	logger             lib.Logger
	txManager          lib.TxManager
	menuRepository     repository.MenuRepository
	roleMenuRepository repository.RoleMenuRepository
}

// NewMenuService creates a new menu service
func NewMenuService(
	logger lib.Logger,
	txManager lib.TxManager,
	menuRepository repository.MenuRepository,
	roleMenuRepository repository.RoleMenuRepository,
) MenuService {
	return MenuService{
		logger:             logger,
		txManager:          txManager,
		menuRepository:     menuRepository,
		roleMenuRepository: roleMenuRepository,
	}
}

func (a MenuService) Check(item *system.Menu) error {
	result, err := a.menuRepository.Query(&system.MenuQueryParam{
		Name:     item.Name,
		ParentID: &item.ParentID,
	})

	if err != nil {
		return err
	} else if len(result.List) > 0 {
		return errors.MenuAlreadyExists
	}

	return nil
}

func (a MenuService) Query(param *system.MenuQueryParam) (*system.MenuQueryResult, error) {
	return a.menuRepository.Query(param)
}

func (a MenuService) Get(id string) (*system.Menu, error) {
	return a.menuRepository.Get(id)
}

func (a MenuService) Create(menu *system.Menu) (string, error) {
	if err := a.Check(menu); err != nil {
		return "", err
	}

	var err error
	if menu.TreePath, err = a.GetTreePath(menu.ParentID); err != nil {
		return "", err
	}

	if err = a.menuRepository.Create(menu); err != nil {
		return "", err
	}

	return menu.ID, nil
}

func (a MenuService) CreateMenus(parentID string, mTrees system.MenuTrees) error {
	for _, mTree := range mTrees {
		// 处理 visible 字段：nil 时默认为 1（按钮除外），显式设置时使用设置的值
		visible := 1
		if mTree.Visible != nil {
			visible = *mTree.Visible
		} else if mTree.Type.Value() == constants.MenuTypeButton {
			visible = 0
		}

		menu := &system.Menu{
			ParentID:   parentID,
			Name:       mTree.Name,
			Type:       mTree.Type.Value(),
			RouteName:  mTree.RouteName,
			RoutePath:  mTree.RoutePath,
			Component:  mTree.Component,
			Perm:       mTree.Perm,
			AlwaysShow: mTree.AlwaysShow,
			KeepAlive:  mTree.KeepAlive,
			Visible:    visible,
			Sort:       mTree.Sort,
			Icon:       mTree.Icon,
			Redirect:   mTree.Redirect,
			Params:     mTree.Params,
		}

		menuID, err := a.Create(menu)
		if err != nil {
			// If menu already exists, find its ID and continue with children
			if err == errors.MenuAlreadyExists {
				existingMenu, findErr := a.FindByNameAndParent(mTree.Name, parentID)
				if findErr != nil {
					return findErr
				}
				menuID = existingMenu.ID
			} else {
				return err
			}
		}

		if mTree.Children != nil && len(mTree.Children) > 0 {
			err := a.CreateMenus(menuID, mTree.Children)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a MenuService) FindByNameAndParent(name string, parentID string) (*system.Menu, error) {
	result, err := a.menuRepository.Query(&system.MenuQueryParam{
		Name:     name,
		ParentID: &parentID,
	})
	if err != nil {
		return nil, err
	}
	if len(result.List) == 0 {
		return nil, errors.DatabaseRecordNotFound
	}
	return result.List[0], nil
}

func (a MenuService) Update(id string, menu *system.Menu) error {
	if id == menu.ParentID {
		return errors.MenuInvalidParent
	}

	// get old menu
	oMenu, err := a.Get(id)
	if err != nil {
		return err
	} else if oMenu.Name != menu.Name {
		if err = a.Check(menu); err != nil {
			return err
		}
	}

	menu.ID = oMenu.ID
	menu.CreateTime = oMenu.CreateTime

	if menu.ParentID != oMenu.ParentID {
		treePath, err := a.GetTreePath(menu.ParentID)
		if err != nil {
			return err
		}

		menu.TreePath = treePath
	} else {
		menu.TreePath = oMenu.TreePath
	}

	// Rewriting descendant tree paths and updating the menu itself must be atomic.
	return a.txManager.RunInTx(context.Background(), func(q store.Store) error {
		menuRepo := a.menuRepository.WithTx(q)

		if err := a.updateChildTreePath(menuRepo, oMenu, menu); err != nil {
			return err
		}

		return menuRepo.Update(id, menu)
	})
}

func (a MenuService) Delete(id string) error {
	_, err := a.menuRepository.Get(id)
	if err != nil {
		return err
	}

	menuQR, err := a.menuRepository.Query(&system.MenuQueryParam{
		ParentID: &id,
	})

	if err != nil {
		return err
	} else if menuQR.Pagination.Total > 0 {
		return errors.MenuNotAllowDeleteWithChild
	}

	// Remove role_menu associations and delete the menu atomically.
	return a.txManager.RunInTx(context.Background(), func(q store.Store) error {
		if err := a.roleMenuRepository.WithTx(q).DeleteByMenuID(id); err != nil {
			return err
		}
		return a.menuRepository.WithTx(q).Delete(id)
	})
}

func (a MenuService) UpdateVisible(id string, visible int) error {
	_, err := a.menuRepository.Get(id)
	if err != nil {
		return err
	}

	return a.menuRepository.UpdateVisible(id, visible)
}

func (a MenuService) GetTreePath(parentID string) (string, error) {
	if parentID == "" {
		return "", nil
	}

	parentMenu, err := a.menuRepository.Get(parentID)
	if err != nil {
		return "", err
	}

	return a.JoinTreePath(parentMenu.TreePath, parentMenu.ID), nil
}

func (a MenuService) JoinTreePath(parent string, id string) string {
	idStr := id
	if parent != "" {
		return parent + "," + idStr
	}

	return idStr
}

// updateChildTreePath rewrites the tree_path of every descendant when a menu is
// re-parented. It runs on the supplied (transaction-bound) menu repository so the
// rewrites share the caller's transaction.
func (a MenuService) updateChildTreePath(menuRepo repository.MenuRepository, oMenu, nMenu *system.Menu) error {
	if oMenu.ParentID == nMenu.ParentID {
		return nil
	}

	oPath := a.JoinTreePath(oMenu.TreePath, oMenu.ID)
	menuQR, err := menuRepo.Query(&system.MenuQueryParam{
		PrefixTreePath: oPath,
	})

	if err != nil {
		return err
	}

	nPath := a.JoinTreePath(nMenu.TreePath, nMenu.ID)
	for _, menu := range menuQR.List {
		err = menuRepo.UpdateTreePath(menu.ID, nPath+menu.TreePath[len(oPath):])
		if err != nil {
			return err
		}
	}

	return nil
}

// ListMenuOptions 获取菜单下拉选项（用于父级菜单选择）
func (a MenuService) ListMenuOptions(onlyParent bool) ([]dto.MenuOption, error) {
	param := &system.MenuQueryParam{
		PaginationParam: dto.PaginationParam{PageNum: 1, PageSize: 1000},
		OrderParam:      dto.OrderParam{Key: "sort", Direction: dto.OrderByASC},
	}

	menuQR, err := a.menuRepository.Query(param)
	if err != nil {
		return nil, err
	}

	// Filter out buttons if onlyParent
	menus := menuQR.List
	if onlyParent {
		filtered := make(system.Menus, 0)
		for _, m := range menus {
			if m.Type != constants.MenuTypeButton {
				filtered = append(filtered, m)
			}
		}
		menus = filtered
	}

	// 预构建 parentID -> children 映射（O(n) 复杂度）
	childMap := buildMenuChildMap(menus)
	return buildMenuOptions("", childMap), nil
}

// buildMenuChildMap 预构建 parentID -> children 映射
func buildMenuChildMap(menus system.Menus) map[string][]*system.Menu {
	childMap := make(map[string][]*system.Menu, len(menus))
	for _, menu := range menus {
		childMap[menu.ParentID] = append(childMap[menu.ParentID], menu)
	}
	return childMap
}

func buildMenuOptions(parentID string, childMap map[string][]*system.Menu) []dto.MenuOption {
	children, ok := childMap[parentID]
	if !ok {
		return nil
	}

	options := make([]dto.MenuOption, 0, len(children))
	for _, menu := range children {
		option := dto.MenuOption{
			Value: menu.ID,
			Label: menu.Name,
		}

		subChildren := buildMenuOptions(menu.ID, childMap)
		if len(subChildren) > 0 {
			option.Children = subChildren
		}

		options = append(options, option)
	}

	return options
}

// GetUserRoutes 获取用户的路由列表
func (a MenuService) GetUserRoutes(roleIDs []string, isSuperAdmin bool) ([]*dto.RouteVO, error) {
	var menus system.Menus
	var err error

	if isSuperAdmin {
		// 超级管理员获取所有菜单（包括隐藏的，前端自己判断）
		menuQR, err := a.menuRepository.Query(&system.MenuQueryParam{
			PaginationParam: dto.PaginationParam{PageNum: 1, PageSize: 1000},
			OrderParam:      dto.OrderParam{Key: "sort", Direction: dto.OrderByASC},
		})
		if err != nil {
			return nil, err
		}
		menus = menuQR.List
	} else {
		// 普通用户根据角色获取菜单
		menus, err = a.menuRepository.GetMenusByRoleIDs(roleIDs)
		if err != nil {
			return nil, err
		}

		// 补充父级菜单
		menus, err = a.fillParentMenus(menus)
		if err != nil {
			return nil, err
		}
	}

	// 过滤出非按钮类型的菜单用于构建路由（隐藏菜单也返回，前端自己判断）
	routeMenus := make(system.Menus, 0)
	for _, menu := range menus {
		if menu.Type != constants.MenuTypeButton {
			routeMenus = append(routeMenus, menu)
		}
	}

	// 按 sort 排序
	sort.Slice(routeMenus, func(i, j int) bool {
		return routeMenus[i].Sort < routeMenus[j].Sort
	})

	// 预构建 parentID -> children 映射（O(n) 复杂度）
	routeChildMap := buildMenuChildMap(routeMenus)
	return a.buildRoutes(routeChildMap, ""), nil
}

func (a MenuService) fillParentMenus(menus system.Menus) (system.Menus, error) {
	menuMap := menus.ToMap()
	parentIDs := menus.SplitParentIDs()

	var missingIDs []string
	for _, parentID := range parentIDs {
		if _, ok := menuMap[parentID]; !ok {
			missingIDs = append(missingIDs, parentID)
		}
	}

	if len(missingIDs) > 0 {
		parentMenuQR, err := a.menuRepository.Query(&system.MenuQueryParam{
			IDs: missingIDs,
		})
		if err != nil {
			return nil, err
		}
		menus = append(menus, parentMenuQR.List...)
	}

	return menus, nil
}

func (a MenuService) buildRoutes(childMap map[string][]*system.Menu, parentID string) []*dto.RouteVO {
	menuChildren, ok := childMap[parentID]
	if !ok {
		return nil
	}

	routes := make([]*dto.RouteVO, 0, len(menuChildren))

	for _, menu := range menuChildren {
		children := a.buildRoutes(childMap, menu.ID)

		// 顶级路由处理
		if parentID == "" {
			route := a.buildTopLevelRoute(menu, children)
			routes = append(routes, route)
		} else {
			// 非顶级路由，正常构建
			route := &dto.RouteVO{
				Name:      menu.RouteName,
				Path:      menu.RoutePath,
				Component: menu.Component,
				Redirect:  menu.Redirect,
				Meta: dto.RouteMeta{
					Title:      menu.Name,
					Icon:       menu.Icon,
					Hidden:     menu.Visible != 1,
					KeepAlive:  menu.KeepAlive == 1,
					AlwaysShow: menu.AlwaysShow == 1,
					Params:     menu.GetParamsMap(),
				},
			}
			if len(children) > 0 {
				route.Children = children
			}
			routes = append(routes, route)
		}
	}

	return routes
}

// buildTopLevelRoute 构建顶级路由（需要 Layout 包裹）
func (a MenuService) buildTopLevelRoute(menu *system.Menu, children []*dto.RouteVO) *dto.RouteVO {
	// 如果是目录类型（type=2）或者已有子路由，使用 Layout 包裹
	if menu.Type == constants.MenuTypeCatalog || len(children) > 0 {
		route := &dto.RouteVO{
			Name:      menu.RouteName,
			Path:      menu.RoutePath,
			Component: "Layout",
			Redirect:  menu.Redirect,
			Meta: dto.RouteMeta{
				Title:      menu.Name,
				Icon:       menu.Icon,
				Hidden:     menu.Visible != 1,
				KeepAlive:  menu.KeepAlive == 1,
				AlwaysShow: menu.AlwaysShow == 1,
				Params:     menu.GetParamsMap(),
			},
			Children: children,
		}
		return route
	}

	// 如果是菜单类型（type=1）且没有子路由，需要创建一个包含 Layout 和子路由的结构
	// 例如：Dashboard 页面需要包裹在 Layout 中
	childRoute := &dto.RouteVO{
		Name:      menu.RouteName + "Index",
		Path:      "index",
		Component: menu.Component,
		Meta: dto.RouteMeta{
			Title:      menu.Name,
			Icon:       menu.Icon,
			Hidden:     menu.Visible != 1,
			KeepAlive:  menu.KeepAlive == 1,
			AlwaysShow: menu.AlwaysShow == 1,
			Params:     menu.GetParamsMap(),
		},
	}

	redirect := menu.Redirect
	if redirect == "" {
		redirect = menu.RoutePath + "/index"
	}

	route := &dto.RouteVO{
		Name:      menu.RouteName,
		Path:      menu.RoutePath,
		Component: "Layout",
		Redirect:  redirect,
		Meta: dto.RouteMeta{
			Title:      menu.Name,
			Icon:       menu.Icon,
			Hidden:     menu.Visible != 1,
			KeepAlive:  menu.KeepAlive == 1,
			AlwaysShow: menu.AlwaysShow == 1,
			Params:     menu.GetParamsMap(),
		},
		Children: []*dto.RouteVO{childRoute},
	}

	return route
}
