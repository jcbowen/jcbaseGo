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
type GormDebuggerLogger struct {
	debuggerLogger debugger.LoggerInterface // debugger日志记录器
	logLevel       logger.LogLevel          // 日志级别
	slowThreshold  time.Duration            // 慢查询阈值
}

// NewGormDebuggerLogger 创建GORM调试日志记录器
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

// LogMode 设置日志级别
func (l *GormDebuggerLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info 记录信息级别日志
func (l *GormDebuggerLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		l.debuggerLogger.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn 记录警告级别日志
func (l *GormDebuggerLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		l.debuggerLogger.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error 记录错误级别日志
func (l *GormDebuggerLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		l.debuggerLogger.Error(fmt.Sprintf(msg, data...))
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

	// 预分配map容量，避免多次内存分配
	fields := make(map[string]interface{}, 6)
	fields["duration_ms"] = elapsed.Milliseconds()
	fields["rows_affected"] = rowsAffected

	// 添加错误信息（如果有）
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		fields["error"] = err.Error()
		fields["error_type"] = fmt.Sprintf("%T", err)
		fields["message_text"] = "SQL执行失败"

		// 记录错误级别日志
		l.debuggerLogger.Error(fmt.Sprintf("SQL执行失败:\n %s", sql), fields)
		return
	}

	// 判断是否为慢查询
	if elapsed > l.slowThreshold {
		fields["slow_query"] = true
		fields["slow_threshold_ms"] = l.slowThreshold.Milliseconds()
		l.debuggerLogger.Warn(fmt.Sprintf("慢SQL查询:\n %s", sql), fields)
		return
	}

	// 记录调试级别日志
	if l.debuggerLogger.GetLevel() >= debugger.LevelInfo {
		l.debuggerLogger.Info(fmt.Sprintf("SQL执行成功:\n %s", sql), fields)
	}
}

// SetDebuggerLogger 设置debugger日志记录器
// 用于在运行时动态更换日志记录器
func (l *GormDebuggerLogger) SetDebuggerLogger(debuggerLogger debugger.LoggerInterface) {
	l.debuggerLogger = debuggerLogger
}

// GetDebuggerLogger 获取当前的debugger日志记录器
func (l *GormDebuggerLogger) GetDebuggerLogger() debugger.LoggerInterface {
	return l.debuggerLogger
}

// EnableSQLLogging 为GORM实例启用SQL日志记录
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
	var (
		slowThreshold time.Duration = 200 * time.Millisecond
		logLevel                    = debuggerLogger.GetLevel()
	)

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

	// 创建GORM调试日志记录器，使用统一的日志级别
	gormLogger := NewGormDebuggerLogger(debuggerLogger, logLevel, slowThreshold)

	// 配置GORM日志
	db.Config.Logger = gormLogger

	// 记录配置信息（仅在非静默模式下）
	if logLevel > debugger.LevelSilent {
		debuggerLogger.Info("SQL日志记录已启用", map[string]interface{}{
			"debugger_level": logLevel,
			"slow_threshold": slowThreshold.String(),
		})
	}

	return db
}

// WithSQLLogging GORM配置选项，用于启用SQL日志记录
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

// sqlLoggingOption SQL日志记录配置选项
type sqlLoggingOption struct {
	debuggerLogger debugger.LoggerInterface
	opts           []interface{}
}

// AfterInitialize 在GORM初始化后配置日志记录器
func (o *sqlLoggingOption) AfterInitialize(db *gorm.DB) error {
	EnableSQLLogging(db, o.debuggerLogger, o.opts...)
	return nil
}

// Apply 应用配置选项
func (o *sqlLoggingOption) Apply(*gorm.Config) error {
	return nil
}
