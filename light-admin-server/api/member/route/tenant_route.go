package route

import (
	"github.com/top-system/light-admin/api/member/controller"
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/lib"
)

type TenantRoutes struct {
	logger         lib.Logger
	handler        lib.HttpHandler
	controller     controller.TenantController
	permMiddleware middlewares.PermissionMiddleware
}

func NewTenantRoutes(logger lib.Logger, handler lib.HttpHandler, c controller.TenantController, perm middlewares.PermissionMiddleware) TenantRoutes {
	return TenantRoutes{logger: logger, handler: handler, controller: c, permMiddleware: perm}
}

func (a TenantRoutes) Setup() {
	api := a.handler.RouterV1.Group("/tenants")
	{
		api.GET("", a.controller.Query, a.permMiddleware.RequirePerm("member:tenant:query"))
		api.POST("", a.controller.Create, a.permMiddleware.RequirePerm("member:tenant:add"))
		api.PUT("/:id", a.controller.Update, a.permMiddleware.RequirePerm("member:tenant:edit"))
		api.DELETE("/:id", a.controller.Delete, a.permMiddleware.RequirePerm("member:tenant:delete"))
	}
}
