package service

import (
	"context"
	"strings"

	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/db/store"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/system"
)

// DeptService service layer
type DeptService struct {
	logger         lib.Logger
	txManager      lib.TxManager
	deptRepository repository.DeptRepository
}

// NewDeptService creates a new dept service
func NewDeptService(
	logger lib.Logger,
	txManager lib.TxManager,
	deptRepository repository.DeptRepository,
) DeptService {
	return DeptService{
		logger:         logger,
		txManager:      txManager,
		deptRepository: deptRepository,
	}
}

// GetDeptList 获取部门列表（树形）
func (a DeptService) GetDeptList(param *system.DeptQueryParam) ([]*system.DeptVO, error) {
	deptList, err := a.deptRepository.Query(param)
	if err != nil {
		return nil, err
	}

	if len(deptList) == 0 {
		return []*system.DeptVO{}, nil
	}

	// 获取所有部门ID
	deptIds := make(map[string]bool)
	for _, dept := range deptList {
		deptIds[dept.ID] = true
	}

	// 获取父节点ID
	parentIds := make(map[string]bool)
	for _, dept := range deptList {
		parentIds[dept.ParentID] = true
	}

	// 获取根节点ID（父节点ID中不包含在部门ID中的节点）
	var rootIds []string
	for parentId := range parentIds {
		if !deptIds[parentId] {
			rootIds = append(rootIds, parentId)
		}
	}

	// 预构建 parentID -> children 映射（O(n) 复杂度）
	childMap := buildDeptChildMap(deptList)

	var result []*system.DeptVO
	for _, rootId := range rootIds {
		children := buildDeptTree(rootId, childMap)
		result = append(result, children...)
	}

	return result, nil
}

// buildDeptTree 使用 map 预构建 O(n) 复杂度的部门树
func buildDeptTree(parentId string, childMap map[string][]*system.Dept) []*system.DeptVO {
	children, ok := childMap[parentId]
	if !ok {
		return nil
	}

	result := make([]*system.DeptVO, 0, len(children))
	for _, dept := range children {
		deptVO := &system.DeptVO{
			ID:         dept.ID,
			Name:       dept.Name,
			Code:       dept.Code,
			ParentID:   dept.ParentID,
			Sort:       dept.Sort,
			Status:     dept.Status,
			CreateTime: dept.CreateTime,
			UpdateTime: dept.UpdateTime,
		}
		subChildren := buildDeptTree(dept.ID, childMap)
		if len(subChildren) > 0 {
			deptVO.Children = subChildren
		}
		result = append(result, deptVO)
	}

	return result
}

// buildChildMap 预构建 parentID -> children 映射
func buildDeptChildMap(deptList system.Depts) map[string][]*system.Dept {
	childMap := make(map[string][]*system.Dept, len(deptList))
	for _, dept := range deptList {
		childMap[dept.ParentID] = append(childMap[dept.ParentID], dept)
	}
	return childMap
}

// ListDeptOptions 部门下拉选项
func (a DeptService) ListDeptOptions() ([]*system.DeptOption, error) {
	deptList, err := a.deptRepository.GetAllEnabled()
	if err != nil {
		return nil, err
	}

	if len(deptList) == 0 {
		return []*system.DeptOption{}, nil
	}

	// 获取所有部门ID
	deptIds := make(map[string]bool)
	for _, dept := range deptList {
		deptIds[dept.ID] = true
	}

	// 获取父节点ID
	parentIds := make(map[string]bool)
	for _, dept := range deptList {
		parentIds[dept.ParentID] = true
	}

	// 获取根节点ID
	var rootIds []string
	for parentId := range parentIds {
		if !deptIds[parentId] {
			rootIds = append(rootIds, parentId)
		}
	}

	// 预构建 parentID -> children 映射（O(n) 复杂度）
	childMap := buildDeptChildMap(deptList)

	var result []*system.DeptOption
	for _, rootId := range rootIds {
		children := buildDeptOptions(rootId, childMap)
		result = append(result, children...)
	}

	return result, nil
}

// buildDeptOptions 使用 map 预构建 O(n) 复杂度的部门下拉选项
func buildDeptOptions(parentId string, childMap map[string][]*system.Dept) []*system.DeptOption {
	children, ok := childMap[parentId]
	if !ok {
		return nil
	}

	result := make([]*system.DeptOption, 0, len(children))
	for _, dept := range children {
		option := &system.DeptOption{
			Value: dept.ID,
			Label: dept.Name,
		}
		subChildren := buildDeptOptions(dept.ID, childMap)
		if len(subChildren) > 0 {
			option.Children = subChildren
		}
		result = append(result, option)
	}

	return result
}

// SaveDept 新增部门
func (a DeptService) SaveDept(form *system.DeptForm, createdBy string) (string, error) {
	// 校验部门编号是否存在
	existDept, err := a.deptRepository.GetByCode(form.Code)
	if err != nil {
		return "", err
	}
	if existDept != nil {
		return "", errors.New("部门编号已存在")
	}

	// 生成部门路径
	parentID := form.ParentID
	treePath, err := a.generateDeptTreePath(parentID)
	if err != nil {
		return "", err
	}

	dept := &system.Dept{
		Name:     form.Name,
		Code:     form.Code,
		ParentID: parentID,
		TreePath: treePath,
		Sort:     form.Sort,
		Status:   form.Status,
		CreateBy: createdBy,
	}

	if err := a.deptRepository.Create(dept); err != nil {
		return "", err
	}

	return dept.ID, nil
}

// GetDeptForm 获取部门表单数据
func (a DeptService) GetDeptForm(id string) (*system.DeptForm, error) {
	dept, err := a.deptRepository.Get(id)
	if err != nil {
		return nil, err
	}

	return &system.DeptForm{
		ID:       dept.ID,
		Name:     dept.Name,
		Code:     dept.Code,
		ParentID: dept.ParentID,
		Sort:     dept.Sort,
		Status:   dept.Status,
	}, nil
}

// UpdateDept 更新部门
func (a DeptService) UpdateDept(id string, form *system.DeptForm, updatedBy string) (string, error) {
	// 检查部门是否存在
	_, err := a.deptRepository.Get(id)
	if err != nil {
		return "", err
	}

	// 校验部门编号是否存在（排除自身）
	existDept, err := a.deptRepository.GetByCode(form.Code, id)
	if err != nil {
		return "", err
	}
	if existDept != nil {
		return "", errors.New("部门编号已存在")
	}

	// 生成部门路径
	parentID := form.ParentID
	treePath, err := a.generateDeptTreePath(parentID)
	if err != nil {
		return "", err
	}

	dept := &system.Dept{
		ID:       id,
		Name:     form.Name,
		Code:     form.Code,
		ParentID: parentID,
		TreePath: treePath,
		Sort:     form.Sort,
		Status:   form.Status,
		UpdateBy: updatedBy,
	}

	if err := a.deptRepository.Update(id, dept); err != nil {
		return "", err
	}

	return id, nil
}

// DeleteByIds 删除部门
func (a DeptService) DeleteByIds(ids string, deletedBy string) error {
	if ids == "" {
		return errors.New("删除的部门数据为空")
	}

	idStrs := strings.Split(ids, ",")

	// Delete every requested department (and its subtree) atomically.
	return a.txManager.RunInTx(context.Background(), func(q store.Store) error {
		deptRepo := a.deptRepository.WithTx(q)
		for _, idStr := range idStrs {
			id := strings.TrimSpace(idStr)
			if id == "" {
				continue
			}

			// 删除部门及子部门
			if err := deptRepo.DeleteByTreePath(id, deletedBy); err != nil {
				return err
			}
		}
		return nil
	})
}

// generateDeptTreePath 生成部门路径
func (a DeptService) generateDeptTreePath(parentId string) (string, error) {
	if parentId == "" {
		return "0", nil
	}

	parent, err := a.deptRepository.Get(parentId)
	if err != nil {
		return "", err
	}

	return parent.TreePath + "," + parent.ID, nil
}
