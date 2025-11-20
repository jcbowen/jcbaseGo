package debugger

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
)

type LogLevel int

// 日志级别常量
const (
	LevelSilent LogLevel = iota + 1 // 信息级别：不记录任何日志
	LevelError                      // 错误级别：只记录错误信息
	LevelWarn                       // 警告级别：记录错误+警告信息
	LevelInfo                       // 调试级别：记录所有详细信息
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case LevelSilent:
		return "silent"
	case LevelError:
		return "error"
	case LevelWarn:
		return "warn"
	case LevelInfo:
		return "info"
	default:
		return "unknown"
	}
}

// LoggerInterface 日志记录器接口
// 支持不同级别的日志记录，可以在控制器中直接使用
type LoggerInterface interface {
	// Info 记录信息级别日志
	Info(msg any, fields ...map[string]interface{})

	// Warn 记录警告级别日志
	Warn(msg any, fields ...map[string]interface{})

	// Error 记录错误级别日志
	Error(msg any, fields ...map[string]interface{})

	// WithFields 创建带有字段的日志记录器
	WithFields(fields map[string]interface{}) LoggerInterface

	// GetLevel 获取当前日志记录器的日志级别
	GetLevel() LogLevel
}

// ----- DefaultLogger 方法实现

// DefaultLogger 调试器内置的日志记录器实现（默认实现）
// 不能直接创建DefaultLogger实例，只能通过NewDefaultLogger创建
type DefaultLogger struct {
	debugger *Debugger
	level    LogLevel // 当前日志记录器的日志级别
	fields   map[string]interface{}
	logs     []LoggerLog // 存储收集的日志
}

// NewDefaultLogger 创建默认日志记录器实例
func NewDefaultLogger(debugger *Debugger) LoggerInterface {
	return &DefaultLogger{
		debugger: debugger,
		level:    debugger.config.Level,
		fields:   map[string]interface{}{},
		logs:     []LoggerLog{},
	}
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
func (l *DefaultLogger) GetLevel() LogLevel {
	return l.debugger.config.Level
}

// log 内部日志记录方法
// - level: 日志级别（info/warn/error/silent）
// - msg: 日志消息（字符串、结构体、map、数组、实现了Stringer接口的类型等）
// - fields: 可选的附加字段（键值对）
func (l *DefaultLogger) log(level LogLevel, msg any, fields ...map[string]interface{}) {
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

	// 获取调用位置信息
	fileName, line, function := getCallerInfo()

	// 收集日志信息到logs字段
	loggerLog := LoggerLog{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    allFields,
		FileName:  fileName,
		Line:      line,
		Function:  function,
	}
	l.logs = append(l.logs, loggerLog)

	// 输出包含位置信息的日志（支持IDE可点击链接）
	// 格式：[级别] 文件名:行号 - 函数名: 消息内容
	// 注意：文件名:行号 格式是IDEA控制台可点击链接的标准格式
	fmt.Printf("[%s] %s:%d - %s\n", level.String(), fileName, line, message)

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

// getCallerInfo 获取调用位置信息
// 返回文件名、行号和函数名
func getCallerInfo() (fileName string, line int, function string) {
	// 跳过当前函数和log方法，获取调用者的信息
	pc, file, lineNo, ok := runtime.Caller(3)
	if !ok {
		return "unknown", 0, "unknown"
	}

	// 获取函数名
	funcName := runtime.FuncForPC(pc).Name()

	// 简化文件名，只保留最后一部分
	parts := strings.Split(file, "/")
	if len(parts) > 0 {
		fileName = parts[len(parts)-1]
	} else {
		fileName = file
	}

	return fileName, lineNo, funcName
}

// shouldLog 检查是否应该记录指定级别的日志
func (l *DefaultLogger) shouldLog(level LogLevel) bool {
	// 根据配置的日志级别决定是否记录
	switch l.GetLevel() {
	case LevelInfo:
		// 调试级别记录所有日志
		return true
	case LevelWarn:
		// 警告级别记录warn、error
		return level == LevelWarn || level == LevelError
	case LevelError:
		// 错误级别只记录error
		return level == LevelError
	default:
		// 默认不记录日志
		return false
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
	Level     LogLevel               `json:"level"`     // 日志级别：debug/info/warn/error
	Message   string                 `json:"message"`   // 日志消息
	Fields    map[string]interface{} `json:"fields"`    // 日志附加字段

	// 位置信息字段（新增）
	FileName string `json:"file_name,omitempty"` // 文件名
	Line     int    `json:"line,omitempty"`      // 行号
	Function string `json:"function,omitempty"`  // 函数名
}

// LogEntry 调试日志条目结构
// 用于记录单个HTTP请求或进程的完整调试信息
type LogEntry struct {
	ID         string        `json:"id"`          // 日志唯一标识
	Timestamp  time.Time     `json:"timestamp"`   // 请求/进程开始时间戳
	Method     string        `json:"method"`      // HTTP方法（HTTP记录专用）
	URL        string        `json:"url"`         // 请求URL（HTTP记录专用）
	StatusCode int           `json:"status_code"` // HTTP状态码（HTTP记录专用）
	Duration   time.Duration `json:"duration"`    // 处理耗时
	ClientIP   string        `json:"client_ip"`   // 客户端IP（HTTP记录专用）
	UserAgent  string        `json:"user_agent"`  // 用户代理（HTTP记录专用）
	Host       string        `json:"host"`        // 请求域名（HTTP记录专用）
	RequestID  string        `json:"request_id"`  // 请求ID（用于追踪）

	// 记录类型标识
	RecordType string `json:"record_type" default:"http"` // 记录类型：http/process

	// 进程记录专用字段
	ProcessID   string    `json:"process_id,omitempty"`   // 进程唯一标识
	ProcessName string    `json:"process_name,omitempty"` // 进程名称
	ProcessType string    `json:"process_type,omitempty"` // 进程类型（background/worker/cron等）
	EndTime     time.Time `json:"end_time,omitempty"`     // 进程结束时间
	Status      string    `json:"status,omitempty"`       // 进程状态（running/completed/failed）

	// 请求信息（HTTP记录专用）
	RequestHeaders map[string]string `json:"request_headers"` // 请求头
	QueryParams    map[string]string `json:"query_params"`    // 查询参数
	RequestBody    string            `json:"request_body"`    // 请求体内容

	// 响应信息（HTTP记录专用）
	ResponseHeaders map[string]string `json:"response_headers"` // 响应头
	ResponseBody    string            `json:"response_body"`    // 响应体内容

	// 会话数据（可选）
	SessionData map[string]interface{} `json:"session_data,omitempty"` // 会话数据

	// 错误信息
	Error string `json:"error,omitempty"` // 错误信息

	// Logger日志信息（新增）
	LoggerLogs []LoggerLog `json:"logger_logs,omitempty"` // 通过logger记录的日志

	// 流式响应元数据（新增）
	IsStreamingResponse bool   `json:"is_streaming_response,omitempty"` // 是否为流式响应
	StreamingChunks     int    `json:"streaming_chunks,omitempty"`      // 流式响应分块数量
	StreamingChunkSize  int    `json:"streaming_chunk_size,omitempty"`  // 流式响应分块大小限制（字节）
	MaxStreamingChunks  int    `json:"max_streaming_chunks,omitempty"`  // 流式响应最大分块数量限制
	StreamingData       string `json:"streaming_data,omitempty"`        // 流式响应数据摘要（格式化显示）

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

	// 计算流式响应元数据大小（新增）
	totalSize += len(e.StreamingData)
	if e.IsStreamingResponse {
		// 流式响应相关的布尔值和整数字段占用固定大小
		totalSize += 32 // 布尔值和整数的大致存储大小
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

// ProcessLogger 进程级日志记录器
// 用于在进程中记录不同级别的debugger日志，所有日志属于同一条进程记录
// 该记录器适用于后台任务、批处理作业、定时任务等非HTTP进程场景
type ProcessLogger struct {
	debugger    *Debugger
	processID   string
	processName string
	processType string
	startTime   time.Time
	logger      *DefaultLogger
}

// ProcessLoggerInterface 进程级日志记录器接口
// 提供进程级日志记录功能，继承基础日志接口并添加进程管理方法
type ProcessLoggerInterface interface {
	LoggerInterface // 继承基础日志接口

	// GetProcessID 获取进程ID
	GetProcessID() string

	// GetProcessName 获取进程名称
	GetProcessName() string

	// GetProcessType 获取进程类型
	GetProcessType() string

	// SetProcessInfo 设置进程信息
	SetProcessInfo(info map[string]interface{})

	// EndProcess 结束进程记录
	EndProcess(status string) error
}

// NewProcessLogger 创建新的进程级日志记录器
// 该方法会创建进程记录并初始化进程级日志记录器，适用于需要监控的非HTTP进程
//
// 参数:
//
//	debugger: 调试器实例，用于存储和配置管理
//	processName: 进程名称，用于标识进程用途（如"数据同步任务"）
//	processType: 进程类型，可选值：background/worker/cron/batch等
//
// 返回值:
//
//	*ProcessLogger: 进程级日志记录器实例
//
// 示例:
//
//	logger := NewProcessLogger(dbg, "数据同步任务", "batch")
//	defer logger.EndProcess("completed")
func NewProcessLogger(debugger *Debugger, processName, processType string) *ProcessLogger {
	processID := GenerateID()

	// 创建默认日志记录器并进行类型断言
	defaultLogger := NewDefaultLogger(debugger).WithFields(map[string]interface{}{
		"process_id":   processID,
		"process_name": processName,
		"process_type": processType,
	})

	logger := &ProcessLogger{
		debugger:    debugger,
		processID:   processID,
		processName: processName,
		processType: processType,
		startTime:   time.Now(),
		logger:      defaultLogger.(*DefaultLogger),
	}

	// 创建进程记录条目
	entry := &LogEntry{
		ID:          processID,
		Timestamp:   logger.startTime,
		RecordType:  "process",
		ProcessID:   processID,
		ProcessName: processName,
		ProcessType: processType,
		Status:      "running",
	}

	// 保存初始进程记录
	if err := debugger.GetStorage().Save(entry); err != nil {
		fmt.Printf("创建进程记录失败: %v\n", err)
	}

	return logger
}

// Info 记录信息级别日志
// 记录进程执行的关键信息，适用于监控和状态跟踪
//
// 参数:
//
//	msg: 日志消息，可以是字符串、结构体或实现了Stringer接口的类型
//	fields: 可选的附加字段，用于记录额外的状态信息
//
// 示例:
//
//	logger.Info("数据同步完成", map[string]interface{}{"processed": 1000})
func (p *ProcessLogger) Info(msg any, fields ...map[string]interface{}) {
	p.logger.Info(msg, fields...)
}

// Warn 记录警告级别日志
// 记录可能影响进程正常执行的警告信息
//
// 参数:
//
//	msg: 日志消息，可以是字符串、结构体或实现了Stringer接口的类型
//	fields: 可选的附加字段，用于记录警告相关的上下文信息
//
// 示例:
//
//	logger.Warn("磁盘空间不足", map[string]interface{}{"available": "1GB"})
func (p *ProcessLogger) Warn(msg any, fields ...map[string]interface{}) {
	p.logger.Warn(msg, fields...)
}

// Error 记录错误级别日志
// 记录进程执行过程中发生的错误信息
//
// 参数:
//
//	msg: 日志消息，可以是字符串、结构体或实现了Stringer接口的类型
//	fields: 可选的附加字段，用于记录错误相关的详细信息
//
// 示例:
//
//	logger.Error("数据库连接失败", map[string]interface{}{"error": err.Error()})
func (p *ProcessLogger) Error(msg any, fields ...map[string]interface{}) {
	p.logger.Error(msg, fields...)
}

// WithFields 创建带有字段的日志记录器
// 创建一个新的日志记录器实例，继承当前记录器的所有字段并添加新字段
//
// 参数:
//
//	fields: 要添加的字段映射
//
// 返回值:
//
//	LoggerInterface: 新的日志记录器实例
//
// 示例:
//
//	subLogger := logger.WithFields(map[string]interface{}{"module": "data_processor"})
func (p *ProcessLogger) WithFields(fields map[string]interface{}) LoggerInterface {
	return p.logger.WithFields(fields)
}

// GetLevel 获取当前日志记录器的日志级别
// 返回当前调试器配置的日志级别
//
// 返回值:
//
//	string: 日志级别（debug/info/warn/error）
func (p *ProcessLogger) GetLevel() LogLevel {
	return p.logger.GetLevel()
}

// GetProcessID 获取进程ID
// 返回当前进程的唯一标识符，可用于后续查询和监控
//
// 返回值:
//
//	string: 进程ID
//
// 示例:
//
//	processID := logger.GetProcessID()
func (p *ProcessLogger) GetProcessID() string {
	return p.processID
}

// GetProcessName 获取进程名称
// 返回当前进程的名称，用于标识进程用途
//
// 返回值:
//
//	string: 进程名称
func (p *ProcessLogger) GetProcessName() string {
	return p.processName
}

// GetProcessType 获取进程类型
// 返回当前进程的类型，如background/worker/cron等
//
// 返回值:
//
//	string: 进程类型
func (p *ProcessLogger) GetProcessType() string {
	return p.processType
}

// SetProcessInfo 设置进程信息
// 动态更新进程的附加信息，适用于进程执行过程中需要记录额外上下文信息的场景
//
// 参数:
//
//	info: 进程信息映射，可以包含任意键值对
//
// 示例:
//
//	logger.SetProcessInfo(map[string]interface{}{"progress": "50%", "current_file": "data.csv"})
func (p *ProcessLogger) SetProcessInfo(info map[string]interface{}) {
	// 更新logger的字段
	for k, v := range info {
		p.logger.fields[k] = v
	}
}

// 进程状态常量定义
const (
	ProcessStatusCompleted = "completed" // 进程正常完成
	ProcessStatusFailed    = "failed"    // 进程执行失败
	ProcessStatusCancelled = "cancelled" // 进程被取消
)

// EndProcess 结束进程记录
// 结束当前进程记录，记录进程的结束时间、状态和所有收集的日志
// 该方法应在进程执行完成时调用，以确保记录完整的执行时间线
//
// 参数:
//
//	status: 进程结束状态，使用预定义的常量：ProcessStatusCompleted、ProcessStatusFailed、ProcessStatusCancelled
//
// 返回值:
//
//	error: 错误信息，当进程记录不存在或保存失败时返回错误
//
// 示例:
//
//	err := logger.EndProcess(ProcessStatusCompleted)
//	if err != nil {
//	    // 处理错误
//	}
func (p *ProcessLogger) EndProcess(status string) error {
	endTime := time.Now()

	// 获取进程记录
	entry, err := p.debugger.GetStorage().FindByID(p.processID)
	if err != nil {
		return fmt.Errorf("获取进程记录失败: %w", err)
	}

	// 更新进程记录
	entry.EndTime = endTime
	entry.Duration = endTime.Sub(p.startTime)
	entry.Status = status

	// 添加logger收集的日志
	entry.LoggerLogs = p.logger.GetLogs()

	// 保存更新后的记录
	if err := p.debugger.GetStorage().Save(entry); err != nil {
		return fmt.Errorf("保存进程记录失败: %w", err)
	}

	return nil
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
		Level:   LevelInfo,
	}
	// 设置默认值
	if err := helper.CheckAndSetDefault(config); err != nil {
		fmt.Printf("设置默认值失败: %v\n", err)
	}
	debugger, _ := New(config)

	var host string
	if c != nil && c.Request != nil {
		host = c.Request.Header.Get("X-Forwarded-Host")
		if host == "" {
			host = c.Request.Host
		}
	}
	return debugger.GetLogger().WithFields(map[string]interface{}{
		"context_error": "logger_not_found_in_context",
		"host":          host,
	})
}
