package debugger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"time"
)

// DetailViewer 日志详情查看器
// 提供丰富的日志详情展示功能，支持格式化显示请求和响应信息
type DetailViewer struct {
	queryManager *QueryManager // 查询管理器
}

// NewDetailViewer 创建新的详情查看器
// queryManager: 查询管理器实例
func NewDetailViewer(queryManager *QueryManager) *DetailViewer {
	return &DetailViewer{
		queryManager: queryManager,
	}
}

// DetailView 详情视图
// 包含日志条目的完整信息和格式化后的内容
type DetailView struct {
	LogEntry *LogEntry `json:"log_entry"` // 原始日志条目

	// 格式化后的内容
	FormattedRequestHeaders  string `json:"formatted_request_headers"`  // 格式化请求头
	FormattedResponseHeaders string `json:"formatted_response_headers"` // 格式化响应头
	FormattedRequestBody     string `json:"formatted_request_body"`     // 格式化请求体
	FormattedResponseBody    string `json:"formatted_response_body"`    // 格式化响应体
	FormattedSessionData     string `json:"formatted_session_data"`     // 格式化会话数据
	FormattedQueryParams     string `json:"formatted_query_params"`     // 格式化查询参数

	// 进程记录专用格式化字段
	FormattedProcessInfo   string `json:"formatted_process_info"`   // 格式化进程信息
	FormattedEndTime       string `json:"formatted_end_time"`       // 格式化结束时间
	FormattedProcessStatus string `json:"formatted_process_status"` // 格式化进程状态

	// 统计信息
	RequestSize  int `json:"request_size"`  // 请求大小（字节）
	ResponseSize int `json:"response_size"` // 响应大小（字节）

	// 时间信息
	FormattedTimestamp string `json:"formatted_timestamp"` // 格式化时间戳
	FormattedDuration  string `json:"formatted_duration"`  // 格式化持续时间

	// 相关日志
	RelatedEntries []*LogEntry `json:"related_entries"` // 相关日志条目
}

// GetDetail 获取日志详情
// id: 日志条目ID
func (dv *DetailViewer) GetDetail(id string) (*DetailView, error) {
	// 获取日志条目
	entry, err := dv.queryManager.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取日志条目失败: %v", err)
	}

	// 创建详情视图
	detailView := &DetailView{
		LogEntry: entry,
	}

	// 格式化内容
	dv.formatDetailView(detailView)

	// 获取相关日志
	dv.getRelatedEntries(detailView)

	return detailView, nil
}

// formatDetailView 格式化详情视图
func (dv *DetailViewer) formatDetailView(detailView *DetailView) {
	entry := detailView.LogEntry

	// 格式化时间戳
	detailView.FormattedTimestamp = entry.Timestamp.Format("2006-01-02 15:04:05.000")

	// 格式化持续时间
	detailView.FormattedDuration = formatDuration(entry.Duration)

	// 根据记录类型进行不同的格式化处理
	if entry.RecordType == "process" {
		// 进程记录格式化
		dv.formatProcessDetail(detailView)
	} else {
		// HTTP记录格式化（默认）
		dv.formatHTTPDetail(detailView)
	}
}

// formatHTTPDetail 格式化HTTP记录详情
func (dv *DetailViewer) formatHTTPDetail(detailView *DetailView) {
	entry := detailView.LogEntry

	// 格式化请求头
	detailView.FormattedRequestHeaders = formatHeaders(entry.RequestHeaders)

	// 格式化响应头
	detailView.FormattedResponseHeaders = formatHeaders(entry.ResponseHeaders)

	// 格式化查询参数
	detailView.FormattedQueryParams = formatQueryParams(entry.QueryParams)

	// 格式化请求体
	detailView.FormattedRequestBody = formatBody(entry.RequestBody, "请求体")
	detailView.RequestSize = len(entry.RequestBody)

	// 格式化响应体
	detailView.FormattedResponseBody = formatBody(entry.ResponseBody, "响应体")
	detailView.ResponseSize = len(entry.ResponseBody)

	// 格式化会话数据
	detailView.FormattedSessionData = formatSessionData(entry.SessionData)
}

// formatProcessDetail 格式化进程记录详情
func (dv *DetailViewer) formatProcessDetail(detailView *DetailView) {
	entry := detailView.LogEntry

	// 格式化进程信息
	detailView.FormattedProcessInfo = dv.formatProcessInfo(entry)

	// 格式化结束时间
	if !entry.EndTime.IsZero() {
		detailView.FormattedEndTime = entry.EndTime.Format("2006-01-02 15:04:05.000")
	} else {
		detailView.FormattedEndTime = "进行中"
	}

	// 格式化进程状态
	detailView.FormattedProcessStatus = formatProcessStatus(entry.Status)

	// 对于进程记录，清空HTTP相关的格式化字段
	detailView.FormattedRequestHeaders = ""
	detailView.FormattedResponseHeaders = ""
	detailView.FormattedQueryParams = ""
	detailView.FormattedRequestBody = ""
	detailView.FormattedResponseBody = ""
	detailView.RequestSize = 0
	detailView.ResponseSize = 0
}

// formatProcessInfo 格式化进程信息
func (dv *DetailViewer) formatProcessInfo(entry *LogEntry) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("进程ID: %s\n", entry.ProcessID))
	buf.WriteString(fmt.Sprintf("进程名称: %s\n", entry.ProcessName))
	buf.WriteString(fmt.Sprintf("进程类型: %s\n", entry.ProcessType))

	if !entry.EndTime.IsZero() {
		buf.WriteString(fmt.Sprintf("开始时间: %s\n", entry.Timestamp.Format("2006-01-02 15:04:05.000")))
		buf.WriteString(fmt.Sprintf("结束时间: %s\n", entry.EndTime.Format("2006-01-02 15:04:05.000")))
	} else {
		buf.WriteString(fmt.Sprintf("开始时间: %s\n", entry.Timestamp.Format("2006-01-02 15:04:05.000")))
		buf.WriteString("结束时间: 进行中\n")
	}

	buf.WriteString(fmt.Sprintf("状态: %s\n", formatProcessStatus(entry.Status)))
	buf.WriteString(fmt.Sprintf("耗时: %s\n", formatDuration(entry.Duration)))

	return buf.String()
}

// formatProcessStatus 格式化进程状态
func formatProcessStatus(status string) string {
	switch status {
	case "running":
		return "运行中"
	case "completed":
		return "已完成"
	case "failed":
		return "失败"
	case "cancelled":
		return "已取消"
	default:
		return status
	}
}

// getRelatedEntries 获取相关日志条目
func (dv *DetailViewer) getRelatedEntries(detailView *DetailView) {
	entry := detailView.LogEntry

	// 获取相同请求ID的日志
	if entry.RequestID != "" {
		options := QueryOptions{
			Page:     1,
			PageSize: 10,
			Filters: map[string]interface{}{
				"request_id": entry.RequestID,
			},
			SortBy:    "timestamp",
			SortOrder: "desc",
		}

		result, err := dv.queryManager.Query(options)
		if err == nil && len(result.Entries) > 1 {
			// 过滤掉当前条目
			related := make([]*LogEntry, 0)
			for _, e := range result.Entries {
				if e.ID != entry.ID {
					related = append(related, e)
				}
			}
			detailView.RelatedEntries = related
		}
	}

	// 获取相同客户端IP的最近日志
	if entry.ClientIP != "" {
		options := QueryOptions{
			Page:     1,
			PageSize: 5,
			Filters: map[string]interface{}{
				"client_ip": entry.ClientIP,
			},
			SortBy:    "timestamp",
			SortOrder: "desc",
		}

		result, err := dv.queryManager.Query(options)
		if err == nil && len(result.Entries) > 0 {
			// 如果还没有相关日志，添加这些
			if len(detailView.RelatedEntries) == 0 {
				detailView.RelatedEntries = result.Entries
			} else {
				// 合并相关日志，去重
				existingIDs := make(map[string]bool)
				for _, e := range detailView.RelatedEntries {
					existingIDs[e.ID] = true
				}

				for _, e := range result.Entries {
					if !existingIDs[e.ID] && e.ID != entry.ID {
						detailView.RelatedEntries = append(detailView.RelatedEntries, e)
					}
				}
			}
		}
	}
}

// formatDuration 格式化持续时间
func formatDuration(duration time.Duration) string {
	if duration < time.Millisecond {
		return fmt.Sprintf("%dµs", duration.Microseconds())
	} else if duration < time.Second {
		return fmt.Sprintf("%.2fms", float64(duration.Microseconds())/1000)
	} else {
		return fmt.Sprintf("%.2fs", duration.Seconds())
	}
}

// formatHeaders 格式化头部信息
func formatHeaders(headers map[string]string) string {
	if len(headers) == 0 {
		return "无"
	}

	var buf bytes.Buffer
	for key, value := range headers {
		buf.WriteString(fmt.Sprintf("%s: %s\n", key, html.EscapeString(value)))
	}

	return buf.String()
}

// formatQueryParams 格式化查询参数
func formatQueryParams(params map[string]string) string {
	if len(params) == 0 {
		return "无"
	}

	var buf bytes.Buffer
	for key, value := range params {
		buf.WriteString(fmt.Sprintf("%s: %s\n", key, html.EscapeString(value)))
	}

	return buf.String()
}

// formatBody 格式化请求体或响应体
func formatBody(body, defaultText string) string {
	if body == "" {
		return defaultText
	}

	// 尝试解析为JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(body), &jsonData); err == nil {
		// 是JSON，进行格式化
		formatted, err := json.MarshalIndent(jsonData, "", "  ")
		if err == nil {
			return string(formatted)
		}
	}

	// 不是JSON，直接返回（进行HTML转义）
	return html.EscapeString(body)
}

// formatSessionData 格式化会话数据
func formatSessionData(sessionData map[string]interface{}) string {
	if len(sessionData) == 0 {
		return "无"
	}

	// 转换为JSON格式
	jsonData, err := json.MarshalIndent(sessionData, "", "  ")
	if err != nil {
		return fmt.Sprintf("格式化会话数据失败: %v", err)
	}

	return string(jsonData)
}

// GetDetailHTML 获取HTML格式的详情
func (dv *DetailViewer) GetDetailHTML(id string) (string, error) {
	detailView, err := dv.GetDetail(id)
	if err != nil {
		return "", err
	}

	return dv.generateHTML(detailView), nil
}

// generateHTML 生成HTML格式的详情
func (dv *DetailViewer) generateHTML(detailView *DetailView) string {
	var htmlBuilder strings.Builder

	htmlBuilder.WriteString(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>调试日志详情</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .section { margin-bottom: 20px; border: 1px solid #ddd; padding: 15px; border-radius: 5px; }
        .section-title { font-weight: bold; margin-bottom: 10px; color: #333; }
        .key-value { margin: 5px 0; }
        .key { font-weight: bold; display: inline-block; width: 120px; }
        .value { display: inline-block; }
        .preformatted { background: #f5f5f5; padding: 10px; border-radius: 3px; white-space: pre-wrap; font-family: monospace; }
        .error { color: #d32f2f; }
        .success { color: #388e3c; }
        .warning { color: #f57c00; }
        .related-entry { border: 1px solid #e0e0e0; padding: 10px; margin: 5px 0; border-radius: 3px; }
        .related-entry:hover { background: #f5f5f5; }
    </style>
</head>
<body>
    <h1>调试日志详情</h1>
`)

	entry := detailView.LogEntry

	// 基本信息
	htmlBuilder.WriteString(`<div class="section">
        <div class="section-title">基本信息</div>
        <div class="key-value"><span class="key">ID:</span><span class="value">` + entry.ID + `</span></div>
        <div class="key-value"><span class="key">时间:</span><span class="value">` + detailView.FormattedTimestamp + `</span></div>
        <div class="key-value"><span class="key">方法:</span><span class="value">` + entry.Method + `</span></div>
        <div class="key-value"><span class="key">URL:</span><span class="value">` + html.EscapeString(entry.URL) + `</span></div>
        <div class="key-value"><span class="key">状态码:</span><span class="value">` + getStatusCodeHTML(entry.StatusCode) + `</span></div>
        <div class="key-value"><span class="key">持续时间:</span><span class="value">` + detailView.FormattedDuration + `</span></div>
        <div class="key-value"><span class="key">客户端IP:</span><span class="value">` + entry.ClientIP + `</span></div>
        <div class="key-value"><span class="key">User Agent:</span><span class="value">` + html.EscapeString(entry.UserAgent) + `</span></div>
        <div class="key-value"><span class="key">请求ID:</span><span class="value">` + entry.RequestID + `</span></div>
    </div>
`)

	// 错误信息
	if entry.Error != "" {
		htmlBuilder.WriteString(`<div class="section">
            <div class="section-title error">错误信息</div>
            <div class="preformatted">` + html.EscapeString(entry.Error) + `</div>
        </div>
`)
	}

	// 查询参数
	htmlBuilder.WriteString(`<div class="section">
        <div class="section-title">查询参数</div>
        <div class="preformatted">` + detailView.FormattedQueryParams + `</div>
    </div>
`)

	// 请求头
	htmlBuilder.WriteString(`<div class="section">
        <div class="section-title">请求头</div>
        <div class="preformatted">` + detailView.FormattedRequestHeaders + `</div>
    </div>
`)

	// 请求体
	htmlBuilder.WriteString(`<div class="section">
        <div class="section-title">请求体 (` + fmt.Sprintf("%d 字节", detailView.RequestSize) + `)</div>
        <div class="preformatted">` + detailView.FormattedRequestBody + `</div>
    </div>
`)

	// 响应头
	htmlBuilder.WriteString(`<div class="section">
        <div class="section-title">响应头</div>
        <div class="preformatted">` + detailView.FormattedResponseHeaders + `</div>
    </div>
`)

	// 响应体
	htmlBuilder.WriteString(`<div class="section">
        <div class="section-title">响应体 (` + fmt.Sprintf("%d 字节", detailView.ResponseSize) + `)</div>
        <div class="preformatted">` + detailView.FormattedResponseBody + `</div>
    </div>
`)

	// 会话数据
	htmlBuilder.WriteString(`<div class="section">
        <div class="section-title">会话数据</div>
        <div class="preformatted">` + detailView.FormattedSessionData + `</div>
    </div>
`)

	// 相关日志
	if len(detailView.RelatedEntries) > 0 {
		htmlBuilder.WriteString(`<div class="section">
            <div class="section-title">相关日志 (` + fmt.Sprintf("%d 条", len(detailView.RelatedEntries)) + `)</div>
`)

		for _, related := range detailView.RelatedEntries {
			htmlBuilder.WriteString(`<div class="related-entry">
                <div class="key-value"><span class="key">时间:</span><span class="value">` + related.Timestamp.Format("2006-01-02 15:04:05") + `</span></div>
                <div class="key-value"><span class="key">方法:</span><span class="value">` + related.Method + `</span></div>
                <div class="key-value"><span class="key">URL:</span><span class="value">` + html.EscapeString(related.URL) + `</span></div>
                <div class="key-value"><span class="key">状态码:</span><span class="value">` + getStatusCodeHTML(related.StatusCode) + `</span></div>
            </div>
`)
		}

		htmlBuilder.WriteString(`</div>
`)
	}

	htmlBuilder.WriteString(`</body>
</html>`)

	return htmlBuilder.String()
}

// getStatusCodeHTML 获取状态码的HTML表示
func getStatusCodeHTML(statusCode int) string {
	class := ""
	if statusCode >= 200 && statusCode < 300 {
		class = "success"
	} else if statusCode >= 400 && statusCode < 500 {
		class = "warning"
	} else if statusCode >= 500 {
		class = "error"
	}

	if class != "" {
		return fmt.Sprintf(`<span class="%s">%d</span>`, class, statusCode)
	}

	return fmt.Sprintf("%d", statusCode)
}

// GetDetailJSON 获取JSON格式的详情
func (dv *DetailViewer) GetDetailJSON(id string) ([]byte, error) {
	detailView, err := dv.GetDetail(id)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(detailView, "", "  ")
}

// CompareEntries 比较两个日志条目
func (dv *DetailViewer) CompareEntries(id1, id2 string) (map[string]interface{}, error) {
	entry1, err := dv.queryManager.GetByID(id1)
	if err != nil {
		return nil, fmt.Errorf("获取第一个日志条目失败: %v", err)
	}

	entry2, err := dv.queryManager.GetByID(id2)
	if err != nil {
		return nil, fmt.Errorf("获取第二个日志条目失败: %v", err)
	}

	comparison := map[string]interface{}{
		"entry1":      entry1,
		"entry2":      entry2,
		"differences": dv.findDifferences(entry1, entry2),
	}

	return comparison, nil
}

// findDifferences 查找两个日志条目的差异
func (dv *DetailViewer) findDifferences(entry1, entry2 *LogEntry) map[string]interface{} {
	differences := make(map[string]interface{})

	// 比较基本信息
	if entry1.Method != entry2.Method {
		differences["method"] = map[string]string{
			"entry1": entry1.Method,
			"entry2": entry2.Method,
		}
	}

	if entry1.URL != entry2.URL {
		differences["url"] = map[string]string{
			"entry1": entry1.URL,
			"entry2": entry2.URL,
		}
	}

	if entry1.StatusCode != entry2.StatusCode {
		differences["status_code"] = map[string]int{
			"entry1": entry1.StatusCode,
			"entry2": entry2.StatusCode,
		}
	}

	if entry1.Duration != entry2.Duration {
		differences["duration"] = map[string]time.Duration{
			"entry1": entry1.Duration,
			"entry2": entry2.Duration,
		}
	}

	// 比较请求体
	if entry1.RequestBody != entry2.RequestBody {
		differences["request_body"] = map[string]string{
			"entry1": entry1.RequestBody,
			"entry2": entry2.RequestBody,
		}
	}

	// 比较响应体
	if entry1.ResponseBody != entry2.ResponseBody {
		differences["response_body"] = map[string]string{
			"entry1": entry1.ResponseBody,
			"entry2": entry2.ResponseBody,
		}
	}

	return differences
}
