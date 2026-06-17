package route

import (
	"github.com/top-system/light-admin/api/member/controller"
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/lib"
)

type MemberAuthRoutes struct {
	logger     lib.Logger
	handler    lib.HttpHandler
	controller controller.MemberAuthController
	tenantMw   middlewares.TenantMiddleware
	memberMw   middlewares.MemberAuthMiddleware
}

func NewMemberAuthRoutes(
	logger lib.Logger,
	handler lib.HttpHandler,
	c controller.MemberAuthController,
	tenantMw middlewares.TenantMiddleware,
	memberMw middlewares.MemberAuthMiddleware,
) MemberAuthRoutes {
	return MemberAuthRoutes{logger: logger, handler: handler, controller: c, tenantMw: tenantMw, memberMw: memberMw}
}

func (a MemberAuthRoutes) Setup() {
	// 公开（仅需租户解析）
	auth := a.handler.RouterV1.Group("/member/auth", a.tenantMw.Resolve())
	{
		auth.POST("/register", a.controller.Register)
		auth.POST("/login", a.controller.Login)
		auth.DELETE("/logout", a.controller.Logout, a.memberMw.Require())
	}

	// 需要会员登录
	profile := a.handler.RouterV1.Group("/member/profile", a.memberMw.Require())
	{
		profile.GET("", a.controller.Profile)
		profile.PUT("", a.controller.UpdateProfile)
	}
}
