package debugger

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/security"
	"github.com/jcbowen/jcbaseGo/middleware"
)

// idRand 包内独立的随机数生成器，用于 GenerateID 生成唯一标识的随机后缀。
// 使用 rand.New(rand.NewSource(...)) 显式播种，避免依赖全局 math/rand 状态：
// 全局源虽在 Go 1.20+ 自动播种，但若进程内其他代码调用已废弃的 rand.Seed 会影响此处随机性。
// rand.NewSource 返回的 *rand.Rand 非并发安全，GenerateID 会被中间件并发调用，故用互斥锁保护。
var (
	idRandMu sync.Mutex
	idRand   = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// Config 调试器配置结构
type Config struct {
	Enabled         bool          `json:"enabled" default:"true" preserve:"true"` // 是否启用调试器
	MaxBodySize     int64         `json:"max_body_size" default:"1024"`           // 最大请求/响应体大小（KB），默认1MB
	RetentionPeriod time.Duration `json:"retention_period" default:"168h"`        // 日志保留期限，默认7天
	// CleanupInterval 过期日志自动清理的执行间隔，默认24小时。
	// 仅在调用方显式调用 StartPeriodicCleanup 时生效；<=0 归一化为默认值。
	CleanupInterval time.Duration `json:"cleanup_interval" default:"24h"`
	Level           LogLevel      `json:"level" default:"3"`         // 日志级别：LevelWarn
	MaxRecords      int           `json:"max_records" default:"150"` // 最大记录数量，默认150条
	// EnableCallerInfo 是否启用调用位置信息（关闭可提高高并发场景下的性能）。
	// 与 AutoSkipStatic 同理使用 *bool 三态：nil（JSON 未配置）默认开启，显式 false 生效关闭，
	// 避免 bool + default:"true" 的默认值机制把显式设置的 false 强制改回 true 导致开关失效
	EnableCallerInfo *bool `json:"enable_caller_info"` // 是否启用调用位置信息；nil默认开启，显式false生效关闭

	// 服务标识配置
	ServiceName string `json:"service_name" default:""` // 服务标识，用于多服务共享同一存储时区分日志来源，默认空 = 不标识

	// 过滤配置
	SkipPaths   []string `json:"skip_paths" default:""`          // 跳过的路径（如静态文件："/static/,/favicon.ico"）
	SkipMethods []string `json:"skip_methods" default:"OPTIONS"` // 跳过的HTTP方法
	// AutoSkipStatic 是否自动跳过静态资源请求。
	// 使用 *bool 表达三态语义：nil（JSON 未配置）默认开启；显式 false 生效关闭。
	// 不用 bool + default:"true" 的原因：CheckAndSetDefault 的布尔默认值机制会把
	// 显式设置的 false 也强制改回 true（无法区分「未配置」与「显式关闭」），导致该开关失效；
	// 也不用 preserve:"true" 的原因：preserve 会把「JSON 未配置」（零值 false）同样按显式值保留，
	// 使未写该键的下游配置默认行为从开启静默变为关闭。
	AutoSkipStatic   *bool    `json:"auto_skip_static"`
	StaticExtensions []string `json:"static_extensions" default:".css,.js,.map,.png,.jpg,.jpeg,.gif,.svg,.ico,.woff,.woff2,.ttf,.eot,.otf,.webp,.txt,.xml,.pdf,.mp4,.mp3"` // 静态资源扩展名列表

	// 采样配置
	SampleRate float64 `json:"sample_rate" default:"1.0"` // 采样率（0-1之间），默认记录所有请求

	// 凭据掩码配置
	// HeaderMaskKeys 追加到内置默认头名集合（authorization、cookie 等）之后参与掩码，
	// 不提供覆盖能力：默认集合按「宁可多掩」取向固化，避免下游配置写错造成明文入库
	HeaderMaskKeys []string `json:"header_mask_keys" default:""`
	// BodyMaskKeys 追加到内置默认字段名集合（password、token、sms_code 等）之后，
	// 作用于查询参数与请求/响应体
	BodyMaskKeys []string `json:"body_mask_keys" default:""`
	// BodyMaskEnabled 是否启用凭据掩码。
	// 与 AutoSkipStatic 同理使用 *bool 三态：nil（JSON 未配置）默认开启，显式 false 生效关闭；
	// 关闭会让密码、验证码、令牌原文写入存储与页面，仅建议本地一次性排查时使用
	BodyMaskEnabled *bool `json:"body_mask_enabled"`

	// IP访问控制配置
	AllowedIPs []string `json:"allowed_ips" default:""` // 页面访问IP白名单（支持IP与CIDR）；为空时默认不注册页面路由
	// AllowEmptyWhitelist 白名单为空时是否仍然开放调试页面（默认 false = fail-closed）。
	// 调试页会输出全量请求日志，无白名单等于把凭据交给任意访问者；需要沿用旧行为的消费方显式置 true
	AllowEmptyWhitelist bool `json:"allow_empty_whitelist" default:"false" preserve:"true"`
	UseCDN              bool `json:"use_cdn" default:"false" preserve:"true"` // 是否使用CDN，仅影响记录（entry.ClientIP）的取IP方式，不参与页面鉴权

	// 流式请求配置
	EnableStreamingSupport bool  `json:"enable_streaming_support" default:"false" preserve:"true"` // 是否启用流式请求支持
	StreamingChunkSize     int64 `json:"streaming_chunk_size" default:"1024"`                      // 流式响应分块大小（KB），默认1MB，0表示无限制
	MaxStreamingChunks     int   `json:"max_streaming_chunks" default:"10"`                        // 最大流式响应分块数量，默认10个，0表示无限制
	MaxStreamingMemory     int64 `json:"max_streaming_memory" default:"10485760"`                  // 流式响应最大内存使用量（字节），默认10MB，0表示无限制

	// Multipart请求配置
	EnableMultipartSupport bool  `json:"enable_multipart_support" default:"true" preserve:"true"` // 是否启用multipart支持
	MultipartMaxPartSize   int64 `json:"multipart_max_part_size" default:"64"`                    // multipart单个部分最大大小（KB），默认64KB
	// MultipartSkipFiles 是否跳过文件内容记录，只记录元数据（filename / Content-Type）。
	// 默认开启：上传的文件字节既不该进数据库 LONGTEXT，也不该被调试日志带走；需要内容时显式置 false。
	// 与 AutoSkipStatic 同理使用 *bool 三态：preserve 标签下的「值与默认不同即视为显式提供」
	// 会把未配置的零值 false 当成显式关闭，导致 default:"true" 永远不生效，故只能用 *bool 表达默认开启
	MultipartSkipFiles *bool `json:"multipart_skip_files"`
	// MultipartPreserveState 是否保持 Gin 上下文状态，避免破坏后续中间件与业务处理。
	// 同样必须是 *bool 三态（nil = 默认开启）：关掉该开关会走 MultipartReader 直接消费原始请求体，
	// 之后 gin 的 ParseMultipartForm 返回 ErrMultipartAlreadyWritten，业务侧 c.FormFile 拿不到文件
	MultipartPreserveState *bool `json:"multipart_preserve_state"`

	// 中间件执行顺序配置
	MiddlewareOrder string `json:"middleware_order" default:"normal"` // 中间件执行顺序：normal（正常）、early（优先）、late（最后）

	// 核心组件配置 - 必须传入实例化的存储器
	Storage Storage         `json:"-"` // 存储实现（必须传入实例化的存储器）
	Logger  LoggerInterface `json:"-"` // 日志记录器（推荐直接传入实例化的日志记录器）

	// 主进程日志配置
	EnableMainLogger  bool   `json:"enable_main_logger" default:"false"`      // 是否启用主进程日志
	MainLogPath       string `json:"main_log_path" default:"./runtime/logs/"` // 主进程日志文件路径
	MainLogSplitMode  string `json:"main_log_split_mode" default:"size"`      // 日志分割模式：size（按大小）、date（按日期）
	MainLogMaxSize    int64  `json:"main_log_max_size" default:"100"`         // 日志文件最大大小（MB），按大小分割时有效
	MainLogMaxBackups int    `json:"main_log_max_backups" default:"7"`        // 最大备份文件数量
	MainLogCompress   bool   `json:"main_log_compress" default:"false"`       // 是否压缩备份日志
}

// Debugger 调试器主结构
type Debugger struct {
	config     *Config
	storage    Storage
	controller *Controller
	logger     LoggerInterface // 日志记录器实例
	mainLogger LoggerInterface // 主进程日志记录器实例

	// 进程日志记录器注册表：processID -> *ProcessLogger
	// 用于通过 Debugger.EndProcess（仅凭processID）结束时补充内存中已收集的日志，
	// 避免该路径下进程日志丢失；条目在任一路径的EndProcess时移除，
	// 从未结束的进程会滞留注册表，与存储中running状态的记录一致
	processLoggers  map[string]*ProcessLogger
	processLoggerMu sync.RWMutex

	// 自动清理协程状态：保证 StartPeriodicCleanup 幂等，并允许 stop 被重复调用
	cleanupMu      sync.Mutex
	cleanupRunning bool
	cleanupStop    func()

	// 凭据掩码键集合（小写归一化）：New 中一次性构建，避免每请求重复分配
	maskSetsOnce  sync.Once
	maskHeaderSet map[string]struct{}
	maskBodySet   map[string]struct{}
}

// defaultCleanupInterval StartPeriodicCleanup 在 CleanupInterval 缺失或非正时使用的兜底间隔
const defaultCleanupInterval = 24 * time.Hour

// New 创建新的调试器实例
// config: 调试器配置，如果没有传入存储实例，将使用默认的内存存储
func New(config *Config) (*Debugger, error) {
	if config == nil {
		return nil, fmt.Errorf("配置不能为nil，请提供有效的Config实例")
	}

	d := &Debugger{
		config: config,
	}

	// 使用CheckAndSetDefaultWithPreserveTag方法设置默认值，并保留关键布尔字段的用户显式设置
	if err := helper.CheckAndSetDefaultWithPreserveTag(d.config); err != nil {
		return nil, fmt.Errorf("设置配置默认值失败: %w", err)
	}

	// *bool 三态字段归一化：nil（JSON 未配置）→ 默认开启，显式 false 保持不变。
	// 目的：1) 保证后续读取（shouldSkip / isCallerInfoEnabled 等）无需判空；
	// 2) 配置导出（JSON）时语义明确为 true/false 而非 null。
	// *bool 字段不会被 CheckAndSetDefault 自动填充默认值（nil 指针不实例化的约定），
	// 故必须在 New 中显式兜底，确保所有经 New 构造的实例这些字段始终非 nil。
	if d.config.AutoSkipStatic == nil {
		autoSkipStatic := true
		d.config.AutoSkipStatic = &autoSkipStatic
	}
	if d.config.EnableCallerInfo == nil {
		callerInfo := true
		d.config.EnableCallerInfo = &callerInfo
	}
	if d.config.BodyMaskEnabled == nil {
		bodyMask := true
		d.config.BodyMaskEnabled = &bodyMask
	}
	if d.config.MultipartSkipFiles == nil {
		skipFiles := true
		d.config.MultipartSkipFiles = &skipFiles
	}
	if d.config.MultipartPreserveState == nil {
		preserveState := true
		d.config.MultipartPreserveState = &preserveState
	}

	// 构建掩码键集合（sync.Once，后续请求只读）
	d.ensureMaskSets()

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
		d.logger = NewDefaultLogger(d)
	}

	// 数据库存储器的裁剪错误经实例 logger 记录，避免直接写 stdout；
	// 消费方仍可通过 (*DatabaseStorage).SetLogger 覆盖
	if dbStore, ok := d.storage.(*DatabaseStorage); ok {
		dbStore.SetLogger(d.logger)
	}

	// 初始化主进程日志记录器
	if d.config.EnableMainLogger {
		d.mainLogger = NewMainLogger(d)
	}

	return d, nil
}

// StartPeriodicCleanup 启动过期日志的自动清理：先立即执行一次（覆盖长期停机造成的积压），
// 之后按 CleanupInterval 周期执行；返回的 stop 用于停止协程并释放定时器。
//
// 参数：无（全部取实例配置）
// 返回值：
//   - stop: 停止清理协程的函数，可安全重复调用；未启动清理时为空操作
//
// 异常：不触发 panic；单次清理失败仅经实例 logger 记录，不影响后续周期
// 使用示例：
//
//	stop := dbg.StartPeriodicCleanup()
//	defer stop()
//
// 语义约定：RetentionPeriod <= 0 表示「调用方不希望按时间清理」，此时不启动任何协程。
// 过期清理的删除条件是 timestamp < now-RetentionPeriod，保留期为 0 会退化成「删除全表」，
// 必须显式拒绝而不是静默执行。注意 New 会经 CheckAndSetDefault 把零值 RetentionPeriod
// 归一化为默认 168h，所以「不清理」的权威表达是调用方不启动本方法（如业务侧
// retention_days<=0 时跳过调用），本函数的判空用于防御绕过 New 直接构造的实例。
// 幂等：重复调用返回同一个 stop，不会启动第二个协程。
func (d *Debugger) StartPeriodicCleanup() (stop func()) {
	d.cleanupMu.Lock()
	defer d.cleanupMu.Unlock()

	if d.cleanupRunning {
		return d.cleanupStop
	}

	retention := d.config.RetentionPeriod
	if retention <= 0 {
		return func() {}
	}

	interval := d.config.CleanupInterval
	if interval <= 0 {
		interval = defaultCleanupInterval
	}

	done := make(chan struct{})
	stopCh := make(chan struct{})
	stopped := false

	go func() {
		defer close(done)

		d.cleanupOnce(retention)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				d.cleanupOnce(retention)
			}
		}
	}()

	d.cleanupRunning = true
	d.cleanupStop = func() {
		d.cleanupMu.Lock()
		if stopped {
			d.cleanupMu.Unlock()
			return
		}
		stopped = true
		d.cleanupRunning = false
		close(stopCh)
		d.cleanupMu.Unlock()

		<-done
	}
	return d.cleanupStop
}

// cleanupOnce 执行一次过期清理，删除 retention 之前的记录
// 参数：
//   - retention: 日志保留时长，必须为正值
//
// 返回值：无（失败仅记录，交由下一周期重试）
// 异常：不触发 panic
func (d *Debugger) cleanupOnce(retention time.Duration) {
	before := time.Now().Add(-retention)
	if err := d.storage.Cleanup(before); err != nil {
		d.logger.Error("debugger 过期日志清理失败: " + err.Error())
		return
	}
}

// Middleware 创建Gin中间件
// 用于拦截HTTP请求并记录调试信息
func (d *Debugger) Middleware() gin.HandlerFunc {
	return d.createMiddlewareHandler()
}

// createMiddlewareHandler 创建中间件处理函数
// 根据配置的执行顺序提供不同的处理逻辑
func (d *Debugger) createMiddlewareHandler() gin.HandlerFunc {
	switch d.config.MiddlewareOrder {
	case "early":
		return d.createEarlyMiddleware()
	case "late":
		return d.createLateMiddleware()
	default:
		return d.createNormalMiddleware()
	}
}

// createNormalMiddleware 创建标准中间件处理函数
// 在正常位置执行，适用于大多数场景
func (d *Debugger) createNormalMiddleware() gin.HandlerFunc {
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
		// 计算主机信息（优先使用X-Forwarded-Host）
		host := c.Request.Header.Get("X-Forwarded-Host")
		if host == "" {
			host = c.Request.Host
		}

		// 构造请求级Logger：固定携带 request_id 等字段，请求结束时由 recordLoggerLogs 统一收集
		requestLogger := d.logger.WithFields(map[string]interface{}{
			"request_id": entry.ID,
			"method":     c.Request.Method,
			"url":        c.Request.URL.String(),
			"host":       host,
			"client_ip":  middleware.GetRealIP(c, d.config.UseCDN),
		})

		// 设置Logger到 gin 上下文，供控制器通过 GetLoggerFromContext 获取
		c.Set("debugger_logger", requestLogger)

		// 同时注入 request 的标准 context，使 ORM 等拿不到 gin.Context 的组件
		// 也能沿调用链取到请求级Logger，业务代码无需修改 WithContext 调用
		c.Request = c.Request.WithContext(ContextWithLogger(c.Request.Context(), requestLogger))

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

		// 恢复multipart请求体状态（如果适用）
		d.restoreMultipartRequestBody(c)

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

// createEarlyMiddleware 创建优先执行的中间件处理函数
// 在其他中间件之前执行，适用于需要记录完整请求信息的场景
func (d *Debugger) createEarlyMiddleware() gin.HandlerFunc {
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
		// 构造请求级Logger：固定携带 request_id 等字段，请求结束时由 recordLoggerLogs 统一收集
		requestLogger := d.logger.WithFields(map[string]interface{}{
			"request_id": entry.ID,
			"method":     c.Request.Method,
			"url":        c.Request.URL.String(),
			"client_ip":  middleware.GetRealIP(c, d.config.UseCDN),
		})

		// 设置Logger到 gin 上下文，供控制器通过 GetLoggerFromContext 获取
		c.Set("debugger_logger", requestLogger)

		// 同时注入 request 的标准 context，使 ORM 等拿不到 gin.Context 的组件
		// 也能沿调用链取到请求级Logger，业务代码无需修改 WithContext 调用
		c.Request = c.Request.WithContext(ContextWithLogger(c.Request.Context(), requestLogger))

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

		// 恢复multipart请求体状态（如果适用）
		d.restoreMultipartRequestBody(c)

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

// createLateMiddleware 创建最后执行的中间件处理函数
// 在其他中间件之后执行，适用于需要记录完整响应信息的场景
func (d *Debugger) createLateMiddleware() gin.HandlerFunc {
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
		// 构造请求级Logger：固定携带 request_id 等字段，请求结束时由 recordLoggerLogs 统一收集
		requestLogger := d.logger.WithFields(map[string]interface{}{
			"request_id": entry.ID,
			"method":     c.Request.Method,
			"url":        c.Request.URL.String(),
			"client_ip":  middleware.GetRealIP(c, d.config.UseCDN),
		})

		// 设置Logger到 gin 上下文，供控制器通过 GetLoggerFromContext 获取
		c.Set("debugger_logger", requestLogger)

		// 同时注入 request 的标准 context，使 ORM 等拿不到 gin.Context 的组件
		// 也能沿调用链取到请求级Logger，业务代码无需修改 WithContext 调用
		c.Request = c.Request.WithContext(ContextWithLogger(c.Request.Context(), requestLogger))

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

		// 恢复multipart请求体状态（如果适用）
		d.restoreMultipartRequestBody(c)

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

// restoreMultipartRequestBody 恢复multipart请求体状态
// 当debugger处理multipart请求后，确保后续中间件能正常访问请求体
func (d *Debugger) restoreMultipartRequestBody(c *gin.Context) {
	// 只有在启用multipart支持且需要保持状态时才进行恢复
	if !d.config.EnableMultipartSupport || !d.preserveMultipartState() {
		return
	}

	// 检查是否为multipart请求
	if !isMultipartFormData(c) {
		return
	}

	// 重新设置multipart reader，确保Gin能正确解析
	contentType := c.Request.Header.Get("Content-Type")
	boundary := extractBoundary(contentType)
	if boundary != "" {
		// 重置Gin的multipart状态
		c.Request.MultipartForm = nil
		// 强制Gin重新解析multipart表单
		c.Request.ParseMultipartForm(32 << 20) // 32MB
	}
}

// createLogEntry 创建日志条目
func (d *Debugger) createLogEntry(c *gin.Context, startTime time.Time) *LogEntry {
	entry := &LogEntry{
		ID:        GenerateID(),
		Timestamp: startTime,
		Method:    c.Request.Method,
		URL:       c.Request.URL.String(),
		Host: func() string {
			h := c.Request.Header.Get("X-Forwarded-Host")
			if h == "" {
				h = c.Request.Host
			}
			return h
		}(),
		ClientIP:       middleware.GetRealIP(c, d.config.UseCDN),
		UserAgent:      c.Request.UserAgent(),
		RequestID:      c.GetHeader("X-Request-ID"),
		ServiceName:    d.config.ServiceName, // 携带服务标识（空值时为不标识）
		RecordType:     "http",               // 设置记录类型为HTTP
		RequestHeaders: helper.ExtractHeaders(c.Request.Header),
		QueryParams:    extractQueryParams(c.Request.URL.Query()),
	}

	// 凭据掩码统一在条目生成处执行，保证存储、页面、导出三条下游路径都拿不到原文
	entry.RequestHeaders = d.maskHeaders(entry.RequestHeaders)
	entry.QueryParams = d.maskQueryParams(entry.QueryParams)

	// 记录请求体；掩码已在提取路径内完成（有界正文必须先掩码再追加截断说明）
	body, _ := d.extractRequestBody(c) // 忽略错误，因为extractRequestBody已经处理了错误情况
	entry.RequestBody = body

	return entry
}

// recordResponseInfo 记录响应信息
// 自动识别二进制响应数据并适当处理，避免文件下载内容显示为乱码
// 支持流式响应记录，对实时流式请求进行特殊处理
func (d *Debugger) recordResponseInfo(entry *LogEntry, writer *responseWriter, startTime time.Time) {
	entry.StatusCode = writer.Status()
	entry.Duration = time.Since(startTime)
	entry.ResponseHeaders = d.maskHeaders(helper.ExtractHeaders(writer.Header()))

	// 检查是否为流式响应（只有在启用流式支持时才处理流式响应）
	if d.config.EnableStreamingSupport && writer.isStreaming {
		// 流式响应处理
		entry.RecordType = "streaming" // 标记为流式请求
		entry.ResponseBody = d.maskBody(d.formatStreamingResponse(writer), writer.Header().Get("Content-Type"))

		// 记录流式响应元数据
		entry.IsStreamingResponse = true
		entry.StreamingChunks = len(writer.streamingChunks)
		entry.StreamingChunkSize = int(writer.chunkSizeLimit)
		entry.MaxStreamingChunks = writer.maxChunks
		entry.StreamingData = fmt.Sprintf("Streaming Response: %d chunks, total size: %d bytes",
			len(writer.streamingChunks), d.calculateTotalStreamingSize(writer))
	} else {
		// 非流式响应处理
		entry.RecordType = "http" // 标记为普通HTTP请求

		// 记录响应体（捕获窗口已在 responseWriter.Write 中按上限截断）
		if writer.body.Len() > 0 {
			limit := d.bodyCaptureLimit()
			contentType := writer.Header().Get("Content-Type")

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
					// 解压同样限量：捕获窗口本身可能被截断，整包解压会放大内存占用
					decompressed, err := io.ReadAll(io.LimitReader(reader, limit+1))
					if err == nil || len(decompressed) > 0 {
						// 检查解压后的数据是否为二进制
						if isBinaryData(decompressed) {
							entry.ResponseBody = capBodyText(formatBinaryData(decompressed), int64(len(decompressed)), limit)
						} else {
							entry.ResponseBody = d.maskAndCapText(string(decompressed), int64(len(decompressed)), limit, contentType)
						}
					} else {
						// 解压缩失败，记录原始数据并添加错误标记
						entry.ResponseBody = "[GZIP解压缩失败] " + formatBinaryData(writer.body.Bytes())
					}
				} else {
					// 创建gzip读取器失败（含捕获窗口截断导致流不完整），记录原始数据
					entry.ResponseBody = "[GZIP格式错误或已被截断] " + formatBinaryData(writer.body.Bytes())
				}
			} else {
				// 非gzip压缩的响应，检查是否为二进制数据
				responseBytes := writer.body.Bytes()
				if isBinaryData(responseBytes) {
					entry.ResponseBody = capBodyText(formatBinaryData(responseBytes), writer.totalSize, limit)
				} else {
					entry.ResponseBody = d.maskAndCapText(string(responseBytes), writer.totalSize, limit, contentType)
				}
			}
		}
	}
}

// maskAndCapText 响应/请求正文的统一出口：先掩码再按上限截断
// 顺序不可颠倒：capBodyText 会在正文末尾追加截断说明，
// 被截半的凭据值只有在位于文本末尾时才命中尾部掩码规则，说明行先附上去就匹配不到
// 参数：
//   - text: 原始正文
//   - total: 原始总字节数
//   - limit: 捕获上限字节数
//   - contentType: 对应的 Content-Type
//
// 返回值：
//   - string: 掩码并按需截断后的正文
//
// 异常：不触发 panic
func (d *Debugger) maskAndCapText(text string, total, limit int64, contentType string) string {
	return capBodyText(d.maskBody(text, contentType), total, limit)
}

// capBodyText 按捕获上限裁剪展示用正文，并在发生截断时补充大小说明
// 参数：
//   - text: 待裁剪的正文
//   - total: 原始总字节数（响应取 writer.totalSize，解压场景取解压后长度）
//   - limit: 上限字节数
//
// 返回值：
//   - string: 长度不超过 limit 的正文；截断时末尾附 [Body truncated: ...]
//
// 异常：不触发 panic
// 使用示例：
//
//	entry.ResponseBody = capBodyText(string(bodyBytes), total, d.bodyCaptureLimit())
func capBodyText(text string, total int64, limit int64) string {
	if int64(len(text)) <= limit && total <= limit {
		return text
	}

	captured := int64(len(text))
	if captured > limit {
		captured = limit
	}

	// 按字节上限切分可能把多字节字符截半，产生非法 UTF-8：MySQL 严格模式下
	// 连 LONGTEXT 列都会以 Incorrect string value 拒绝整条写入，故回退到最近的完整 rune 边界
	text = trimToRuneBoundary(text[:captured])
	captured = int64(len(text))

	if total < captured {
		total = captured
	}

	return fmt.Sprintf("%s\n[Body truncated: captured %d of %d bytes, max_body_size_kb=%d]",
		text, captured, total, limit/1024)
}

// trimToRuneBoundary 回退掉字符串末尾被截断的不完整 UTF-8 序列
// 参数：
//   - text: 已按字节上限切过的文本
//
// 返回值：
//   - string: 末尾不含残缺 rune 的文本
//
// 异常：不触发 panic
// 使用示例：
//
//	safe := trimToRuneBoundary(body[:maxBytes])
func trimToRuneBoundary(text string) string {
	// 单个 rune 最长 4 字节，最多回退 3 次即可越过截半的多字节序列
	for i := 0; i < utf8.UTFMax-1 && len(text) > 0; i++ {
		if r, size := utf8.DecodeLastRuneInString(text); r != utf8.RuneError || size > 1 {
			break
		}
		text = text[:len(text)-1]
	}
	return text
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
		// 使用接口断言而非具体类型断言，因为logger可能通过接口返回
		if logger, ok := loggerValue.(LoggerInterface); ok {
			// 获取logger收集的所有日志
			entry.LoggerLogs = logger.GetLogs()
			// 清空logger的日志记录，避免内存泄漏
			logger.ClearLogs()
		}
	}
}

// saveLogEntry 保存日志条目
// 保存前统计派生严重度字段，确保列表中可直观识别含error日志的记录
func (d *Debugger) saveLogEntry(entry *LogEntry) {
	entry.UpdateSeverity()
	if err := d.storage.Save(entry); err != nil {
		// 记录保存错误，但不影响正常请求处理
		fmt.Printf("保存调试日志失败: %v\n", err)
	}
}

// skipMultipartFileContent 返回 multipart 文件部件是否只记元数据
// MultipartSkipFiles 为 *bool 三态（nil = 未配置，默认跳过；New 已归一化为非 nil，
// 判空仅为防御绕过 New 直接构造的实例），显式 false 才记录文件内容
func (d *Debugger) skipMultipartFileContent() bool {
	return d.config.MultipartSkipFiles == nil || *d.config.MultipartSkipFiles
}

// preserveMultipartState 返回是否走「复制请求体解析 multipart」的安全路径
// MultipartPreserveState 同样为 *bool 三态（nil = 默认开启），显式 false 才回退到旧路径
func (d *Debugger) preserveMultipartState() bool {
	return d.config.MultipartPreserveState == nil || *d.config.MultipartPreserveState
}

// isCallerInfoEnabled 返回是否启用调用位置信息
// EnableCallerInfo 为 *bool 三态字段（nil = 未配置，默认开启；New 已归一化为非 nil，
// 判空仅防御绕过 New 直接构造的实例），显式 false 生效关闭
func (d *Debugger) isCallerInfoEnabled() bool {
	return d.config.EnableCallerInfo != nil && *d.config.EnableCallerInfo
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

	// 自动跳过静态资源请求（New 已将 nil 归一化为 &true，判空仅为防御直接构造的实例）
	if d.config.AutoSkipStatic != nil && *d.config.AutoSkipStatic && d.isStaticRequest(c) {
		return true
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

// isStaticRequest 检查是否为静态资源请求
// 函数名：isStaticRequest
// 参数：
// - c *gin.Context — Gin上下文
// 返回值：
// - bool — 是静态资源请求返回 true，否则返回 false
// 异常：不触发 panic
// 使用示例：
//
//	if d.isStaticRequest(c) { return }
//
// 说明：通过目录提示和扩展名判断静态资源，仅针对 GET/HEAD 方法
func (d *Debugger) isStaticRequest(c *gin.Context) bool {
	p := c.Request.URL.Path
	m := c.Request.Method
	if m != http.MethodGet && m != http.MethodHead {
		return false
	}
	// for _, hint := range []string{"/static/", "/assets/", "/public/", "/uploads/", "/favicon.ico"} {
	// 	if strings.Contains(p, hint) || strings.HasSuffix(p, hint) {
	// 		return true
	// 	}
	// }
	ext := strings.ToLower(path.Ext(p))
	if ext == "" {
		return false
	}
	for _, e := range d.config.StaticExtensions {
		if strings.ToLower(e) == ext {
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

// defaultMaxBodySizeKB 与 Config.MaxBodySize 的 default 标签保持一致（KB），
// 用于调用方传入非正值时的兜底，避免上限为 0 导致完全不记录请求体
const defaultMaxBodySizeKB = 1024

// defaultMultipartMaxPartSizeKB 与 Config.MultipartMaxPartSize 的 default 标签保持一致（KB），
// 用于单个 multipart 部件读取量的兜底上限
const defaultMultipartMaxPartSizeKB = 64

// bodyCaptureLimit 返回请求/响应体的捕获上限（字节）
// 参数：无（取实例配置，MaxBodySize 单位为 KB）
// 返回值：
//   - int64: 上限字节数，非正配置时回退为默认 1MB
//
// 异常：不触发 panic
func (d *Debugger) bodyCaptureLimit() int64 {
	if d.config.MaxBodySize <= 0 {
		return defaultMaxBodySizeKB * 1024
	}
	return d.config.MaxBodySize * 1024
}

// readRequestBodyBounded 有界读取请求体
// 最多读 limit+1 字节用于判断是否超限，其余部分不落地；已读片段与未读部分通过 MultiReader
// 原样回灌给后续处理链，因此大文件上传、批量导入不会把整份请求体驻留在内存
// 参数：
//   - c: gin 上下文，函数会把 c.Request.Body 替换为回灌后的读取器
//
// 返回值：
//   - []byte: 捕获到的字节，长度不超过上限
//   - bool: 是否发生截断（真实请求体大于上限）
//   - error: 读取错误；出错时 Body 被替换为空读取器，避免后续处理挂起
//
// 异常：不触发 panic
// 使用示例：
//
//	captured, truncated, err := d.readRequestBodyBounded(c)
func (d *Debugger) readRequestBodyBounded(c *gin.Context) ([]byte, bool, error) {
	limit := d.bodyCaptureLimit()

	// 多读 1 字节用于区分「正好等于上限」与「超过上限」；
	// 缓冲按实际内容增长，不像定长数组那样每请求预留一整块上限大小的内存
	buf := &bytes.Buffer{}
	if _, err := buf.ReadFrom(io.LimitReader(c.Request.Body, limit+1)); err != nil {
		c.Request.Body = io.NopCloser(bytes.NewReader(nil))
		return nil, false, err
	}

	read := int64(buf.Len())
	truncated := read > limit

	captured := buf.Bytes()
	if truncated {
		captured = captured[:limit]
	}

	// 已读出的字节要整体回灌，后续中间件与业务处理器才能读到完整请求体；
	// 未读到的剩余部分仍是原 Body 的当前位置
	c.Request.Body = io.NopCloser(io.MultiReader(bytes.NewReader(buf.Bytes()), c.Request.Body))

	return captured, truncated, nil
}

// extractRequestBody 提取请求体内容
// 优化性能：实现流式处理，避免大文件上传时的内存压力
// 改进：避免破坏Gin上下文状态，特别是multipart请求的处理
func (d *Debugger) extractRequestBody(c *gin.Context) (string, error) {
	if c.Request.Body == nil {
		return "", nil
	}

	// 检查是否为multipart/form-data文件上传
	if isMultipartFormData(c) {
		// 检查是否启用multipart支持
		if !d.config.EnableMultipartSupport {
			return "[Multipart request - disabled]", nil
		}

		contentType := c.Request.Header.Get("Content-Type")
		// 对于multipart/form-data，使用安全的流式处理
		text, err := d.extractMultipartRequestBodySafe(c)
		if err != nil {
			return "", err
		}
		// 掩码在提取路径内完成：有界正文必须先掩码再追加截断说明，
		// 若在条目生成处统一重掩，说明行会被当成字段值一起掩掉
		return d.maskBody(text, contentType), nil
	}

	// 对于其他类型，使用原有逻辑但添加大小限制
	return d.extractBoundedBody(c)
}

// extractBoundedBody 有界提取非 multipart 的请求体
// 捕获上限由 bodyCaptureLimit 决定，超限只记录捕获量与上限值，不把超出部分带入日志或数据库
// 参数：
//   - c: gin 上下文（请求体在读取后原样回灌）
//
// 返回值：
//   - string: 用于展示的请求体内容或截断/错误标记
//   - error: 恒为 nil，读取失败以标记文本形式返回，便于 createLogEntry 记录
//
// 异常：不触发 panic
// 使用示例：
//
//	body, _ := d.extractBoundedBody(c)
func (d *Debugger) extractBoundedBody(c *gin.Context) (string, error) {
	bodyBytes, truncated, err := d.readRequestBodyBounded(c)
	if err != nil {
		// 返回错误信息，而不是空字符串，这样createLogEntry可以记录错误情况
		return fmt.Sprintf("[Error reading request body: %v]", err), nil
	}

	contentType := c.Request.Header.Get("Content-Type")

	text := string(bodyBytes)
	if isBinaryData(bodyBytes) {
		// 二进制数据（文件内容等）转为可读摘要，不含凭据文本
		text = formatBinaryData(bodyBytes)
	} else {
		// 掩码必须在追加截断说明之前完成：被截半的凭据值只有位于文本末尾时才会命中尾部正则，
		// 说明行一旦先附上去就再也匹配不到
		text = d.maskBody(text, contentType)
	}

	if !truncated {
		return text, nil
	}

	// 超限部分不入库，保留上限内正文并附截断说明，避免 LONGTEXT 无界增长。
	// 不整段丢弃正文：刚好越过上限的请求体若只剩一行说明，日志会失去排查价值
	// captured 报告原始捕获字节数而非掩码后的文本长度：掩码把命中值替换为 ***
	// 会缩短文本，用 len(text) 会低估实际捕获量
	text = trimToRuneBoundary(text)
	return fmt.Sprintf("%s\n[Body truncated: captured %d bytes, max_body_size=%d KB]",
		text, len(bodyBytes), d.config.MaxBodySize), nil
}

// extractMultipartRequestBody 流式提取multipart/form-data请求体
// 避免大文件上传时的内存压力，只读取元数据不读取文件内容
func (d *Debugger) extractMultipartRequestBody(c *gin.Context) (string, error) {
	contentType := c.Request.Header.Get("Content-Type")

	// 提取boundary参数
	boundary := extractBoundary(contentType)
	if boundary == "" {
		// 如果没有boundary，回退到二进制数据显示
		return d.extractRequestBodyWithSizeLimit(c)
	}

	// 创建multipart reader进行流式处理
	mr, err := c.Request.MultipartReader()
	if err != nil {
		// 如果创建失败，回退到普通处理
		return d.extractRequestBodyWithSizeLimit(c)
	}

	var result strings.Builder
	result.WriteString("[Multipart Form Data - Stream Processing]\n")

	partCount := 0
	fileCount := 0
	totalSize := 0

	// 流式处理multipart的各个部分
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

		// 文件部件默认只记元数据：连内容都不读取，避免上传字节进入调试日志与数据库
		if fileName != "" && d.skipMultipartFileContent() {
			fileCount++
			result.WriteString(fmt.Sprintf("  [File] Name: %s, Field: %s, Content-Type: %s (content not recorded)\n",
				fileName, partName, part.Header.Get("Content-Type")))
			continue
		}

		// 读取部分内容（限制大小，避免大文件内存占用）
		maxPartSize := int(d.config.MultipartMaxPartSize * 1024)
		if maxPartSize <= 0 {
			maxPartSize = int(defaultMultipartMaxPartSizeKB * 1024)
		}
		partBytes := make([]byte, maxPartSize)
		n, err := part.Read(partBytes)
		if err != nil && err != io.EOF {
			result.WriteString(fmt.Sprintf("  Error reading part content: %v\n", err))
			continue
		}

		totalSize += n

		if fileName != "" {
			// 文件上传部分
			fileCount++
			result.WriteString(fmt.Sprintf("  [File] Name: %s, Field: %s, Size: %d bytes, Type: %s\n",
				fileName, partName, n, detectFileType(partBytes[:n])))
		} else {
			// 文本字段部分
			if isBinaryData(partBytes[:n]) {
				result.WriteString(fmt.Sprintf("  [Field] Name: %s, Size: %d bytes, Type: Binary\n",
					partName, n))
			} else {
				// 限制文本长度，避免显示过长
				textContent := string(partBytes[:n])
				if len(textContent) > 100 {
					textContent = textContent[:100] + "..."
				}
				result.WriteString(fmt.Sprintf("  [Field] Name: %s, Size: %d bytes, Content: %s\n",
					partName, n, textContent))
			}
		}

		// 检查总大小是否超过限制
		if totalSize > int(d.config.MaxBodySize*1024) {
			result.WriteString(fmt.Sprintf("  [Truncated] Total size exceeds limit: %d bytes\n", totalSize))
			break
		}
	}

	result.WriteString(fmt.Sprintf("Total Parts: %d, Files: %d, Total Size: %d bytes\n", partCount, fileCount, totalSize))
	return result.String(), nil
}

// extractMultipartRequestBodySafe 安全的multipart请求体提取方法
// 避免破坏Gin上下文状态，通过复制请求体进行处理
func (d *Debugger) extractMultipartRequestBodySafe(c *gin.Context) (string, error) {
	contentType := c.Request.Header.Get("Content-Type")

	// 提取boundary参数
	boundary := extractBoundary(contentType)
	if boundary == "" {
		// 如果没有boundary，回退到安全的二进制数据显示
		return d.extractRequestBodyWithSizeLimitSafe(c)
	}

	// 检查是否保持Gin上下文状态
	if d.preserveMultipartState() {
		// 方法1：通过复制请求体进行处理，避免影响原始请求
		return d.extractMultipartWithBodyCopy(c, boundary)
	} else {
		// 方法2：使用原始方法（不推荐，但保持向后兼容）
		return d.extractMultipartRequestBody(c)
	}
}

// extractMultipartWithBodyCopy 通过复制请求体进行multipart处理
// 避免破坏Gin上下文状态
// 只解析捕获窗口内的部件：上限由 bodyCaptureLimit 决定，
// 大文件上传不会把整份请求体读进内存，也不会把文件字节写进 stdout 日志
func (d *Debugger) extractMultipartWithBodyCopy(c *gin.Context, boundary string) (string, error) {
	// 有界读取请求体（已读部分会回灌给后续处理链）
	bodyBytes, truncated, err := d.readRequestBodyBounded(c)
	if err != nil {
		return "", err
	}

	// 使用复制的请求体进行multipart解析
	reader := bytes.NewReader(bodyBytes)
	mr := multipart.NewReader(reader, boundary)
	if mr == nil {
		return "[Multipart parse error]", nil
	}

	var result strings.Builder
	result.WriteString("[Multipart Form Data - Safe Processing]\n")
	if truncated {
		// 只解析捕获窗口内的部件，其余部分不读取也不入库
		result.WriteString(fmt.Sprintf("  [Truncated] body exceeds max_body_size=%d KB, parts after the capture window are not recorded\n",
			d.config.MaxBodySize))
	}

	partCount := 0
	fileCount := 0
	totalSize := 0
	maxPartSize := int(d.config.MultipartMaxPartSize * 1024)
	if maxPartSize <= 0 {
		maxPartSize = int(defaultMultipartMaxPartSizeKB * 1024)
	}

	// 流式处理multipart的各个部分
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

		// 文件部件默认只记元数据：连内容都不读取，避免上传字节进入调试日志与数据库
		if fileName != "" && d.skipMultipartFileContent() {
			fileCount++
			result.WriteString(fmt.Sprintf("  [File] Name: %s, Field: %s, Content-Type: %s (content not recorded)\n",
				fileName, partName, part.Header.Get("Content-Type")))
			continue
		}

		// 读取部分内容（限制大小，避免大文件内存占用）
		partBytes := make([]byte, maxPartSize)
		n, err := part.Read(partBytes)
		if err != nil && err != io.EOF {
			result.WriteString(fmt.Sprintf("  Error reading part content: %v\n", err))
			continue
		}

		totalSize += n

		if fileName != "" {
			// 文件上传部分（显式关闭 MultipartSkipFiles 才会记录内容）
			fileCount++
			result.WriteString(fmt.Sprintf("  [File] Name: %s, Field: %s, Size: %d bytes, Type: %s, Content: %s\n",
				fileName, partName, n, detectFileType(partBytes[:n]), formatPartContentPreview(partBytes[:n])))
		} else {
			// 文本字段部分
			if isBinaryData(partBytes[:n]) {
				result.WriteString(fmt.Sprintf("  [Field] Name: %s, Size: %d bytes, Type: Binary\n",
					partName, n))
			} else {
				result.WriteString(fmt.Sprintf("  [Field] Name: %s, Size: %d bytes, Content: %s\n",
					partName, n, formatPartContentPreview(partBytes[:n])))
			}
		}

		// 检查总大小是否超过限制
		if int64(totalSize) > d.bodyCaptureLimit() {
			result.WriteString(fmt.Sprintf("  [Truncated] Total size exceeds limit: %d bytes\n", totalSize))
			break
		}
	}

	result.WriteString(fmt.Sprintf("Total Parts: %d, Files: %d, Total Size: %d bytes\n", partCount, fileCount, totalSize))
	return result.String(), nil
}

// formatPartContentPreview 输出 multipart 部件内容预览
// 二进制内容不展开（避免乱码与超长文本进入日志），文本最多保留 100 字符
// 参数：
//   - data: 已按 MultipartMaxPartSize 截断的部件字节
//
// 返回值：
//   - string: 可直接拼进日志行的一行内容
//
// 异常：不触发 panic
// 使用示例：
//
//	fmt.Sprintf("  [File] ... Content: %s", formatPartContentPreview(partBytes[:n]))
func formatPartContentPreview(data []byte) string {
	if isBinaryData(data) {
		return "(binary content not displayed)"
	}
	text := string(data)
	if len(text) > 100 {
		text = trimToRuneBoundary(text[:100]) + "..."
	}
	return text
}

// extractRequestBodyWithSizeLimit 带大小限制的请求体提取
// 作为multipart处理失败时的降级方案
func (d *Debugger) extractRequestBodyWithSizeLimit(c *gin.Context) (string, error) {
	return d.extractBoundedBody(c)
}

// extractRequestBodyWithSizeLimitSafe 安全的带大小限制请求体提取
// 避免破坏Gin上下文状态，通过复制请求体进行处理
func (d *Debugger) extractRequestBodyWithSizeLimitSafe(c *gin.Context) (string, error) {
	return d.extractBoundedBody(c)
}

// isBinaryData 检查数据是否为二进制格式
// 优化性能：使用更高效的二进制检测算法，减少内存访问和计算
// 通过检测不可打印字符的比例来判断是否为二进制数据
func isBinaryData(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// 优化：对于小数据直接检查，避免不必要的循环
	if len(data) <= 256 {
		return isBinaryDataQuick(data)
	}

	// 优化：采样检查，而不是检查整个1KB
	sampleSize := min(len(data), 512)          // 减少采样大小
	sampleStep := max(1, len(data)/sampleSize) // 采样步长

	nonPrintableCount := 0
	sampledCount := 0

	// 采样检查，减少内存访问次数
	for i := 0; i < len(data) && sampledCount < sampleSize; i += sampleStep {
		b := data[i]
		// 不可打印字符：ASCII码小于32且不是制表符、换行符、回车符
		if b < 32 && b != '\t' && b != '\n' && b != '\r' {
			nonPrintableCount++
		}
		sampledCount++
	}

	// 如果不可打印字符超过10%，则认为是二进制数据
	return float64(nonPrintableCount)/float64(sampledCount) > 0.1
}

// isBinaryDataQuick 快速二进制检测（适用于小数据）
func isBinaryDataQuick(data []byte) bool {
	nonPrintableCount := 0
	for i := 0; i < len(data); i++ {
		b := data[i]
		if b < 32 && b != '\t' && b != '\n' && b != '\r' {
			nonPrintableCount++
		}
	}
	return float64(nonPrintableCount)/float64(len(data)) > 0.1
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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
// isMultipartFormData 检查请求是否为multipart/form-data类型
func isMultipartFormData(c *gin.Context) bool {
	contentType := c.Request.Header.Get("Content-Type")
	isMultipart := strings.Contains(contentType, "multipart/form-data")
	fmt.Printf("[DEBUG isMultipartFormData] Content-Type: %s, isMultipart: %v\n", contentType, isMultipart)
	return isMultipart
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

		// 读取部分内容：整份请求体已在上游按捕获上限读入，这里再对单个部件限量，
		// 避免个别大部件把已捕获的字节全部复制一遍
		partBytes, err := io.ReadAll(io.LimitReader(part, defaultMultipartMaxPartSizeKB*1024))
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
	totalSize       int64 // 响应实际写出的总字节数（body 只保留到捕获上限，总数在此累计）
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
		w.totalSize += int64(len(b))
		// 有界捕获：响应体可能远大于 MaxBodySize（导出、附件、整页 HTML），
		// 只保留到上限的字节，其余仅累计长度，避免整份响应在内存中驻留
		if remaining := w.debugger.bodyCaptureLimit() - int64(w.body.Len()); remaining > 0 {
			if int64(len(b)) <= remaining {
				w.body.Write(b)
			} else {
				w.body.Write(b[:remaining])
			}
		}
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
// 支持无限制配置：当MaxStreamingChunks为0时表示无分块数量限制
// 优化内存管理：限制总内存使用量，避免内存泄漏
func (w *responseWriter) recordStreamingChunk(data []byte) {
	// 检查是否超过最大分块数量（仅当MaxStreamingChunks > 0时生效）
	if w.maxChunks > 0 && len(w.streamingChunks) >= w.maxChunks {
		// 超过最大分块数量时，移除最旧的分块（LRU策略）
		if len(w.streamingChunks) > 0 {
			w.streamingChunks = w.streamingChunks[1:]
		}
	}

	chunk := StreamingChunk{
		Timestamp: time.Now(),
		Size:      len(data),
		IsBinary:  w.isBinaryData(data),
	}

	// 限制单个分块的数据大小（仅当StreamingChunkSize > 0时生效）
	if w.chunkSizeLimit > 0 {
		maxChunkSize := w.chunkSizeLimit * 1024 // 转换为字节
		if int64(len(data)) > maxChunkSize {
			chunk.Data = string(data[:maxChunkSize]) + "... [truncated]"
		} else {
			chunk.Data = string(data)
		}
	} else {
		// 无限制时记录完整数据
		chunk.Data = string(data)
	}

	w.streamingChunks = append(w.streamingChunks, chunk)

	// 检查总内存使用量，如果超过限制则清理最旧的分块
	w.cleanupExcessiveMemory()
}

// cleanupExcessiveMemory 清理过量的内存使用
// 当流式分块总大小超过阈值时，自动清理最旧的分块
// 支持无限制配置：当MaxStreamingMemory为0时表示无内存限制
func (w *responseWriter) cleanupExcessiveMemory() {
	// 获取配置中的内存限制
	maxMemory := w.debugger.config.MaxStreamingMemory

	// 如果内存限制为0，表示无限制，直接返回
	if maxMemory == 0 {
		return
	}

	// 计算当前总内存使用量（使用原始字节大小，而不是截断后的字符串大小）
	totalSize := 0
	for _, chunk := range w.streamingChunks {
		totalSize += chunk.Size
	}

	// 如果超过阈值，清理最旧的分块直到满足限制
	for totalSize > int(maxMemory) && len(w.streamingChunks) > 0 {
		// 移除最旧的分块
		removedChunk := w.streamingChunks[0]
		w.streamingChunks = w.streamingChunks[1:]
		totalSize -= removedChunk.Size
	}
}

// isBinaryData 检测数据是否为二进制数据
// 优化：复用全局的isBinaryData函数，避免重复实现
// 通过不可打印字符比例判断是否为二进制数据
func (w *responseWriter) isBinaryData(data []byte) bool {
	return isBinaryData(data)
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

// GetMainLogger 获取主进程日志记录器
// 可以在应用程序的任何地方通过此方法获取主进程日志记录器实例
func (d *Debugger) GetMainLogger() LoggerInterface {
	if d == nil {
		// 如果Debugger实例为nil，返回空操作日志记录器
		return noopLogger{}
	}
	if d.mainLogger == nil {
		// 如果主进程日志未启用，返回空操作日志记录器
		return noopLogger{}
	}
	return d.mainLogger
}

// GetMainLoggerWithFields 获取带有指定字段的主进程日志记录器
// 可以在应用程序的任何地方通过此方法获取带有指定字段的主进程日志记录器实例
func (d *Debugger) GetMainLoggerWithFields(fields map[string]interface{}) LoggerInterface {
	if d == nil {
		// 如果Debugger实例为nil，返回空操作日志记录器
		return noopLogger{}
	}
	if d.mainLogger == nil {
		// 如果主进程日志未启用，返回空操作日志记录器
		return noopLogger{}
	}
	return d.mainLogger.WithFields(fields)
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
// 使用纳秒时间戳、进程ID和随机数确保唯一性，经过SM4加密后输出，密文结果不包含=字符
// 注意：Windows等平台计时器精度有限，连续调用的纳秒时间戳可能相同，
// 纯"时间戳+进程ID"组合在同进程内会重复（如多服务共享存储时日志互相覆盖），故追加随机数保证唯一性
func GenerateID() string {
	// 使用纳秒时间戳的后8位作为基础
	timestamp := time.Now().UnixNano()
	idSuffix := timestamp % 100000000

	// 添加进程ID的后2位作为额外标识，避免高并发下的重复
	pid := os.Getpid() % 100

	// 添加4位随机数，避免同进程内时间戳相同导致的ID重复
	idRandMu.Lock()
	random := idRand.Intn(10000)
	idRandMu.Unlock()

	// 组合格式：d_时间戳后8位_进程ID后2位_随机数4位
	originalID := fmt.Sprintf("d_%d_%02d_%04d", idSuffix, pid, random)

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
		Level:      LevelInfo,
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
		Level:      LevelInfo,
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
		Level:   LevelInfo,
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

// registerProcessLogger 将进程日志记录器登记到注册表
// 供 Debugger.EndProcess 仅凭 processID 结束时取回其内存中收集的日志
func (d *Debugger) registerProcessLogger(pl *ProcessLogger) {
	d.processLoggerMu.Lock()
	if d.processLoggers == nil {
		d.processLoggers = make(map[string]*ProcessLogger)
	}
	d.processLoggers[pl.processID] = pl
	d.processLoggerMu.Unlock()
}

// takeProcessLogger 从注册表取出并移除指定进程的日志记录器
// 返回nil表示该进程未登记或已被此前的EndProcess取走
func (d *Debugger) takeProcessLogger(processID string) *ProcessLogger {
	d.processLoggerMu.Lock()
	defer d.processLoggerMu.Unlock()
	pl := d.processLoggers[processID]
	delete(d.processLoggers, processID)
	return pl
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

	// 优先返回注册表中的原始记录器，保证同一进程的日志汇聚到同一实例
	d.processLoggerMu.RLock()
	registered := d.processLoggers[processID]
	d.processLoggerMu.RUnlock()
	if registered != nil {
		return registered, nil
	}

	// 创建进程记录器实例
	defaultLogger := NewDefaultLogger(d).WithFields(map[string]interface{}{
		"process_id":   entry.ProcessID,
		"process_name": entry.ProcessName,
		"process_type": entry.ProcessType,
	})

	logger := &ProcessLogger{
		debugger:    d,
		processID:   entry.ProcessID,
		processName: entry.ProcessName,
		processType: entry.ProcessType,
		startTime:   entry.Timestamp,
		logger:      defaultLogger.(*DefaultLogger),
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

	// 补充注册表中logger已收集的日志（该调用路径仅持有processID，不持有logger实例）
	if pl := d.takeProcessLogger(processID); pl != nil {
		entry.LoggerLogs = pl.logger.GetLogs()
	}

	// 统计派生严重度（error/warn数量），供列表直观展示与过滤
	entry.UpdateSeverity()

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
