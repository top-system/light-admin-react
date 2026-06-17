package repository

import (
	"time"

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

// nowDateTime 返回当前时间的 dto.DateTime（用于登录信息更新）
func nowDateTime() dto.DateTime {
	return dto.DateTime(time.Now())
}
