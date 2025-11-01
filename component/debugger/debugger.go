package debugger

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/security"
)

// LogEntry 调试日志条目结构
// 用于记录单个HTTP请求的完整调试信息
type LogEntry struct {
	ID         string        `json:"id"`          // 日志唯一标识
	Timestamp  time.Time     `json:"timestamp"`   // 请求时间戳
	Method     string        `json:"method"`      // HTTP方法
	URL        string        `json:"url"`         // 请求URL
	StatusCode int           `json:"status_code"` // HTTP状态码
	Duration   time.Duration `json:"duration"`    // 处理耗时
	ClientIP   string        `json:"client_ip"`   // 客户端IP
	UserAgent  string        `json:"user_agent"`  // 用户代理
	RequestID  string        `json:"request_id"`  // 请求ID（用于追踪）

	// 请求信息
	RequestHeaders map[string]string `json:"request_headers"` // 请求头
	QueryParams    map[string]string `json:"query_params"`    // 查询参数
	RequestBody    string            `json:"request_body"`    // 请求体内容

	// 响应信息
	ResponseHeaders map[string]string `json:"response_headers"` // 响应头
	ResponseBody    string            `json:"response_body"`    // 响应体内容

	// 会话数据（可选）
	SessionData map[string]interface{} `json:"session_data,omitempty"` // 会话数据

	// 错误信息
	Error string `json:"error,omitempty"` // 错误信息

	// Logger日志信息（新增）
	LoggerLogs []LoggerLog `json:"logger_logs,omitempty"` // 通过logger记录的日志

	// 存储大小（计算字段，不持久化到存储）
	StorageSize string `json:"storage_size,omitempty"` // 存储大小（格式化显示）
}

// LoggerLog 记录通过logger打印的日志信息
type LoggerLog struct {
	Timestamp time.Time              `json:"timestamp"` // 日志时间戳
	Level     string                 `json:"level"`     // 日志级别：debug/info/warn/error
	Message   string                 `json:"message"`   // 日志消息
	Fields    map[string]interface{} `json:"fields"`    // 日志附加字段
}

// Storage 存储接口定义
// 支持多种存储方式：文件、内存、数据库
type Storage interface {
	// Save 保存日志条目
	Save(entry *LogEntry) error

	// FindByID 根据ID查找日志条目
	FindByID(id string) (*LogEntry, error)

	// FindAll 查找所有日志条目，支持分页和过滤
	FindAll(page, pageSize int, filters map[string]interface{}) ([]*LogEntry, int, error)

	// Search 搜索日志内容
	Search(keyword string, page, pageSize int) ([]*LogEntry, int, error)

	// Cleanup 清理过期日志
	Cleanup(before time.Time) error

	// GetStats 获取统计信息
	GetStats() (map[string]interface{}, error)

	// GetMethods 获取HTTP方法统计
	GetMethods() (map[string]int, error)

	// GetStatusCodes 获取状态码统计
	GetStatusCodes() (map[int]int, error)

	// Close 关闭存储
	Close() error
}

// 日志级别常量
const (
	LevelDebug = "debug" // 调试级别：记录所有详细信息
	LevelInfo  = "info"  // 信息级别：只记录基本信息
	LevelWarn  = "warn"  // 警告级别：记录警告信息
	LevelError = "error" // 错误级别：记录错误信息
)

// Logger 日志记录器接口
// 支持不同级别的日志记录，可以在控制器中直接使用
type Logger interface {
	// Debug 记录调试级别日志
	Debug(msg string, fields ...map[string]interface{})

	// Info 记录信息级别日志
	Info(msg string, fields ...map[string]interface{})

	// Warn 记录警告级别日志
	Warn(msg string, fields ...map[string]interface{})

	// Error 记录错误级别日志
	Error(msg string, fields ...map[string]interface{})

	// WithFields 创建带有字段的日志记录器
	WithFields(fields map[string]interface{}) Logger
}

// DebugLogger 调试器内置的日志记录器实现
type DebugLogger struct {
	debugger *Debugger
	fields   map[string]interface{}
	logs     []LoggerLog // 存储收集的日志
}

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

	// 核心组件配置 - 必须传入实例化的存储器
	Storage Storage `json:"-"` // 存储实现（必须传入实例化的存储器）
	Logger  Logger  `json:"-"` // 日志记录器（推荐直接传入实例化的日志记录器）
}

// Debugger 调试器主结构
type Debugger struct {
	config     *Config
	storage    Storage
	controller *Controller
	logger     Logger // 日志记录器实例
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
		d.logger = &DebugLogger{
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
		entry := &LogEntry{
			ID:             GenerateID(),
			Timestamp:      startTime,
			Method:         c.Request.Method,
			URL:            c.Request.URL.String(),
			ClientIP:       c.ClientIP(),
			UserAgent:      c.Request.UserAgent(),
			RequestID:      c.GetHeader("X-Request-ID"),
			RequestHeaders: extractHeaders(c.Request.Header),
			QueryParams:    extractQueryParams(c.Request.URL.Query()),
		}

		// 记录请求体
		if body, err := d.extractRequestBody(c); err == nil {
			entry.RequestBody = body
		}

		// 设置请求ID到上下文
		c.Set("debugger_request_id", entry.ID)

		// 设置Logger到上下文，供控制器使用
		c.Set("debugger_logger", d.logger.WithFields(map[string]interface{}{
			"request_id": entry.ID,
			"method":     c.Request.Method,
			"url":        c.Request.URL.String(),
			"client_ip":  c.ClientIP(),
		}))

		// 创建自定义的ResponseWriter来捕获响应
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		// 处理请求
		c.Next()

		// 记录响应信息
		entry.StatusCode = writer.Status()
		entry.Duration = time.Since(startTime)
		entry.ResponseHeaders = extractHeaders(writer.Header())

		// 记录响应体
		if writer.body.Len() > 0 {
			// 检查响应头是否包含gzip压缩
			contentEncoding := writer.Header().Get("Content-Encoding")

			// 如果是gzip压缩的响应，尝试解压缩
			if strings.Contains(contentEncoding, "gzip") {
				// 解压缩gzip数据
				reader, err := gzip.NewReader(bytes.NewReader(writer.body.Bytes()))
				if err == nil {
					defer reader.Close()
					decompressed, err := io.ReadAll(reader)
					if err == nil {
						entry.ResponseBody = string(decompressed)
					} else {
						// 解压缩失败，记录原始数据并添加错误标记
						entry.ResponseBody = "[GZIP解压缩失败] " + string(writer.body.Bytes())
					}
				} else {
					// 创建gzip读取器失败，记录原始数据
					entry.ResponseBody = "[GZIP格式错误] " + string(writer.body.Bytes())
				}
			} else {
				// 非gzip压缩的响应，直接使用原始数据
				entry.ResponseBody = string(writer.body.Bytes())
			}
		}

		// 记录会话数据（如果存在）
		if sessionData, exists := c.Get("session_data"); exists {
			if data, ok := sessionData.(map[string]interface{}); ok {
				entry.SessionData = data
			}
		}

		// 记录错误信息
		if len(c.Errors) > 0 {
			var errorMsgs []string
			for _, err := range c.Errors {
				errorMsgs = append(errorMsgs, err.Error())
			}
			entry.Error = strings.Join(errorMsgs, "; ")
		}

		// 从上下文中获取logger并保存其收集的日志
		if loggerValue, exists := c.Get("debugger_logger"); exists {
			if logger, ok := loggerValue.(*DebugLogger); ok {
				// 获取logger收集的所有日志
				entry.LoggerLogs = logger.GetLogs()
				// 清空logger的日志记录，避免内存泄漏
				logger.ClearLogs()
			}
		}

		// 保存日志条目
		if err := d.storage.Save(entry); err != nil {
			// 记录保存错误，但不影响正常请求处理
			fmt.Printf("保存调试日志失败: %v\n", err)
		}
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

	return string(bodyBytes), nil
}

// extractHeaders 提取HTTP头信息
func extractHeaders(header http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range header {
		result[key] = strings.Join(values, ", ")
	}
	return result
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

// responseWriter 自定义ResponseWriter用于捕获响应
type responseWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

// Write 重写Write方法以捕获响应体
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
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

// JSONString 返回日志条目的JSON字符串表示
func (e *LogEntry) JSONString() string {
	data, _ := json.MarshalIndent(e, "", "  ")
	return string(data)
}

// Summary 返回日志条目的摘要信息
func (e *LogEntry) Summary() string {
	return fmt.Sprintf("%s %s %d %s", e.Method, e.URL, e.StatusCode, e.Duration)
}

// CalculateStorageSize 计算日志条目的存储大小并格式化显示
// 计算内容包括：基本字段、请求体、响应体、请求头、响应头、查询参数、会话数据等
// 与总存储计算逻辑保持一致
func (e *LogEntry) CalculateStorageSize() string {
	totalSize := 0

	// 计算基本字段大小（与总存储计算保持一致）
	totalSize += len(e.ID) + len(e.URL) + len(e.Method) + len(e.ClientIP) + len(e.UserAgent) + len(e.RequestID)

	// 计算请求体大小
	totalSize += len(e.RequestBody)

	// 计算响应体大小
	totalSize += len(e.ResponseBody)

	// 计算请求头大小
	for key, value := range e.RequestHeaders {
		totalSize += len(key) + len(value)
	}

	// 计算响应头大小
	for key, value := range e.ResponseHeaders {
		totalSize += len(key) + len(value)
	}

	// 计算查询参数大小
	for key, value := range e.QueryParams {
		totalSize += len(key) + len(value)
	}

	// 计算会话数据大小（JSON格式）
	if e.SessionData != nil {
		if sessionData, err := json.Marshal(e.SessionData); err == nil {
			totalSize += len(sessionData)
		}
	}

	// 计算错误信息大小
	totalSize += len(e.Error)

	// 计算Logger日志大小
	for _, log := range e.LoggerLogs {
		totalSize += len(log.Message)
		if log.Fields != nil {
			if fieldsData, err := json.Marshal(log.Fields); err == nil {
				totalSize += len(fieldsData)
			}
		}
	}

	// 格式化显示
	if totalSize < 1024 {
		return fmt.Sprintf("%d B", totalSize)
	} else if totalSize < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(totalSize)/1024)
	} else {
		return fmt.Sprintf("%.2f MB", float64(totalSize)/(1024*1024))
	}
}

// DebugLogger 方法实现

// Debug 记录调试级别日志
func (l *DebugLogger) Debug(msg string, fields ...map[string]interface{}) {
	l.log(LevelDebug, msg, fields...)
}

// Info 记录信息级别日志
func (l *DebugLogger) Info(msg string, fields ...map[string]interface{}) {
	l.log(LevelInfo, msg, fields...)
}

// Warn 记录警告级别日志
func (l *DebugLogger) Warn(msg string, fields ...map[string]interface{}) {
	l.log(LevelWarn, msg, fields...)
}

// Error 记录错误级别日志
func (l *DebugLogger) Error(msg string, fields ...map[string]interface{}) {
	l.log(LevelError, msg, fields...)
}

// WithFields 创建带有字段的日志记录器
func (l *DebugLogger) WithFields(fields map[string]interface{}) Logger {
	// 合并现有字段和新字段
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &DebugLogger{
		debugger: l.debugger,
		fields:   newFields,
		logs:     l.logs, // 继承父logger的日志
	}
}

// log 内部日志记录方法
func (l *DebugLogger) log(level, msg string, fields ...map[string]interface{}) {
	// 检查日志级别是否启用
	if !l.shouldLog(level) {
		return
	}

	// 合并所有字段
	allFields := make(map[string]interface{})

	// 添加基础字段
	allFields["level"] = level
	allFields["message"] = msg
	allFields["timestamp"] = time.Now().Format(time.RFC3339)

	// 添加实例字段
	for k, v := range l.fields {
		allFields[k] = v
	}

	// 添加调用方传入的字段
	if len(fields) > 0 {
		for k, v := range fields[0] {
			allFields[k] = v
		}
	}

	// 收集日志信息到logs字段
	loggerLog := LoggerLog{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Fields:    allFields,
	}
	l.logs = append(l.logs, loggerLog)

	// 格式化日志输出
	logEntry := map[string]interface{}{
		"debug_log": allFields,
	}

	// 转换为JSON格式输出
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		// 如果JSON转换失败，使用简单格式输出
		log.Printf("[%s] %s: %s", level, time.Now().Format("2006-01-02 15:04:05"), msg)
		if len(l.fields) > 0 {
			log.Printf(" fields=%v", l.fields)
		}
		if len(fields) > 0 {
			log.Printf(" extra_fields=%v", fields[0])
		}
		return
	}

	log.Println(string(jsonData))
}

// shouldLog 检查是否应该记录指定级别的日志
func (l *DebugLogger) shouldLog(level string) bool {
	// 根据配置的日志级别决定是否记录
	switch l.debugger.config.Level {
	case LevelDebug:
		// 调试级别记录所有日志
		return true
	case LevelInfo:
		// 信息级别记录info、warn、error
		return level == LevelInfo || level == LevelWarn || level == LevelError
	case LevelWarn:
		// 警告级别记录warn、error
		return level == LevelWarn || level == LevelError
	case LevelError:
		// 错误级别只记录error
		return level == LevelError
	default:
		// 默认记录所有日志
		return true
	}
}

// GetLogs 获取收集的日志信息
func (l *DebugLogger) GetLogs() []LoggerLog {
	return l.logs
}

// ClearLogs 清空收集的日志信息
func (l *DebugLogger) ClearLogs() {
	l.logs = []LoggerLog{}
}

// GetLogger 获取调试器的日志记录器
// 可以在控制器中通过此方法获取Logger实例来记录日志
func (d *Debugger) GetLogger() Logger {
	return d.logger
}

// GetLoggerWithFields 获取带有指定字段的日志记录器
func (d *Debugger) GetLoggerWithFields(fields map[string]interface{}) Logger {
	return d.logger.WithFields(fields)
}

// GetLoggerFromContext 从Gin上下文中获取Logger实例
// 控制器可以通过此函数获取Logger来记录日志
func GetLoggerFromContext(c *gin.Context) Logger {
	if logger, exists := c.Get("debugger_logger"); exists {
		if l, ok := logger.(Logger); ok {
			return l
		}
	}

	// 如果上下文中没有Logger，创建一个默认的调试器实例
	// 这种情况通常发生在调试器未启用或中间件未正确设置时
	memoryStorage, _ := NewMemoryStorage(150)
	config := &Config{
		Enabled: true,
		Storage: memoryStorage,
		Level:   LevelDebug,
	}
	// 设置默认值
	if err := helper.CheckAndSetDefault(config); err != nil {
		fmt.Printf("设置默认值失败: %v\n", err)
	}
	debugger, _ := New(config)

	return debugger.GetLogger().WithFields(map[string]interface{}{
		"context_error": "logger_not_found_in_context",
	})
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
		d.controller = NewController(d, router, nil)
	} else {
		d.controller.RegisterRoutes(router)
	}
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
