package debugger

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// QueryManager 日志查询管理器
// 提供丰富的查询和过滤功能，支持多种查询条件组合
type QueryManager struct {
	storage Storage // 存储接口
}

// NewQueryManager 创建新的查询管理器
// storage: 存储接口实例
func NewQueryManager(storage Storage) *QueryManager {
	return &QueryManager{
		storage: storage,
	}
}

// QueryOptions 查询选项
// 定义查询的各种过滤条件和分页参数
type QueryOptions struct {
	Page      int                    `json:"page"`       // 页码，从1开始
	PageSize  int                    `json:"page_size"`  // 每页大小
	Filters   map[string]interface{} `json:"filters"`    // 过滤条件
	SortBy    string                 `json:"sort_by"`    // 排序字段
	SortOrder string                 `json:"sort_order"` // 排序方向：asc/desc
}

// QueryResult 查询结果
// 包含查询到的日志条目和分页信息
type QueryResult struct {
	Entries    []*LogEntry `json:"entries"`     // 日志条目列表
	Total      int         `json:"total"`       // 总记录数
	Page       int         `json:"page"`        // 当前页码
	PageSize   int         `json:"page_size"`   // 每页大小
	TotalPages int         `json:"total_pages"` // 总页数
}

// Query 执行查询
// 根据查询选项获取日志条目
func (qm *QueryManager) Query(options QueryOptions) (*QueryResult, error) {
	// 设置默认值
	if options.Page <= 0 {
		options.Page = 1
	}
	if options.PageSize <= 0 {
		options.PageSize = 20
	}
	if options.PageSize > 100 {
		options.PageSize = 100 // 限制最大页大小
	}

	// 执行查询
	entries, total, err := qm.storage.FindAll(options.Page, options.PageSize, options.Filters)
	if err != nil {
		return nil, fmt.Errorf("查询日志失败: %v", err)
	}

	// 应用排序
	if options.SortBy != "" {
		qm.sortEntries(entries, options.SortBy, options.SortOrder)
	}

	// 计算总页数
	totalPages := (total + options.PageSize - 1) / options.PageSize

	return &QueryResult{
		Entries:    entries,
		Total:      total,
		Page:       options.Page,
		PageSize:   options.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Search 执行搜索
// 根据关键词搜索日志内容
func (qm *QueryManager) Search(keyword string, page, pageSize int) (*QueryResult, error) {
	// 设置默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大页大小
	}

	// 执行搜索
	entries, total, err := qm.storage.Search(keyword, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("搜索日志失败: %v", err)
	}

	// 计算总页数
	totalPages := (total + pageSize - 1) / pageSize

	return &QueryResult{
		Entries:    entries,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetByID 根据ID获取日志条目
func (qm *QueryManager) GetByID(id string) (*LogEntry, error) {
	return qm.storage.FindByID(id)
}

// GetRecent 获取最近的日志条目
// limit: 限制返回的条目数量
func (qm *QueryManager) GetRecent(limit int) ([]*LogEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	options := QueryOptions{
		Page:      1,
		PageSize:  limit,
		SortBy:    "timestamp",
		SortOrder: "desc",
	}

	result, err := qm.Query(options)
	if err != nil {
		return nil, err
	}

	return result.Entries, nil
}

// GetByTimeRange 根据时间范围查询日志
// startTime: 开始时间
// endTime: 结束时间
func (qm *QueryManager) GetByTimeRange(startTime, endTime time.Time, page, pageSize int) (*QueryResult, error) {
	filters := map[string]interface{}{
		"start_time": startTime,
		"end_time":   endTime,
	}

	options := QueryOptions{
		Page:      page,
		PageSize:  pageSize,
		Filters:   filters,
		SortBy:    "timestamp",
		SortOrder: "desc",
	}

	return qm.Query(options)
}

// GetByMethod 根据HTTP方法查询日志
func (qm *QueryManager) GetByMethod(method string, page, pageSize int) (*QueryResult, error) {
	filters := map[string]interface{}{
		"method": strings.ToUpper(method),
	}

	options := QueryOptions{
		Page:      page,
		PageSize:  pageSize,
		Filters:   filters,
		SortBy:    "timestamp",
		SortOrder: "desc",
	}

	return qm.Query(options)
}

// GetByStatusCode 根据状态码查询日志
func (qm *QueryManager) GetByStatusCode(statusCode int, page, pageSize int) (*QueryResult, error) {
	filters := map[string]interface{}{
		"status_code": statusCode,
	}

	options := QueryOptions{
		Page:      page,
		PageSize:  pageSize,
		Filters:   filters,
		SortBy:    "timestamp",
		SortOrder: "desc",
	}

	return qm.Query(options)
}

// GetByURL 根据URL查询日志（模糊匹配）
func (qm *QueryManager) GetByURL(urlPattern string, page, pageSize int) (*QueryResult, error) {
	filters := map[string]interface{}{
		"url": urlPattern,
	}

	options := QueryOptions{
		Page:      page,
		PageSize:  pageSize,
		Filters:   filters,
		SortBy:    "timestamp",
		SortOrder: "desc",
	}

	return qm.Query(options)
}

// GetErrors 获取包含错误的日志
func (qm *QueryManager) GetErrors(page, pageSize int) (*QueryResult, error) {
	filters := map[string]interface{}{
		"has_error": true,
	}

	options := QueryOptions{
		Page:      page,
		PageSize:  pageSize,
		Filters:   filters,
		SortBy:    "timestamp",
		SortOrder: "desc",
	}

	return qm.Query(options)
}

// GetSlowRequests 获取慢请求日志
// threshold: 慢请求阈值
func (qm *QueryManager) GetSlowRequests(threshold time.Duration, page, pageSize int) (*QueryResult, error) {
	filters := map[string]interface{}{
		"min_duration": threshold,
	}

	options := QueryOptions{
		Page:      page,
		PageSize:  pageSize,
		Filters:   filters,
		SortBy:    "duration",
		SortOrder: "desc",
	}

	return qm.Query(options)
}

// GetStats 获取统计信息
func (qm *QueryManager) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取存储统计
	if storageStats, err := qm.storage.GetStats(); err == nil {
		stats["storage"] = storageStats
	}

	// 获取方法统计
	if methods, err := qm.storage.GetMethods(); err == nil {
		stats["methods"] = methods
	}

	// 获取状态码统计
	if statusCodes, err := qm.storage.GetStatusCodes(); err == nil {
		stats["status_codes"] = statusCodes
	}

	// 获取时间范围统计
	timeStats, err := qm.getTimeStats()
	if err == nil {
		stats["time_stats"] = timeStats
	}

	return stats, nil
}

// getTimeStats 获取时间统计信息
func (qm *QueryManager) getTimeStats() (map[string]interface{}, error) {
	timeStats := make(map[string]interface{})

	// 获取今天的日志
	todayStart := time.Now().Truncate(24 * time.Hour)
	todayEnd := todayStart.Add(24 * time.Hour)

	todayResult, err := qm.GetByTimeRange(todayStart, todayEnd, 1, 1)
	if err == nil {
		timeStats["today_count"] = todayResult.Total
	}

	// 获取本周的日志
	weekStart := todayStart.AddDate(0, 0, -int(todayStart.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 7)

	weekResult, err := qm.GetByTimeRange(weekStart, weekEnd, 1, 1)
	if err == nil {
		timeStats["week_count"] = weekResult.Total
	}

	// 获取本月的日志
	monthStart := time.Date(todayStart.Year(), todayStart.Month(), 1, 0, 0, 0, 0, todayStart.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	monthResult, err := qm.GetByTimeRange(monthStart, monthEnd, 1, 1)
	if err == nil {
		timeStats["month_count"] = monthResult.Total
	}

	return timeStats, nil
}

// Export 导出日志数据
// format: 导出格式（json, csv等）
func (qm *QueryManager) Export(format string, options QueryOptions) ([]byte, error) {
	// 获取所有符合条件的日志（不分页）
	options.Page = 1
	options.PageSize = 10000 // 最大导出数量

	result, err := qm.Query(options)
	if err != nil {
		return nil, err
	}

	// 根据格式导出
	switch strings.ToLower(format) {
	case "json":
		return qm.exportJSON(result.Entries)
	case "csv":
		return qm.exportCSV(result.Entries)
	default:
		return nil, fmt.Errorf("不支持的导出格式: %s", format)
	}
}

// exportJSON 导出为JSON格式
func (qm *QueryManager) exportJSON(entries []*LogEntry) ([]byte, error) {
	type ExportEntry struct {
		ID         string    `json:"id"`
		Timestamp  time.Time `json:"timestamp"`
		Method     string    `json:"method"`
		URL        string    `json:"url"`
		StatusCode int       `json:"status_code"`
		Duration   string    `json:"duration"`
		ClientIP   string    `json:"client_ip"`
		UserAgent  string    `json:"user_agent"`
		RequestID  string    `json:"request_id"`
		Error      string    `json:"error,omitempty"`
	}

	exportEntries := make([]ExportEntry, len(entries))
	for i, entry := range entries {
		exportEntries[i] = ExportEntry{
			ID:         entry.ID,
			Timestamp:  entry.Timestamp,
			Method:     entry.Method,
			URL:        entry.URL,
			StatusCode: entry.StatusCode,
			Duration:   entry.Duration.String(),
			ClientIP:   entry.ClientIP,
			UserAgent:  entry.UserAgent,
			RequestID:  entry.RequestID,
			Error:      entry.Error,
		}
	}

	return json.MarshalIndent(exportEntries, "", "  ")
}

// exportCSV 导出为CSV格式
func (qm *QueryManager) exportCSV(entries []*LogEntry) ([]byte, error) {
	csvData := "ID,Timestamp,Method,URL,StatusCode,Duration,ClientIP,UserAgent,RequestID,Error\n"

	for _, entry := range entries {
		// 转义CSV特殊字符
		url := strings.ReplaceAll(entry.URL, "\"", "\"\"")
		userAgent := strings.ReplaceAll(entry.UserAgent, "\"", "\"\"")
		errorMsg := strings.ReplaceAll(entry.Error, "\"", "\"\"")

		csvData += fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",%d,\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n",
			entry.ID,
			entry.Timestamp.Format(time.RFC3339),
			entry.Method,
			url,
			entry.StatusCode,
			entry.Duration.String(),
			entry.ClientIP,
			userAgent,
			entry.RequestID,
			errorMsg,
		)
	}

	return []byte(csvData), nil
}

// sortEntries 对日志条目进行排序
func (qm *QueryManager) sortEntries(entries []*LogEntry, sortBy, sortOrder string) {
	sort.Slice(entries, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "timestamp":
			less = entries[i].Timestamp.Before(entries[j].Timestamp)
		case "method":
			less = entries[i].Method < entries[j].Method
		case "url":
			less = entries[i].URL < entries[j].URL
		case "status_code":
			less = entries[i].StatusCode < entries[j].StatusCode
		case "duration":
			less = entries[i].Duration < entries[j].Duration
		case "client_ip":
			less = entries[i].ClientIP < entries[j].ClientIP
		default:
			// 默认按时间排序
			less = entries[i].Timestamp.Before(entries[j].Timestamp)
		}

		// 根据排序方向调整
		if sortOrder == "desc" {
			return !less
		}

		return less
	})
}

// Cleanup 清理过期日志
// before: 清理此时间之前的日志
func (qm *QueryManager) Cleanup(before time.Time) error {
	return qm.storage.Cleanup(before)
}

// GetFilterOptions 获取可用的过滤选项
func (qm *QueryManager) GetFilterOptions() map[string]interface{} {
	return map[string]interface{}{
		"methods":      []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"},
		"status_codes": []int{200, 201, 204, 301, 302, 400, 401, 403, 404, 500, 502, 503},
		"time_ranges": map[string]string{
			"today":      "今天",
			"yesterday":  "昨天",
			"this_week":  "本周",
			"last_week":  "上周",
			"this_month": "本月",
			"last_month": "上月",
		},
	}
}
