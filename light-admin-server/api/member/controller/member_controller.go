package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/top-system/light-admin/api/member/service"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/member"
	"github.com/top-system/light-admin/pkg/echox"
)

type MemberController struct {
	memberService service.MemberService
	logger        lib.Logger
}

func NewMemberController(memberService service.MemberService, logger lib.Logger) MemberController {
	return MemberController{memberService: memberService, logger: logger}
}

func (a MemberController) Query(ctx echo.Context) error {
	param := new(member.MemberQueryParam)
	if err := ctx.Bind(param); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	result, err := a.memberService.Query(param)
	if err != nil {
		return echox.Response{Code: http.StatusInternalServerError, Message: err}.JSON(ctx)
	}
	return echox.Response{
		Code: http.StatusOK,
		Data: result.List,
		Page: &echox.PageInfo{Total: result.Pagination.Total, PageNum: result.Pagination.PageNum, PageSize: result.Pagination.PageSize},
	}.JSON(ctx)
}

type statusForm struct {
	TenantID string `json:"tenantId" validate:"required"`
	Status   int    `json:"status"`
}

func (a MemberController) SetStatus(ctx echo.Context) error {
	id := ctx.Param("id")
	form := new(statusForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	if err := a.memberService.SetStatus(form.TenantID, id, form.Status); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

type resetPwdForm struct {
	TenantID string `json:"tenantId" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

func (a MemberController) ResetPassword(ctx echo.Context) error {
	id := ctx.Param("id")
	form := new(resetPwdForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	if err := a.memberService.ResetPassword(form.TenantID, id, form.Password); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}
