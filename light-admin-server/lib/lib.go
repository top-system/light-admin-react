package lib

import "go.uber.org/fx"

// Module exports dependency
var Module = fx.Options(
	fx.Provide(NewHttpHandler),
	fx.Provide(NewConfig),
	fx.Provide(NewLogger),
	fx.Provide(NewDataLayer),           // engine-neutral store.Store + TxManager (postgres/sqlite)
	fx.Provide(NewQueueTaskRepository), // store-backed queue.TaskRepository (injected into pkg/queue)
	fx.Provide(NewCache),
	fx.Provide(NewCaptcha),
	fx.Provide(NewWebSocket),
	ExtrasModule, // 启用扩展模块（队列、定时任务、下载器）
)
