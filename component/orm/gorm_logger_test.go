package orm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jcbowen/jcbaseGo/component/debugger"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

func (l *recordingLogger) ClearLogs() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = nil
}

// count 返回已收集的日志条数
func (l *recordingLogger) count() int {
	return len(l.GetLogs())
}

// traceSQL 模拟一次GORM的SQL执行回调
func traceSQL(l *GormDebuggerLogger, ctx context.Context) {
	l.Trace(ctx, time.Now(), func() (string, int64) { return "SELECT 1", 1 }, nil)
}

// TestGormDebuggerLogger_固定模式忽略上下文 验证模式A始终写入构造时注入的Logger
func TestGormDebuggerLogger_固定模式忽略上下文(t *testing.T) {
	fixed := newRecordingLogger("fixed")
	ctxLogger := newRecordingLogger("ctx")

	gl := NewGormDebuggerLogger(fixed, debugger.LevelInfo, time.Second)
	traceSQL(gl, debugger.ContextWithLogger(context.Background(), ctxLogger))

	if fixed.count() != 1 {
		t.Fatalf("固定模式应写入注入的Logger, got=%d", fixed.count())
	}
	if ctxLogger.count() != 0 {
		t.Fatalf("固定模式不应写入上下文Logger, got=%d", ctxLogger.count())
	}
}

// TestGormDebuggerLogger_上下文模式优先取上下文 验证模式B按 context 路由日志
func TestGormDebuggerLogger_上下文模式优先取上下文(t *testing.T) {
	fallback := newRecordingLogger("fallback")
	ctxLogger := newRecordingLogger("ctx")

	gl := NewContextAwareGormDebuggerLogger(fallback, debugger.LevelInfo, time.Second)
	traceSQL(gl, debugger.ContextWithLogger(context.Background(), ctxLogger))

	if ctxLogger.count() != 1 {
		t.Fatalf("上下文模式应优先写入上下文Logger, got=%d", ctxLogger.count())
	}
	if fallback.count() != 0 {
		t.Fatalf("取到上下文Logger时不应写入兜底Logger, got=%d", fallback.count())
	}
}

// TestGormDebuggerLogger_上下文模式回退兜底 验证模式B在无上下文Logger时使用兜底
func TestGormDebuggerLogger_上下文模式回退兜底(t *testing.T) {
	fallback := newRecordingLogger("fallback")

	gl := NewContextAwareGormDebuggerLogger(fallback, debugger.LevelInfo, time.Second)
	traceSQL(gl, context.Background())

	if fallback.count() != 1 {
		t.Fatalf("无上下文Logger时应回退兜底Logger, got=%d", fallback.count())
	}
}

// TestGormDebuggerLogger_上下文模式兜底为空不panic 验证兜底缺失时安全丢弃
func TestGormDebuggerLogger_上下文模式兜底为空不panic(t *testing.T) {
	gl := NewContextAwareGormDebuggerLogger(nil, debugger.LevelInfo, time.Second)
	traceSQL(gl, context.Background())
}

// TestGormDebuggerLogger_并发上下文互不串台 验证模式B在高并发下日志归属正确
func TestGormDebuggerLogger_并发上下文互不串台(t *testing.T) {
	const concurrency = 100

	fallback := newRecordingLogger("fallback")
	gl := NewContextAwareGormDebuggerLogger(fallback, debugger.LevelInfo, time.Second)

	loggers := make([]*recordingLogger, concurrency)
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		loggers[i] = newRecordingLogger(fmt.Sprintf("req-%d", i))

		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ctx := debugger.ContextWithLogger(context.Background(), loggers[idx])
			traceSQL(gl, ctx)
		}(i)
	}
	wg.Wait()

	for i, l := range loggers {
		if l.count() != 1 {
			t.Fatalf("第 %d 个请求的日志条数应为 1, got=%d", i, l.count())
		}
		if l.GetLogs()[0].Message != "SELECT 1" {
			t.Fatalf("第 %d 个请求收到的SQL不属于自己: %s", i, l.GetLogs()[0].Message)
		}
	}
	if fallback.count() != 0 {
		t.Fatalf("全部命中上下文Logger时兜底不应收到日志, got=%d", fallback.count())
	}
}

// TestGormDebuggerLogger_LogMode保留模式标志 验证调整日志级别后上下文感知模式不丢失
func TestGormDebuggerLogger_LogMode保留模式标志(t *testing.T) {
	fallback := newRecordingLogger("fallback")
	ctxLogger := newRecordingLogger("ctx")

	gl := NewContextAwareGormDebuggerLogger(fallback, debugger.LevelInfo, time.Second)
	converted, ok := gl.LogMode(logger.Warn).(*GormDebuggerLogger)
	if !ok {
		t.Fatal("LogMode 应返回 GormDebuggerLogger")
	}
	if !converted.IsContextAware() {
		t.Fatal("LogMode 后应保留上下文感知模式")
	}

	// 降为 Warn 级别后，普通SQL不再记录（Info 被过滤）
	traceSQL(converted, debugger.ContextWithLogger(context.Background(), ctxLogger))
	if ctxLogger.count() != 0 {
		t.Fatalf("Warn级别下不应记录普通SQL, got=%d", ctxLogger.count())
	}
}

// TestGormDebuggerLogger_慢查询与错误分级 验证慢查询走Warn、失败走Error
func TestGormDebuggerLogger_慢查询与错误分级(t *testing.T) {
	slow := newRecordingLogger("slow")
	gl := NewGormDebuggerLogger(slow, debugger.LevelInfo, 10*time.Millisecond)

	// 慢查询：begin 设为 500ms 前，超过阈值
	gl.Trace(context.Background(), time.Now().Add(-500*time.Millisecond),
		func() (string, int64) { return "SELECT slow", 1 }, nil)

	// 执行失败：附带错误
	gl.Trace(context.Background(), time.Now(),
		func() (string, int64) { return "SELECT bad", 0 }, errors.New("connection reset"))

	logs := slow.GetLogs()
	if len(logs) != 2 {
		t.Fatalf("期望记录 2 条日志, got=%d", len(logs))
	}
	if logs[0].Level != debugger.LevelWarn {
		t.Fatalf("慢查询应为 Warn 级别, got=%v", logs[0].Level)
	}
	if logs[1].Level != debugger.LevelError {
		t.Fatalf("执行失败应为 Error 级别, got=%v", logs[1].Level)
	}
}

// TestGormDebuggerLogger_记录不存在不算错误 验证 ErrRecordNotFound 不升级为错误日志
func TestGormDebuggerLogger_记录不存在不算错误(t *testing.T) {
	target := newRecordingLogger("target")
	gl := NewGormDebuggerLogger(target, debugger.LevelInfo, time.Second)

	gl.Trace(context.Background(), time.Now(),
		func() (string, int64) { return "SELECT empty", 0 }, gorm.ErrRecordNotFound)

	logs := target.GetLogs()
	if len(logs) != 1 {
		t.Fatalf("期望记录 1 条日志, got=%d", len(logs))
	}
	if logs[0].Level != debugger.LevelInfo {
		t.Fatalf("记录不存在应仍按 Info 记录, got=%v", logs[0].Level)
	}
}
