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

// MemberAuthMiddleware 校验会员 JWT，注入 constants.CurrentMember
type MemberAuthMiddleware struct {
	logger      lib.Logger
	authService memberService.MemberAuthService
}

func NewMemberAuthMiddleware(logger lib.Logger, authService memberService.MemberAuthService) MemberAuthMiddleware {
	return MemberAuthMiddleware{logger: logger, authService: authService}
}

// Require 返回需要会员登录的路由组级中间件
func (a MemberAuthMiddleware) Require() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			auth := ctx.Request().Header.Get("Authorization")
			prefix := "Bearer "
			var token string
			if auth != "" && strings.HasPrefix(auth, prefix) {
				token = auth[len(prefix):]
			}

			claims, err := a.authService.ParseToken(token)
			if err != nil {
				return echox.Response{Code: http.StatusUnauthorized, Message: err}.JSON(ctx)
			}

			ctx.Set(constants.CurrentMember, claims)
			// 以 token 内租户为准
			ctx.Set(constants.CurrentTenantID, claims.TenantID)
			return next(ctx)
		}
	}
}
