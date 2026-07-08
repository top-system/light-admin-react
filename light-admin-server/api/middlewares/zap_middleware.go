package middlewares

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/top-system/light-admin/lib"
)

// ZapMiddleware is the structured request-logging middleware. It is now backed
// by the standard library slog logger (the zap core lives behind it via
// zapslog); the type name is retained to avoid churning its registration.
type ZapMiddleware struct {
	handler lib.HttpHandler
	logger  lib.Logger
}

// NewZapMiddleware creates new request-logging middleware
func NewZapMiddleware(handler lib.HttpHandler, logger lib.Logger) ZapMiddleware {
	return ZapMiddleware{
		handler: handler,
		logger:  logger,
	}
}

func (a ZapMiddleware) core() echo.MiddlewareFunc {
	logger := a.logger.With(slog.String("module", "log-mw"))

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// 跳过 WebSocket 请求，WebSocket 会劫持响应
			if ctx.Request().URL.Path == "/ws" {
				return next(ctx)
			}

			start := time.Now()

			// 使用局部变量避免污染闭包外的 logger
			reqLogger := logger
			if err := next(ctx); err != nil {
				reqLogger = reqLogger.With(slog.Any("error", err))
				ctx.Error(err)
			}

			request := ctx.Request()
			response := ctx.Response()

			fields := []any{
				slog.String("remote_ip", ctx.RealIP()),
				slog.String("time", time.Since(start).String()),
				slog.String("host", request.Host),
				slog.String("request", fmt.Sprintf("%s %s", request.Method, request.RequestURI)),
				slog.Int("status", response.Status),
				slog.Int64("size", response.Size),
				slog.String("user_agent", request.UserAgent()),
			}

			id := request.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = response.Header().Get(echo.HeaderXRequestID)
				fields = append(fields, slog.String("request_id", id))
			}

			n := response.Status
			switch {
			case n >= 500:
				reqLogger.Error("Server error", fields...)
			case n >= 400:
				reqLogger.Warn("Client error", fields...)
			case n >= 300:
				reqLogger.Info("Redirection", fields...)
			default:
				reqLogger.Info("Success", fields...)
			}

			return nil
		}
	}
}

func (a ZapMiddleware) Setup() {
	a.handler.Engine.Use(a.core())
}
