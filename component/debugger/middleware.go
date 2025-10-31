package debugger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware Gin中间件，用于拦截请求并记录调试信息
// 该中间件会捕获请求和响应的所有关键信息，包括请求头、请求体、响应头、响应体等
type Middleware struct {
	debugger *Debugger // 调试器实例
}

// NewMiddleware 创建新的中间件实例
// debugger: 调试器实例，用于存储日志
func NewMiddleware(debugger *Debugger) *Middleware { // 创建中间件
	middleware := &Middleware{
		debugger: debugger,
	}

	return middleware
}

// Handler 中间件处理函数
// 实现gin.HandlerFunc接口，用于拦截HTTP请求
func (m *Middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果调试器未启用，直接跳过
		if !m.debugger.config.Enabled {
			c.Next()
			return
		}

		// 创建日志条目
		entry := &LogEntry{
			ID:        generateID(),
			Timestamp: time.Now(),
			Method:    c.Request.Method,
			URL:       c.Request.URL.String(),
			ClientIP:  getClientIP(c),
			UserAgent: c.Request.UserAgent(),
			RequestID: getRequestID(c),
		}

		// 记录开始时间
		startTime := time.Now()

		// 捕获请求头
		entry.RequestHeaders = make(map[string]string)
		for key, values := range c.Request.Header {
			entry.RequestHeaders[key] = strings.Join(values, ", ")
		}

		// 捕获查询参数
		entry.QueryParams = make(map[string]string)
		for key, values := range c.Request.URL.Query() {
			entry.QueryParams[key] = strings.Join(values, ", ")
		}

		// 捕获请求体（如果存在且可读取）
		if c.Request.Body != nil {
			// 读取请求体内容
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				// 恢复请求体，以便后续处理
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// 尝试格式化JSON，如果失败则保持原样
				var formattedBody bytes.Buffer
				if err := json.Indent(&formattedBody, bodyBytes, "", "  "); err == nil {
					entry.RequestBody = formattedBody.String()
				} else {
					// 如果不是JSON，直接存储原始内容
					entry.RequestBody = string(bodyBytes)
				}
			}
		}

		// 创建自定义的ResponseWriter来捕获响应
		responseWriter := &responseCaptureWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// 处理请求
		c.Next()

		// 记录结束时间和持续时间
		entry.Duration = time.Since(startTime)

		// 捕获响应状态码
		entry.StatusCode = responseWriter.statusCode
		if entry.StatusCode == 0 {
			entry.StatusCode = http.StatusOK // 默认状态码
		}

		// 捕获响应头
		entry.ResponseHeaders = make(map[string]string)
		for key, values := range c.Writer.Header() {
			entry.ResponseHeaders[key] = strings.Join(values, ", ")
		}

		// 捕获响应体
		responseBody := responseWriter.body.Bytes()
		if len(responseBody) > 0 {
			// 尝试格式化JSON，如果失败则保持原样
			var formattedBody bytes.Buffer
			if err := json.Indent(&formattedBody, responseBody, "", "  "); err == nil {
				entry.ResponseBody = formattedBody.String()
			} else {
				// 如果不是JSON，直接存储原始内容
				entry.ResponseBody = string(responseBody)
			}
		}

		// 捕获会话数据（如果存在）
		if sessionData := getSessionData(c); sessionData != nil {
			entry.SessionData = sessionData
		}

		// 捕获错误信息（如果存在）
		if len(c.Errors) > 0 {
			errors := make([]string, len(c.Errors))
			for i, err := range c.Errors {
				errors[i] = err.Error()
			}
			entry.Error = strings.Join(errors, "; ")
		}

		// 保存日志条目
		if err := m.debugger.storage.Save(entry); err != nil {
			// 记录保存失败的错误，但不影响请求处理
			fmt.Printf("保存调试日志失败: %v\n", err)
		}
	}
}

// responseCaptureWriter 自定义ResponseWriter，用于捕获响应内容
type responseCaptureWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer // 捕获响应体
	statusCode int           // 捕获状态码
}

// Write 重写Write方法，捕获响应体
func (w *responseCaptureWriter) Write(data []byte) (int, error) {
	// 捕获响应体
	w.body.Write(data)
	// 调用原始的Write方法
	return w.ResponseWriter.Write(data)
}

// WriteHeader 重写WriteHeader方法，捕获状态码
func (w *responseCaptureWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// getClientIP 获取客户端IP地址
// 支持从X-Forwarded-For等代理头中获取真实IP
func getClientIP(c *gin.Context) string {
	// 尝试从X-Forwarded-For获取
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从X-Real-IP获取
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	// 使用远程地址
	return c.ClientIP()
}

// getRequestID 获取请求ID
// 如果请求头中有X-Request-ID，则使用它，否则生成新的ID
func getRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}

	// 如果没有请求ID，生成一个
	return generateMiddlewareID()
}

// getSessionData 获取会话数据
// 从Gin上下文中提取会话信息
func getSessionData(c *gin.Context) map[string]interface{} {
	sessionData := make(map[string]interface{})

	// 尝试从上下文中获取用户信息
	if user, exists := c.Get("user"); exists {
		sessionData["user"] = user
	}

	// 尝试从上下文中获取会话ID
	if sessionID := c.GetHeader("Authorization"); sessionID != "" {
		sessionData["session_id"] = sessionID
	}

	// 尝试从Cookie中获取会话信息
	if sessionCookie, err := c.Cookie("session"); err == nil {
		sessionData["session_cookie"] = sessionCookie
	}

	// 如果没有任何会话数据，返回nil
	if len(sessionData) == 0 {
		return nil
	}

	return sessionData
}

// generateMiddlewareID 生成中间件唯一ID
// 使用时间戳和随机数生成唯一标识符
func generateMiddlewareID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("middleware_%d", timestamp)
}

// EnableMiddleware 启用中间件的快捷方法
// 创建一个已启用调试功能的中间件
func EnableMiddleware(storage Storage, config ...Config) gin.HandlerFunc {
	// 创建默认配置
	cfg := Config{
		Enabled: true,
	}

	// 如果提供了自定义配置，使用它
	if len(config) > 0 {
		cfg = config[0]
		cfg.Enabled = true // 确保启用
	}

	// 创建调试器
	cfg.Storage = storage
	debugger, _ := New(&cfg)

	// 创建中间件并返回其处理器
	return NewMiddleware(debugger).Handler()
}

// DisableMiddleware 禁用中间件的快捷方法
// 创建一个空中间件，不进行任何调试记录
func DisableMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// MiddlewareWithConfig 使用配置创建中间件
// 提供更灵活的配置选项
func MiddlewareWithConfig(storage Storage, config Config) gin.HandlerFunc {
	config.Storage = storage
	debugger, _ := New(&config)
	// 创建中间件并返回其处理器
	return NewMiddleware(debugger).Handler()
}

// RequestLogger 简化的请求日志中间件
// 只记录基本的请求信息，适用于生产环境
func RequestLogger(storage Storage) gin.HandlerFunc {
	config := Config{
		Enabled: true,
	}

	config.Storage = storage
	debugger, _ := New(&config)

	return func(c *gin.Context) {
		// 只记录基本信息
		startTime := time.Now()

		c.Next()

		// 创建简化的日志条目
		entry := &LogEntry{
			ID:         generateID(),
			Timestamp:  startTime,
			Method:     c.Request.Method,
			URL:        c.Request.URL.String(),
			ClientIP:   getClientIP(c),
			StatusCode: c.Writer.Status(),
			Duration:   time.Since(startTime),
			UserAgent:  c.Request.UserAgent(),
			RequestID:  getRequestID(c),
		}

		// 保存简化的日志
		_ = debugger.storage.Save(entry)
	}
}

// DetailedDebugLogger 详细的调试日志中间件
// 记录完整的请求和响应信息，适用于开发环境
func DetailedDebugLogger(storage Storage) gin.HandlerFunc {
	config := Config{
		Enabled: true,
	}

	config.Storage = storage
	debugger, _ := New(&config)

	return debugger.Middleware()
}
