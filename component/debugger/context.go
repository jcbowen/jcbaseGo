package debugger

import "context"

// loggerContextKey 日志记录器在 context.Context 中存储所使用的私有键类型。
// 使用私有结构体类型而非字符串，避免与其他包的 context 键发生冲突。
type loggerContextKey struct{}

// ContextWithLogger 将日志记录器注入到 context.Context 中。
//
// 注入后，下游代码可通过 LoggerFromContext 取回该日志记录器，
// 常用于把「请求级 logger」沿着调用链透传给 ORM、RPC 等不具备 gin.Context 的组件。
//
// 参数：
//   - ctx: 父级上下文，允许为 nil（此时等价于 context.Background()）
//   - logger: 待注入的日志记录器；传 nil 时等价于 ContextWithoutLogger，即移除已有 logger
//
// 返回：
//   - context.Context: 携带（或不携带）日志记录器的新上下文
//
// 使用示例：
//
//	ctx := debugger.ContextWithLogger(c.Request.Context(), debugger.GetLoggerFromContext(c))
//	db.WithContext(ctx).Find(&list)
func ContextWithLogger(ctx context.Context, logger LoggerInterface) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	// 传入 nil 时视为移除，避免下游拿到 nil 接口后调用 panic
	if logger == nil {
		return ContextWithoutLogger(ctx)
	}

	return context.WithValue(ctx, loggerContextKey{}, logger)
}

// ContextWithoutLogger 从 context.Context 中移除日志记录器。
//
// 适用于 WebSocket 长连接、后台 goroutine 等生命周期超出单次请求的场景：
// 这类场景若继续复用请求级 logger，日志会持续累积到一个早已归档的请求日志中，
// 既无法在调试面板查看，也会造成内存增长。
//
// 参数：
//   - ctx: 父级上下文，允许为 nil
//
// 返回：
//   - context.Context: 不含日志记录器的新上下文；原上下文中不存在 logger 时原样返回
//
// 使用示例：
//
//	ctx = debugger.ContextWithoutLogger(ctx) // 长连接建立后调用一次
func ContextWithoutLogger(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}

	if !HasLoggerFromContext(ctx) {
		return ctx
	}

	return context.WithValue(ctx, loggerContextKey{}, nil)
}

// NoopLogger 返回一个空操作日志记录器实例。
//
// 适用于「日志目标可选」的场景：调用方允许传入 nil 表示不记录日志时，
// 用本函数的返回值占位，可避免下游出现 nil 接口调用 panic。
//
// 返回：
//   - LoggerInterface: 所有方法均为空实现的日志记录器
//
// 使用示例：
//
//	if fallback == nil {
//	    fallback = debugger.NoopLogger()
//	}
func NoopLogger() LoggerInterface {
	return noopLogger{}
}

// LoggerFromContext 从 context.Context 中获取日志记录器。
//
// 参数：
//   - ctx: 待读取的上下文，允许为 nil
//
// 返回：
//   - LoggerInterface: 上下文中携带的日志记录器；不存在或已被移除时返回 noopLogger（空操作实现，可安全调用）
//
// 使用示例：
//
//	logger := debugger.LoggerFromContext(ctx)
//	logger.Info("执行完成", map[string]interface{}{"cost": 12})
func LoggerFromContext(ctx context.Context) LoggerInterface {
	if ctx == nil {
		return noopLogger{}
	}

	if logger, ok := ctx.Value(loggerContextKey{}).(LoggerInterface); ok && logger != nil {
		return logger
	}

	return noopLogger{}
}

// HasLoggerFromContext 判断 context.Context 中是否存在有效的日志记录器。
//
// 与 LoggerFromContext 的区别：后者在缺失时返回 noopLogger，无法区分「取不到」与「取到空实现」；
// 需要「取不到则回退兜底 logger」的场景（如 ORM 的 SQL 日志）应使用本函数判断。
//
// 参数：
//   - ctx: 待判断的上下文，允许为 nil
//
// 返回：
//   - bool: 存在有效日志记录器返回 true，否则返回 false
//
// 使用示例：
//
//	if debugger.HasLoggerFromContext(ctx) {
//	    logger = debugger.LoggerFromContext(ctx)
//	} else {
//	    logger = fallbackLogger
//	}
func HasLoggerFromContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	logger, ok := ctx.Value(loggerContextKey{}).(LoggerInterface)
	return ok && logger != nil
}
