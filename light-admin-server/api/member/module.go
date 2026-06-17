package member

import (
	"github.com/top-system/light-admin/api/member/controller"
	"github.com/top-system/light-admin/api/member/repository"
	"github.com/top-system/light-admin/api/member/route"
	"github.com/top-system/light-admin/api/member/service"

	"go.uber.org/fx"
)

// Module exports member domain module
var Module = fx.Options(
	controller.Module,
	service.Module,
	repository.Module,
	route.Module,
)
