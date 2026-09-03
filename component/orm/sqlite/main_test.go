package sqlite

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/debugger"
)

// recordingLogger 测试用日志记录器，按调用顺序收集日志以便断言
type recordingLogger struct {
	mu    sync.Mutex
	name  string
	level debugger.LogLevel
	logs  []debugger.LoggerLog
}

// newRecordingLogger 创建一个可收集日志的测试Logger
func newRecordingLogger(name string) *recordingLogger {
	return &recordingLogger{name: name, level: debugger.LevelInfo}
}

func (l *recordingLogger) Info(msg any, fields ...map[string]interface{}) {
	l.record(debugger.LevelInfo, msg, fields...)
}

func (l *recordingLogger) Warn(msg any, fields ...map[string]interface{}) {
	l.record(debugger.LevelWarn, msg, fields...)
}

func (l *recordingLogger) Error(msg any, fields ...map[string]interface{}) {
	l.record(debugger.LevelError, msg, fields...)
}

// record 记录一条日志并合并附加字段
func (l *recordingLogger) record(level debugger.LogLevel, msg any, fields ...map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	merged := make(map[string]interface{})
	for _, f := range fields {
		for k, v := range f {
			merged[k] = v
		}
	}

	l.logs = append(l.logs, debugger.LoggerLog{
		Level:   level,
		Message: fmt.Sprint(msg),
		Fields:  merged,
	})
}

func (l *recordingLogger) WithFields(fields map[string]interface{}) debugger.LoggerInterface {
	return l
}
func (l *recordingLogger) GetLevel() debugger.LogLevel { return l.level }

// GetLogs 返回已收集日志的副本
func (l *recordingLogger) GetLogs() []debugger.LoggerLog {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]debugger.LoggerLog{}, l.logs...)
}

// ClearLogs 清空已收集日志
func (l *recordingLogger) ClearLogs() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = nil
}

// count 返回已收集的日志条数
func (l *recordingLogger) count() int {
	return len(l.GetLogs())
}

// newTestInstance 创建用于测试的 SQLite 实例（临时文件库）
//
// 参数：
//   - t: 测试实例
//
// 返回：
//   - *Instance: 已建立连接的 SQLite 实例
func newTestInstance(t *testing.T) *Instance {
	t.Helper()

	conf := jcbaseGo.SqlLiteStruct{
		DbFile:                                   filepath.Join(t.TempDir(), "test.db"),
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	inst, err := New(conf)
	if err != nil {
		t.Fatalf("创建 SQLite 实例失败: %v", err)
	}
	return inst
}

// TestInstance_上下文感知模式SQL按请求路由 验证 SQLite 实例的上下文感知 SQL 日志：
// 带请求级 Logger 的 SQL 进入该 Logger，无请求上下文的 SQL 回退兜底 Logger
func TestInstance_上下文感知模式SQL按请求路由(t *testing.T) {
	fallback := newRecordingLogger("fallback")
	ctxLogger := newRecordingLogger("ctx")

	inst := newTestInstance(t)
	inst.EnableContextAwareSQLLogging(fallback)

	if !inst.IsContextAware() {
		t.Fatal("启用后应处于上下文感知模式")
	}

	// 清空调试日志注入时的记录，避免影响后续断言
	fallback.ClearLogs()

	// 带请求级 Logger 的查询：SQL 应进入 ctxLogger
	ctx := debugger.ContextWithLogger(context.Background(), ctxLogger)
	if err := inst.GetDb().WithContext(ctx).Exec("SELECT 1").Error; err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	// 无请求上下文的查询：SQL 应回退到兜底 Logger
	if err := inst.GetDb().Exec("SELECT 2").Error; err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	if ctxLogger.count() != 1 {
		t.Fatalf("请求级Logger应收到1条SQL, got=%d", ctxLogger.count())
	}
	if got := ctxLogger.GetLogs()[0].Message; got != "SELECT 1" {
		t.Fatalf("请求级Logger收到的SQL不符: %s", got)
	}
	if fallback.count() != 1 {
		t.Fatalf("兜底Logger应收到1条SQL, got=%d", fallback.count())
	}
	if got := fallback.GetLogs()[0].Message; got != "SELECT 2" {
		t.Fatalf("兜底Logger收到的SQL不符: %s", got)
	}
}

// TestInstance_固定模式覆盖上下文感知 验证 SetDebuggerLogger 会关闭上下文感知模式，
// 之后 SQL 固定写入指定 Logger，忽略 context 中的请求级 Logger
func TestInstance_固定模式覆盖上下文感知(t *testing.T) {
	fixed := newRecordingLogger("fixed")
	ctxLogger := newRecordingLogger("ctx")

	inst := newTestInstance(t)
	inst.EnableContextAwareSQLLogging(newRecordingLogger("old-fallback"))
	inst.SetDebuggerLogger(fixed)

	if inst.IsContextAware() {
		t.Fatal("SetDebuggerLogger 后应回到固定模式")
	}

	// 清空调试日志注入时的记录
	fixed.ClearLogs()

	ctx := debugger.ContextWithLogger(context.Background(), ctxLogger)
	if err := inst.GetDb().WithContext(ctx).Exec("SELECT 3").Error; err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	if fixed.count() != 1 {
		t.Fatalf("固定Logger应收到1条SQL, got=%d", fixed.count())
	}
	if ctxLogger.count() != 0 {
		t.Fatalf("固定模式下不应写入context中的Logger, got=%d", ctxLogger.count())
	}
}

// TestInstance_EnableSQLLogging保持固定模式 验证便捷方法 EnableSQLLogging 默认固定模式
func TestInstance_EnableSQLLogging保持固定模式(t *testing.T) {
	logger := newRecordingLogger("fixed")
	ctxLogger := newRecordingLogger("ctx")

	inst := newTestInstance(t)
	inst.EnableSQLLogging(logger)

	if inst.IsContextAware() {
		t.Fatal("EnableSQLLogging 不应启用上下文感知模式")
	}

	// 清空调试日志注入时的记录
	logger.ClearLogs()

	ctx := debugger.ContextWithLogger(context.Background(), ctxLogger)
	if err := inst.GetDb().WithContext(ctx).Exec("SELECT 4").Error; err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	if logger.count() != 1 {
		t.Fatalf("固定Logger应收到1条SQL, got=%d", logger.count())
	}
	if ctxLogger.count() != 0 {
		t.Fatalf("固定模式下不应写入context中的Logger, got=%d", ctxLogger.count())
	}
}

// TestInstance_上下文感知模式nil兜底 验证 nil 兜底时，带请求级 Logger 的 SQL 仍能正确路由，
// 无请求上下文的 SQL 被丢弃，不会 panic
func TestInstance_上下文感知模式nil兜底(t *testing.T) {
	ctxLogger := newRecordingLogger("ctx")

	inst := newTestInstance(t)
	// nil 兜底 + 显式 Info 级别：无请求上下文的 SQL 丢弃，有请求级 Logger 的 SQL 记录
	inst.EnableContextAwareSQLLogging(nil, debugger.LevelInfo)

	if !inst.IsContextAware() {
		t.Fatal("启用后应处于上下文感知模式")
	}

	// 带请求级 Logger 的查询：SQL 应进入 ctxLogger
	ctx := debugger.ContextWithLogger(context.Background(), ctxLogger)
	if err := inst.GetDb().WithContext(ctx).Exec("SELECT 5").Error; err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	// 无请求上下文的查询：SQL 应被丢弃
	if err := inst.GetDb().Exec("SELECT 6").Error; err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	if ctxLogger.count() != 1 {
		t.Fatalf("请求级Logger应收到1条SQL, got=%d", ctxLogger.count())
	}
	if got := ctxLogger.GetLogs()[0].Message; got != "SELECT 5" {
		t.Fatalf("请求级Logger收到的SQL不符: %s", got)
	}
}
