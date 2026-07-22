package debugger

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestFileStorageSearchWithHeaderFilters 验证文件存储在关键词搜索时，
// 同时使用 host、client_ip、is_streaming 等仅存在于完整日志中的字段进行筛选，
// 返回的总数统计是否准确。
//
// 该测试用于覆盖 readLogFileHeader 快速读取时遗漏字段导致的 total 统计错误。
func TestFileStorageSearchWithHeaderFilters(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "debugger_file_storage_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	storage, err := NewFileStorage(tempDir, 0)
	assert.NoError(t, err)

	now := time.Now()
	entries := []*LogEntry{
		{
			ID:                  "host-test-1",
			Timestamp:           now,
			URL:                 "/api/test",
			Host:                "host1.example.com",
			ClientIP:            "192.168.1.1",
			IsStreamingResponse: false,
		},
		{
			ID:                  "host-test-2",
			Timestamp:           now.Add(-time.Second),
			URL:                 "/api/test",
			Host:                "host2.example.com",
			ClientIP:            "192.168.1.2",
			IsStreamingResponse: true,
		},
		{
			ID:                  "host-test-3",
			Timestamp:           now.Add(-2 * time.Second),
			URL:                 "/api/other",
			Host:                "host1.example.com",
			ClientIP:            "192.168.1.3",
			IsStreamingResponse: false,
		},
	}

	for _, entry := range entries {
		assert.NoError(t, storage.Save(entry))
	}

	tests := []struct {
		name     string
		keyword  string
		filters  map[string]interface{}
		expected int
	}{
		{
			name:     "按 host 筛选",
			keyword:  "api",
			filters:  map[string]interface{}{"host": "host1.example.com"},
			expected: 2,
		},
		{
			name:     "按 client_ip 筛选",
			keyword:  "api",
			filters:  map[string]interface{}{"client_ip": "192.168.1.2"},
			expected: 1,
		},
		{
			name:     "按 is_streaming 筛选",
			keyword:  "api",
			filters:  map[string]interface{}{"is_streaming": "true"},
			expected: 1,
		},
		{
			name:     "组合筛选",
			keyword:  "api",
			filters:  map[string]interface{}{"host": "host1.example.com", "is_streaming": "false"},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, total, err := storage.Search(tt.keyword, 1, 10, tt.filters)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, total, "总数统计应正确")
			assert.Len(t, result, tt.expected, "返回结果数量应与总数一致")
		})
	}
}
