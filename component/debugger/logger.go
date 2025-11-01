package debugger

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
)

// 日志级别常量
const (
	LevelDebug = "debug" // 调试级别：记录所有详细信息
	LevelInfo  = "info"  // 信息级别：只记录基本信息
	LevelWarn  = "warn"  // 警告级别：记录警告信息
	LevelError = "error" // 错误级别：记录错误信息
)

// LoggerInterface 日志记录器接口
// 支持不同级别的日志记录，可以在控制器中直接使用
type LoggerInterface interface {
	// Debug 记录调试级别日志
	Debug(msg any, fields ...map[string]interface{})

	// Info 记录信息级别日志
	Info(msg any, fields ...map[string]interface{})

	// Warn 记录警告级别日志
	Warn(msg any, fields ...map[string]interface{})

	// Error 记录错误级别日志
	Error(msg any, fields ...map[string]interface{})

	// WithFields 创建带有字段的日志记录器
	WithFields(fields map[string]interface{}) LoggerInterface

	// GetLevel 获取当前日志记录器的日志级别
	GetLevel() string
}

// ----- DefaultLogger 方法实现

// DefaultLogger 调试器内置的日志记录器实现（默认实现）
type DefaultLogger struct {
	debugger *Debugger
	level    string // 当前日志记录器的日志级别
	fields   map[string]interface{}
	logs     []LoggerLog // 存储收集的日志
}

// Debug 记录调试级别日志
func (l *DefaultLogger) Debug(msg any, fields ...map[string]interface{}) {
	l.log(LevelDebug, msg, fields...)
}

// Info 记录信息级别日志
func (l *DefaultLogger) Info(msg any, fields ...map[string]interface{}) {
	l.log(LevelInfo, msg, fields...)
}

// Warn 记录警告级别日志
func (l *DefaultLogger) Warn(msg any, fields ...map[string]interface{}) {
	l.log(LevelWarn, msg, fields...)
}

// Error 记录错误级别日志
func (l *DefaultLogger) Error(msg any, fields ...map[string]interface{}) {
	l.log(LevelError, msg, fields...)
}

// WithFields 创建带有字段的日志记录器
func (l *DefaultLogger) WithFields(fields map[string]interface{}) LoggerInterface {
	// 合并现有字段和新字段
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &DefaultLogger{
		debugger: l.debugger,
		fields:   newFields,
		logs:     l.logs,  // 继承父logger的日志
		level:    l.level, // 设置新的日志级别
	}
}

// GetLevel 获取当前日志记录器的日志级别
func (l *DefaultLogger) GetLevel() string {
	return l.debugger.config.Level
}

// log 内部日志记录方法
// - level: 日志级别（debug/info/warn/error）
// - msg: 日志消息（字符串、结构体、map、数组、实现了Stringer接口的类型等）
// - fields: 可选的附加字段（键值对）
func (l *DefaultLogger) log(level string, msg any, fields ...map[string]interface{}) {
	// 检查日志级别是否启用
	if !l.shouldLog(level) {
		return
	}

	// 处理msg参数
	var message string
	switch v := msg.(type) {
	case string:
		message = v
	case fmt.Stringer:
		message = v.String()
	default:
		helper.Json(msg).ToString(&message)
	}

	// 合并所有字段
	allFields := make(map[string]interface{})

	// 添加基础字段
	allFields["level"] = level
	allFields["message"] = message
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
		Message:   message,
		Fields:    allFields,
	}
	l.logs = append(l.logs, loggerLog)

	log.Println(msg)

	// // 格式化日志输出
	// logEntry := map[string]interface{}{
	// 	"debug_log": allFields,
	// }

	// // 转换为JSON格式输出
	// jsonData, err := json.Marshal(logEntry)
	// if err != nil {
	// 	// 如果JSON转换失败，使用简单格式输出
	// 	log.Printf("[%s] %s: %s", level, time.Now().Format("2006-01-02 15:04:05"), msg)
	// 	if len(l.fields) > 0 {
	// 		log.Printf(" fields=%v", l.fields)
	// 	}
	// 	if len(fields) > 0 {
	// 		log.Printf(" extra_fields=%v", fields[0])
	// 	}
	// 	return
	// }

	// log.Println(string(jsonData))
}

// shouldLog 检查是否应该记录指定级别的日志
func (l *DefaultLogger) shouldLog(level string) bool {
	// 根据配置的日志级别决定是否记录
	switch l.GetLevel() {
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
func (l *DefaultLogger) GetLogs() []LoggerLog {
	return l.logs
}

// ClearLogs 清空收集的日志信息
func (l *DefaultLogger) ClearLogs() {
	l.logs = []LoggerLog{}
}

// LoggerLog 记录通过logger打印的日志信息
type LoggerLog struct {
	Timestamp time.Time              `json:"timestamp"` // 日志时间戳
	Level     string                 `json:"level"`     // 日志级别：debug/info/warn/error
	Message   string                 `json:"message"`   // 日志消息
	Fields    map[string]interface{} `json:"fields"`    // 日志附加字段
}

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

// GetLoggerFromContext 从Gin上下文中获取Logger实例
// 控制器可以通过此函数获取Logger来记录日志
func GetLoggerFromContext(c *gin.Context) LoggerInterface {
	if logger, exists := c.Get("debugger_logger"); exists {
		if l, ok := logger.(LoggerInterface); ok {
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
