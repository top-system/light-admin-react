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

// ConfigRepository is the sqlc/pgx-backed persistence layer for system configs.
type ConfigRepository struct {
	q      store.Store
	logger lib.Logger
}

// NewConfigRepository creates a new config repository bound to the pool-level Queries.
func NewConfigRepository(q store.Store, logger lib.Logger) ConfigRepository {
	return ConfigRepository{
		q:      q,
		logger: logger,
	}
}

// WithTx returns a copy bound to the given transaction-scoped Queries.
func (a ConfigRepository) WithTx(q store.Store) ConfigRepository {
	a.q = q
	return a
}

// Query lists configs matching the keyword filter. Ordering is fixed to id DESC.
func (a ConfigRepository) Query(param *system.ConfigQueryParam) (*system.ConfigQueryResult, error) {
	ctx := context.Background()

	var keywords *string
	if param.Keywords != "" {
		keywords = ptr("%" + param.Keywords + "%")
	}

	total, err := a.q.CountConfigs(ctx, keywords)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.Configs, 0)
	if total > 0 {
		limit, offset := pageBounds(param.PaginationParam)
		rows, err := a.q.ListConfigs(ctx, store.ListConfigsParams{Keywords: keywords, Limit: limit, Offset: offset})
		if err != nil {
			return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
		}
		for _, r := range rows {
			list = append(list, toDomainConfig(r))
		}
	}

	return &system.ConfigQueryResult{
		Pagination: &dto.Pagination{
			PageNum:  param.PaginationParam.GetPageNum(),
			PageSize: param.PaginationParam.GetPageSize(),
			Total:    total,
		},
		List: list,
	}, nil
}

func (a ConfigRepository) Get(id string) (*system.Config, error) {
	row, err := a.q.GetConfig(context.Background(), id)
	if err != nil {
		if errors.Is(err, store.ErrNoRows) {
			return nil, apperrors.DatabaseRecordNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainConfig(row), nil
}

func (a ConfigRepository) GetByKey(key string) (*system.Config, error) {
	row, err := a.q.GetConfigByKey(context.Background(), key)
	if err != nil {
		if errors.Is(err, store.ErrNoRows) {
			return nil, apperrors.DatabaseRecordNotFound
		}
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return toDomainConfig(row), nil
}

func (a ConfigRepository) GetAll() (system.Configs, error) {
	rows, err := a.q.ListAllConfigs(context.Background())
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}

	list := make(system.Configs, 0, len(rows))
	for _, r := range rows {
		list = append(list, toDomainConfig(r))
	}
	return list, nil
}

func (a ConfigRepository) ExistsByKey(key string, excludeID string) (bool, error) {
	params := store.CountConfigsByKeyParams{ConfigKey: key}
	if excludeID != "" {
		params.ExcludeID = ptr(excludeID)
	}

	count, err := a.q.CountConfigsByKey(context.Background(), params)
	if err != nil {
		return false, apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return count > 0, nil
}

// Create inserts a new config, assigning a UUID when the ID is empty.
func (a ConfigRepository) Create(config *system.Config) error {
	if config.ID == "" {
		config.ID = uuid.NewID()
	}

	err := a.q.CreateConfig(context.Background(), store.CreateConfigParams{
		ID:          config.ID,
		ConfigName:  config.ConfigName,
		ConfigKey:   config.ConfigKey,
		ConfigValue: config.ConfigValue,
		Remark:      config.Remark,
		CreateBy:    config.CreateBy,
		UpdateBy:    config.UpdateBy,
		IsDeleted:   int32(config.IsDeleted),
		Now:         time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Update updates the mutable fields of a config.
func (a ConfigRepository) Update(id string, config *system.Config) error {
	err := a.q.UpdateConfig(context.Background(), store.UpdateConfigParams{
		ID:          id,
		ConfigName:  config.ConfigName,
		ConfigKey:   config.ConfigKey,
		ConfigValue: config.ConfigValue,
		Remark:      config.Remark,
		UpdateBy:    config.UpdateBy,
		Now:         time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Delete soft-deletes a config.
func (a ConfigRepository) Delete(id string, deletedBy string) error {
	err := a.q.SoftDeleteConfig(context.Background(), store.SoftDeleteConfigParams{
		ID:       id,
		UpdateBy: deletedBy,
		Now:      time.Now(),
	})
	if err != nil {
		return apperrors.Wrap(apperrors.DatabaseInternalError, err.Error())
	}
	return nil
}

func toDomainConfig(r store.TConfig) *system.Config {
	return &system.Config{
		ID:          r.ID,
		ConfigName:  r.ConfigName,
		ConfigKey:   r.ConfigKey,
		ConfigValue: r.ConfigValue,
		Remark:      r.Remark,
		CreateTime:  dto.DateTime(r.CreateTime),
		CreateBy:    r.CreateBy,
		UpdateTime:  dto.DateTime(r.UpdateTime),
		UpdateBy:    r.UpdateBy,
		IsDeleted:   int(r.IsDeleted),
	}
}
