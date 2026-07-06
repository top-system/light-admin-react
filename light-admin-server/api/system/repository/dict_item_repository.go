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

// DictItemRepository is the sqlc/pgx-backed persistence layer for dictionary items.
type DictItemRepository struct {
	q      sqlc.Querier
	logger lib.Logger
}

// NewDictItemRepository creates a new dict item repository bound to the pool-level
// Queries.
func NewDictItemRepository(q *sqlc.Queries, logger lib.Logger) DictItemRepository {
	return DictItemRepository{
		q:      q,
		logger: logger,
	}
}

// WithTx returns a copy bound to the given transaction-scoped Queries.
func (a DictItemRepository) WithTx(q *sqlc.Queries) DictItemRepository {
	a.q = q
	return a
}

// Query 查询字典项分页列表. Ordering is fixed to sort ASC (the previous default).
func (a DictItemRepository) Query(param *system.DictItemQueryParam) (*system.DictItemQueryResult, error) {
	ctx := context.Background()

	var dictCode *string
	if param.DictCode != "" {
		dictCode = ptr(param.DictCode)
	}
	var keywords *string
	if param.Keywords != "" {
		keywords = ptr("%" + param.Keywords + "%")
	}

	total, err := a.q.CountDictItems(ctx, sqlc.CountDictItemsParams{DictCode: dictCode, Keywords: keywords})
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.DictItems, 0)
	if total > 0 {
		limit, offset := pageBounds(param.PaginationParam)
		rows, err := a.q.ListDictItems(ctx, sqlc.ListDictItemsParams{
			DictCode: dictCode,
			Keywords: keywords,
			Limit:    limit,
			Offset:   offset,
		})
		if err != nil {
			return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
		}
		for _, r := range rows {
			list = append(list, toDomainDictItem(r))
		}
	}

	return &system.DictItemQueryResult{
		Pagination: &dto.Pagination{
			PageNum:  param.PaginationParam.GetPageNum(),
			PageSize: param.PaginationParam.GetPageSize(),
			Total:    total,
		},
		List: list,
	}, nil
}

// GetByDictCode 根据字典编码获取字典项列表
func (a DictItemRepository) GetByDictCode(dictCode string) (system.DictItems, error) {
	rows, err := a.q.ListDictItemsByDictCode(context.Background(), dictCode)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.DictItems, 0, len(rows))
	for _, r := range rows {
		list = append(list, toDomainDictItem(r))
	}
	return list, nil
}

// Get 获取字典项
func (a DictItemRepository) Get(id string) (*system.DictItem, error) {
	row, err := a.q.GetDictItem(context.Background(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.DatabaseRecordNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainDictItem(row), nil
}

// Create 创建字典项, assigning a UUID when the ID is empty.
func (a DictItemRepository) Create(item *system.DictItem) error {
	if item.ID == "" {
		item.ID = uuid.NewID()
	}

	err := a.q.CreateDictItem(context.Background(), sqlc.CreateDictItemParams{
		ID:        item.ID,
		DictCode:  item.DictCode,
		Label:     item.Label,
		Value:     item.Value,
		TagType:   item.TagType,
		Sort:      int32(item.Sort),
		Status:    int32(item.Status),
		Remark:    item.Remark,
		CreateBy:  item.CreateBy,
		UpdateBy:  item.UpdateBy,
		IsDeleted: int32(item.IsDeleted),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Update 更新字典项
func (a DictItemRepository) Update(id string, item *system.DictItem) error {
	err := a.q.UpdateDictItem(context.Background(), sqlc.UpdateDictItemParams{
		ID:       id,
		DictCode: item.DictCode,
		Label:    item.Label,
		Value:    item.Value,
		TagType:  item.TagType,
		Sort:     int32(item.Sort),
		Status:   int32(item.Status),
		Remark:   item.Remark,
		UpdateBy: item.UpdateBy,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Delete 删除字典项（软删除）
func (a DictItemRepository) Delete(id string, deletedBy string) error {
	err := a.q.SoftDeleteDictItem(context.Background(), sqlc.SoftDeleteDictItemParams{
		ID:       id,
		UpdateBy: deletedBy,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// DeleteByIDs 批量删除字典项
func (a DictItemRepository) DeleteByIDs(ids []string, deletedBy string) error {
	if len(ids) == 0 {
		return nil
	}
	err := a.q.SoftDeleteDictItemsByIDs(context.Background(), sqlc.SoftDeleteDictItemsByIDsParams{
		Ids:      ids,
		UpdateBy: deletedBy,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// DeleteByDictCodes 根据字典编码删除字典项
func (a DictItemRepository) DeleteByDictCodes(dictCodes []string, deletedBy string) error {
	if len(dictCodes) == 0 {
		return nil
	}
	err := a.q.SoftDeleteDictItemsByDictCodes(context.Background(), sqlc.SoftDeleteDictItemsByDictCodesParams{
		DictCodes: dictCodes,
		UpdateBy:  deletedBy,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func toDomainDictItem(r sqlc.TDictItem) *system.DictItem {
	return &system.DictItem{
		ID:         r.ID,
		DictCode:   r.DictCode,
		Label:      r.Label,
		Value:      r.Value,
		TagType:    r.TagType,
		Sort:       int(r.Sort),
		Status:     int(r.Status),
		Remark:     r.Remark,
		CreateBy:   r.CreateBy,
		CreateTime: dto.DateTime(r.CreateTime.Time),
		UpdateBy:   r.UpdateBy,
		UpdateTime: dto.DateTime(r.UpdateTime.Time),
		IsDeleted:  int(r.IsDeleted),
	}
}
