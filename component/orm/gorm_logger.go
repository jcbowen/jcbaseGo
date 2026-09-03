package orm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jcbowen/jcbaseGo/component/debugger"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormDebuggerLogger 实现GORM的logger.Interface接口，将SQL执行情况记录到debugger中
//
// 支持两种工作模式：
//  1. 固定Logger模式（contextAware=false，默认）：所有SQL日志写入构造时注入的Logger，
//     适用于只有单一日志目标、或按开关临时打开SQL日志的场景；
//  2. 上下文感知模式（contextAware=true）：每次记录时优先从 context.Context 中取请求级Logger，
//     取不到才回退到兜底Logger。全局只需注入一次，天然并发安全，
//     适用于需要把SQL日志归属到具体HTTP请求的场景。
type GormDebuggerLogger struct {
	debuggerLogger debugger.LoggerInterface // debugger日志记录器（上下文感知模式下作为兜底）
	logLevel       logger.LogLevel          // 日志级别
	slowThreshold  time.Duration            // 慢查询阈值
	contextAware   bool                     // 是否启用上下文感知模式
}

// NewGormDebuggerLogger 创建GORM调试日志记录器（固定Logger模式）
// 参数：
//   - debuggerLogger: debugger组件的日志记录器实例
//   - logLevel: GORM日志级别
//   - slowThreshold: 慢查询阈值
//
// 返回：
//   - *GormDebuggerLogger: GORM调试日志记录器实例
func NewGormDebuggerLogger(debuggerLogger debugger.LoggerInterface, logLevel debugger.LogLevel, slowThreshold time.Duration) *GormDebuggerLogger {
	return &GormDebuggerLogger{
		debuggerLogger: debuggerLogger,
		logLevel:       logger.LogLevel(logLevel),
		slowThreshold:  slowThreshold,
	}
}

// NewContextAwareGormDebuggerLogger 创建上下文感知模式的GORM调试日志记录器
//
// 与固定Logger模式的区别：本模式下SQL日志会写入 context.Context 中携带的请求级Logger，
// 因此可在高并发场景下把每条SQL正确归属到对应的请求详情，不会被并发请求互相覆盖。
//
// 参数：
//   - fallbackLogger: 兜底日志记录器，当 context 中不存在Logger时使用；
//     建议传入 debugger.Debugger 的 GetMainLogger()，使定时任务、启动迁移等
//     无请求上下文的SQL也能落盘；传 nil 表示此类SQL直接丢弃
//   - logLevel: GORM日志级别
//   - slowThreshold: 慢查询阈值
//
// 返回：
//   - *GormDebuggerLogger: 上下文感知模式的GORM调试日志记录器实例
//
// 使用示例：
//
//	gormLogger := orm.NewContextAwareGormDebuggerLogger(dbg.GetMainLogger(), debugger.LevelInfo, 200*time.Millisecond)
//	db.Config.Logger = gormLogger
func NewContextAwareGormDebuggerLogger(fallbackLogger debugger.LoggerInterface, logLevel debugger.LogLevel, slowThreshold time.Duration) *GormDebuggerLogger {
	if fallbackLogger == nil {
		fallbackLogger = debugger.NoopLogger()
	}

	return &GormDebuggerLogger{
		debuggerLogger: fallbackLogger,
		logLevel:       logger.LogLevel(logLevel),
		slowThreshold:  slowThreshold,
		contextAware:   true,
	}
}

// resolveLogger 根据实际工作模式解析本次日志应写入的日志记录器
//
// 上下文感知模式下优先取 context 中携带的请求级Logger，取不到时回退到兜底Logger；
// 固定Logger模式下始终返回构造时注入的Logger。
//
// 参数：
//   - ctx: GORM 回调传入的上下文
//
// 返回：
//   - debugger.LoggerInterface: 本次日志的写入目标，必定非 nil
func (l *GormDebuggerLogger) resolveLogger(ctx context.Context) debugger.LoggerInterface {
	if l.contextAware && debugger.HasLoggerFromContext(ctx) {
		return debugger.LoggerFromContext(ctx)
	}

	if l.debuggerLogger == nil {
		return debugger.NoopLogger()
	}
	return l.debuggerLogger
}

// LogMode 设置日志级别
func (l *GormDebuggerLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info 记录信息级别日志
func (l *GormDebuggerLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		l.resolveLogger(ctx).Info(fmt.Sprintf(msg, data...))
	}
}

// Warn 记录警告级别日志
func (l *GormDebuggerLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		l.resolveLogger(ctx).Warn(fmt.Sprintf(msg, data...))
	}
}

// Error 记录错误级别日志
func (l *GormDebuggerLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		l.resolveLogger(ctx).Error(fmt.Sprintf(msg, data...))
	}
}

// Trace 记录SQL跟踪日志
// 这是记录SQL执行情况的核心方法
func (l *GormDebuggerLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel == logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rowsAffected := fc()
	target := l.resolveLogger(ctx)

	// 预分配map容量，避免多次内存分配
	fields := make(map[string]interface{}, 6)
	fields["duration_ms"] = elapsed.Milliseconds()
	fields["rows_affected"] = rowsAffected
	fields["status_text"] = "SQL执行成功"

	// 添加错误信息（如果有）
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		fields["error"] = err.Error()
		fields["error_type"] = fmt.Sprintf("%T", err)
		fields["status_text"] = "SQL执行失败"

		// 记录错误级别日志
		target.Error(sql, fields)
		return
	}

	// 判断是否为慢查询
	if elapsed > l.slowThreshold {
		fields["slow_query"] = true
		fields["slow_threshold_ms"] = l.slowThreshold.Milliseconds()
		fields["status_text"] = "慢SQL查询"
		target.Warn(sql, fields)
		return
	}

	// 记录调试级别日志：需同时满足GORM日志级别与目标Logger级别均达到Info
	if l.logLevel >= logger.Info && target.GetLevel() >= debugger.LevelInfo {
		target.Info(sql, fields)
	}
}

// SetDebuggerLogger 设置debugger日志记录器
// 用于在运行时动态更换日志记录器
//
// 注意：本方法只替换兜底Logger，不会关闭上下文感知模式；
// 固定Logger模式下等价于更换全部SQL日志的写入目标。
func (l *GormDebuggerLogger) SetDebuggerLogger(debuggerLogger debugger.LoggerInterface) {
	l.debuggerLogger = debuggerLogger
}

// GetDebuggerLogger 获取当前的debugger日志记录器
func (l *GormDebuggerLogger) GetDebuggerLogger() debugger.LoggerInterface {
	return l.debuggerLogger
}

// IsContextAware 判断当前是否处于上下文感知模式
//
// 返回：
//   - bool: 上下文感知模式返回 true，固定Logger模式返回 false
func (l *GormDebuggerLogger) IsContextAware() bool {
	return l.contextAware
}

// EnableSQLLogging 为GORM实例启用SQL日志记录（固定Logger模式）
// 支持统一的日志级别配置，优先使用debugger的日志级别设置
// 参数：
//   - db: GORM数据库实例
//   - debuggerLogger: debugger日志记录器
//   - opts: 可选参数，可以是logger.LogLevel或time.Duration
//   - 第一个参数：日志级别（可选，如果不提供则使用debugger的日志级别）
//   - 第二个参数：慢查询阈值（可选，默认200ms）
//
// 返回：
//   - *gorm.DB: 配置好SQL日志的GORM实例
func EnableSQLLogging(db *gorm.DB, debuggerLogger debugger.LoggerInterface, opts ...interface{}) *gorm.DB {
	return enableSQLLogging(db, debuggerLogger, false, opts...)
}

// EnableContextAwareSQLLogging 为GORM实例启用上下文感知的SQL日志记录
//
// 与 EnableSQLLogging 的区别：本函数全局只需在初始化阶段调用一次，
// 运行期每条SQL会依据 context.Context 中携带的请求级Logger路由到对应请求的日志中，
// 不会因并发请求互相覆盖而导致日志串台。
//
// 参数：
//   - db: GORM数据库实例
//   - fallbackLogger: 兜底日志记录器，context 中无Logger时使用，建议传 GetMainLogger()
//   - opts: 可选参数，语义同 EnableSQLLogging（日志级别、慢查询阈值）
//
// 返回：
//   - *gorm.DB: 配置好SQL日志的GORM实例
//
// 使用示例：
//
//	orm.EnableContextAwareSQLLogging(db, dbg.GetMainLogger())
//	orm.EnableContextAwareSQLLogging(db, dbg.GetMainLogger(), debugger.LevelInfo, 100*time.Millisecond)
func EnableContextAwareSQLLogging(db *gorm.DB, fallbackLogger debugger.LoggerInterface, opts ...interface{}) *gorm.DB {
	return enableSQLLogging(db, fallbackLogger, true, opts...)
}

// enableSQLLogging SQL日志启用逻辑的统一实现
// 参数：
//   - db: GORM数据库实例
//   - debuggerLogger: debugger日志记录器（上下文感知模式下作为兜底）
//   - contextAware: 是否启用上下文感知模式
//   - opts: 可选参数（日志级别、慢查询阈值）
//
// 返回：
//   - *gorm.DB: 配置好SQL日志的GORM实例
func enableSQLLogging(db *gorm.DB, debuggerLogger debugger.LoggerInterface, contextAware bool, opts ...interface{}) *gorm.DB {
	var (
		slowThreshold time.Duration = 200 * time.Millisecond
		logLevel                    = debugger.LevelSilent
	)

	// 上下文感知模式允许兜底Logger为空，此时取不到请求级Logger的SQL将被丢弃
	if debuggerLogger == nil {
		debuggerLogger = debugger.NoopLogger()
	} else {
		logLevel = debuggerLogger.GetLevel()
	}

	// 处理可选参数
	if len(opts) > 0 {
		switch v := opts[0].(type) {
		case debugger.LogLevel:
			logLevel = v
		case logger.LogLevel:
			logLevel = debugger.LogLevel(v)
		case string:
			switch v {
			case "silent":
				logLevel = debugger.LevelSilent
			case "error":
				logLevel = debugger.LevelError
			case "warn", "warning":
				logLevel = debugger.LevelWarn
			case "info", "debug":
				logLevel = debugger.LevelInfo
			}
		case int:
			if v >= int(debugger.LevelSilent) && v <= int(debugger.LevelInfo) {
				logLevel = debugger.LogLevel(v)
			}
		}
	}

	// 处理慢查询阈值参数
	if len(opts) > 1 {
		if threshold, ok := opts[1].(time.Duration); ok {
			slowThreshold = threshold
		}
	}

	// 创建GORM调试日志记录器，根据contextAware直接选择构造方式，避免重复创建导致的内存浪费
	var gormLogger *GormDebuggerLogger
	if contextAware {
		gormLogger = NewContextAwareGormDebuggerLogger(debuggerLogger, logLevel, slowThreshold)
	} else {
		gormLogger = NewGormDebuggerLogger(debuggerLogger, logLevel, slowThreshold)
	}

	// 配置GORM日志
	db.Config.Logger = gormLogger

	// 记录配置信息（仅在非静默模式下）
	if logLevel > debugger.LevelSilent {
		fields := map[string]interface{}{
			"debugger_level": logLevel,
			"slow_threshold": slowThreshold.String(),
		}
		if contextAware {
			fields["context_aware"] = true
		}
		debuggerLogger.Info("SQL日志记录已启用", fields)
	}

	return db
}

// WithSQLLogging GORM配置选项，用于启用SQL日志记录（固定Logger模式）
// 支持统一的日志级别配置，优先使用debugger的日志级别设置
// 使用示例：
//
//	db, err := gorm.Open(mysql.Open(dsn), WithSQLLogging(debuggerLogger))
//	db, err := gorm.Open(mysql.Open(dsn), WithSQLLogging(debuggerLogger, logger.Info))
//	db, err := gorm.Open(mysql.Open(dsn), WithSQLLogging(debuggerLogger, "debug", 100*time.Millisecond))
func WithSQLLogging(debuggerLogger debugger.LoggerInterface, opts ...interface{}) gorm.Option {
	return &sqlLoggingOption{
		debuggerLogger: debuggerLogger,
		opts:           opts,
	}
}

// WithContextAwareSQLLogging GORM配置选项，用于启用上下文感知的SQL日志记录
// 使用示例：
//
//	db, err := gorm.Open(mysql.Open(dsn), WithContextAwareSQLLogging(dbg.GetMainLogger()))
//	db, err := gorm.Open(mysql.Open(dsn), WithContextAwareSQLLogging(dbg.GetMainLogger(), debugger.LevelInfo, 100*time.Millisecond))
func WithContextAwareSQLLogging(fallbackLogger debugger.LoggerInterface, opts ...interface{}) gorm.Option {
	return &sqlLoggingOption{
		debuggerLogger: fallbackLogger,
		opts:           opts,
		contextAware:   true,
	}
}

// sqlLoggingOption SQL日志记录配置选项
type sqlLoggingOption struct {
	debuggerLogger debugger.LoggerInterface
	opts           []interface{}
	contextAware   bool // 是否启用上下文感知模式
}

// AfterInitialize 在GORM初始化后配置日志记录器
func (o *sqlLoggingOption) AfterInitialize(db *gorm.DB) error {
	enableSQLLogging(db, o.debuggerLogger, o.contextAware, o.opts...)
	return nil
}

// Apply 应用配置选项
func (o *sqlLoggingOption) Apply(*gorm.Config) error {
	return nil
}
