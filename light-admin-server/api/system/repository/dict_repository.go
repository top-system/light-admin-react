package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/top-system/light-admin/db/sqlc"
	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
	"github.com/top-system/light-admin/pkg/uuid"
)

// DictRepository is the sqlc/pgx-backed persistence layer for dictionaries.
type DictRepository struct {
	q      sqlc.Querier
	logger lib.Logger
}

// NewDictRepository creates a new dict repository bound to the pool-level Queries.
func NewDictRepository(q *sqlc.Queries, logger lib.Logger) DictRepository {
	return DictRepository{
		q:      q,
		logger: logger,
	}
}

// WithTx returns a copy bound to the given transaction-scoped Queries.
func (a DictRepository) WithTx(q *sqlc.Queries) DictRepository {
	a.q = q
	return a
}

// Query 查询字典分页列表. Ordering is fixed to create_time DESC (the previous default).
func (a DictRepository) Query(param *system.DictQueryParam) (*system.DictQueryResult, error) {
	ctx := context.Background()

	var keywords *string
	if param.Keywords != "" {
		keywords = ptr("%" + param.Keywords + "%")
	}

	total, err := a.q.CountDicts(ctx, keywords)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.Dicts, 0)
	if total > 0 {
		limit, offset := pageBounds(param.PaginationParam)
		rows, err := a.q.ListDicts(ctx, sqlc.ListDictsParams{Keywords: keywords, Limit: limit, Offset: offset})
		if err != nil {
			return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
		}
		for _, r := range rows {
			list = append(list, toDomainDict(r))
		}
	}

	return &system.DictQueryResult{
		Pagination: &dto.Pagination{
			PageNum:  param.PaginationParam.GetPageNum(),
			PageSize: param.PaginationParam.GetPageSize(),
			Total:    total,
		},
		List: list,
	}, nil
}

// GetAll 获取所有启用的字典
func (a DictRepository) GetAll() (system.Dicts, error) {
	rows, err := a.q.ListEnabledDicts(context.Background())
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.Dicts, 0, len(rows))
	for _, r := range rows {
		list = append(list, toDomainDict(r))
	}
	return list, nil
}

// Get 获取字典
func (a DictRepository) Get(id string) (*system.Dict, error) {
	row, err := a.q.GetDict(context.Background(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.DatabaseRecordNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainDict(row), nil
}

// GetByCode 根据编码获取字典. Returns (nil, nil) when not found, preserving the
// previous uniqueness-check behaviour.
func (a DictRepository) GetByCode(dictCode string, excludeID ...string) (*system.Dict, error) {
	params := sqlc.GetDictByCodeParams{DictCode: dictCode}
	if len(excludeID) > 0 && excludeID[0] != "" {
		params.ExcludeID = ptr(excludeID[0])
	}

	row, err := a.q.GetDictByCode(context.Background(), params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainDict(row), nil
}

// GetByIDs 根据ID列表获取字典列表
func (a DictRepository) GetByIDs(ids []string) (system.Dicts, error) {
	if len(ids) == 0 {
		return make(system.Dicts, 0), nil
	}

	rows, err := a.q.ListDictsByIDs(context.Background(), ids)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.Dicts, 0, len(rows))
	for _, r := range rows {
		list = append(list, toDomainDict(r))
	}
	return list, nil
}

// Create 创建字典, assigning a UUID when the ID is empty.
func (a DictRepository) Create(dict *system.Dict) error {
	if dict.ID == "" {
		dict.ID = uuid.NewID()
	}

	err := a.q.CreateDict(context.Background(), sqlc.CreateDictParams{
		ID:        dict.ID,
		DictCode:  dict.DictCode,
		Name:      dict.Name,
		Status:    int32(dict.Status),
		Remark:    dict.Remark,
		CreateBy:  dict.CreateBy,
		UpdateBy:  dict.UpdateBy,
		IsDeleted: int32(dict.IsDeleted),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Update 更新字典 (dict_code, name, status, remark, update_by).
func (a DictRepository) Update(id string, dict *system.Dict) error {
	err := a.q.UpdateDict(context.Background(), sqlc.UpdateDictParams{
		ID:       id,
		DictCode: dict.DictCode,
		Name:     dict.Name,
		Status:   int32(dict.Status),
		Remark:   dict.Remark,
		UpdateBy: dict.UpdateBy,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Delete 删除字典（软删除）
func (a DictRepository) Delete(id string, deletedBy string) error {
	err := a.q.SoftDeleteDict(context.Background(), sqlc.SoftDeleteDictParams{
		ID:       id,
		UpdateBy: deletedBy,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// DeleteByIDs 批量删除字典
func (a DictRepository) DeleteByIDs(ids []string, deletedBy string) error {
	if len(ids) == 0 {
		return nil
	}
	err := a.q.SoftDeleteDictsByIDs(context.Background(), sqlc.SoftDeleteDictsByIDsParams{
		Ids:      ids,
		UpdateBy: deletedBy,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// UpdateDictItemsCode 更新字典项的字典编码 (cascades a dictionary code rename).
func (a DictRepository) UpdateDictItemsCode(oldCode, newCode string) error {
	err := a.q.UpdateDictItemsDictCode(context.Background(), sqlc.UpdateDictItemsDictCodeParams{
		OldCode: oldCode,
		NewCode: newCode,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func toDomainDict(r sqlc.TDict) *system.Dict {
	return &system.Dict{
		ID:         r.ID,
		DictCode:   r.DictCode,
		Name:       r.Name,
		Status:     int(r.Status),
		Remark:     r.Remark,
		CreateBy:   r.CreateBy,
		CreateTime: dto.DateTime(r.CreateTime.Time),
		UpdateBy:   r.UpdateBy,
		UpdateTime: dto.DateTime(r.UpdateTime.Time),
		IsDeleted:  int(r.IsDeleted),
	}
}
