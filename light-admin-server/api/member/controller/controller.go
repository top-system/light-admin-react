package controller

import "go.uber.org/fx"

// Module exports member controllers
var Module = fx.Options(
	fx.Provide(NewMemberAuthController),
	fx.Provide(NewTenantController),
	fx.Provide(NewMemberController),
)
