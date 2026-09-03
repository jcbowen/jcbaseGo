package debugger

import (
	"context"
	"testing"
)

// stubLogger 测试用的最小日志记录器实现
// 仅用于验证Logger在 context 中的存取语义，不关心日志内容
type stubLogger struct {
	level LogLevel
}

func (l *stubLogger) Info(msg any, fields ...map[string]interface{})  {}
func (l *stubLogger) Warn(msg any, fields ...map[string]interface{})  {}
func (l *stubLogger) Error(msg any, fields ...map[string]interface{}) {}

func (l *stubLogger) WithFields(fields map[string]interface{}) LoggerInterface { return l }
func (l *stubLogger) GetLevel() LogLevel                                       { return l.level }
func (l *stubLogger) GetLogs() []LoggerLog                                     { return nil }
func (l *stubLogger) ClearLogs()                                               {}

// TestContextWithLogger_注入与取回 验证Logger可沿 context 透传
func TestContextWithLogger_注入与取回(t *testing.T) {
	base := &stubLogger{level: LevelInfo}
	ctx := ContextWithLogger(context.Background(), base)

	if !HasLoggerFromContext(ctx) {
		t.Fatal("期望 context 中存在Logger")
	}

	if got := LoggerFromContext(ctx); got != LoggerInterface(base) {
		t.Fatalf("取回的Logger与注入的不一致: got=%v, want=%v", got, base)
	}
}

// TestContextWithLogger_覆盖已有Logger 验证重复注入时以最后一次为准
func TestContextWithLogger_覆盖已有Logger(t *testing.T) {
	first := &stubLogger{level: LevelInfo}
	second := &stubLogger{level: LevelWarn}

	ctx := ContextWithLogger(ContextWithLogger(context.Background(), first), second)
	if got := LoggerFromContext(ctx); got != LoggerInterface(second) {
		t.Fatal("期望取回最后一次注入的Logger")
	}
}

// TestContextWithLogger_传入nil等价于移除 验证nil参数不会产生nil接口
func TestContextWithLogger_传入nil等价于移除(t *testing.T) {
	ctx := ContextWithLogger(ContextWithLogger(context.Background(), &stubLogger{level: LevelInfo}), nil)

	if HasLoggerFromContext(ctx) {
		t.Fatal("传入 nil 后期望 context 中不存在Logger")
	}
	if _, ok := LoggerFromContext(ctx).(noopLogger); !ok {
		t.Fatal("缺失Logger时期望返回 noopLogger")
	}
}

// TestContextWithoutLogger_移除Logger 验证可主动摘除上下文中的Logger
func TestContextWithoutLogger_移除Logger(t *testing.T) {
	ctx := ContextWithLogger(context.Background(), &stubLogger{level: LevelInfo})
	ctx = ContextWithoutLogger(ctx)

	if HasLoggerFromContext(ctx) {
		t.Fatal("期望Logger已被移除")
	}
}

// TestContextWithoutLogger_无Logger时原样返回 验证不产生多余的 context 包装
func TestContextWithoutLogger_无Logger时原样返回(t *testing.T) {
	base := context.Background()
	if got := ContextWithoutLogger(base); got != base {
		t.Fatal("无Logger时期望原样返回同一个 context")
	}
}

// TestLoggerFromContext_空上下文返回noopLogger 验证缺失时返回空实现而非nil
func TestLoggerFromContext_空上下文返回noopLogger(t *testing.T) {
	cases := map[string]context.Context{
		"nil上下文":     nil,
		"无Logger上下文": context.Background(),
	}

	for name, ctx := range cases {
		logger := LoggerFromContext(ctx)
		if logger == nil {
			t.Fatalf("%s: 期望返回空实现Logger，实际返回 nil", name)
		}
		if _, ok := logger.(noopLogger); !ok {
			t.Fatalf("%s: 期望返回 noopLogger", name)
		}
	}
}

// TestNoopLogger_可安全调用 验证空操作Logger不会panic且级别为静默
func TestNoopLogger_可安全调用(t *testing.T) {
	logger := NoopLogger()

	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	if len(logger.GetLogs()) != 0 {
		t.Fatal("noopLogger 不应收集任何日志")
	}
	if logger.GetLevel() != LevelSilent {
		t.Fatalf("noopLogger 级别应为静默, got=%v", logger.GetLevel())
	}
}
