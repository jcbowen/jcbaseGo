package debugger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMemoryStorageHostFilter 验证内存存储支持 host 字段筛选，
// 与文件/数据库存储保持一致的筛选语义。
func TestMemoryStorageHostFilter(t *testing.T) {
	storage, err := NewMemoryStorage(100)
	assert.NoError(t, err)

	now := time.Now()
	entries := []*LogEntry{
		{
			ID:        "mem-host-1",
			Timestamp: now,
			URL:       "/api/test",
			Host:      "host1.example.com",
			Method:    "GET",
		},
		{
			ID:        "mem-host-2",
			Timestamp: now.Add(-time.Second),
			URL:       "/api/test",
			Host:      "host2.example.com",
			Method:    "POST",
		},
		{
			ID:        "mem-host-3",
			Timestamp: now.Add(-2 * time.Second),
			URL:       "/api/other",
			Host:      "host1.example.com",
			Method:    "GET",
		},
	}

	for _, entry := range entries {
		assert.NoError(t, storage.Save(entry))
	}

	// 验证按 host 精确/包含匹配
	result, total, err := storage.FindAll(1, 10, map[string]interface{}{"host": "host1.example.com"})
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, result, 2)

	// 验证 host 包含匹配
	result, total, err = storage.FindAll(1, 10, map[string]interface{}{"host": "example.com"})
	assert.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, result, 3)

	// 验证关键词搜索能命中 host 字段
	result, total, err = storage.Search("host1", 1, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, result, 2)
}
