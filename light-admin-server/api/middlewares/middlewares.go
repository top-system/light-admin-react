package middlewares

import "go.uber.org/fx"

// Module Middleware exported
var Module = fx.Options(
	fx.Provide(NewCoreMiddleware),
	fx.Provide(NewCorsMiddleware),
	fx.Provide(NewZapMiddleware),
	fx.Provide(NewAuthMiddleware),
	fx.Provide(NewCasbinMiddleware),
	fx.Provide(NewLogMiddleware),
	fx.Provide(NewRateLimitMiddleware),
	fx.Provide(NewMiddlewares),
	// 路由组级中间件（不进入全局 NewMiddlewares 列表，不通过 engine.Use 执行）
	fx.Provide(NewTenantMiddleware),
	fx.Provide(NewMemberAuthMiddleware),
)

// IMiddleware middleware interface
type IMiddleware interface {
	Setup()
}

// Middlewares contains multiple middleware
type Middlewares []IMiddleware

// NewMiddlewares creates new middlewares
// Register the middleware that should be applied directly (globally)
func NewMiddlewares(
	coreMiddleware CoreMiddleware,
	corsMiddleware CorsMiddleware,
	zapMiddleware ZapMiddleware,
	authMiddleware AuthMiddleware,
	casbinMiddleware CasbinMiddleware,
	logMiddleware LogMiddleware,
	rateLimitMiddleware RateLimitMiddleware,
) Middlewares {
	return Middlewares{
		coreMiddleware,
		rateLimitMiddleware,
		zapMiddleware,
		corsMiddleware,
		authMiddleware,
		casbinMiddleware,
		logMiddleware,
	}
}

// Setup sets up middlewares
func (a Middlewares) Setup() {
	for _, middleware := range a {
		middleware.Setup()
	}
}

func isIgnorePath(path string, prefixes ...string) bool {
	pathLen := len(path)

	for _, p := range prefixes {
		if pl := len(p); pathLen >= pl && path[:pl] == p {
			return true
		}
	}

	return false
}
