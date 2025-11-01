package debugger

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemoryStorage 内存存储器实现
// 将调试日志保存在内存中，适合开发环境使用
type MemoryStorage struct {
	entries []*LogEntry    // 日志条目列表
	index   map[string]int // ID到索引的映射
	mutex   sync.RWMutex   // 读写锁
	maxSize int            // 最大存储条目数
}

// NewMemoryStorage 创建新的内存存储器
// maxSize: 最大存储条目数（0表示无限制）
func NewMemoryStorage(maxSize ...int) (*MemoryStorage, error) {
	size := 10000 // 默认最大10000条
	if len(maxSize) > 0 && maxSize[0] > 0 {
		size = maxSize[0]
	}

	return &MemoryStorage{
		entries: make([]*LogEntry, 0),
		index:   make(map[string]int),
		maxSize: size,
	}, nil
}

// Save 保存日志条目到内存
func (ms *MemoryStorage) Save(entry *LogEntry) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// 检查是否已存在相同ID的条目
	if idx, exists := ms.index[entry.ID]; exists {
		// 更新现有条目
		ms.entries[idx] = entry
		return nil
	}

	// 检查是否超过最大存储限制
	if ms.maxSize > 0 && len(ms.entries) >= ms.maxSize {
		// 删除最旧的条目
		oldestID := ms.entries[0].ID
		ms.entries = ms.entries[1:]
		delete(ms.index, oldestID)

		// 重新构建索引
		ms.rebuildIndex()
	}

	// 添加新条目
	ms.entries = append(ms.entries, entry)
	ms.index[entry.ID] = len(ms.entries) - 1

	return nil
}

// FindByID 根据ID查找日志条目
func (ms *MemoryStorage) FindByID(id string) (*LogEntry, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	if idx, exists := ms.index[id]; exists {
		return ms.entries[idx], nil
	}

	return nil, fmt.Errorf("未找到ID为 %s 的日志条目", id)
}

// FindAll 查找所有日志条目，支持分页和过滤
func (ms *MemoryStorage) FindAll(page, pageSize int, filters map[string]interface{}) ([]*LogEntry, int, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	// 应用过滤器
	var filtered []*LogEntry
	for _, entry := range ms.entries {
		if ms.filterEntry(entry, filters) {
			filtered = append(filtered, entry)
		}
	}

	// 按时间倒序排序
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// 计算分页
	total := len(filtered)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []*LogEntry{}, total, nil
	}

	if end > total {
		end = total
	}

	return filtered[start:end], total, nil
}

// Search 搜索日志内容
func (ms *MemoryStorage) Search(keyword string, page, pageSize int) ([]*LogEntry, int, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	// 搜索匹配的日志条目
	var results []*LogEntry
	for _, entry := range ms.entries {
		if ms.containsKeyword(entry, keyword) {
			results = append(results, entry)
		}
	}

	// 按时间倒序排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

	// 计算分页
	total := len(results)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []*LogEntry{}, total, nil
	}

	if end > total {
		end = total
	}

	return results[start:end], total, nil
}

// Cleanup 清理过期日志
func (ms *MemoryStorage) Cleanup(before time.Time) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// 过滤出未过期的条目
	var remaining []*LogEntry
	for _, entry := range ms.entries {
		if !entry.Timestamp.Before(before) {
			remaining = append(remaining, entry)
		}
	}

	// 更新条目列表和索引
	ms.entries = remaining
	ms.rebuildIndex()

	return nil
}

// filterEntry 应用过滤器
func (ms *MemoryStorage) filterEntry(entry *LogEntry, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "method":
			if entry.Method != value {
				return false
			}
		case "status_code":
			if entry.StatusCode != value {
				return false
			}
		case "url_contains":
			if !strings.Contains(entry.URL, value.(string)) {
				return false
			}
		case "start_time":
			if entry.Timestamp.Before(value.(time.Time)) {
				return false
			}
		case "end_time":
			if entry.Timestamp.After(value.(time.Time)) {
				return false
			}
		case "client_ip":
			if !strings.Contains(entry.ClientIP, value.(string)) {
				return false
			}
		case "has_error":
			if value.(bool) && entry.Error == "" {
				return false
			}
			if !value.(bool) && entry.Error != "" {
				return false
			}
		case "min_duration":
			if entry.Duration < value.(time.Duration) {
				return false
			}
		case "max_duration":
			if entry.Duration > value.(time.Duration) {
				return false
			}
		}
	}

	return true
}

// containsKeyword 检查日志条目是否包含关键词
func (ms *MemoryStorage) containsKeyword(entry *LogEntry, keyword string) bool {
	// 检查URL
	if strings.Contains(strings.ToLower(entry.URL), strings.ToLower(keyword)) {
		return true
	}

	// 检查请求体
	if strings.Contains(strings.ToLower(entry.RequestBody), strings.ToLower(keyword)) {
		return true
	}

	// 检查响应体
	if strings.Contains(strings.ToLower(entry.ResponseBody), strings.ToLower(keyword)) {
		return true
	}

	// 检查错误信息
	if strings.Contains(strings.ToLower(entry.Error), strings.ToLower(keyword)) {
		return true
	}

	// 检查请求头
	for _, value := range entry.RequestHeaders {
		if strings.Contains(strings.ToLower(value), strings.ToLower(keyword)) {
			return true
		}
	}

	// 检查响应头
	for _, value := range entry.ResponseHeaders {
		if strings.Contains(strings.ToLower(value), strings.ToLower(keyword)) {
			return true
		}
	}

	// 检查用户代理
	if strings.Contains(strings.ToLower(entry.UserAgent), strings.ToLower(keyword)) {
		return true
	}

	return false
}

// rebuildIndex 重新构建索引
func (ms *MemoryStorage) rebuildIndex() {
	ms.index = make(map[string]int)
	for i, entry := range ms.entries {
		ms.index[entry.ID] = i
	}
}

// GetStats 获取存储统计信息
func (ms *MemoryStorage) GetStats() (map[string]interface{}, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_requests": len(ms.entries), // 总请求数
		"max_size":       ms.maxSize,
		"storage_type":   "memory",
	}

	// 计算存储大小（精确计算，与单个条目计算逻辑保持一致）
	var storageSize int64
	for _, entry := range ms.entries {
		// 使用与CalculateStorageSize相同的计算逻辑
		entrySize := int64(len(entry.ID) + len(entry.URL) + len(entry.Method) +
			len(entry.RequestBody) + len(entry.ResponseBody) + len(entry.Error) +
			len(entry.UserAgent) + len(entry.ClientIP))

		// 计算请求头大小
		for key, value := range entry.RequestHeaders {
			entrySize += int64(len(key) + len(value))
		}

		// 计算响应头大小
		for key, value := range entry.ResponseHeaders {
			entrySize += int64(len(key) + len(value))
		}

		// 计算查询参数大小
		for key, value := range entry.QueryParams {
			entrySize += int64(len(key) + len(value))
		}

		// 计算会话数据大小（JSON格式）
		if entry.SessionData != nil {
			if sessionData, err := json.Marshal(entry.SessionData); err == nil {
				entrySize += int64(len(sessionData))
			}
		}

		// 计算Logger日志大小
		for _, log := range entry.LoggerLogs {
			entrySize += int64(len(log.Message))
			if log.Fields != nil {
				if fieldsData, err := json.Marshal(log.Fields); err == nil {
					entrySize += int64(len(fieldsData))
				}
			}
		}

		storageSize += entrySize
	}

	// 转换为KB
	stats["storage_size"] = fmt.Sprintf("%.2f KB", float64(storageSize)/1024)

	// 计算平均响应时间
	if len(ms.entries) > 0 {
		var totalDuration time.Duration
		var errorCount int

		for _, entry := range ms.entries {
			totalDuration += entry.Duration
			if entry.Error != "" {
				errorCount++
			}
		}

		stats["avg_duration"] = totalDuration / time.Duration(len(ms.entries))
		stats["error_rate"] = float64(errorCount) / float64(len(ms.entries))
		stats["error_count"] = errorCount
	}

	return stats, nil
}

// Clear 清空所有日志条目
func (ms *MemoryStorage) Clear() error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	ms.entries = make([]*LogEntry, 0)
	ms.index = make(map[string]int)

	return nil
}

// GetRecent 获取最近的日志条目
func (ms *MemoryStorage) GetRecent(count int) ([]*LogEntry, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	if count <= 0 || count > len(ms.entries) {
		count = len(ms.entries)
	}

	// 返回最近的count条记录
	start := len(ms.entries) - count
	if start < 0 {
		start = 0
	}

	return ms.entries[start:], nil
}

// GetByTimeRange 根据时间范围获取日志条目
func (ms *MemoryStorage) GetByTimeRange(startTime, endTime time.Time) ([]*LogEntry, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	var results []*LogEntry
	for _, entry := range ms.entries {
		if (entry.Timestamp.Equal(startTime) || entry.Timestamp.After(startTime)) &&
			(entry.Timestamp.Equal(endTime) || entry.Timestamp.Before(endTime)) {
			results = append(results, entry)
		}
	}

	return results, nil
}

// GetMethods 获取所有HTTP方法统计
func (ms *MemoryStorage) GetMethods() (map[string]int, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	methods := make(map[string]int)
	for _, entry := range ms.entries {
		methods[entry.Method]++
	}

	return methods, nil
}

// GetStatusCodes 获取所有状态码统计
func (ms *MemoryStorage) GetStatusCodes() (map[int]int, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	statusCodes := make(map[int]int)
	for _, entry := range ms.entries {
		statusCodes[entry.StatusCode]++
	}

	return statusCodes, nil
}

// ExportAll 导出所有日志到内存（返回副本）
func (ms *MemoryStorage) ExportAll() ([]*LogEntry, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	// 创建副本
	entries := make([]*LogEntry, len(ms.entries))
	copy(entries, ms.entries)

	return entries, nil
}

// ImportEntries 导入日志条目
func (ms *MemoryStorage) ImportEntries(entries []*LogEntry) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for _, entry := range entries {
		// 检查是否已存在
		if _, exists := ms.index[entry.ID]; !exists {
			// 检查是否超过最大存储限制
			if ms.maxSize > 0 && len(ms.entries) >= ms.maxSize {
				// 删除最旧的条目
				oldestID := ms.entries[0].ID
				ms.entries = ms.entries[1:]
				delete(ms.index, oldestID)

				// 重新构建索引
				ms.rebuildIndex()
			}

			// 添加新条目
			ms.entries = append(ms.entries, entry)
			ms.index[entry.ID] = len(ms.entries) - 1
		}
	}

	return nil
}

// Close 关闭内存存储器
func (ms *MemoryStorage) Close() error {
	// 内存存储不需要特殊关闭操作
	return nil
}
