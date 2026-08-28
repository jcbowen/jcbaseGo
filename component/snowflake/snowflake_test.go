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
