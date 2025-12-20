package debugger

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
)

type LogLevel int

// 日志级别常量
const (
	LevelSilent LogLevel = iota + 1 // 静默级别：不记录任何日志
	LevelError                      // 错误级别：只记录错误信息
	LevelWarn                       // 警告级别：记录错误+警告信息
	LevelInfo                       // 信息级别：记录所有详细信息
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
	mutex    sync.Mutex  // 用于保护并发访问的互斥锁
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

	l.mutex.Lock()
	for k, v := range l.fields {
		newFields[k] = v
	}
	logsCopy := make([]LoggerLog, len(l.logs))
	copy(logsCopy, l.logs)
	level := l.level
	l.mutex.Unlock()

	for k, v := range fields {
		newFields[k] = v
	}

	return &DefaultLogger{
		debugger: l.debugger,
		fields:   newFields,
		logs:     logsCopy, // 继承父logger的日志
		level:    level,    // 设置新的日志级别
	}
}

// GetLevel 获取当前日志记录器的日志级别
func (l *DefaultLogger) GetLevel() LogLevel {
	return l.level
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
	l.mutex.Lock()
	for k, v := range l.fields {
		allFields[k] = v
	}
	l.mutex.Unlock()

	// 添加调用方传入的字段
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			allFields[k] = v
		}
	}

	// 根据配置决定是否获取调用位置信息
	var fileName string
	var line int
	var function string
	if l.debugger.config.EnableCallerInfo {
		// 获取调用位置信息
		fileName, line, function = getCallerInfo()
	}

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

	l.mutex.Lock()
	l.logs = append(l.logs, loggerLog)
	l.mutex.Unlock()

	// 输出包含位置信息的日志（支持IDE可点击链接）
	// 格式：[级别] 文件名:行号 - 函数名: 消息内容
	// 注意：文件名:行号 格式是IDEA控制台可点击链接的标准格式
	if l.debugger.config.EnableCallerInfo {
		fmt.Printf("[%s] %s:%d - %s\n", level.String(), fileName, line, message)
	} else {
		fmt.Printf("[%s] %s\n", level.String(), message)
	}

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
	// 从调用栈中查找第一个非调试器/非GORM的调用位置
	// 跳过 runtime.Callers 和本函数自身
	pcs := make([]uintptr, 32)
	n := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:n])

	var found runtime.Frame
	for {
		f, more := frames.Next()
		file := f.File
		fn := f.Function

		// 过滤调试器自身和GORM内部调用栈
		if !(strings.Contains(file, "/component/debugger/logger.go") ||
			strings.Contains(file, "/component/debugger/debugger.go") ||
			strings.Contains(file, "/component/orm/gorm_logger.go") ||
			strings.Contains(fn, "gorm.io/gorm") ||
			strings.Contains(fn, "gorm.io/driver") ||
			strings.Contains(fn, "database/sql")) {
			found = f
			break
		}

		if !more {
			// 未找到合适的帧，使用最后一个可用帧作为回退
			found = f
			break
		}
	}

	// 保留完整的文件名路径，便于精确定位日志位置
	// 对于多个包中存在相同文件名的情况，可以更准确地区分
	fileName = found.File

	return fileName, found.Line, found.Function
}

// shouldLog 检查是否应该记录指定级别的日志
func (l *DefaultLogger) shouldLog(level LogLevel) bool {
	// 根据配置的日志级别决定是否记录
	switch l.GetLevel() {
	case LevelInfo:
		// 信息级别记录所有日志
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
	l.mutex.Lock()
	logsCopy := make([]LoggerLog, len(l.logs))
	copy(logsCopy, l.logs)
	l.mutex.Unlock()
	return logsCopy
}

// ClearLogs 清空收集的日志信息
func (l *DefaultLogger) ClearLogs() {
	l.mutex.Lock()
	l.logs = []LoggerLog{}
	l.mutex.Unlock()
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

// MainLogger 主进程日志记录器
// 用于记录应用程序主进程的日志，支持文件输出和日志分割
// 该记录器适用于需要持久化保存的应用程序级日志
// 支持按日期或大小分割日志文件
// 复用debugger的日志级别和调用位置信息配置

type MainLogger struct {
	debugger    *Debugger
	level       LogLevel               // 当前日志记录器的日志级别（复用debugger的Level配置）
	fields      map[string]interface{} // 固定字段
	mutex       sync.Mutex             // 用于保护并发访问的互斥锁
	logFile     *os.File               // 当前日志文件句柄
	logPath     string                 // 日志文件路径
	splitMode   string                 // 日志分割模式：size（按大小）、date（按日期）
	maxSize     int64                  // 日志文件最大大小（字节）
	maxBackups  int                    // 最大备份文件数量
	compress    bool                   // 是否压缩备份日志
	currentDate string                 // 当前日志日期，用于按日期分割
	currentSize int64                  // 当前日志文件大小
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

// NewMainLogger 创建新的主进程日志记录器
// 该方法会创建主进程日志记录器，适用于需要持久化保存的应用程序级日志
// 支持按日期或大小分割日志文件
//
// 参数:
//
//	debugger: 调试器实例，用于配置管理
//
// 返回值:
//
//	LoggerInterface: 主进程日志记录器实例
//
// 示例:
//
//	mainLogger := NewMainLogger(dbg)
//	mainLogger.Info("应用程序启动")
func NewMainLogger(debugger *Debugger) LoggerInterface {
	mainLogger := &MainLogger{
		debugger:    debugger,
		level:       debugger.config.Level, // 复用debugger的日志级别配置
		fields:      make(map[string]interface{}),
		logPath:     debugger.config.MainLogPath,
		splitMode:   debugger.config.MainLogSplitMode,
		maxSize:     debugger.config.MainLogMaxSize * 1024 * 1024, // 转换为字节
		maxBackups:  debugger.config.MainLogMaxBackups,
		compress:    debugger.config.MainLogCompress,
		currentDate: time.Now().Format("2006-01-02"),
		currentSize: 0,
	}

	// 初始化日志文件
	if err := mainLogger.initLogFile(); err != nil {
		fmt.Printf("初始化主进程日志文件失败: %v\n", err)
	}

	return mainLogger
}

// initLogFile 初始化日志文件
// 创建日志目录（如果不存在）
// 打开或创建日志文件
func (l *MainLogger) initLogFile() error {
	// 确保日志目录存在
	if err := os.MkdirAll(l.logPath, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 获取日志文件路径
	logFilePath := l.getLogFilePath()

	// 打开或创建日志文件
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("获取日志文件信息失败: %w", err)
	}

	l.logFile = file
	l.currentSize = fileInfo.Size()

	return nil
}

// getLogFilePath 获取当前日志文件路径
// 根据分割模式生成不同的文件名
func (l *MainLogger) getLogFilePath() string {
	baseName := "process.log"
	logFile := path.Join(l.logPath, baseName)

	// 按日期分割时，添加日期后缀
	if l.splitMode == "date" {
		logFile = fmt.Sprintf("%s.%s", logFile, l.currentDate)
	}

	return logFile
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
//	LogLevel: 日志级别（LevelInfo/LevelWarn/LevelError/LevelSilent）
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
// logger.SetProcessInfo(map[string]interface{}{"progress": "50%", "current_file": "data.csv"})
func (p *ProcessLogger) SetProcessInfo(info map[string]interface{}) {
	// 更新logger的字段
	p.logger.mutex.Lock()
	for k, v := range info {
		p.logger.fields[k] = v
	}
	p.logger.mutex.Unlock()
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

// -------------------- MainLogger 方法实现 --------------------

// Info 记录信息级别日志
func (l *MainLogger) Info(msg any, fields ...map[string]interface{}) {
	l.log(LevelInfo, msg, fields...)
}

// Warn 记录警告级别日志
func (l *MainLogger) Warn(msg any, fields ...map[string]interface{}) {
	l.log(LevelWarn, msg, fields...)
}

// Error 记录错误级别日志
func (l *MainLogger) Error(msg any, fields ...map[string]interface{}) {
	l.log(LevelError, msg, fields...)
}

// WithFields 创建带有字段的日志记录器
func (l *MainLogger) WithFields(fields map[string]interface{}) LoggerInterface {
	// 合并现有字段和新字段
	newFields := make(map[string]interface{})

	l.mutex.Lock()
	for k, v := range l.fields {
		newFields[k] = v
	}
	level := l.level
	l.mutex.Unlock()

	for k, v := range fields {
		newFields[k] = v
	}

	// 创建新的MainLogger实例
	newLogger := &MainLogger{
		debugger:    l.debugger,
		level:       level,
		fields:      newFields,
		logPath:     l.logPath,
		splitMode:   l.splitMode,
		maxSize:     l.maxSize,
		maxBackups:  l.maxBackups,
		compress:    l.compress,
		currentDate: l.currentDate,
		currentSize: 0,         // 新实例的当前大小为0
		logFile:     l.logFile, // 共享同一个日志文件句柄
	}

	return newLogger
}

// GetLevel 获取当前日志记录器的日志级别
func (l *MainLogger) GetLevel() LogLevel {
	return l.level
}

// log 内部日志记录方法
// - level: 日志级别（info/warn/error/silent）
// - msg: 日志消息（字符串、结构体、map、数组、实现了Stringer接口的类型等）
// - fields: 可选的附加字段（键值对）
func (l *MainLogger) log(level LogLevel, msg any, fields ...map[string]interface{}) {
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

	// 检查并执行日志分割
	l.checkRotate()

	// 根据配置决定是否获取调用位置信息
	var fileName string
	var line int
	var function string
	if l.debugger.config.EnableCallerInfo {
		// 获取调用位置信息
		fileName, line, function = getCallerInfo()
	}

	// 格式化日志
	logLine := l.formatLog(level, message, fileName, line, function)

	// 写入日志文件
	l.writeLog(logLine)

	// 同时输出到控制台
	if l.debugger.config.EnableCallerInfo {
		fmt.Printf("[%s] %s:%d - %s: %s\n", level.String(), fileName, line, function, message)
	} else {
		fmt.Printf("[%s] %s\n", level.String(), message)
	}
}

// formatLog 格式化日志输出
// 只支持文本格式，简化实现
func (l *MainLogger) formatLog(level LogLevel, message string, fileName string, line int, function string) string {
	// 文本格式
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	if l.debugger.config.EnableCallerInfo {
		return fmt.Sprintf("%s [%s] %s:%d - %s: %s\n", timestamp, level.String(), fileName, line, function, message)
	} else {
		return fmt.Sprintf("%s [%s] %s\n", timestamp, level.String(), message)
	}
}

// writeLog 写入日志到文件
func (l *MainLogger) writeLog(logLine string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.logFile == nil {
		// 如果日志文件未初始化，尝试重新初始化
		if err := l.initLogFile(); err != nil {
			fmt.Printf("写入日志时重新初始化日志文件失败: %v\n", err)
			return
		}
	}

	// 写入日志
	n, err := l.logFile.WriteString(logLine)
	if err != nil {
		fmt.Printf("写入日志文件失败: %v\n", err)
		return
	}

	// 更新当前文件大小
	l.currentSize += int64(n)
}

// checkRotate 检查是否需要进行日志分割
// 根据分割模式（日期或大小）检查是否需要创建新的日志文件
func (l *MainLogger) checkRotate() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 检查是否按日期分割
	if l.splitMode == "date" {
		currentDate := time.Now().Format("2006-01-02")
		if currentDate != l.currentDate {
			// 日期已变更，执行日志分割
			l.rotateLog()
			l.currentDate = currentDate
		}
		return
	}

	// 检查是否按大小分割
	if l.splitMode == "size" {
		if l.currentSize >= l.maxSize {
			// 文件大小已超过限制，执行日志分割
			l.rotateLog()
		}
	}
}

// rotateLog 执行日志分割
// 关闭当前日志文件，创建新的日志文件，并处理旧日志文件
func (l *MainLogger) rotateLog() {
	// 关闭当前日志文件
	if l.logFile != nil {
		l.logFile.Close()
	}

	// 处理旧日志文件（备份、压缩等）
	l.handleOldLogs()

	// 创建新的日志文件
	if err := l.initLogFile(); err != nil {
		fmt.Printf("创建新日志文件失败: %v\n", err)
	}

	// 重置当前文件大小
	l.currentSize = 0
}

// handleOldLogs 处理旧日志文件
// 根据配置进行备份、压缩等操作
func (l *MainLogger) handleOldLogs() {
	// 获取日志文件列表
	logFiles, err := l.getLogFiles()
	if err != nil {
		fmt.Printf("获取日志文件列表失败: %v\n", err)
		return
	}

	// 如果日志文件数量不超过最大备份数量，直接返回
	if len(logFiles) <= l.maxBackups {
		return
	}

	// 按修改时间排序，保留最新的l.maxBackups个文件
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].ModTime().Before(logFiles[j].ModTime())
	})

	// 删除多余的旧日志文件
	for i := 0; i < len(logFiles)-l.maxBackups; i++ {
		filePath := path.Join(l.logPath, logFiles[i].Name())
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("删除旧日志文件失败: %v\n", err)
		}
	}
}

// getLogFiles 获取日志文件列表
func (l *MainLogger) getLogFiles() ([]os.FileInfo, error) {
	// 打开日志目录
	dir, err := os.Open(l.logPath)
	if err != nil {
		return nil, fmt.Errorf("打开日志目录失败: %w", err)
	}
	defer dir.Close()

	// 读取目录内容
	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("读取日志目录失败: %w", err)
	}

	// 过滤出日志文件
	var logFiles []os.FileInfo
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "process.log") {
			logFiles = append(logFiles, file)
		}
	}

	return logFiles, nil
}

// shouldLog 检查是否应该记录指定级别的日志
func (l *MainLogger) shouldLog(level LogLevel) bool {
	// 根据配置的日志级别决定是否记录
	switch l.GetLevel() {
	case LevelInfo:
		// 信息级别记录所有日志
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

// noopLogger 空操作日志记录器实现
// 当上下文中没有Logger实例时返回此实现，避免创建不必要的资源
// 所有方法都是空操作，不会产生任何日志输出

type noopLogger struct{}

// Info 空操作实现
func (l noopLogger) Info(msg any, fields ...map[string]interface{}) {}

// Warn 空操作实现
func (l noopLogger) Warn(msg any, fields ...map[string]interface{}) {}

// Error 空操作实现
func (l noopLogger) Error(msg any, fields ...map[string]interface{}) {}

// WithFields 空操作实现，返回自身
func (l noopLogger) WithFields(fields map[string]interface{}) LoggerInterface {
	return l
}

// GetLevel 空操作实现，返回LevelSilent
func (l noopLogger) GetLevel() LogLevel {
	return LevelSilent
}

// GetLoggerFromContext 从Gin上下文中获取Logger实例
// 控制器可以通过此函数获取Logger来记录日志
func GetLoggerFromContext(c *gin.Context) LoggerInterface {
	if c == nil {
		return noopLogger{}
	}

	if logger, exists := c.Get("debugger_logger"); exists {
		if l, ok := logger.(LoggerInterface); ok {
			return l
		}
	}

	// 如果上下文中没有Logger，返回空操作日志记录器
	// 避免创建不必要的资源和潜在的内存泄漏
	return noopLogger{}
}
