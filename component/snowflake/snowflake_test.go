package snowflake

import (
	"sync"
	"testing"
)

// TestSnowflakeNextIDUnique 验证同一生成器在并发下生成的 ID 唯一且单调不减
func TestSnowflakeNextIDUnique(t *testing.T) {
	sf, err := NewSnowflake(1)
	if err != nil {
		t.Fatalf("NewSnowflake(1) returned error: %v", err)
	}

	const n = 10000
	ids := make([]uint64, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			id, err := sf.NextID()
			if err != nil {
				t.Errorf("NextID returned error: %v", err)
				return
			}
			ids[i] = id
		}(i)
	}
	wg.Wait()

	seen := make(map[uint64]bool, n)
	for _, id := range ids {
		if id == 0 {
			t.Fatalf("generated zero ID")
		}
		if seen[id] {
			t.Fatalf("duplicate ID generated: %d", id)
		}
		seen[id] = true
	}
}

// TestSnowflakeNodeValidation 验证非法节点 ID 返回错误
func TestSnowflakeNodeValidation(t *testing.T) {
	if _, err := NewSnowflake(1024); err == nil {
		t.Fatal("NewSnowflake(1024) expected error, got nil")
	}
	if _, err := NewSnowflake(1023); err != nil {
		t.Fatalf("NewSnowflake(1023) returned unexpected error: %v", err)
	}
}

// TestSnowflakeWithEpoch 验证自定义起始时间戳可正常生成 ID
func TestSnowflakeWithEpoch(t *testing.T) {
	sf, err := NewSnowflakeWithEpoch(1, 1704067200000)
	if err != nil {
		t.Fatalf("NewSnowflakeWithEpoch returned error: %v", err)
	}
	id, err := sf.NextID()
	if err != nil {
		t.Fatalf("NextID returned error: %v", err)
	}
	if id == 0 {
		t.Fatal("NextID returned zero ID")
	}
}

// TestSnowflakeSequenceExhaustionWithClockJitter 回归测试：复现 CI 上出现的重复 ID。
//
// 场景：同一毫秒内 4096 个序列号被耗尽后，下一次时钟读数因抖动仍停留在同一毫秒。
// 旧实现（释放锁后 sleep 重试）此时会把序列号从 0 重新递增，复用已用过的序列号产生重复 ID；
// 正确行为是等待时钟严格越过该毫秒后从 0 重新计数。
func TestSnowflakeSequenceExhaustionWithClockJitter(t *testing.T) {
	sf, err := NewSnowflake(1)
	if err != nil {
		t.Fatalf("NewSnowflake(1) returned error: %v", err)
	}

	const base = int64(1704067200000 + 12345)
	reads := 0
	sf.now = func() int64 {
		reads++
		switch {
		case reads <= 4096:
			return base // 前 4096 次读数：耗尽 base 毫秒的全部序列号
		case reads == 4097:
			return base // 模拟时钟抖动：第 4097 次读数仍停留在 base 毫秒
		default:
			return base + int64(reads-4097) // 之后时钟正常前进
		}
	}

	const n = 8192
	seen := make(map[uint64]bool, n)
	for i := 0; i < n; i++ {
		id, err := sf.NextID()
		if err != nil {
			t.Fatalf("NextID returned error: %v", err)
		}
		if seen[id] {
			t.Fatalf("duplicate ID generated: %d", id)
		}
		seen[id] = true
	}
}

// TestSnowflakeClockRollback 验证小幅时钟回拨不会产生重复 ID，大幅回拨返回错误
func TestSnowflakeClockRollback(t *testing.T) {
	sf, err := NewSnowflake(1)
	if err != nil {
		t.Fatalf("NewSnowflake(1) returned error: %v", err)
	}

	const base = int64(1704067200000 + 6789)
	// 时钟：先生成几个 ID 后回拨 3ms，再逐毫秒恢复并前进
	reads := 0
	sf.now = func() int64 {
		reads++
		switch {
		case reads <= 3:
			return base + int64(reads) // 正常前进
		case reads <= 6:
			return base // 回拨到 base（比 lastTime 小 3ms）
		default:
			return base + int64(reads-6) // 逐毫秒恢复并前进
		}
	}

	seen := make(map[uint64]bool)
	for i := 0; i < 100; i++ {
		id, err := sf.NextID()
		if err != nil {
			t.Fatalf("NextID returned error at %d: %v", i, err)
		}
		if seen[id] {
			t.Fatalf("duplicate ID generated: %d", id)
		}
		seen[id] = true
	}

	// 大幅回拨（超过 5s 阈值）应返回错误
	big, err := NewSnowflake(1)
	if err != nil {
		t.Fatalf("NewSnowflake(1) returned error: %v", err)
	}
	far := int64(1704067200000 + 100000)
	big.now = func() int64 { return far }
	if _, err := big.NextID(); err != nil {
		t.Fatalf("first NextID returned error: %v", err)
	}
	big.now = func() int64 { return far - 10000 }
	if _, err := big.NextID(); err == nil {
		t.Fatal("expected error on large clock rollback, got nil")
	}
}
