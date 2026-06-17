package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/top-system/light-admin/api/member/service"
	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/tenant"
	"github.com/top-system/light-admin/pkg/echox"
)

type TenantController struct {
	tenantService service.TenantService
	logger        lib.Logger
}

func NewTenantController(tenantService service.TenantService, logger lib.Logger) TenantController {
	return TenantController{tenantService: tenantService, logger: logger}
}

func operatorID(ctx echo.Context) string {
	if claims, ok := ctx.Get(constants.CurrentUser).(*dto.JwtClaims); ok && claims != nil {
		return claims.ID
	}
	return ""
}

func (a TenantController) Query(ctx echo.Context) error {
	param := new(tenant.TenantQueryParam)
	if err := ctx.Bind(param); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	result, err := a.tenantService.Query(param)
	if err != nil {
		return echox.Response{Code: http.StatusInternalServerError, Message: err}.JSON(ctx)
	}
	return echox.Response{
		Code: http.StatusOK,
		Data: result.List,
		Page: &echox.PageInfo{Total: result.Pagination.Total, PageNum: result.Pagination.PageNum, PageSize: result.Pagination.PageSize},
	}.JSON(ctx)
}

func (a TenantController) Create(ctx echo.Context) error {
	form := new(tenant.TenantForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	id, err := a.tenantService.Create(form, operatorID(ctx))
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	return echox.Response{Code: http.StatusOK, Data: map[string]string{"id": id}}.JSON(ctx)
}

func (a TenantController) Update(ctx echo.Context) error {
	id := ctx.Param("id")
	form := new(tenant.TenantForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	if err := a.tenantService.Update(id, form, operatorID(ctx)); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

func (a TenantController) Delete(ctx echo.Context) error {
	if err := a.tenantService.Delete(ctx.Param("id")); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}
