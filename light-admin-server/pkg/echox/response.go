package echox

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/top-system/light-admin/errors"
)

// Context key constants
const (
	KeyCurrentUser = "current_user"
)

// Response in order to unify the returned response structure
type Response struct {
	Code    int         `json:"-"`
	Pretty  bool        `json:"-"`
	BizCode string      `json:"code"`
	Data    interface{} `json:"data,omitempty"`
	Page    *PageInfo   `json:"page,omitempty"`
	Message interface{} `json:"message"`
}

// PageInfo 分页信息
type PageInfo struct {
	Total    int64 `json:"total"`
	PageNum  int   `json:"pageNum"`
	PageSize int   `json:"pageSize"`
}

// sends a JSON response with status code.
func (a Response) JSON(ctx echo.Context) error {
	var responseErr error
	if a.Message == "" || a.Message == nil {
		a.Message = http.StatusText(a.Code)
	}

	if err, ok := a.Message.(error); ok {
		responseErr = err
		if status := errors.HTTPStatusCode(err); status != 0 {
			a.Code = status
		}
		a.Message = err.Error()
	}

	// Derive a stable business code after resolving any domain HTTP status.
	if a.BizCode == "" {
		a.BizCode = errors.BusinessCode(responseErr, a.Code)
	}

	if a.Pretty {
		return ctx.JSONPretty(a.Code, a, "\t")
	}

	return ctx.JSON(a.Code, a)
}

// ============ 辅助函数 ============

// OK 返回成功响应
func OK(ctx echo.Context, data interface{}) error {
	return Response{Code: http.StatusOK, Data: data}.JSON(ctx)
}

// OKWithPage 返回带分页的成功响应
func OKWithPage(ctx echo.Context, data interface{}, total int64, pageNum, pageSize int) error {
	return Response{
		Code: http.StatusOK,
		Data: data,
		Page: &PageInfo{Total: total, PageNum: pageNum, PageSize: pageSize},
	}.JSON(ctx)
}

// Fail 返回失败响应
func Fail(ctx echo.Context, err error) error {
	return Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
}

// FailWithCode 返回带状态码的失败响应
func FailWithCode(ctx echo.Context, code int, err error) error {
	return Response{Code: code, Message: err}.JSON(ctx)
}

// GetPathID 从路径参数获取 ID
func GetPathID(ctx echo.Context, param string) (uint64, error) {
	return strconv.ParseUint(ctx.Param(param), 10, 64)
}
