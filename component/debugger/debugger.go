package debugger

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/security"
	"github.com/jcbowen/jcbaseGo/middleware"
)

// Config 调试器配置结构
type Config struct {
	Enabled         bool          `json:"enabled" default:"false"`         // 是否启用调试器
	MaxBodySize     int64         `json:"max_body_size" default:"1024"`    // 最大请求/响应体大小（KB），默认1MB
	RetentionPeriod time.Duration `json:"retention_period" default:"168h"` // 日志保留期限，默认7天
	Level           string        `json:"level" default:"debug"`           // 日志级别：debug/info
	MaxRecords      int           `json:"max_records" default:"150"`       // 最大记录数量，默认150条

	// 过滤配置
	SkipPaths   []string `json:"skip_paths" default:""`          // 跳过的路径（如静态文件："/static/,/favicon.ico"）
	SkipMethods []string `json:"skip_methods" default:"OPTIONS"` // 跳过的HTTP方法

	// 采样配置
	SampleRate float64 `json:"sample_rate" default:"1.0"` // 采样率（0-1之间），默认记录所有请求

	// IP访问控制配置
	AllowedIPs []string `json:"allowed_ips" default:""`  // 允许访问的IP白名单，空数组表示不限制
	UseCDN     bool     `json:"use_cdn" default:"false"` // 是否使用CDN，影响真实IP获取方式

	// 流式请求配置
	EnableStreamingSupport bool  `json:"enable_streaming_support" default:"false"` // 是否启用流式请求支持
	StreamingChunkSize     int64 `json:"streaming_chunk_size" default:"1024"`      // 流式响应分块大小（KB），默认1MB
	MaxStreamingChunks     int   `json:"max_streaming_chunks" default:"10"`        // 最大流式响应分块数量，默认10个

	// 核心组件配置 - 必须传入实例化的存储器
	Storage Storage         `json:"-"` // 存储实现（必须传入实例化的存储器）
	Logger  LoggerInterface `json:"-"` // 日志记录器（推荐直接传入实例化的日志记录器）
}

// Debugger 调试器主结构
type Debugger struct {
	config     *Config
	storage    Storage
	controller *Controller
	logger     LoggerInterface // 日志记录器实例
}

// New 创建新的调试器实例
// config: 调试器配置，如果没有传入存储实例，将使用默认的内存存储
func New(config *Config) (*Debugger, error) {
	if config == nil {
		return nil, fmt.Errorf("配置不能为nil，请提供有效的Config实例")
	}

	d := &Debugger{
		config: config,
	}

	// 使用CheckAndSetDefault方法设置默认值，符合jcbaseGo规范
	if err := helper.CheckAndSetDefault(d.config); err != nil {
		return nil, fmt.Errorf("设置配置默认值失败: %w", err)
	}

	// 如果没有传入存储实例，则使用默认的内存存储
	if d.config.Storage == nil {
		// 创建默认内存存储，使用配置中的MaxRecords作为最大记录数
		memoryStorage, err := NewMemoryStorage(d.config.MaxRecords)
		if err != nil {
			return nil, fmt.Errorf("创建默认内存存储失败: %w", err)
		}
		d.storage = memoryStorage
	} else {
		d.storage = d.config.Storage
	}

	// 优先使用传入的日志记录器实例
	if d.config.Logger != nil {
		d.logger = d.config.Logger
	} else {
		// 创建默认日志记录器
		d.logger = &DefaultLogger{
			debugger: d,
			fields:   make(map[string]interface{}),
		}
	}

	return d, nil
}

// Middleware 创建Gin中间件
// 用于拦截HTTP请求并记录调试信息
func (d *Debugger) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用调试器
		if !d.config.Enabled {
			c.Next()
			return
		}

		// 检查是否跳过当前请求
		if d.shouldSkip(c) {
			c.Next()
			return
		}

		// 检查采样率
		if d.config.SampleRate < 1.0 && !d.shouldSample() {
			c.Next()
			return
		}

		// 记录请求开始时间
		startTime := time.Now()

		// 创建日志条目
		entry := d.createLogEntry(c, startTime)

		// 设置请求ID到上下文
		c.Set("debugger_request_id", entry.ID)

		// 设置Logger到上下文，供控制器使用
		c.Set("debugger_logger", d.logger.WithFields(map[string]interface{}{
			"request_id": entry.ID,
			"method":     c.Request.Method,
			"url":        c.Request.URL.String(),
			"client_ip":  middleware.GetRealIP(c, d.config.UseCDN),
		}))

		// 创建自定义的ResponseWriter来捕获响应
		writer := &responseWriter{
			ResponseWriter:  c.Writer,
			debugger:        d, // 保存Debugger实例引用
			body:            &bytes.Buffer{},
			isStreaming:     false, // 初始化为非流式响应
			chunkSizeLimit:  d.config.StreamingChunkSize,
			maxChunks:       d.config.MaxStreamingChunks,
			streamingChunks: make([]StreamingChunk, 0),
		}
		c.Writer = writer

		// 处理请求
		c.Next()

		// 记录响应信息
		d.recordResponseInfo(entry, writer, startTime)

		// 记录会话数据（如果存在）
		d.recordSessionData(entry, c)

		// 记录错误信息
		d.recordErrorInfo(entry, c)

		// 从上下文中获取logger并保存其收集的日志
		d.recordLoggerLogs(entry, c)

		// 保存日志条目
		d.saveLogEntry(entry)
	}
}

// createLogEntry 创建日志条目
func (d *Debugger) createLogEntry(c *gin.Context, startTime time.Time) *LogEntry {
	entry := &LogEntry{
		ID:             GenerateID(),
		Timestamp:      startTime,
		Method:         c.Request.Method,
		URL:            c.Request.URL.String(),
		ClientIP:       middleware.GetRealIP(c, d.config.UseCDN),
		UserAgent:      c.Request.UserAgent(),
		RequestID:      c.GetHeader("X-Request-ID"),
		RecordType:     "http", // 设置记录类型为HTTP
		RequestHeaders: helper.ExtractHeaders(c.Request.Header),
		QueryParams:    extractQueryParams(c.Request.URL.Query()),
	}

	// 记录请求体
	if body, err := d.extractRequestBody(c); err == nil {
		entry.RequestBody = body
	}

	return entry
}

// recordResponseInfo 记录响应信息
// 自动识别二进制响应数据并适当处理，避免文件下载内容显示为乱码
// 支持流式响应记录，对实时流式请求进行特殊处理
func (d *Debugger) recordResponseInfo(entry *LogEntry, writer *responseWriter, startTime time.Time) {
	entry.StatusCode = writer.Status()
	entry.Duration = time.Since(startTime)
	entry.ResponseHeaders = helper.ExtractHeaders(writer.Header())

	// 检查是否为流式响应（只有在启用流式支持时才处理流式响应）
	if d.config.EnableStreamingSupport && writer.isStreaming {
		// 流式响应处理
		entry.RecordType = "streaming" // 标记为流式请求
		entry.ResponseBody = d.formatStreamingResponse(writer)

		// 记录流式响应元数据
		entry.IsStreamingResponse = true
		entry.StreamingChunks = len(writer.streamingChunks)
		entry.StreamingChunkSize = int(writer.chunkSizeLimit)
		entry.MaxStreamingChunks = int(writer.maxChunks)
		entry.StreamingData = fmt.Sprintf("Streaming Response: %d chunks, total size: %d bytes",
			len(writer.streamingChunks), d.calculateTotalStreamingSize(writer))
	} else {
		// 非流式响应处理
		entry.RecordType = "http" // 标记为普通HTTP请求

		// 记录响应体
		if writer.body.Len() > 0 {
			// 检查响应头是否包含gzip压缩
			contentEncoding := writer.Header().Get("Content-Encoding")

			// 如果是gzip压缩的响应，尝试解压缩
			if strings.Contains(contentEncoding, "gzip") {
				// 解压缩gzip数据
				reader, err := gzip.NewReader(bytes.NewReader(writer.body.Bytes()))
				if err == nil {
					defer func(reader *gzip.Reader) {
						err = reader.Close()
						if err != nil {
							d.logger.Error("关闭gzip读取器失败" + err.Error())
						}
					}(reader)
					decompressed, err := io.ReadAll(reader)
					if err == nil {
						// 检查解压后的数据是否为二进制
						if isBinaryData(decompressed) {
							entry.ResponseBody = formatBinaryData(decompressed)
						} else {
							entry.ResponseBody = string(decompressed)
						}
					} else {
						// 解压缩失败，记录原始数据并添加错误标记
						entry.ResponseBody = "[GZIP解压缩失败] " + formatBinaryData(writer.body.Bytes())
					}
				} else {
					// 创建gzip读取器失败，记录原始数据
					entry.ResponseBody = "[GZIP格式错误] " + formatBinaryData(writer.body.Bytes())
				}
			} else {
				// 非gzip压缩的响应，检查是否为二进制数据
				responseBytes := writer.body.Bytes()
				if isBinaryData(responseBytes) {
					entry.ResponseBody = formatBinaryData(responseBytes)
				} else {
					entry.ResponseBody = string(responseBytes)
				}
			}
		}
	}
}

// formatStreamingResponse 格式化流式响应数据
// 对流式响应的分块数据进行格式化显示
func (d *Debugger) formatStreamingResponse(writer *responseWriter) string {
	if len(writer.streamingChunks) == 0 {
		return "[Streaming Response: No chunks recorded]"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("[Streaming Response: %d chunks, total size: %d bytes]\n",
		len(writer.streamingChunks), d.calculateTotalStreamingSize(writer)))

	for i, chunk := range writer.streamingChunks {
		result.WriteString(fmt.Sprintf("Chunk %d (size: %d bytes, time: %s): ",
			i+1, chunk.Size, chunk.Timestamp.Format("15:04:05.000")))

		if chunk.IsBinary {
			result.WriteString("[Binary Data] ")
		}

		// 限制显示的数据长度
		if len(chunk.Data) > 200 {
			result.WriteString(chunk.Data[:200] + "...")
		} else {
			result.WriteString(chunk.Data)
		}
		result.WriteString("\n")
	}

	return result.String()
}

// calculateTotalStreamingSize 计算流式响应的总大小
func (d *Debugger) calculateTotalStreamingSize(writer *responseWriter) int {
	total := 0
	for _, chunk := range writer.streamingChunks {
		total += chunk.Size
	}
	return total
}

// recordSessionData 记录会话数据
func (d *Debugger) recordSessionData(entry *LogEntry, c *gin.Context) {
	if sessionData, exists := c.Get("session_data"); exists {
		if data, ok := sessionData.(map[string]interface{}); ok {
			entry.SessionData = data
		}
	}
}

// recordErrorInfo 记录错误信息
func (d *Debugger) recordErrorInfo(entry *LogEntry, c *gin.Context) {
	if len(c.Errors) > 0 {
		var errorMsgs []string
		for _, err := range c.Errors {
			errorMsgs = append(errorMsgs, err.Error())
		}
		entry.Error = strings.Join(errorMsgs, "; ")
	}
}

// recordLoggerLogs 记录logger收集的日志
func (d *Debugger) recordLoggerLogs(entry *LogEntry, c *gin.Context) {
	if loggerValue, exists := c.Get("debugger_logger"); exists {
		if logger, ok := loggerValue.(*DefaultLogger); ok {
			// 获取logger收集的所有日志
			entry.LoggerLogs = logger.GetLogs()
			// 清空logger的日志记录，避免内存泄漏
			logger.ClearLogs()
		}
	}
}

// saveLogEntry 保存日志条目
func (d *Debugger) saveLogEntry(entry *LogEntry) {
	if err := d.storage.Save(entry); err != nil {
		// 记录保存错误，但不影响正常请求处理
		fmt.Printf("保存调试日志失败: %v\n", err)
	}
}

// shouldSkip 检查是否应该跳过当前请求
func (d *Debugger) shouldSkip(c *gin.Context) bool {
	// 自动跳过debugger控制器自身的请求，避免无限循环（放在最前面提高性能）
	if d.controller != nil && d.controller.config != nil {
		// 检查当前请求路径是否以debugger控制器的基础路径开头
		if strings.HasPrefix(c.Request.URL.Path, d.controller.config.BasePath) {
			return true
		}
	}

	// 检查路径
	for _, path := range d.config.SkipPaths {
		if strings.HasPrefix(c.Request.URL.Path, path) {
			return true
		}
	}

	// 检查方法
	for _, method := range d.config.SkipMethods {
		if c.Request.Method == method {
			return true
		}
	}

	return false
}

// shouldSample 根据采样率决定是否记录当前请求
func (d *Debugger) shouldSample() bool {
	if d.config.SampleRate >= 1.0 {
		return true
	}
	return time.Now().UnixNano()%100 < int64(d.config.SampleRate*100)
}

// extractRequestBody 提取请求体内容
func (d *Debugger) extractRequestBody(c *gin.Context) (string, error) {
	if c.Request.Body == nil {
		return "", nil
	}

	// 读取请求体
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", err
	}

	// 恢复请求体，以便后续处理
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 限制体大小（MaxBodySize单位为KB，需要转换为字节）
	if int64(len(bodyBytes)) > d.config.MaxBodySize*1024 {
		return fmt.Sprintf("[Body too large: %d bytes]", len(bodyBytes)), nil
	}

	// 检查是否为multipart/form-data文件上传
	if isMultipartFormData(c) {
		return formatMultipartFormData(c, bodyBytes), nil
	}

	// 检查是否为二进制数据（文件上传等）
	if isBinaryData(bodyBytes) {
		// 对于二进制数据，显示为十六进制格式或文件信息
		return formatBinaryData(bodyBytes), nil
	}

	// 对于文本数据，直接转换为字符串
	return string(bodyBytes), nil
}

// isBinaryData 检查数据是否为二进制格式
// 通过检测不可打印字符的比例来判断是否为二进制数据
func isBinaryData(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// 统计不可打印字符的数量
	nonPrintableCount := 0
	for i := 0; i < len(data) && i < 1024; i++ { // 只检查前1KB以提高性能
		b := data[i]
		// 不可打印字符：ASCII码小于32且不是制表符、换行符、回车符
		if b < 32 && b != '\t' && b != '\n' && b != '\r' {
			nonPrintableCount++
		}
	}

	// 如果不可打印字符超过10%，则认为是二进制数据
	totalChecked := min(len(data), 1024)
	return float64(nonPrintableCount)/float64(totalChecked) > 0.1
}

// formatBinaryData 格式化二进制数据显示
// 对于小文件显示十六进制预览和文件类型，对于大文件显示文件信息
func formatBinaryData(data []byte) string {
	fileType := detectFileType(data)

	if len(data) <= 512 {
		// 小文件：显示十六进制预览和文件类型
		hexPreview := fmt.Sprintf("%x", data)
		if len(hexPreview) > 200 {
			hexPreview = hexPreview[:200] + "..."
		}
		return fmt.Sprintf("[Binary Data: %d bytes, Type: %s, Hex: %s]", len(data), fileType, hexPreview)
	}

	// 大文件：显示文件信息
	return fmt.Sprintf("[Binary File: %d bytes, Type: %s]", len(data), fileType)
}

// detectFileType 检测文件类型
// 通过文件魔数识别常见文件类型
func detectFileType(data []byte) string {
	if len(data) < 8 {
		return "Unknown"
	}

	// 常见文件类型的魔数检测
	switch {
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}):
		return "JPEG"
	case bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
		return "PNG"
	case bytes.HasPrefix(data, []byte{0x47, 0x49, 0x46, 0x38}):
		return "GIF"
	case bytes.HasPrefix(data, []byte{0x25, 0x50, 0x44, 0x46}):
		return "PDF"
	case bytes.HasPrefix(data, []byte{0x50, 0x4B, 0x03, 0x04}):
		return "ZIP/Office Document"
	case bytes.HasPrefix(data, []byte{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07, 0x00}):
		return "RAR"
	case bytes.HasPrefix(data, []byte{0x49, 0x44, 0x33}):
		return "MP3"
	case bytes.HasPrefix(data, []byte{0x66, 0x74, 0x79, 0x70, 0x69, 0x73, 0x6F, 0x6D}):
		return "MP4"
	default:
		// 检查是否为文本文件（大部分字符可打印）
		printableCount := 0
		for i := 0; i < min(len(data), 100); i++ {
			b := data[i]
			if b >= 32 && b <= 126 {
				printableCount++
			}
		}

		if float64(printableCount)/float64(min(len(data), 100)) > 0.8 {
			return "Text"
		}
		return "Binary"
	}
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isMultipartFormData 检查是否为multipart/form-data文件上传请求
// 通过Content-Type头识别multipart/form-data类型
func isMultipartFormData(c *gin.Context) bool {
	contentType := c.Request.Header.Get("Content-Type")
	return strings.Contains(contentType, "multipart/form-data")
}

// extractBoundary 从Content-Type头中提取boundary参数
// 用于multipart/form-data请求的解析
func extractBoundary(contentType string) string {
	if !strings.Contains(contentType, "multipart/form-data") {
		return ""
	}

	// 查找boundary参数
	boundaryPrefix := "boundary="
	boundaryIndex := strings.Index(contentType, boundaryPrefix)
	if boundaryIndex == -1 {
		return ""
	}

	boundaryIndex += len(boundaryPrefix)
	boundaryEnd := strings.Index(contentType[boundaryIndex:], ";")

	if boundaryEnd == -1 {
		// 如果没有分号，取到字符串末尾
		return strings.Trim(contentType[boundaryIndex:], " \"")
	}

	return strings.Trim(contentType[boundaryIndex:boundaryIndex+boundaryEnd], " \"")
}

// formatMultipartFormData 格式化multipart/form-data请求体
// 解析multipart数据并显示文件上传信息，避免显示二进制乱码
func formatMultipartFormData(c *gin.Context, bodyBytes []byte) string {
	contentType := c.Request.Header.Get("Content-Type")

	// 提取boundary参数
	boundary := extractBoundary(contentType)
	if boundary == "" {
		// 如果没有boundary，回退到二进制数据显示
		return formatBinaryData(bodyBytes)
	}

	// 创建multipart reader
	reader := bytes.NewReader(bodyBytes)
	mr := multipart.NewReader(reader, boundary)
	if mr == nil {
		// 如果创建失败，回退到二进制数据显示
		return formatBinaryData(bodyBytes)
	}

	var result strings.Builder
	result.WriteString("[Multipart Form Data]\n")
	result.WriteString(fmt.Sprintf("Total Size: %d bytes\n", len(bodyBytes)))
	result.WriteString("Parts:\n")

	// 解析multipart的各个部分
	partCount := 0
	fileCount := 0
	textCount := 0

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.WriteString(fmt.Sprintf("  Error reading part: %v\n", err))
			continue
		}

		partCount++
		partName := part.FormName()
		fileName := part.FileName()

		// 读取部分内容
		partBytes, err := io.ReadAll(part)
		if err != nil {
			result.WriteString(fmt.Sprintf("  Error reading part content: %v\n", err))
			continue
		}

		if fileName != "" {
			// 文件上传部分
			fileCount++
			result.WriteString(fmt.Sprintf("  [File] Name: %s, Field: %s, Size: %d bytes, Type: %s\n",
				fileName, partName, len(partBytes), detectFileType(partBytes)))
		} else {
			// 文本字段部分
			textCount++
			if isBinaryData(partBytes) {
				result.WriteString(fmt.Sprintf("  [Field] Name: %s, Size: %d bytes, Type: Binary\n",
					partName, len(partBytes)))
			} else {
				// 限制文本长度，避免显示过长
				textContent := string(partBytes)
				if len(textContent) > 100 {
					textContent = textContent[:100] + "..."
				}
				result.WriteString(fmt.Sprintf("  [Field] Name: %s, Value: %s\n",
					partName, textContent))
			}
		}
	}

	result.WriteString(fmt.Sprintf("\nSummary: %d parts total (%d files, %d fields)",
		partCount, fileCount, textCount))

	return result.String()
}

// responseWriter 自定义ResponseWriter用于捕获响应
type responseWriter struct {
	gin.ResponseWriter
	debugger        *Debugger // 指向Debugger实例的引用
	body            *bytes.Buffer
	status          int
	isStreaming     bool             // 是否为流式响应
	streamingChunks []StreamingChunk // 流式响应分块记录
	chunkSizeLimit  int64            // 单个分块大小限制
	maxChunks       int              // 最大分块数量
}

// StreamingChunk 流式响应分块记录
// 用于记录流式响应的分块数据，避免内存溢出
type StreamingChunk struct {
	Timestamp time.Time `json:"timestamp"` // 分块接收时间
	Size      int       `json:"size"`      // 分块大小
	Data      string    `json:"data"`      // 分块数据（截断后）
	IsBinary  bool      `json:"is_binary"` // 是否为二进制数据
}

// Write 重写Write方法以捕获响应体
// 支持流式响应处理，自动检测流式请求并分块记录
func (w *responseWriter) Write(b []byte) (int, error) {
	// 检查是否为流式响应（只有在启用流式支持时才处理流式响应）
	if w.debugger.config.EnableStreamingSupport && w.isStreamingResponse() {
		w.isStreaming = true
		w.recordStreamingChunk(b)
	} else {
		// 非流式响应，使用原有逻辑
		w.isStreaming = false // 确保禁用流式支持时标记为非流式响应
		w.body.Write(b)
	}

	return w.ResponseWriter.Write(b)
}

// isStreamingResponse 检测是否为流式响应
// 通过响应头信息判断是否为流式传输
func (w *responseWriter) isStreamingResponse() bool {
	return isStreamingResponseFromHeaders(w.Header())
}

// isStreamingResponseFromHeaders 根据响应头检测是否为流式响应
// 独立的函数，便于测试使用
func isStreamingResponseFromHeaders(headers http.Header) bool {
	contentType := headers.Get("Content-Type")
	transferEncoding := headers.Get("Transfer-Encoding")

	// 检查是否为分块传输编码
	if transferEncoding == "chunked" {
		return true
	}

	// 检查是否为Server-Sent Events
	if contentType == "text/event-stream" {
		return true
	}

	// 检查是否为流式JSON响应
	if contentType == "application/x-ndjson" || contentType == "application/json-seq" {
		return true
	}

	// 检查是否为二进制流式响应
	if contentType == "application/octet-stream" {
		return true
	}

	// 检查是否为WebSocket升级响应
	if headers.Get("Upgrade") == "websocket" {
		return true
	}

	return false
}

// recordStreamingChunk 记录流式响应分块
// 对每个分块进行大小限制和数量限制，避免内存溢出
func (w *responseWriter) recordStreamingChunk(data []byte) {
	// 检查是否超过最大分块数量
	if len(w.streamingChunks) >= w.maxChunks {
		return
	}

	chunk := StreamingChunk{
		Timestamp: time.Now(),
		Size:      len(data),
		IsBinary:  w.isBinaryData(data),
	}

	// 限制单个分块的数据大小
	maxChunkSize := w.chunkSizeLimit * 1024 // 转换为字节
	if int64(len(data)) > maxChunkSize {
		chunk.Data = string(data[:maxChunkSize]) + "... [truncated]"
	} else {
		chunk.Data = string(data)
	}

	w.streamingChunks = append(w.streamingChunks, chunk)
}

// isBinaryData 检测数据是否为二进制数据
// 通过不可打印字符比例判断是否为二进制数据
func (w *responseWriter) isBinaryData(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	binaryCount := 0
	for _, b := range data {
		if b < 32 && b != 9 && b != 10 && b != 13 { // 排除制表符、换行符、回车符
			binaryCount++
		}
	}

	// 如果不可打印字符超过30%，认为是二进制数据
	binaryRatio := float64(binaryCount) / float64(len(data))
	return binaryRatio > 0.3
}

// WriteHeader 重写WriteHeader方法以捕获状态码
func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Status 获取状态码
func (w *responseWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

// GetStorage 获取存储实例（用于外部访问）
func (d *Debugger) GetStorage() Storage {
	return d.storage
}

// GetConfig 获取配置信息
func (d *Debugger) GetConfig() *Config {
	return d.config
}

// SetSessionData 设置会话数据（供外部调用）
func (d *Debugger) SetSessionData(c *gin.Context, data map[string]interface{}) {
	c.Set("session_data", data)
}

// GetLogger 获取调试器的日志记录器
// 可以在控制器中通过此方法获取Logger实例来记录日志
func (d *Debugger) GetLogger() LoggerInterface {
	return d.logger
}

// GetLoggerWithFields 获取带有指定字段的日志记录器
func (d *Debugger) GetLoggerWithFields(fields map[string]interface{}) LoggerInterface {
	return d.logger.WithFields(fields)
}

// WithController 为调试器添加控制器支持
// router: Gin引擎，用于注册调试器页面路由
// config: 控制器配置（可选）
func (d *Debugger) WithController(router *gin.Engine, config *ControllerConfig) *Debugger {
	d.controller = NewController(d, router, config)
	return d
}

// GetController 获取控制器实例
func (d *Debugger) GetController() *Controller {
	return d.controller
}

// RegisterRoutes 手动注册路由到指定的Gin引擎
// 只支持 *gin.Engine 类型，使用更简洁
func (d *Debugger) RegisterRoutes(router *gin.Engine) {
	// 使用根路径创建路由组
	if d.controller == nil {
		d.controller = NewController(d, router, &ControllerConfig{UseCDN: d.config.UseCDN})
	} else {
		d.controller.RegisterRoutes(router)
	}
}

// DefaultConfig 返回默认配置
// 使用helper.CheckAndSetDefault方法设置默认值，符合jcbaseGo规范
func DefaultConfig() *Config {
	config := &Config{}

	// 使用CheckAndSetDefault方法设置默认值
	if err := helper.CheckAndSetDefault(config); err != nil {
		panic(fmt.Sprintf("DefaultConfig failed: %v", err))
	}

	return config
}

// extractQueryParams 提取查询参数
func extractQueryParams(values map[string][]string) map[string]string {
	result := make(map[string]string)
	for key, vals := range values {
		result[key] = strings.Join(vals, ", ")
	}
	return result
}

// GenerateID 生成唯一请求标识
// 使用纳秒时间戳和进程ID确保唯一性，经过SM4加密后输出，密文结果不包含=字符
func GenerateID() string {
	// 使用纳秒时间戳的后8位作为基础
	timestamp := time.Now().UnixNano()
	idSuffix := timestamp % 100000000

	// 添加进程ID的后2位作为额外标识，避免高并发下的重复
	pid := os.Getpid() % 100

	// 组合格式：d_时间戳后8位_进程ID后2位
	originalID := fmt.Sprintf("d_%d_%02d", idSuffix, pid)

	// 使用SM4加密原始ID，确保安全性和唯一性
	sm4Instance := security.SM4{
		Text:     originalID,
		Encoding: "RawURL", // 使用RawURL编码避免=字符
	}

	var encryptedID string
	err := sm4Instance.Encrypt(&encryptedID)
	if err != nil {
		// 如果加密失败，返回原始ID作为降级方案
		return originalID
	}

	return encryptedID
}

// ==================== 便捷存储器构造函数 ====================

// NewWithMemoryStorage 创建使用内存存储器的调试器
// 适用于开发和测试环境
func NewWithMemoryStorage(maxRecords int) (*Debugger, error) {
	memoryStorage, err := NewMemoryStorage(maxRecords)
	if err != nil {
		return nil, err
	}

	config := &Config{
		Enabled:    true,
		MaxRecords: maxRecords,
		Storage:    memoryStorage,
		Level:      LevelDebug,
	}
	// 设置默认值
	if err := helper.CheckAndSetDefault(config); err != nil {
		return nil, err
	}

	return New(config)
}

// NewWithFileStorage 创建使用文件存储器的调试器
// 适用于生产环境，日志会持久化到文件系统
func NewWithFileStorage(storagePath string, maxRecords int) (*Debugger, error) {
	fileStorage, err := NewFileStorage(storagePath, maxRecords)
	if err != nil {
		return nil, err
	}

	config := &Config{
		Enabled:    true,
		MaxRecords: maxRecords,
		Storage:    fileStorage,
		Level:      LevelDebug,
	}
	// 设置默认值
	if err := helper.CheckAndSetDefault(config); err != nil {
		return nil, err
	}

	return New(config)
}

// NewWithCustomStorage 创建使用自定义存储器的调试器
// 适用于需要自定义存储逻辑的场景
func NewWithCustomStorage(customStorage Storage) (*Debugger, error) {
	config := &Config{
		Enabled: true,
		Storage: customStorage,
		Level:   LevelDebug,
	}
	// 设置默认值
	if err := helper.CheckAndSetDefault(config); err != nil {
		return nil, err
	}

	return New(config)
}

// NewSimpleDebugger 创建简单的调试器实例
// 使用默认配置，适合快速开始
func NewSimpleDebugger() (*Debugger, error) {
	return NewWithMemoryStorage(150)
}

// NewProductionDebugger 创建生产环境调试器实例
// 使用文件存储，适合生产环境
func NewProductionDebugger(storagePath string) (*Debugger, error) {
	return NewWithFileStorage(storagePath, 1000)
}

// ==================== 进程级Debugger支持 ====================

// StartProcess 开始进程记录
// 创建一个进程级日志记录器，用于在进程中记录不同级别的debugger日志
// 该方法适用于后台任务、批处理作业、定时任务等非HTTP进程场景
//
// 参数:
//
//	processName: 进程名称，用于标识进程用途（如"数据同步任务"）
//	processType: 进程类型，可选值：background/worker/cron/batch等
//
// 返回值:
//
//	ProcessLoggerInterface: 进程级日志记录器实例
//
// 示例:
//
//	logger := dbg.StartProcess("数据同步任务", "batch")
//	defer dbg.EndProcess(logger.GetProcessID(), "completed")
func (d *Debugger) StartProcess(processName, processType string) ProcessLoggerInterface {
	return NewProcessLogger(d, processName, processType)
}

// GetProcessLogger 获取进程级日志记录器
// 通过进程ID获取已存在的进程记录器，用于在进程执行过程中获取记录器实例
// 该方法适用于需要跨函数或协程共享同一进程记录器的场景
//
// 参数:
//
//	processID: 进程ID，由StartProcess方法返回
//
// 返回值:
//
//	ProcessLoggerInterface: 进程级日志记录器实例
//	error: 错误信息，当进程记录不存在、类型不匹配或进程已结束时返回错误
//
// 示例:
//
//	logger, err := dbg.GetProcessLogger("process-123456")
//	if err != nil {
//	    // 处理错误
//	}
//	logger.Info("继续处理任务")
func (d *Debugger) GetProcessLogger(processID string) (ProcessLoggerInterface, error) {
	// 查找进程记录
	entry, err := d.storage.FindByID(processID)
	if err != nil {
		return nil, fmt.Errorf("进程记录不存在: %w", err)
	}

	// 检查记录类型
	if entry.RecordType != "process" {
		return nil, fmt.Errorf("记录类型不是进程记录")
	}

	// 检查进程状态
	if entry.Status != "running" {
		return nil, fmt.Errorf("进程已结束，无法获取记录器")
	}

	// 创建进程记录器实例
	logger := &ProcessLogger{
		debugger:    d,
		processID:   entry.ProcessID,
		processName: entry.ProcessName,
		processType: entry.ProcessType,
		startTime:   entry.Timestamp,
		logger: &DefaultLogger{
			debugger: d,
			fields: map[string]interface{}{
				"process_id":   entry.ProcessID,
				"process_name": entry.ProcessName,
				"process_type": entry.ProcessType,
			},
		},
	}

	return logger, nil
}

// EndProcess 结束进程记录
// 通过进程ID结束指定的进程记录，记录进程的结束时间和状态
// 该方法应在进程执行完成时调用，以确保记录完整的执行时间线
//
// 参数:
//
//	processID: 进程ID，由StartProcess方法返回
//	status: 进程结束状态，使用预定义的常量：ProcessStatusCompleted、ProcessStatusFailed、ProcessStatusCancelled
//
// 返回值:
//
//	error: 错误信息，当进程记录不存在、类型不匹配或保存失败时返回错误
//
// 示例:
//
//	err := dbg.EndProcess("process-123456", ProcessStatusCompleted)
//	if err != nil {
//	    // 处理错误
//	}
func (d *Debugger) EndProcess(processID, status string) error {
	// 查找进程记录
	entry, err := d.storage.FindByID(processID)
	if err != nil {
		return fmt.Errorf("进程记录不存在: %w", err)
	}

	// 检查记录类型
	if entry.RecordType != "process" {
		return fmt.Errorf("记录类型不是进程记录")
	}

	// 更新进程记录
	entry.EndTime = time.Now()
	entry.Duration = entry.EndTime.Sub(entry.Timestamp)
	entry.Status = status

	// 保存更新后的记录
	if err := d.storage.Save(entry); err != nil {
		return fmt.Errorf("保存进程记录失败: %w", err)
	}

	return nil
}

// GetProcessRecords 获取进程记录列表
// 支持分页和过滤查询进程记录，可用于监控和分析进程执行情况
//
// 参数:
//
//	page: 页码，从1开始
//	pageSize: 每页记录数
//	filters: 过滤条件，支持record_type/process_name/process_id等字段
//
// 返回值:
//
//	[]*LogEntry: 进程记录列表
//	int: 总记录数
//	error: 错误信息
//
// 示例:
//
//	records, total, err := dbg.GetProcessRecords(1, 20, map[string]interface{}{
//	    "process_name": "数据同步任务",
//	})
func (d *Debugger) GetProcessRecords(page, pageSize int, filters map[string]interface{}) ([]*LogEntry, int, error) {
	// 添加记录类型过滤
	if filters == nil {
		filters = make(map[string]interface{})
	}
	filters["record_type"] = "process"

	return d.storage.FindAll(page, pageSize, filters)
}
