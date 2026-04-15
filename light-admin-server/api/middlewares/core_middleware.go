package middlewares

import (
	"fmt"
	"runtime"

	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/lib"
	"github.com/labstack/echo/v4"

	"go.uber.org/zap"
)

// core middleware is a functional extension to "echo",
// including database transactions and panic recovery
// and more
type CoreMiddleware struct {
	handler lib.HttpHandler
	logger  lib.Logger
	db      lib.Database
}

// statusInList function checks if context writer status is in provided list
func statusInList(status int, statusList []int) bool {
	for _, i := range statusList {
		if i == status {
			return true
		}
	}
	return false
}

// NewCoreMiddleware creates new database transactions middleware
func NewCoreMiddleware(handler lib.HttpHandler, logger lib.Logger, db lib.Database) CoreMiddleware {
	return CoreMiddleware{
		handler: handler,
		logger:  logger,
		db:      db,
	}
}

func (a CoreMiddleware) core() echo.MiddlewareFunc {
	logger := a.logger.DesugarZap.With(zap.String("module", "core-mw"))

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// 跳过 WebSocket 请求，WebSocket 不需要数据库事务
			if ctx.Request().URL.Path == "/ws" {
				return next(ctx)
			}

			// Controllers always read `constants.DBTransaction` and assert to *gorm.DB;
			// for branches where we bypass an explicit transaction we still need to set
			// the bare ORM handle so the type assertion doesn't panic.

			// For SQLite: disable auto-transaction completely to avoid database locking
			// SQLite has limited concurrency support and auto-transactions cause deadlocks
			if lib.IsSQLite() {
				ctx.Set(constants.DBTransaction, a.db.ORM)
				return next(ctx)
			}

			// 只读请求不需要数据库事务
			method := ctx.Request().Method
			if method == "GET" || method == "HEAD" || method == "OPTIONS" {
				ctx.Set(constants.DBTransaction, a.db.ORM)
				return next(ctx)
			}

			txHandle := a.db.ORM.Begin()
			logger.Debug("beginning database transaction")

			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					// recovery stack
					stack := make([]byte, 4<<10)
					length := runtime.Stack(stack, false)
					msg := fmt.Sprintf("[PANIC RECOVER] %v %s\n", err, stack[:length])
					logger.Error(msg)

					// rollback database transaction
					logger.Debug("rolling back transaction due to panic")
					txHandle.Rollback()
					ctx.Error(err)
				}
			}()

			ctx.Set(constants.DBTransaction, txHandle)

			if err := next(ctx); err != nil {
				ctx.Error(err)
			}

			code := ctx.Response().Status
			// rollback transaction on server errors
			if code >= 400 {
				logger.Debug(fmt.Sprintf("rolling back transaction due to status code: %d", code))
				txHandle.Rollback()
			} else {
				logger.Debug("committing transaction")
				if err := txHandle.Commit().Error; err != nil {
					logger.Error(fmt.Sprintf("trx commit error: %v", err))
				}
			}

			return nil
		}
	}
}

func (a CoreMiddleware) Setup() {
	a.handler.Engine.Use(a.core())
}
