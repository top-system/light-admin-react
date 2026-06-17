package route

import (
	"github.com/top-system/light-admin/api/member/controller"
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/lib"
)

type MemberAdminRoutes struct {
	logger         lib.Logger
	handler        lib.HttpHandler
	controller     controller.MemberController
	permMiddleware middlewares.PermissionMiddleware
}

func NewMemberAdminRoutes(logger lib.Logger, handler lib.HttpHandler, c controller.MemberController, perm middlewares.PermissionMiddleware) MemberAdminRoutes {
	return MemberAdminRoutes{logger: logger, handler: handler, controller: c, permMiddleware: perm}
}

func (a MemberAdminRoutes) Setup() {
	api := a.handler.RouterV1.Group("/members")
	{
		api.GET("", a.controller.Query, a.permMiddleware.RequirePerm("member:member:query"))
		api.PUT("/:id/status", a.controller.SetStatus, a.permMiddleware.RequirePerm("member:member:edit"))
		api.PUT("/:id/password", a.controller.ResetPassword, a.permMiddleware.RequirePerm("member:member:edit"))
	}
}
