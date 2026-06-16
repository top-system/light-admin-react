package middlewares

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	memberService "github.com/top-system/light-admin/api/member/service"
	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/pkg/echox"
)

// TenantMiddleware 解析请求所属租户，注入 constants.CurrentTenantID
type TenantMiddleware struct {
	config        lib.Config
	logger        lib.Logger
	tenantService memberService.TenantService
}

func NewTenantMiddleware(config lib.Config, logger lib.Logger, tenantService memberService.TenantService) TenantMiddleware {
	return TenantMiddleware{config: config, logger: logger, tenantService: tenantService}
}

// Resolve 返回路由组级中间件
func (a TenantMiddleware) Resolve() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			code := a.extractCode(ctx)

			// 多租户关闭：统一落到默认租户
			if !a.config.MultiTenant.IsEnabled() {
				code = a.config.MultiTenant.FallbackTenantCode()
			}

			t, err := a.tenantService.ResolveByCode(code)
			if err != nil {
				return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
			}

			ctx.Set(constants.CurrentTenantID, t.ID)
			return next(ctx)
		}
	}
}

func (a TenantMiddleware) extractCode(ctx echo.Context) string {
	switch a.config.MultiTenant.Resolver {
	case "subdomain":
		host := ctx.Request().Host
		if i := strings.IndexByte(host, '.'); i > 0 {
			return host[:i]
		}
		return ""
	case "path":
		return ctx.Param("tenantCode")
	default: // header
		return ctx.Request().Header.Get(a.config.MultiTenant.ResolverHeaderName())
	}
}
