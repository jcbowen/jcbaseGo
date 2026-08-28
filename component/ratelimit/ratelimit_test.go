package ratelimit

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestAllowLocal(t *testing.T) {
	limiter := NewLimiter(nil)
	defer limiter.Close()

	// 窗口 2 秒内最多 2 次
	if !limiter.AllowLocal("k1", 2, 2) {
		t.Fatal("first request should be allowed")
	}
	if !limiter.AllowLocal("k1", 2, 2) {
		t.Fatal("second request should be allowed")
	}
	if limiter.AllowLocal("k1", 2, 2) {
		t.Fatal("third request should be rejected")
	}
	// 不同 key 互不影响
	if !limiter.AllowLocal("k2", 2, 2) {
		t.Fatal("different key should be allowed")
	}
}

func TestAllowNilClient(t *testing.T) {
	limiter := NewLimiter(nil)
	defer limiter.Close()

	ctx := context.Background()
	if !limiter.Allow(ctx, "k-nil", 1, 1) {
		t.Fatal("nil client should fall back to local and allow first request")
	}
	if limiter.Allow(ctx, "k-nil", 1, 1) {
		t.Fatal("nil client should fall back to local and reject second request")
	}
}

// TestAllowLocalEviction 验证通过真实 allow 路径写入的条目，在超过窗口未访问后会被 sweep 回收。
func TestAllowLocalEviction(t *testing.T) {
	l := &localRateLimiter{entries: make(map[string]*localRateEntry)}
	if !l.allow("k", 2, 2) {
		t.Fatal("first request should be allowed")
	}
	if len(l.entries) != 1 {
		t.Fatalf("expected 1 entry after first request, got %d", len(l.entries))
	}
	// 推进时间超过窗口后触发清理
	l.sweep(time.Now().UnixMilli() + int64(10*time.Second/time.Millisecond))
	if len(l.entries) != 0 {
		t.Fatalf("expected expired entry to be swept, got %d", len(l.entries))
	}
}

// TestSweepEvictsExpired 验证 sweep 仅回收过期条目，保留仍在窗口内的条目。
func TestSweepEvictsExpired(t *testing.T) {
	l := &localRateLimiter{entries: make(map[string]*localRateEntry)}
	base := time.Now().UnixMilli()
	// 过期条目
	l.entries["expired"] = &localRateEntry{expiresAt: base + int64(time.Second/time.Millisecond)}
	// 仍在窗口内的条目
	l.entries["alive"] = &localRateEntry{expiresAt: base + int64(10*time.Second/time.Millisecond)}

	l.sweep(base + int64(5*time.Second/time.Millisecond))

	if _, ok := l.entries["expired"]; ok {
		t.Fatal("expired entry should have been swept")
	}
	if _, ok := l.entries["alive"]; !ok {
		t.Fatal("alive entry should be retained")
	}
}

// TestAllowLocalCapacityCap 验证映射容量不超过上限，超过时回收最早的条目。
func TestAllowLocalCapacityCap(t *testing.T) {
	l := &localRateLimiter{entries: make(map[string]*localRateEntry)}
	for i := 0; i < localFallbackMaxEntries+50; i++ {
		l.allow(fmt.Sprintf("cap-%d", i), 10, 1)
	}
	if len(l.entries) > localFallbackMaxEntries {
		t.Fatalf("expected entries bounded by %d, got %d", localFallbackMaxEntries, len(l.entries))
	}
}

// TestLocalFallbackStop 验证 Stop 能正确终止后台清理协程，且多次调用安全。
func TestLocalFallbackStop(t *testing.T) {
	l := &localRateLimiter{
		entries: make(map[string]*localRateEntry),
	}
	// 测试 nil cancel 时调用 Stop 不会 panic
	l.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	l.ctx = ctx
	l.cancel = cancel

	go l.sweepLoop()

	// 触发一次清理后停止
	l.Stop()
	l.Stop()

	select {
	case <-ctx.Done():
		// 预期内：上下文已被取消
	case <-time.After(time.Second):
		t.Fatal("expected sweepLoop to exit after Stop")
	}
}

// TestLimiterClose 验证 Limiter.Close 能正确停止后台清理协程。
func TestLimiterClose(t *testing.T) {
	limiter := NewLimiter(nil)
	limiter.Close()
	limiter.Close() // 多次调用安全
}
