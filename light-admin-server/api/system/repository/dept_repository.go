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

// DeptRepository is the sqlc/pgx-backed persistence layer for departments.
type DeptRepository struct {
	q      store.Store
	logger lib.Logger
}

// NewDeptRepository creates a new dept repository bound to the pool-level Queries.
func NewDeptRepository(q store.Store, logger lib.Logger) DeptRepository {
	return DeptRepository{
		q:      q,
		logger: logger,
	}
}

// WithTx returns a copy bound to the given transaction-scoped Queries.
func (a DeptRepository) WithTx(q store.Store) DeptRepository {
	a.q = q
	return a
}

// Query 查询部门列表. Ordering is fixed to sort ASC (the previous default).
func (a DeptRepository) Query(param *system.DeptQueryParam) (system.Depts, error) {
	var keywords *string
	if param.Keywords != "" {
		keywords = ptr("%" + param.Keywords + "%")
	}
	var status *int32
	if param.Status != nil {
		status = ptr(int32(*param.Status))
	}

	rows, err := a.q.ListDepts(context.Background(), store.ListDeptsParams{
		Keywords: keywords,
		Status:   status,
	})
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.Depts, 0, len(rows))
	for _, r := range rows {
		list = append(list, toDomainDept(r))
	}
	return list, nil
}

// Get 获取部门
func (a DeptRepository) Get(id string) (*system.Dept, error) {
	row, err := a.q.GetDept(context.Background(), id)
	if err != nil {
		if errors.Is(err, store.ErrNoRows) {
			return nil, apperrors.DatabaseRecordNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainDept(row), nil
}

// GetByCode 根据编码获取部门. Returns (nil, nil) when no matching department
// exists, preserving the previous behaviour used by uniqueness checks.
func (a DeptRepository) GetByCode(code string, excludeID ...string) (*system.Dept, error) {
	params := store.GetDeptByCodeParams{Code: code}
	if len(excludeID) > 0 && excludeID[0] != "" {
		params.ExcludeID = ptr(excludeID[0])
	}

	row, err := a.q.GetDeptByCode(context.Background(), params)
	if err != nil {
		if errors.Is(err, store.ErrNoRows) {
			return nil, nil
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainDept(row), nil
}

// Create 创建部门, assigning a UUID when the ID is empty.
func (a DeptRepository) Create(dept *system.Dept) error {
	if dept.ID == "" {
		dept.ID = uuid.NewID()
	}

	err := a.q.CreateDept(context.Background(), store.CreateDeptParams{
		ID:        dept.ID,
		Name:      dept.Name,
		Code:      dept.Code,
		ParentID:  dept.ParentID,
		TreePath:  dept.TreePath,
		Sort:      int32(dept.Sort),
		Status:    int32(dept.Status),
		CreateBy:  dept.CreateBy,
		UpdateBy:  dept.UpdateBy,
		IsDeleted: int32(dept.IsDeleted),
		Now:       time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Update 更新部门 (name, code, parent_id, tree_path, sort, status, update_by).
func (a DeptRepository) Update(id string, dept *system.Dept) error {
	err := a.q.UpdateDept(context.Background(), store.UpdateDeptParams{
		ID:       id,
		Name:     dept.Name,
		Code:     dept.Code,
		ParentID: dept.ParentID,
		TreePath: dept.TreePath,
		Sort:     int32(dept.Sort),
		Status:   int32(dept.Status),
		UpdateBy: dept.UpdateBy,
		Now:      time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Delete 删除部门（软删除）
func (a DeptRepository) Delete(id string, deletedBy string) error {
	err := a.q.SoftDeleteDept(context.Background(), store.SoftDeleteDeptParams{
		ID:       id,
		UpdateBy: deletedBy,
		Now:      time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// DeleteByTreePath 根据tree_path删除部门及子部门
func (a DeptRepository) DeleteByTreePath(deptId string, deletedBy string) error {
	err := a.q.SoftDeleteDeptByTreePath(context.Background(), store.SoftDeleteDeptByTreePathParams{
		ID:           deptId,
		UpdateBy:     deletedBy,
		TreePathLike: "%," + deptId + ",%",
		Now:          time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// GetAllEnabled 获取所有启用的部门
func (a DeptRepository) GetAllEnabled() (system.Depts, error) {
	rows, err := a.q.ListEnabledDepts(context.Background())
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.Depts, 0, len(rows))
	for _, r := range rows {
		list = append(list, toDomainDept(r))
	}
	return list, nil
}

// GetByIDs 根据ID列表获取部门Map
func (a DeptRepository) GetByIDs(ids []string) (map[string]*system.Dept, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	rows, err := a.q.ListDeptsByIDs(context.Background(), ids)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	result := make(map[string]*system.Dept, len(rows))
	for _, r := range rows {
		dept := toDomainDept(r)
		result[dept.ID] = dept
	}
	return result, nil
}

func toDomainDept(r store.TDept) *system.Dept {
	return &system.Dept{
		ID:         r.ID,
		Name:       r.Name,
		Code:       r.Code,
		ParentID:   r.ParentID,
		TreePath:   r.TreePath,
		Sort:       int(r.Sort),
		Status:     int(r.Status),
		CreateBy:   r.CreateBy,
		CreateTime: dto.DateTime(r.CreateTime),
		UpdateBy:   r.UpdateBy,
		UpdateTime: dto.DateTime(r.UpdateTime),
		IsDeleted:  int(r.IsDeleted),
	}
}
