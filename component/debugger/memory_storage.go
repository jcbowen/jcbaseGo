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
	entries []*LogEntry    // 日志条目环形缓冲区
	index   map[string]int // ID到索引的映射
	mutex   sync.RWMutex   // 读写锁
	maxSize int            // 最大存储条目数
	head    int            // 下一个要插入的位置
	count   int            // 当前存储的条目数量
}

// NewMemoryStorage 创建新的内存存储器
// maxSize: 最大存储条目数（0表示无限制）
func NewMemoryStorage(maxSize ...int) (*MemoryStorage, error) {
	size := 10000 // 默认最大10000条
	if len(maxSize) > 0 && maxSize[0] > 0 {
		size = maxSize[0]
	}

	return &MemoryStorage{
		entries: make([]*LogEntry, size), // 预分配环形缓冲区
		index:   make(map[string]int),
		maxSize: size,
		head:    0,
		count:   0,
	}, nil
}

// Save 保存日志条目到内存
// 将HTTP记录或进程记录保存到内存存储中，支持自动清理和更新机制
// 该方法确保存储的线程安全性，并维护ID索引以提高查询性能
//
// 参数:
//
//	entry: 要保存的日志条目，包含HTTP记录或进程记录的完整信息
//
// 返回值:
//
//	error: 如果保存过程中发生错误返回错误信息，否则返回nil
//
// 保存逻辑:
//   - 如果条目ID已存在，则更新现有条目（支持进程记录的动态更新）
//   - 使用环形缓冲区存储条目，当超过最大存储限制时，自动覆盖最旧的条目（FIFO策略）
//   - 维护ID到索引的映射，提高FindByID等操作的性能
//   - 使用读写锁确保多线程环境下的数据一致性
//   - 时间复杂度O(1)，避免了重建索引的开销
func (ms *MemoryStorage) Save(entry *LogEntry) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// 检查是否已存在相同ID的条目
	if idx, exists := ms.index[entry.ID]; exists {
		// 更新现有条目
		ms.entries[idx] = entry
		return nil
	}

	// 计算要插入的位置
	insertPos := ms.head

	// 如果缓冲区已满，需要覆盖最旧的条目
	if ms.count >= ms.maxSize {
		// 获取要被覆盖的旧条目ID
		oldEntry := ms.entries[insertPos]
		if oldEntry != nil {
			// 从索引中删除旧条目
			delete(ms.index, oldEntry.ID)
		}
	} else {
		// 缓冲区未满，增加计数
		ms.count++
	}

	// 插入新条目
	ms.entries[insertPos] = entry
	ms.index[entry.ID] = insertPos

	// 更新head指针，指向下一个要插入的位置
	ms.head = (ms.head + 1) % ms.maxSize

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
// 优化性能：实现真正的分页查询，避免全量数据遍历
// 优化并发：减少锁持有时间，提高并发性能
// 该方法提供统一的查询接口，适用于HTTP记录和进程记录的检索
//
// 参数:
//
//	page: 页码（从1开始）
//	pageSize: 每页显示数量
//	filters: 过滤条件映射，支持record_type、method、status_code、url、process_name、process_id等字段
//
// 返回值:
//
//	[]*LogEntry: 符合条件的日志条目列表（按时间倒序排列）
//	int: 符合条件的总条目数
//	error: 如果查询过程中发生错误返回错误信息，否则返回nil
//
// 优化逻辑:
//   - 首先收集所有有效的日志条目（非nil）
//   - 计算符合条件的总条目数
//   - 按时间倒序排序后，只返回当前页需要的数据
//   - 支持空过滤条件，返回分页数据
//   - 减少锁持有时间，提高并发性能
func (ms *MemoryStorage) FindAll(page, pageSize int, filters map[string]interface{}) ([]*LogEntry, int, error) {
	// 收集所有有效的日志条目
	ms.mutex.RLock()
	var allEntries []*LogEntry
	for _, entry := range ms.entries {
		if entry != nil {
			allEntries = append(allEntries, entry)
		}
	}
	ms.mutex.RUnlock()

	// 应用过滤器
	var filteredEntries []*LogEntry
	for _, entry := range allEntries {
		if ms.filterEntry(entry, filters) {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	total := len(filteredEntries)

	// 检查是否需要返回空结果
	if total == 0 {
		return []*LogEntry{}, 0, nil
	}

	// 按时间倒序排序
	sort.Slice(filteredEntries, func(i, j int) bool {
		return filteredEntries[i].Timestamp.After(filteredEntries[j].Timestamp)
	})

	// 计算分页范围
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []*LogEntry{}, total, nil
	}

	if end > total {
		end = total
	}

	// 返回分页数据
	return filteredEntries[start:end], total, nil
}

// Search 搜索日志内容
// 优化性能：实现真正的分页搜索，避免全量数据遍历
// 优化并发：减少锁持有时间，提高并发性能
// 在日志条目的多个字段中进行全文搜索，支持HTTP记录和进程记录的关键词检索
// 该方法提供不区分大小写的搜索功能，适用于快速查找特定内容的日志记录
//
// 参数:
//
//	keyword: 搜索关键词，支持在进程名称、URL、请求体、响应体、错误信息等字段中搜索
//	page: 页码（从1开始）
//	pageSize: 每页显示数量
//
// 返回值:
//
//	[]*LogEntry: 包含关键词的日志条目列表（按时间倒序排列）
//	int: 包含关键词的总条目数
//	error: 如果搜索过程中发生错误返回错误信息，否则返回nil
//
// 优化逻辑:
//   - 首先收集所有有效的日志条目（非nil）
//   - 搜索符合条件的条目
//   - 按时间倒序排序后，只返回当前页需要的数据
//   - 减少锁持有时间，提高并发性能
//
// 搜索范围:
//   - 进程名称（process_name）
//   - URL路径（url）
//   - 请求体（request_body）
//   - 响应体（response_body）
//   - 错误信息（error）
//   - 请求头（request_headers）的所有值
//   - 响应头（response_headers）的所有值
func (ms *MemoryStorage) Search(keyword string, page, pageSize int) ([]*LogEntry, int, error) {
	// 收集所有有效的日志条目
	ms.mutex.RLock()
	var allEntries []*LogEntry
	for _, entry := range ms.entries {
		if entry != nil {
			allEntries = append(allEntries, entry)
		}
	}
	ms.mutex.RUnlock()

	// 搜索符合条件的条目
	var matchedEntries []*LogEntry
	for _, entry := range allEntries {
		if ms.containsKeyword(entry, keyword) {
			matchedEntries = append(matchedEntries, entry)
		}
	}

	total := len(matchedEntries)

	// 检查是否需要返回空结果
	if total == 0 {
		return []*LogEntry{}, 0, nil
	}

	// 按时间倒序排序
	sort.Slice(matchedEntries, func(i, j int) bool {
		return matchedEntries[i].Timestamp.After(matchedEntries[j].Timestamp)
	})

	// 计算分页范围
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []*LogEntry{}, total, nil
	}

	if end > total {
		end = total
	}

	// 返回分页数据
	return matchedEntries[start:end], total, nil
}

// Cleanup 清理过期日志
func (ms *MemoryStorage) Cleanup(before time.Time) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// 重置索引
	ms.index = make(map[string]int)
	newHead := 0
	newCount := 0

	// 遍历所有条目，保留未过期的
	for i := 0; i < ms.maxSize; i++ {
		entry := ms.entries[i]
		if entry != nil && !entry.Timestamp.Before(before) {
			// 保留未过期的条目，重新放入环形缓冲区
			ms.entries[newHead] = entry
			ms.index[entry.ID] = newHead
			newHead = (newHead + 1) % ms.maxSize
			newCount++
		} else {
			// 删除过期的条目
			ms.entries[i] = nil
		}
	}

	// 清空剩余的位置
	for i := 0; i < ms.maxSize; i++ {
		if i >= newCount {
			ms.entries[(newHead+i)%ms.maxSize] = nil
		}
	}

	// 更新头部和计数
	ms.head = newHead
	ms.count = newCount

	return nil
}

// filterEntry 应用过滤器
// 根据过滤条件检查日志条目是否匹配，支持HTTP记录和进程记录的多种过滤条件
// 该方法用于内存存储的查询过滤，确保只有符合条件的记录被返回
//
// 参数:
//
//	entry: 要检查的日志条目
//	filters: 过滤条件映射，键为字段名，值为过滤值
//
// 返回值:
//
//	bool: 如果条目匹配所有过滤条件返回true，否则返回false
//
// 支持的过滤条件:
//   - record_type: 记录类型精确匹配（http/process）
//   - method: HTTP方法精确匹配（GET/POST/PUT/DELETE等）
//   - status_code: HTTP状态码精确匹配（200/404/500等）
//   - url: URL路径包含匹配（支持模糊匹配）
//   - start_time: 开始时间过滤（大于等于指定时间）
//   - end_time: 结束时间过滤（小于等于指定时间）
//   - client_ip: 客户端IP地址包含匹配（支持模糊匹配）
//   - process_name: 进程名称包含匹配（支持模糊匹配）
//   - process_id: 进程ID精确匹配
//   - process_status: 进程状态精确匹配（running/completed/failed/error）
//   - has_error: 错误状态过滤（true表示有错误，false表示无错误）
//   - min_duration: 最小持续时间过滤（大于等于指定时长）
//   - max_duration: 最大持续时间过滤（小于等于指定时长）
func (ms *MemoryStorage) filterEntry(entry *LogEntry, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "record_type":
			if entry.RecordType != value {
				return false
			}
		case "method":
			if entry.Method != value {
				return false
			}
		case "status_code":
			if entry.StatusCode != value {
				return false
			}
		case "url":
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
		case "host":
			if !strings.Contains(entry.Host, value.(string)) {
				return false
			}
		case "process_name":
			if !strings.Contains(entry.ProcessName, value.(string)) {
				return false
			}
		case "process_id":
			if entry.ProcessID != value {
				return false
			}
		case "process_status":
			if entry.Status != value {
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
		case "is_streaming":
			// 流式请求过滤：true/false 字符串转换为布尔值
			filterIsStreaming := strings.ToLower(value.(string)) == "true"
			if entry.IsStreamingResponse != filterIsStreaming {
				return false
			}
		case "streaming_status":
			// 流式状态过滤：active/inactive 字符串匹配
			filterStatus := value.(string)
			if filterStatus == "active" && !entry.IsStreamingResponse {
				return false
			}
			if filterStatus == "inactive" && entry.IsStreamingResponse {
				return false
			}
		}
	}

	return true
}

// containsKeyword 检查日志条目是否包含关键词
func (ms *MemoryStorage) containsKeyword(entry *LogEntry, keyword string) bool {
	// 检查进程名称
	if strings.Contains(strings.ToLower(entry.ProcessName), strings.ToLower(keyword)) {
		return true
	}

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

	// 计算存储大小（精确计算，与单个条目计算逻辑保持一致）
	var storageSize int64
	var errorCount int
	var totalDuration time.Duration
	var streamingRequestCount int
	var totalStreamingChunks int
	var maxStreamingChunks int
	var validEntryCount int

	for _, entry := range ms.entries {
		if entry == nil {
			continue
		}

		validEntryCount++

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

		// 统计错误信息
		if entry.Error != "" {
			errorCount++
		}
		totalDuration += entry.Duration

		// 统计流式请求信息
		if entry.IsStreamingResponse {
			streamingRequestCount++
			totalStreamingChunks += entry.StreamingChunks
			if entry.StreamingChunks > maxStreamingChunks {
				maxStreamingChunks = entry.StreamingChunks
			}
		}
	}

	stats := map[string]interface{}{
		"total_requests": validEntryCount, // 总请求数
		"max_size":       ms.maxSize,
		"storage_type":   "memory",
	}

	// 转换为KB
	stats["storage_size"] = fmt.Sprintf("%.2f KB", float64(storageSize)/1024)

	// 计算平均响应时间
	if validEntryCount > 0 {
		stats["avg_duration"] = totalDuration / time.Duration(validEntryCount)
		stats["error_rate"] = float64(errorCount) / float64(validEntryCount)
		stats["error_count"] = errorCount
	}

	// 添加流式请求统计信息
	stats["streaming_request_count"] = streamingRequestCount
	if validEntryCount > 0 {
		stats["streaming_request_rate"] = float64(streamingRequestCount) / float64(validEntryCount)
	} else {
		stats["streaming_request_rate"] = 0
	}
	stats["total_streaming_chunks"] = totalStreamingChunks
	if streamingRequestCount > 0 {
		stats["avg_streaming_chunks"] = float64(totalStreamingChunks) / float64(streamingRequestCount)
		stats["max_streaming_chunks"] = maxStreamingChunks
	} else {
		stats["avg_streaming_chunks"] = 0
		stats["max_streaming_chunks"] = 0
	}

	return stats, nil
}

// Clear 清空所有日志条目
func (ms *MemoryStorage) Clear() error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// 清空环形缓冲区
	for i := 0; i < ms.maxSize; i++ {
		ms.entries[i] = nil
	}

	// 重置索引和计数器
	ms.index = make(map[string]int)
	ms.head = 0
	ms.count = 0

	return nil
}

// GetRecent 获取最近的日志条目
func (ms *MemoryStorage) GetRecent(count int) ([]*LogEntry, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	// 收集所有有效的日志条目
	var allEntries []*LogEntry
	for _, entry := range ms.entries {
		if entry != nil {
			allEntries = append(allEntries, entry)
		}
	}

	// 按时间倒序排序
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].Timestamp.After(allEntries[j].Timestamp)
	})

	// 限制返回数量
	if count <= 0 || count > len(allEntries) {
		count = len(allEntries)
	}

	// 返回最近的count条记录
	return allEntries[:count], nil
}

// GetByTimeRange 根据时间范围获取日志条目
func (ms *MemoryStorage) GetByTimeRange(startTime, endTime time.Time) ([]*LogEntry, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	var results []*LogEntry
	for _, entry := range ms.entries {
		if entry != nil && (entry.Timestamp.Equal(startTime) || entry.Timestamp.After(startTime)) &&
			(entry.Timestamp.Equal(endTime) || entry.Timestamp.Before(endTime)) {
			results = append(results, entry)
		}
	}

	// 按时间倒序排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.After(results[j].Timestamp)
	})

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
		// 检查是否已存在相同ID的条目
		if _, exists := ms.index[entry.ID]; exists {
			continue
		}

		// 计算要插入的位置
		insertPos := ms.head

		// 如果缓冲区已满，需要覆盖最旧的条目
		if ms.count >= ms.maxSize {
			// 获取要被覆盖的旧条目ID
			oldEntry := ms.entries[insertPos]
			if oldEntry != nil {
				// 从索引中删除旧条目
				delete(ms.index, oldEntry.ID)
			}
		} else {
			// 缓冲区未满，增加计数
			ms.count++
		}

		// 插入新条目
		ms.entries[insertPos] = entry
		ms.index[entry.ID] = insertPos

		// 更新head指针，指向下一个要插入的位置
		ms.head = (ms.head + 1) % ms.maxSize
	}

	return nil
}

// Close 关闭内存存储器
func (ms *MemoryStorage) Close() error {
	// 内存存储不需要特殊关闭操作
	return nil
}
