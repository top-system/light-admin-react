package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/top-system/light-admin/api/member/service"
	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/member"
	"github.com/top-system/light-admin/pkg/echox"
)

type MemberAuthController struct {
	memberService service.MemberService
	authService   service.MemberAuthService
	logger        lib.Logger
}

func NewMemberAuthController(memberService service.MemberService, authService service.MemberAuthService, logger lib.Logger) MemberAuthController {
	return MemberAuthController{memberService: memberService, authService: authService, logger: logger}
}

func currentTenantID(ctx echo.Context) string {
	v, _ := ctx.Get(constants.CurrentTenantID).(string)
	return v
}

func currentMember(ctx echo.Context) *dto.MemberClaims {
	v, _ := ctx.Get(constants.CurrentMember).(*dto.MemberClaims)
	return v
}

// Register @router /api/v1/member/auth/register [post]
func (a MemberAuthController) Register(ctx echo.Context) error {
	form := new(dto.MemberRegister)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	m, err := a.memberService.Register(currentTenantID(ctx), form)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	return echox.Response{Code: http.StatusOK, Data: m}.JSON(ctx)
}

// Login @router /api/v1/member/auth/login [post]
func (a MemberAuthController) Login(ctx echo.Context) error {
	form := new(dto.MemberLogin)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	tenantID := currentTenantID(ctx)
	m, err := a.memberService.Verify(tenantID, form.Username, form.Password)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	resp, err := a.authService.GenerateToken(m)
	if err != nil {
		return echox.Response{Code: http.StatusInternalServerError, Message: err}.JSON(ctx)
	}
	a.memberService.RecordLogin(tenantID, m.ID, ctx.RealIP())
	return echox.Response{Code: http.StatusOK, Data: resp}.JSON(ctx)
}

// Profile @router /api/v1/member/profile [get]
func (a MemberAuthController) Profile(ctx echo.Context) error {
	claims := currentMember(ctx)
	if claims == nil {
		return echox.Response{Code: http.StatusUnauthorized, Message: errors.MemberTokenInvalid}.JSON(ctx)
	}
	m, err := a.memberService.GetProfile(claims.TenantID, claims.ID)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	return echox.Response{Code: http.StatusOK, Data: m}.JSON(ctx)
}

// UpdateProfile @router /api/v1/member/profile [put]
func (a MemberAuthController) UpdateProfile(ctx echo.Context) error {
	claims := currentMember(ctx)
	if claims == nil {
		return echox.Response{Code: http.StatusUnauthorized, Message: errors.MemberTokenInvalid}.JSON(ctx)
	}
	form := new(member.MemberProfileForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	if err := a.memberService.UpdateProfile(claims.TenantID, claims.ID, form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// Logout @router /api/v1/member/auth/logout [delete]
func (a MemberAuthController) Logout(ctx echo.Context) error {
	// 无服务端会话，前端丢弃 token 即可
	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}
