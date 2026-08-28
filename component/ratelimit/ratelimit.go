// Package ratelimit 提供基于 Redis 有序集合的滑动窗口限流，并内置本地内存降级。
//
// 核心算法与业务解耦：限流 key 由调用方自行构造，Redis 客户端由调用方注入。
// 当 Redis 不可用或执行失败时自动降级到本地内存限流（多实例不共享计数）。
//
// 使用方式：
//
//	limiter := ratelimit.NewLimiter(redisClient)
//	defer limiter.Close()
//	if !limiter.Allow(ctx, "api:create", 10, 60) {
//	    // 触发限流
//	}
package ratelimit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// localFallbackMaxEntries 本地降级限流映射的最大容量。
	// 超过后新增条目会触发对最早过期条目的回收，作为内存增长的硬上限，
	// 防止 key 基数极高时 entries 映射持续扩张。
	localFallbackMaxEntries = 10000

	// localFallbackSweepInterval 后台清理协程的扫描间隔，用于回收长时间未访问的过期条目。
	localFallbackSweepInterval = time.Minute
)

// Limiter 限流器。由调用方持有实例，并通过 Close 方法释放后台资源。
type Limiter struct {
	client     redis.UniversalClient
	local      *localRateLimiter
	instanceID string
	counter    int64
}

// NewLimiter 创建一个新的限流器实例。
//
// 参数:
//   - client: Redis 客户端；传 nil 时全部降级到本地内存限流。
//
// 返回值:
//   - *Limiter: 限流器实例，使用完毕后需调用 Close 释放资源。
func NewLimiter(client redis.UniversalClient) *Limiter {
	return &Limiter{
		client:     client,
		local:      newLocalFallback(),
		instanceID: generateRedisMemberInstanceID(),
	}
}

// Allow 基于 Redis 有序集合的滑动窗口限流。
// client 为 nil 或执行失败时自动降级到本地内存限流。
//
// 参数:
//   - ctx: 请求上下文
//   - key: 限流 key（业务限流 key 由调用方自行构造）
//   - maxRequests: 窗口内最大请求数
//   - windowSeconds: 窗口长度（秒）
//
// 返回值:
//   - bool: true 表示允许通过，false 表示触发限流
func (l *Limiter) Allow(ctx context.Context, key string, maxRequests int, windowSeconds int) bool {
	if l == nil {
		return true
	}
	if l.client == nil {
		return l.AllowLocal(key, maxRequests, windowSeconds)
	}

	now := time.Now().UnixMilli()
	windowStart := now - int64(windowSeconds)*1000
	// member 使用 "毫秒时间戳:实例标识:自增序号"，确保同一毫秒、多实例下的多次请求不会被覆盖
	member := fmt.Sprintf("%d:%s:%d", now, l.instanceID, atomic.AddInt64(&l.counter, 1))

	pipe := l.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
	pipe.ZAdd(ctx, key, &redis.Z{Score: float64(now), Member: member})
	pipe.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)
	pipe.ZCard(ctx, key)

	res, err := pipe.Exec(ctx)
	if err != nil {
		return l.AllowLocal(key, maxRequests, windowSeconds)
	}
	if len(res) < 4 {
		return l.AllowLocal(key, maxRequests, windowSeconds)
	}
	countCmd, ok := res[3].(*redis.IntCmd)
	if !ok {
		return l.AllowLocal(key, maxRequests, windowSeconds)
	}
	count, err := countCmd.Result()
	if err != nil {
		return l.AllowLocal(key, maxRequests, windowSeconds)
	}
	return int(count) <= maxRequests
}

// AllowLocal 本地内存滑动窗口限流（多实例不共享计数）。
//
// 参数:
//   - key: 限流 key
//   - maxRequests: 窗口内最大请求数
//   - windowSeconds: 窗口长度（秒）
//
// 返回值:
//   - bool: true 表示允许通过，false 表示触发限流
func (l *Limiter) AllowLocal(key string, maxRequests, windowSeconds int) bool {
	if l == nil {
		return true
	}
	return l.local.allow(key, maxRequests, windowSeconds)
}

// Close 释放限流器占用的资源，包括停止本地降级限流器的后台清理协程。
// 多次调用安全。
func (l *Limiter) Close() {
	if l != nil && l.local != nil {
		l.local.Stop()
	}
}

// localRateEntry 本地内存限流条目
type localRateEntry struct {
	mu sync.Mutex
	// expiresAt 该条目最后有效时间 + 窗口毫秒；当前时间超过该值即视为过期可回收。
	// 每次命中 allow 都会刷新，因此持续活跃的条目不会被回收，仅空闲超过窗口的条目会被清理。
	expiresAt  int64
	timestamps []int64
}

// localRateLimiter Redis 不可用时的本地内存降级限流器。
// 仅在当前进程内生效，多实例场景下无法共享计数。
type localRateLimiter struct {
	mu      sync.RWMutex
	entries map[string]*localRateEntry
	// ctx 与 cancel 用于管控后台 sweepLoop 协程的生命周期，
	// 支持调用方在程序退出或测试结束时优雅停止清理协程。
	ctx    context.Context
	cancel context.CancelFunc
}

// newLocalFallback 创建本地降级限流器并启动后台清理协程
func newLocalFallback() *localRateLimiter {
	ctx, cancel := context.WithCancel(context.Background())
	l := &localRateLimiter{
		entries: make(map[string]*localRateEntry),
		ctx:     ctx,
		cancel:  cancel,
	}
	go l.sweepLoop()
	return l
}

// Stop 停止后台清理协程。多次调用安全。
func (l *localRateLimiter) Stop() {
	if l != nil && l.cancel != nil {
		l.cancel()
	}
}

// generateRedisMemberInstanceID 生成一个 8 字节十六进制实例标识
func generateRedisMemberInstanceID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

// allow 使用本地内存滑动窗口判断是否允许通过
func (l *localRateLimiter) allow(key string, maxRequests int, windowSeconds int) bool {
	now := time.Now().UnixMilli()
	windowMs := int64(windowSeconds) * int64(time.Second/time.Millisecond)
	windowStart := now - windowMs

	l.mu.RLock()
	entry, ok := l.entries[key]
	l.mu.RUnlock()
	if !ok {
		l.mu.Lock()
		entry, ok = l.entries[key]
		if !ok {
			// 创建时即写入过期时间，避免被后台清理协程在条目尚未写入 expiresAt 之前误删
			entry = &localRateEntry{expiresAt: now + windowMs}
			l.entries[key] = entry
			// 超过最大容量时回收最早过期的条目，作为内存硬上限
			if len(l.entries) > localFallbackMaxEntries {
				l.evictOldestLocked()
			}
		}
		l.mu.Unlock()
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	filtered := entry.timestamps[:0]
	for _, ts := range entry.timestamps {
		if ts > windowStart {
			filtered = append(filtered, ts)
		}
	}
	filtered = append(filtered, now)
	entry.timestamps = filtered
	// 刷新过期时间：该条目在窗口内仍有效，持续活跃不会被回收
	entry.expiresAt = now + windowMs

	return len(filtered) <= maxRequests
}

// sweepLoop 后台清理循环，按固定间隔回收过期条目，
// 当限流器生命周期结束（Stop/Close 被调用）时立即退出。
func (l *localRateLimiter) sweepLoop() {
	ticker := time.NewTicker(localFallbackSweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.sweep(time.Now().UnixMilli())
		case <-l.ctx.Done():
			return
		}
	}
}

// sweep 回收所有过期条目（now 大于条目 expiresAt 即视为过期）。
// 调用方无需持锁，方法内部自行加锁。now 由外部传入便于单元测试注入时间。
//
// 参数:
//   - now: 当前时间戳（毫秒）
func (l *localRateLimiter) sweep(now int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for key, entry := range l.entries {
		entry.mu.Lock()
		expired := now > entry.expiresAt
		entry.mu.Unlock()
		if expired {
			delete(l.entries, key)
		}
	}
}

// evictOldestLocked 在超过最大容量时回收一个最早过期的条目。
// 调用方需持有 l.mu 写锁。
func (l *localRateLimiter) evictOldestLocked() {
	var oldestKey string
	var oldestExp int64
	first := true
	for k, e := range l.entries {
		e.mu.Lock()
		exp := e.expiresAt
		e.mu.Unlock()
		if first || exp < oldestExp {
			oldestKey = k
			oldestExp = exp
			first = false
		}
	}
	if oldestKey != "" {
		delete(l.entries, oldestKey)
	}
}
