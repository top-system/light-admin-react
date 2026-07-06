package repository

import (
	"go.uber.org/fx"

	"github.com/top-system/light-admin/models/dto"
)

// Module exports member repositories
var Module = fx.Options(
	fx.Provide(NewTenantRepository),
	fx.Provide(NewMemberRepository),
)

// dtoPagination 构造统一分页对象（tenant/member 仓储共用）
func dtoPagination(total int64, pageNum, pageSize int) dto.Pagination {
	return dto.Pagination{Total: total, PageNum: pageNum, PageSize: pageSize}
}

// ptr returns a pointer to v, used to build nullable sqlc arguments.
func ptr[T any](v T) *T { return &v }

// pageBounds converts pagination params into sqlc limit/offset arguments,
// mirroring the previous GORM math: page/size when both are positive, size-only
// limit otherwise, and no bounds when size is non-positive. A zero page number
// therefore yields the first page rather than a negative OFFSET.
func pageBounds(pp dto.PaginationParam) (limit, offset *int32) {
	current, pageSize := pp.GetPageNum(), pp.GetPageSize()
	switch {
	case current > 0 && pageSize > 0:
		l := int32(pageSize)
		o := int32((current - 1) * pageSize)
		return &l, &o
	case pageSize > 0:
		l := int32(pageSize)
		return &l, nil
	default:
		return nil, nil
	}
}
