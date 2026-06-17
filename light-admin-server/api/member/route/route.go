package route

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewMemberAuthRoutes),
	fx.Provide(NewTenantRoutes),
	fx.Provide(NewMemberAdminRoutes),
	fx.Provide(NewRoutes),
)

type Routes []Route

type Route interface{ Setup() }

func NewRoutes(
	memberAuth MemberAuthRoutes,
	tenant TenantRoutes,
	memberAdmin MemberAdminRoutes,
) Routes {
	return Routes{memberAuth, tenant, memberAdmin}
}

func (a Routes) Setup() {
	for _, r := range a {
		r.Setup()
	}
}
