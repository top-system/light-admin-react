package service

import "go.uber.org/fx"

// Module exports member services
var Module = fx.Options(
	fx.Provide(NewMemberAuthService),
	fx.Provide(NewTenantService),
	// fx.Provide(NewMemberService),
)
