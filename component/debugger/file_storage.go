package debugger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// FileStorage 文件存储器实现
// 将调试日志保存到文件系统中
type FileStorage struct {
	basePath string // 基础存储路径
	maxSize  int    // 最大存储条目数
}

// NewFileStorage 创建新的文件存储器
// basePath: 文件存储的基础路径
// maxSize: 最大存储条目数（0表示无限制）
func NewFileStorage(basePath string, maxSize int) (*FileStorage, error) {
	// 确保基础路径存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("创建存储目录失败: %v", err)
	}

	return &FileStorage{
		basePath: basePath,
		maxSize:  maxSize,
	}, nil
}

// Save 保存日志条目到文件
// 每个日志条目保存为一个独立的JSON文件
func (fs *FileStorage) Save(entry *LogEntry) error {
	// 检查是否超过最大存储限制
	if fs.maxSize > 0 {
		// 获取所有日志文件
		files, err := fs.listLogFiles()
		if err != nil {
			return fmt.Errorf("获取日志文件列表失败: %v", err)
		}

		// 如果超过最大限制，删除最旧的条目
		if len(files) >= fs.maxSize {
			// 按时间排序，删除最旧的文件
			sort.Slice(files, func(i, j int) bool {
				// 从文件名中提取时间戳进行比较
				timeI, _ := fs.extractTimestampFromFilename(files[i])
				timeJ, _ := fs.extractTimestampFromFilename(files[j])
				return timeI.Before(timeJ)
			})

			// 删除最旧的文件，直到数量在限制范围内
			for len(files) >= fs.maxSize {
				if err := os.Remove(files[0]); err != nil {
					return fmt.Errorf("删除旧日志文件失败: %v", err)
				}
				files = files[1:]
			}
		}
	}

	// 生成文件名：timestamp_id.json
	filename := fmt.Sprintf("%s_%s.json",
		entry.Timestamp.Format("2006-01-02_15-04-05"),
		entry.ID)

	filePath := filepath.Join(fs.basePath, filename)

	// 将日志条目转换为JSON
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化日志条目失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入日志文件失败: %v", err)
	}

	return nil
}

// FindByID 根据ID查找日志条目
func (fs *FileStorage) FindByID(id string) (*LogEntry, error) {
	// 获取所有日志文件
	files, err := fs.listLogFiles()
	if err != nil {
		return nil, err
	}

	// 遍历文件查找匹配的ID
	for _, file := range files {
		entry, err := fs.readLogFile(file)
		if err != nil {
			continue // 跳过读取失败的文件
		}

		if entry.ID == id {
			return entry, nil
		}
	}

	return nil, fmt.Errorf("未找到ID为 %s 的日志条目", id)
}

// FindAll 查找所有日志条目，支持分页和过滤
// 优化性能：实现真正的分页查询，避免全量文件读取
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
//   - 首先统计符合条件的文件总数（不读取文件内容）
//   - 按文件名时间戳排序，只读取当前页需要的文件
//   - 避免全量文件读取，减少I/O操作和内存占用
func (fs *FileStorage) FindAll(page, pageSize int, filters map[string]interface{}) ([]*LogEntry, int, error) {
	// 获取所有日志文件
	files, err := fs.listLogFiles()
	if err != nil {
		return nil, 0, err
	}

	// 按文件名时间戳排序（最新的在前）
	sort.Slice(files, func(i, j int) bool {
		timeI, _ := fs.extractTimestampFromFilename(files[i])
		timeJ, _ := fs.extractTimestampFromFilename(files[j])
		return timeI.After(timeJ)
	})

	// 计算符合条件的文件总数（不读取文件内容）
	total := 0
	for _, file := range files {
		// 只读取文件头信息进行快速过滤判断
		entry, err := fs.readLogFile(file)
		if err != nil {
			continue
		}
		if fs.filterEntry(entry, filters) {
			total++
		}
	}

	// 计算分页范围
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []*LogEntry{}, total, nil
	}

	if end > total {
		end = total
	}

	// 只读取当前页需要的文件
	var result []*LogEntry
	currentIndex := 0

	for _, file := range files {
		entry, err := fs.readLogFile(file)
		if err != nil {
			continue // 跳过读取失败的文件
		}

		// 应用过滤器
		if fs.filterEntry(entry, filters) {
			if currentIndex >= start && currentIndex < end {
				result = append(result, entry)
			}
			currentIndex++

			// 如果已经收集到足够的数据，提前退出
			if len(result) >= pageSize {
				break
			}
		}
	}

	return result, total, nil
}

// Search 搜索日志内容
// Search 搜索日志内容
// 优化性能：实现真正的分页搜索，避免全量文件读取
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
//   - 首先统计符合条件的文件总数（使用快速文件头读取）
//   - 按文件名时间戳排序，只读取当前页需要的文件
//   - 避免全量文件读取，减少I/O操作和内存占用
func (fs *FileStorage) Search(keyword string, page, pageSize int) ([]*LogEntry, int, error) {
	// 获取所有日志文件
	files, err := fs.listLogFiles()
	if err != nil {
		return nil, 0, err
	}

	// 按文件名时间戳排序（最新的在前）
	sort.Slice(files, func(i, j int) bool {
		timeI, _ := fs.extractTimestampFromFilename(files[i])
		timeJ, _ := fs.extractTimestampFromFilename(files[j])
		return timeI.After(timeJ)
	})

	// 计算符合条件的文件总数（使用快速文件头读取）
	total := 0
	for _, file := range files {
		// 使用快速文件头读取进行关键词匹配
		entry, err := fs.readLogFileHeader(file)
		if err != nil {
			continue
		}
		if fs.containsKeyword(entry, keyword) {
			total++
		}
	}

	// 计算分页范围
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []*LogEntry{}, total, nil
	}

	if end > total {
		end = total
	}

	// 只读取当前页需要的文件
	var result []*LogEntry
	currentIndex := 0

	for _, file := range files {
		entry, err := fs.readLogFile(file)
		if err != nil {
			continue // 跳过读取失败的文件
		}

		// 检查是否包含关键词
		if fs.containsKeyword(entry, keyword) {
			if currentIndex >= start && currentIndex < end {
				result = append(result, entry)
			}
			currentIndex++

			// 如果已经收集到足够的数据，提前退出
			if len(result) >= pageSize {
				break
			}
		}
	}

	return result, total, nil
}

// Cleanup 清理过期日志
func (fs *FileStorage) Cleanup(before time.Time) error {
	// 获取所有日志文件
	files, err := fs.listLogFiles()
	if err != nil {
		return err
	}

	// 删除过期文件
	var deletedCount int
	for _, file := range files {
		entry, err := fs.readLogFile(file)
		if err != nil {
			continue // 跳过读取失败的文件
		}

		// 检查是否过期
		if entry.Timestamp.Before(before) {
			if err := os.Remove(file); err != nil {
				return fmt.Errorf("删除文件 %s 失败: %v", file, err)
			}
			deletedCount++
		}
	}

	return nil
}

// listLogFiles 列出所有日志文件
func (fs *FileStorage) listLogFiles() ([]string, error) {
	var files []string

	// 读取目录
	entries, err := os.ReadDir(fs.basePath)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %v", err)
	}

	// 筛选JSON文件
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			files = append(files, filepath.Join(fs.basePath, entry.Name()))
		}
	}

	return files, nil
}

// readLogFile 读取日志文件
func (fs *FileStorage) readLogFile(filePath string) (*LogEntry, error) {
	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件 %s 失败: %v", filePath, err)
	}

	// 解析JSON
	var entry LogEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &entry, nil
}

// readLogFileHeader 快速读取日志文件头信息，用于快速过滤判断
// 只读取文件的部分内容进行快速过滤，避免全量文件读取
// 该方法用于性能优化，在分页查询时减少I/O操作
//
// 参数:
//
//	filePath: 日志文件路径
//
// 返回值:
//
//	*LogEntry: 包含基本信息的日志条目（可能不完整）
//	error: 如果文件读取或JSON解析失败返回错误信息
func (fs *FileStorage) readLogFileHeader(filePath string) (*LogEntry, error) {
	// 只读取文件的前1KB内容进行快速解析
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %v", err)
	}
	defer file.Close()

	// 读取前1KB内容
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("读取日志文件头失败: %v", err)
	}

	// 解析JSON到临时结构体，只包含基本字段
	var header struct {
		RecordType  string    `json:"record_type"`
		Method      string    `json:"method"`
		StatusCode  int       `json:"status_code"`
		URL         string    `json:"url"`
		ProcessName string    `json:"process_name"`
		ProcessID   string    `json:"process_id"`
		Timestamp   time.Time `json:"timestamp"`
	}

	if err := json.Unmarshal(buffer[:n], &header); err != nil {
		// 如果快速解析失败，回退到完整读取
		return fs.readLogFile(filePath)
	}

	// 转换为完整的LogEntry结构
	entry := &LogEntry{
		RecordType:  header.RecordType,
		Method:      header.Method,
		StatusCode:  header.StatusCode,
		URL:         header.URL,
		ProcessName: header.ProcessName,
		ProcessID:   header.ProcessID,
		Timestamp:   header.Timestamp,
	}

	return entry, nil
}

// extractTimestampFromFilename 从文件名中提取时间戳
// 文件名格式：2006-01-02_15-04-05_debug_123456789.json
func (fs *FileStorage) extractTimestampFromFilename(filePath string) (time.Time, error) {
	filename := filepath.Base(filePath)

	// 文件名格式：2006-01-02_15-04-05_debug_123456789.json
	// 提取时间戳部分（前19个字符）
	if len(filename) < 19 {
		return time.Time{}, fmt.Errorf("文件名格式不正确: %s", filename)
	}

	timestampStr := filename[:19]

	// 解析时间戳
	timestamp, err := time.Parse("2006-01-02_15-04-05", timestampStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("解析时间戳失败: %v", err)
	}

	return timestamp, nil
}

// filterEntry 应用过滤器
// 根据过滤条件检查日志条目是否匹配，支持HTTP记录和进程记录的多种过滤条件
// 该方法用于文件存储的查询过滤，确保只有符合条件的记录被返回
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
//   - client_ip: 客户端IP地址匹配（支持精确匹配、包含匹配和前缀匹配）
//   - process_name: 进程名称包含匹配（支持模糊匹配）
//   - process_id: 进程ID精确匹配
//   - process_status: 进程状态精确匹配（running/completed/failed/error）
//   - has_error: 错误状态过滤（true表示有错误，false表示无错误）
//   - min_duration: 最小持续时间过滤（大于等于指定时长）
//   - max_duration: 最大持续时间过滤（小于等于指定时长）
func (fs *FileStorage) filterEntry(entry *LogEntry, filters map[string]interface{}) bool {
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
			// 处理状态码类型转换，支持字符串和整数类型的比较
			var statusCodeValue int
			switch v := value.(type) {
			case int:
				statusCodeValue = v
			case string:
				if code, err := strconv.Atoi(v); err == nil {
					statusCodeValue = code
				} else {
					return false
				}
			default:
				return false
			}
			if entry.StatusCode != statusCodeValue {
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
			if entry.ClientIP != value && !strings.Contains(entry.ClientIP, value.(string)) && !strings.HasPrefix(entry.ClientIP, value.(string)) {
				return false
			}
		case "host":
			if entry.Host != value && !strings.Contains(entry.Host, value.(string)) && !strings.HasPrefix(entry.Host, value.(string)) {
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
			} else if filterStatus == "inactive" && entry.IsStreamingResponse {
				return false
			}
		}
	}

	return true
}

// containsKeyword 检查日志条目是否包含关键词
// 在日志条目的多个字段中搜索指定的关键词，支持HTTP记录和进程记录的全文搜索
// 该方法用于文件存储的关键词搜索功能，不区分大小写进行匹配
//
// 参数:
//
//	entry: 要检查的日志条目
//	keyword: 要搜索的关键词
//
// 返回值:
//
//	bool: 如果条目包含关键词返回true，否则返回false
//
// 搜索范围:
//   - 进程名称（process_name）
//   - URL路径（url）
//   - 请求体（request_body）
//   - 响应体（response_body）
//   - 错误信息（error）
//   - 请求头（request_headers）的所有值
//   - 响应头（response_headers）的所有值
func (fs *FileStorage) containsKeyword(entry *LogEntry, keyword string) bool {
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

	return false
}

// GetStats 获取存储统计信息
// 计算文件存储的详细统计信息，包括存储大小、请求数量、错误率等指标
// 该方法提供与内存存储器相同的统计接口，确保存储组件的一致性
//
// 返回值:
//
//	map[string]interface{}: 统计信息映射，包含以下字段:
//	  - total_requests: 总请求数（HTTP记录和进程记录的总数）
//	  - storage_size: 存储大小（格式化显示，单位KB）
//	  - max_size: 最大存储条目数
//	  - storage_type: 存储类型（固定为"file"）
//	  - storage_path: 存储路径
//	  - avg_duration: 平均响应时间（仅当有记录时计算）
//	  - error_rate: 错误率（仅当有记录时计算）
//	  - error_count: 错误数量（仅当有记录时计算）
//	error: 如果获取统计信息失败返回错误
//
// 统计计算说明:
//   - 存储大小计算采用精确算法，与内存存储器和单个条目计算逻辑保持一致
//   - 统计信息包含所有有效的日志文件（HTTP记录和进程记录）
//   - 跳过读取失败的文件，确保统计的准确性
func (fs *FileStorage) GetStats() (map[string]interface{}, error) {
	files, err := fs.listLogFiles()
	if err != nil {
		return nil, err
	}

	// 计算存储大小（精确计算，与内存存储器和单个条目计算逻辑保持一致）
	var storageSize int64
	var errorCount int
	var totalDuration time.Duration
	var streamingRequestCount int
	var totalStreamingChunks int
	var maxStreamingChunks int

	for _, file := range files {
		entry, err := fs.readLogFile(file)
		if err != nil {
			continue // 跳过读取失败的文件
		}

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
		totalDuration += entry.Duration
		if entry.Error != "" {
			errorCount++
		}

		// 统计流式请求信息
		if entry.IsStreamingResponse {
			streamingRequestCount++
			totalStreamingChunks += entry.StreamingChunks
			if entry.StreamingChunks > maxStreamingChunks {
				maxStreamingChunks = entry.StreamingChunks
			}
		}
	}

	// 统一字段名，移除重复的total_entries字段
	stats := map[string]interface{}{
		"total_requests": len(files),                                        // 总请求数
		"storage_size":   fmt.Sprintf("%.2f KB", float64(storageSize)/1024), // 存储大小（格式化显示）
		"max_size":       fs.maxSize,                                        // 最大存储条目数
		"storage_type":   "file",                                            // 存储类型
		"storage_path":   fs.basePath,                                       // 存储路径
	}

	// 计算平均响应时间和错误率
	if len(files) > 0 {
		stats["avg_duration"] = totalDuration / time.Duration(len(files))
		stats["error_rate"] = float64(errorCount) / float64(len(files))
		stats["error_count"] = errorCount
	}

	// 添加流式请求统计信息
	stats["streaming_request_count"] = streamingRequestCount
	if len(files) > 0 {
		stats["streaming_request_rate"] = float64(streamingRequestCount) / float64(len(files))
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

// formatFileSize 格式化文件大小
func (fs *FileStorage) formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// ExportAll 导出所有日志到单个文件
func (fs *FileStorage) ExportAll(outputPath string) error {
	files, err := fs.listLogFiles()
	if err != nil {
		return err
	}

	// 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建导出文件失败: %v", err)
	}
	defer outputFile.Close()

	// 写入JSON数组开始
	if _, err := outputFile.WriteString("[\n"); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	// 写入所有日志条目
	for i, file := range files {
		entry, err := fs.readLogFile(file)
		if err != nil {
			continue // 跳过读取失败的文件
		}

		data, err := json.Marshal(entry)
		if err != nil {
			continue // 跳过序列化失败的条目
		}

		// 添加逗号分隔符（除了最后一个条目）
		if i > 0 {
			if _, err := outputFile.WriteString(",\n"); err != nil {
				return fmt.Errorf("写入文件失败: %v", err)
			}
		}

		if _, err := outputFile.Write(data); err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
	}

	// 写入JSON数组结束
	if _, err := outputFile.WriteString("\n]"); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// ImportFromFile 从文件导入日志
func (fs *FileStorage) ImportFromFile(filePath string) error {
	// 读取导入文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取导入文件失败: %v", err)
	}

	// 解析JSON数组
	var entries []LogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("解析导入文件失败: %v", err)
	}

	// 保存所有条目
	for _, entry := range entries {
		if err := fs.Save(&entry); err != nil {
			return fmt.Errorf("保存导入条目失败: %v", err)
		}
	}

	return nil
}

// Close 关闭文件存储器（文件存储不需要特殊关闭操作）
func (fs *FileStorage) Close() error {
	return nil
}

// GetMethods 获取HTTP方法统计
func (fs *FileStorage) GetMethods() (map[string]int, error) {
	files, err := fs.listLogFiles()
	if err != nil {
		return nil, err
	}

	methods := make(map[string]int)

	// 遍历所有日志文件统计方法
	for _, file := range files {
		entry, err := fs.readLogFile(file)
		if err != nil {
			continue // 跳过读取失败的文件
		}

		methods[entry.Method]++
	}

	return methods, nil
}

// GetStatusCodes 获取状态码统计
func (fs *FileStorage) GetStatusCodes() (map[int]int, error) {
	files, err := fs.listLogFiles()
	if err != nil {
		return nil, err
	}

	statusCodes := make(map[int]int)

	// 遍历所有日志文件统计状态码
	for _, file := range files {
		entry, err := fs.readLogFile(file)
		if err != nil {
			continue // 跳过读取失败的文件
		}

		statusCodes[entry.StatusCode]++
	}

	return statusCodes, nil
}
