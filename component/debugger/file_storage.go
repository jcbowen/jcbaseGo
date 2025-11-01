package debugger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
func (fs *FileStorage) FindAll(page, pageSize int, filters map[string]interface{}) ([]*LogEntry, int, error) {
	// 获取所有日志文件
	files, err := fs.listLogFiles()
	if err != nil {
		return nil, 0, err
	}

	// 读取所有日志条目
	var entries []*LogEntry
	for _, file := range files {
		entry, err := fs.readLogFile(file)
		if err != nil {
			continue // 跳过读取失败的文件
		}

		// 应用过滤器
		if fs.filterEntry(entry, filters) {
			entries = append(entries, entry)
		}
	}

	// 按时间倒序排序
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	// 计算分页
	total := len(entries)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []*LogEntry{}, total, nil
	}

	if end > total {
		end = total
	}

	return entries[start:end], total, nil
}

// Search 搜索日志内容
func (fs *FileStorage) Search(keyword string, page, pageSize int) ([]*LogEntry, int, error) {
	// 获取所有日志文件
	files, err := fs.listLogFiles()
	if err != nil {
		return nil, 0, err
	}

	// 搜索匹配的日志条目
	var results []*LogEntry
	for _, file := range files {
		entry, err := fs.readLogFile(file)
		if err != nil {
			continue // 跳过读取失败的文件
		}

		// 检查是否包含关键词
		if fs.containsKeyword(entry, keyword) {
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
func (fs *FileStorage) filterEntry(entry *LogEntry, filters map[string]interface{}) bool {
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
			if entry.ClientIP != value && !strings.Contains(entry.ClientIP, value.(string)) && !strings.HasPrefix(entry.ClientIP, value.(string)) {
				return false
			}
		case "has_error":
			if value.(bool) && entry.Error == "" {
				return false
			}
			if !value.(bool) && entry.Error != "" {
				return false
			}
		}
	}

	return true
}

// containsKeyword 检查日志条目是否包含关键词
func (fs *FileStorage) containsKeyword(entry *LogEntry, keyword string) bool {
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
func (fs *FileStorage) GetStats() (map[string]interface{}, error) {
	files, err := fs.listLogFiles()
	if err != nil {
		return nil, err
	}

	// 计算存储大小（精确计算，与内存存储器和单个条目计算逻辑保持一致）
	var storageSize int64
	var errorCount int
	var totalDuration time.Duration

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
