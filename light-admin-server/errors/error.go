package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// 标准库 errors 别名（保持向后兼容）
var (
	Is     = errors.Is
	As     = errors.As
	New    = errors.New
	Unwrap = errors.Unwrap
)

// Wrap 为 err 添加上下文信息，保留 errors.Is 链
// 替代 pkg/errors.Wrap，使用标准库 fmt.Errorf("%w") 实现
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf 为 err 添加格式化的上下文信息
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// WithStack 兼容保留，标准库不需要手动添加 stack trace（Zap 已处理）
func WithStack(err error) error {
	return err
}

// WithMessage 为 err 添加上下文消息
func WithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WithMessagef 为 err 添加格式化的上下文消息
func WithMessagef(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// Database
var (
	DatabaseInternalError  = errors.New("database internal error")
	DatabaseRecordNotFound = errors.New("database record not found")
)

// Redis
var (
	RedisKeyNoExist = errors.New("redis key does not exist")
)

// Captcha
var (
	CaptchaAnswerCodeNoMatch = errors.New("captcha answer code no match")
)

// Auth
var (
	AuthTokenInvalid      = errors.New("auth token is invalid")
	AuthTokenExpired      = errors.New("auth token is expired")
	AuthTokenNotValidYet  = errors.New("auth token not active yet")
	AuthTokenMalformed    = errors.New("auth token is malformed")
	AuthTokenGenerateFail = errors.New("failed to generate auth token")
)

// errorHTTPStatus 错误到 HTTP 状态码的映射表
var errorHTTPStatus = map[error]int{
	// 500 Internal Server Error
	DatabaseInternalError: http.StatusInternalServerError,
	AuthTokenGenerateFail: http.StatusInternalServerError,

	// 404 Not Found
	DatabaseRecordNotFound: http.StatusNotFound,

	// 401 Unauthorized
	AuthTokenInvalid:     http.StatusUnauthorized,
	AuthTokenExpired:     http.StatusUnauthorized,
	AuthTokenNotValidYet: http.StatusUnauthorized,
	AuthTokenMalformed:   http.StatusUnauthorized,
}

// RegisterHTTPStatus 注册错误到 HTTP 状态码的映射（供各模块 init 时调用）
func RegisterHTTPStatus(err error, status int) {
	errorHTTPStatus[err] = status
}

// HTTPStatusCode 根据错误类型返回对应的 HTTP 状态码，未匹配返回 0
func HTTPStatusCode(err error) int {
	for target, status := range errorHTTPStatus {
		if Is(err, target) {
			return status
		}
	}
	return 0
}
