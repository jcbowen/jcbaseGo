package debugger

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/security"
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
	AllowedIPs []string `json:"allowed_ips" default:""` // 允许访问的IP白名单，空数组表示不限制

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
			if logger, ok := loggerValue.(*DefaultLogger); ok {
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
		d.controller = NewController(d, router, nil)
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
