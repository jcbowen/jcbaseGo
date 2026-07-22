package debugger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestDatabaseStorageStatusCodeFilter 验证数据库存储在使用字符串类型的 status_code 筛选时，
// 能正确转换为整型并与数据库字段匹配。
//
// 该测试覆盖 parseFilters 与 HTTP 控制器传参均为字符串的场景，防止类型不匹配导致查询异常。
func TestDatabaseStorageStatusCodeFilter(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	storage, err := NewDatabaseStorage(db, 0)
	assert.NoError(t, err)

	now := time.Now()
	entries := []*LogEntry{
		{
			ID:         "db-test-1",
			Timestamp:  now,
			URL:        "/api/test",
			StatusCode: 200,
			Method:     "GET",
		},
		{
			ID:         "db-test-2",
			Timestamp:  now.Add(-time.Second),
			URL:        "/api/test",
			StatusCode: 404,
			Method:     "POST",
		},
		{
			ID:         "db-test-3",
			Timestamp:  now.Add(-2 * time.Second),
			URL:        "/api/other",
			StatusCode: 500,
			Method:     "GET",
		},
	}

	for _, entry := range entries {
		assert.NoError(t, storage.Save(entry))
	}

	tests := []struct {
		name     string
		filters  map[string]interface{}
		expected int
	}{
		{
			name:     "字符串 status_code 精确匹配 200",
			filters:  map[string]interface{}{"status_code": "200"},
			expected: 1,
		},
		{
			name:     "字符串 status_code 精确匹配 404",
			filters:  map[string]interface{}{"status_code": "404"},
			expected: 1,
		},
		{
			name:     "整型 status_code 精确匹配 500",
			filters:  map[string]interface{}{"status_code": 500},
			expected: 1,
		},
		{
			name:     "不存在的 status_code",
			filters:  map[string]interface{}{"status_code": "999"},
			expected: 0,
		},
		{
			name:     "method 与 status_code 组合筛选",
			filters:  map[string]interface{}{"method": "GET", "status_code": "200"},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, total, err := storage.FindAll(1, 10, tt.filters)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, total, "总数应正确")
			assert.Len(t, result, tt.expected, "返回结果数量应与总数一致")
		})
	}
}

// TestDatabaseStorageHostFilter 验证数据库存储支持 host 字段筛选，
// 同时验证 host 在 Save/FindByID 过程中被正确持久化与读取。
func TestDatabaseStorageHostFilter(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	storage, err := NewDatabaseStorage(db, 0)
	assert.NoError(t, err)

	now := time.Now()
	entries := []*LogEntry{
		{
			ID:        "host-test-1",
			Timestamp: now,
			URL:       "/api/test",
			Host:      "host1.example.com",
			Method:    "GET",
		},
		{
			ID:        "host-test-2",
			Timestamp: now.Add(-time.Second),
			URL:       "/api/test",
			Host:      "host2.example.com",
			Method:    "POST",
		},
		{
			ID:        "host-test-3",
			Timestamp: now.Add(-2 * time.Second),
			URL:       "/api/other",
			Host:      "host1.example.com",
			Method:    "GET",
		},
	}

	for _, entry := range entries {
		assert.NoError(t, storage.Save(entry))
	}

	// 验证按 host 筛选
	result, total, err := storage.FindAll(1, 10, map[string]interface{}{"host": "host1.example.com"})
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, result, 2)

	// 验证 host 包含匹配（域名的一部分）
	result, total, err = storage.FindAll(1, 10, map[string]interface{}{"host": "example.com"})
	assert.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, result, 3)

	// 验证关键词搜索能命中 host 字段
	result, total, err = storage.Search("host1", 1, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, result, 2)

	// 验证 FindByID 能正确返回 host
	entry, err := storage.FindByID("host-test-2")
	assert.NoError(t, err)
	assert.Equal(t, "host2.example.com", entry.Host)
}
