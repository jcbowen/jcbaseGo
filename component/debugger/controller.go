package debugger

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/middleware"
)

// Controller 调试器控制器结构
type Controller struct {
	debugger *Debugger
	config   *ControllerConfig
	router   *gin.Engine
	basePath string // 基础路径，默认为 "/jcbase/debug"
}

// ControllerConfig 控制器配置
type ControllerConfig struct {
	BasePath string `json:"base_path" default:"/jcbase/debug"` // 基础路径，默认为 "/jcbase/debug"
	UseCDN   bool   `json:"use_cdn" default:"false"`           // 是否使用CDN，默认为 false
	Title    string `json:"title" default:"调试器"`               // 页面标题，默认为 "调试器"
	PageSize int    `json:"page_size" default:"20"`            // 页面大小，默认为 20
}

// Pagination 分页信息
type Pagination struct {
	Page       int  // 当前页码
	PageSize   int  // 每页大小
	Total      int  // 总记录数
	TotalPages int  // 总页数
	HasPrev    bool // 是否有上一页
	HasNext    bool // 是否有下一页
	PrevPage   int  // 上一页页码
	NextPage   int  // 下一页页码
}

// TemplateData 模板数据
type TemplateData struct {
	Title      string
	BasePath   string
	Entries    []*LogEntry
	Entry      *LogEntry
	Pagination *Pagination
	Filters    map[string]string
	Keyword    string
	Stats      map[string]interface{}
}

// NewController 创建新的调试器控制器
func NewController(debugger *Debugger, router *gin.Engine, config *ControllerConfig) *Controller {
	if config == nil {
		config = &ControllerConfig{}
	}

	// 使用CheckAndSetDefault方法设置默认值，符合jcbaseGo规范
	if err := helper.CheckAndSetDefault(config); err != nil {
		// 记录错误但不中断程序执行
		fmt.Printf("设置控制器配置默认值失败: %v\n", err)
	}

	controller := &Controller{
		debugger: debugger,
		config:   config,
		router:   router,
		basePath: config.BasePath,
	}

	// 如果提供了路由引擎，自动注册路由
	if router != nil {
		controller.registerRoutes(config.UseCDN)
	}

	return controller
}

// registerRoutes 注册调试器页面的路由
func (c *Controller) registerRoutes(useCDN bool) {
	// 创建路由组
	routerGroup := c.router.Group("/" + c.basePath)

	// 添加IP访问控制中间件
	routerGroup.Use(c.ipAccessControlMiddleware(useCDN))

	// 重定向根目录到 /list
	routerGroup.GET("", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, helper.GetHostInfo(ctx.Request)+c.basePath+"/list")
	})

	// 调试器主页 - 显示日志列表
	routerGroup.GET("/list", c.indexHandler)

	// 日志详情页面
	routerGroup.GET("/detail/:id", c.detailHandler)

	// 获取日志列表API（JSON格式）
	routerGroup.GET("/api/logs", c.logsAPIHandler)

	// 获取日志详情API（JSON格式）
	routerGroup.GET("/api/logs/:id", c.logDetailAPIHandler)

	// 搜索日志API（JSON格式）
	routerGroup.GET("/api/search", c.searchAPIHandler)

	// 获取统计信息API
	routerGroup.GET("/api/stats", c.statsAPIHandler)

	// 清理过期日志API
	routerGroup.POST("/api/cleanup", c.cleanupAPIHandler)

	// 下载主进程日志API
	routerGroup.GET("/api/download-main-logs", c.downloadMainLogsHandler)
}

// ipAccessControlMiddleware IP访问控制中间件
// 检查客户端IP是否在允许的白名单中，如果配置了白名单但IP不在其中，返回403禁止访问
func (c *Controller) ipAccessControlMiddleware(useCDN bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取调试器配置
		config := c.debugger.GetConfig()

		// 如果没有配置IP白名单，允许所有访问
		if len(config.AllowedIPs) == 0 {
			ctx.Next()
			return
		}

		// 获取客户端IP地址
		clientIP := middleware.GetRealIP(ctx, useCDN) // 默认不使用CDN

		// 检查IP是否在白名单中
		if c.isIPAllowed(clientIP, config.AllowedIPs) {
			ctx.Next()
			return
		}

		log.Println("debugger禁止访问")
		headerJson, _ := json.Marshal(ctx.Request.Header)
		log.Println("Request Header:", string(headerJson))
		bodyJson, _ := json.Marshal(ctx.Request.Body)
		log.Println("Request Body:", string(bodyJson))

		// IP不在白名单中，返回403禁止访问
		ctx.JSON(http.StatusForbidden, gin.H{
			"error":     "禁止访问：您的IP地址不在允许列表中",
			"client_ip": clientIP,
		})
		ctx.Abort()
	}
}

// getClientIP 获取客户端真实IP地址
// 支持从X-Forwarded-For等代理头中获取真实IP
func (c *Controller) getClientIP(ctx *gin.Context) string {
	// 尝试从X-Forwarded-For获取
	if forwardedFor := ctx.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从X-Real-IP获取
	if realIP := ctx.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	// 使用远程地址
	return ctx.ClientIP()
}

// isIPAllowed 检查IP是否在允许的白名单中
// 支持IP地址和CIDR格式（如192.168.1.0/24）
func (c *Controller) isIPAllowed(clientIP string, allowedIPs []string) bool {
	// 如果客户端IP为空，拒绝访问
	if clientIP == "" {
		return false
	}

	// 检查每个允许的IP规则
	for _, allowedIP := range allowedIPs {
		// 如果是CIDR格式
		if strings.Contains(allowedIP, "/") {
			if c.isIPInCIDR(clientIP, allowedIP) {
				return true
			}
		} else {
			// 直接比较IP地址
			if clientIP == allowedIP {
				return true
			}
		}
	}

	return false
}

// isIPInCIDR 检查IP是否在CIDR范围内
func (c *Controller) isIPInCIDR(ip, cidr string) bool {
	// 解析CIDR
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	// 解析IP
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 检查IP是否在CIDR范围内
	return ipNet.Contains(parsedIP)
}

// generateQueryString 生成查询字符串
// 从请求中获取所有查询参数，排除指定的参数，生成完整的查询字符串
func (c *Controller) generateQueryString(ctx *gin.Context, exclude ...string) string {
	query := url.Values{}

	// 从请求中获取所有查询参数
	for key, values := range ctx.Request.URL.Query() {
		// 排除指定的参数
		excluded := false
		for _, ex := range exclude {
			if key == ex {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}
		// 添加参数到查询字符串
		for _, value := range values {
			query.Add(key, value)
		}
	}

	result := query.Encode()
	if result != "" {
		result = "?" + result
	}
	return result
}

// indexHandler 调试器主页处理器（支持搜索功能）
func (c *Controller) indexHandler(ctx *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	// 获取搜索关键词
	keyword := ctx.Query("q")

	// 获取过滤参数
	filters := c.parseFilters(ctx)

	var entries []*LogEntry
	var total int
	var err error

	// 根据是否有搜索关键词选择不同的查询方式
	if keyword != "" {
		// 执行搜索
		entries, total, err = c.debugger.GetStorage().Search(keyword, page, pageSize)
	} else {
		// 获取所有日志列表
		entries, total, err = c.debugger.GetStorage().FindAll(page, pageSize, filters)
	}

	if err != nil {
		c.renderError(ctx, "获取日志列表失败: "+err.Error())
		return
	}

	// 计算分页信息
	pagination := c.calculatePagination(page, pageSize, total)

	// 获取统计信息（仅在非搜索模式下显示）
	var stats map[string]interface{}
	if keyword == "" {
		stats, _ = c.debugger.GetStorage().GetStats()
	}

	// 计算每个日志条目的存储大小
	c.calculateEntriesStorageSize(entries)

	// 生成查询字符串（排除page参数，因为会在分页链接中动态设置）
	queryString := c.generateQueryString(ctx, "page")

	// 渲染页面
	c.renderTemplate(ctx, "index.html", gin.H{
		"Title":       "调试器 - 日志列表",
		"Entries":     entries,
		"Pagination":  pagination,
		"Filters":     filters,
		"Stats":       stats,
		"Keyword":     keyword,
		"BasePath":    c.basePath,
		"QueryString": queryString,
	})
}

// detailHandler 日志详情处理器
func (c *Controller) detailHandler(ctx *gin.Context) {
	id := ctx.Param("id")

	// 获取日志详情
	entry, err := c.debugger.GetStorage().FindByID(id)
	if err != nil {
		// 检查是否是"未找到"的错误
		if strings.Contains(err.Error(), "未找到") {
			// 日志不存在，返回 404 状态码
			ctx.Status(http.StatusNotFound)
			c.renderTemplate(ctx, "error.html", gin.H{
				"Title":    "404 - 页面未找到",
				"Message":  "未找到ID为 " + id + " 的日志条目",
				"BasePath": c.basePath,
			})
			return
		}
		// 其他错误
		c.renderError(ctx, "获取日志详情失败: "+err.Error())
		return
	}

	if entry == nil {
		// 日志不存在，返回 404 状态码
		ctx.Status(http.StatusNotFound)
		c.renderTemplate(ctx, "error.html", gin.H{
			"Title":    "404 - 页面未找到",
			"Message":  "未找到ID为 " + id + " 的日志条目",
			"BasePath": c.basePath,
		})
		return
	}

	// 计算日志条目的存储大小
	entry.StorageSize = entry.CalculateStorageSize()

	// 渲染详情页面
	c.renderTemplate(ctx, "detail.html", gin.H{
		"Title":    "日志详情 - " + entry.ID,
		"Entry":    entry,
		"BasePath": c.basePath,
	})
}

// logsAPIHandler 日志列表API处理器
func (c *Controller) logsAPIHandler(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	filters := c.parseFilters(ctx)

	entries, total, err := c.debugger.GetStorage().FindAll(page, pageSize, filters)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取日志列表失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": entries,
		"pagination": gin.H{
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
			"pages":    (total + pageSize - 1) / pageSize,
		},
	})
}

// logDetailAPIHandler 日志详情API处理器
func (c *Controller) logDetailAPIHandler(ctx *gin.Context) {
	id := ctx.Param("id")

	entry, err := c.debugger.GetStorage().FindByID(id)
	if err != nil {
		// 如果错误消息包含"未找到"，则返回 404 状态码
		if strings.Contains(err.Error(), "未找到") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		// 其他错误返回 500 状态码
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取日志详情失败: " + err.Error(),
		})
		return
	}

	if entry == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "日志不存在",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": entry,
	})
}

// searchAPIHandler 搜索日志API处理器
func (c *Controller) searchAPIHandler(ctx *gin.Context) {
	keyword := ctx.Query("q")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	if keyword == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "搜索关键词不能为空",
		})
		return
	}

	entries, total, err := c.debugger.GetStorage().Search(keyword, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "搜索日志失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": entries,
		"pagination": gin.H{
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
			"pages":    (total + pageSize - 1) / pageSize,
		},
	})
}

// downloadMainLogsHandler 主进程日志下载API处理器
// 支持压缩下载所有主进程日志文件
func (c *Controller) downloadMainLogsHandler(ctx *gin.Context) {
	// 获取调试器配置
	config := c.debugger.GetConfig()

	// 检查主进程日志是否启用
	if config.MainLogPath == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "主进程日志未启用，无法下载",
		})
		return
	}

	// 检查日志路径是否存在，如果不存在则尝试创建
	logPathInfo, err := os.Stat(config.MainLogPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 尝试创建日志目录
			if err := os.MkdirAll(config.MainLogPath, 0755); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error": "创建日志目录失败: " + err.Error(),
				})
				return
			}
			// 目录创建成功，继续执行
		} else {
			// 其他错误
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "检查日志路径失败: " + err.Error(),
			})
			return
		}
	} else if !logPathInfo.IsDir() {
		// 路径存在但不是目录
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "主进程日志路径不是目录: " + config.MainLogPath,
		})
		return
	}

	// 创建临时zip文件
	tempFile, err := os.CreateTemp("", "main-logs-*.zip")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建临时文件失败: " + err.Error(),
		})
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// 创建zip写入器
	zipWriter := zip.NewWriter(tempFile)
	defer zipWriter.Close()

	// 查找所有主进程日志文件
	logFiles, err := findMainLogFiles(config.MainLogPath)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "查找日志文件失败: " + err.Error(),
		})
		return
	}

	if len(logFiles) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "没有找到主进程日志文件",
		})
		return
	}

	// 将所有日志文件添加到zip中
	for _, logFile := range logFiles {
		// 打开日志文件
		srcFile, err := os.Open(logFile)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "打开日志文件失败: " + err.Error(),
			})
			return
		}
		defer srcFile.Close()

		// 在zip中创建文件
		zipFile, err := zipWriter.Create(filepath.Base(logFile))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "在zip中创建文件失败: " + err.Error(),
			})
			return
		}

		// 将日志文件内容复制到zip中
		if _, err := io.Copy(zipFile, srcFile); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "复制日志文件到zip失败: " + err.Error(),
			})
			return
		}
	}

	// 关闭zip写入器，确保所有数据写入完成
	if err := zipWriter.Close(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "关闭zip写入器失败: " + err.Error(),
		})
		return
	}

	// 重置临时文件指针到开始位置
	if _, err := tempFile.Seek(0, io.SeekStart); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "重置文件指针失败: " + err.Error(),
		})
		return
	}

	// 设置响应头
	ctx.Header("Content-Type", "application/zip")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=main-logs-%s.zip", time.Now().Format("20060102150405")))

	// 将zip文件发送给客户端
	if _, err := io.Copy(ctx.Writer, tempFile); err != nil {
		log.Printf("发送zip文件失败: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "发送日志文件失败: " + err.Error(),
		})
		return
	}
}

// findMainLogFiles 查找所有主进程日志文件
// 支持不同的分割模式（按大小、按日期）
func findMainLogFiles(logPath string) ([]string, error) {
	var logFiles []string

	// 定义日志文件名模式
	patterns := []string{
		filepath.Join(logPath, "process.log"),                                               // 当前日志文件
		filepath.Join(logPath, "process.log.[0-9]*"),                                        // 按大小分割的日志文件（仅数字后缀）
		filepath.Join(logPath, "process.log.[0-9]*.gz"),                                     // 压缩的按大小分割日志文件
		filepath.Join(logPath, "process.log.[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]"),    // 按日期分割的日志文件（YYYY-MM-DD格式）
		filepath.Join(logPath, "process.log.[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9].gz"), // 压缩的按日期分割日志文件
	}

	// 遍历所有模式，查找匹配的文件
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("查找日志文件失败: %w", err)
		}
		logFiles = append(logFiles, matches...)
	}

	// 去重，避免重复添加同一文件
	uniqueFiles := make(map[string]bool)
	var uniqueLogFiles []string
	for _, file := range logFiles {
		// 检查文件是否存在
		if _, err := os.Stat(file); err == nil {
			if !uniqueFiles[file] {
				uniqueFiles[file] = true
				uniqueLogFiles = append(uniqueLogFiles, file)
			}
		}
	}

	// 按修改时间排序，最新的日志文件在前
	sort.Slice(uniqueLogFiles, func(i, j int) bool {
		// 获取文件i的信息
		infoI, errI := os.Stat(uniqueLogFiles[i])
		// 获取文件j的信息
		infoJ, errJ := os.Stat(uniqueLogFiles[j])

		// 如果有错误，将错误的文件放在后面
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}

		// 按修改时间降序排序
		if !infoI.ModTime().Equal(infoJ.ModTime()) {
			return infoI.ModTime().After(infoJ.ModTime())
		}

		// 如果修改时间相同，按文件名降序排序
		return uniqueLogFiles[i] > uniqueLogFiles[j]
	})

	return uniqueLogFiles, nil
}

// statsAPIHandler 统计信息API处理器
func (c *Controller) statsAPIHandler(ctx *gin.Context) {
	stats, err := c.debugger.GetStorage().GetStats()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取统计信息失败: " + err.Error(),
		})
		return
	}

	// 获取HTTP方法统计
	methods, _ := c.debugger.GetStorage().GetMethods()

	// 获取状态码统计
	statusCodes, _ := c.debugger.GetStorage().GetStatusCodes()

	ctx.JSON(http.StatusOK, gin.H{
		"stats":       stats,
		"methods":     methods,
		"statusCodes": statusCodes,
	})
}

// cleanupAPIHandler 清理过期日志API处理器
func (c *Controller) cleanupAPIHandler(ctx *gin.Context) {
	// 计算过期时间（保留期限前的日志）
	retentionPeriod := c.debugger.GetConfig().RetentionPeriod
	before := time.Now().Add(-retentionPeriod)

	err := c.debugger.GetStorage().Cleanup(before)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "清理过期日志失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "清理过期日志成功",
		"before":  before.Format(time.RFC3339),
	})
}

// parseFilters 解析过滤参数
// 从HTTP请求中解析查询参数，构建用于日志查询的过滤条件
// 支持HTTP记录和进程记录的多种过滤条件，包括记录类型、方法、状态码、IP地址、进程名称、进程ID、时间范围和URL路径
//
// 参数:
//
//	ctx: Gin上下文，包含HTTP请求信息
//
// 返回值:
//
//	map[string]interface{}: 过滤条件映射，键为字段名，值为过滤值
//
// 支持的查询参数:
//   - record_type: 记录类型（http/process）
//   - method: HTTP方法（GET/POST/PUT/DELETE等）
//   - status_code: HTTP状态码（200/404/500等）
//   - client_ip: 客户端IP地址
//   - host: 域名（支持模糊匹配）
//   - process_name: 进程名称
//   - process_id: 进程ID
//   - process_status: 进程状态（running/completed/failed/error）
//   - start_time: 开始时间（ISO格式）
//   - end_time: 结束时间（ISO格式）
//   - url: URL路径（支持模糊匹配）
//
// 示例:
//
//	GET /logs?record_type=process&process_name=数据同步任务&start_time=2024-01-01T00:00:00Z
func (c *Controller) parseFilters(ctx *gin.Context) map[string]interface{} {
	filters := make(map[string]interface{})

	// 记录类型过滤
	if recordType := ctx.Query("record_type"); recordType != "" {
		filters["record_type"] = recordType
	}

	// 方法过滤
	if method := ctx.Query("method"); method != "" {
		filters["method"] = method
	}

	// 状态码过滤
	if statusCode := ctx.Query("status_code"); statusCode != "" {
		if code, err := strconv.Atoi(statusCode); err == nil {
			filters["status_code"] = code
		} else {
			filters["status_code"] = statusCode
		}
	}

	// IP地址过滤
	if ip := ctx.Query("client_ip"); ip != "" {
		filters["client_ip"] = ip
	}

	// 域名过滤
	if host := ctx.Query("host"); host != "" {
		filters["host"] = host
	}

	// 进程名称过滤
	if processName := ctx.Query("process_name"); processName != "" {
		filters["process_name"] = processName
	}

	// 进程ID过滤
	if processID := ctx.Query("process_id"); processID != "" {
		filters["process_id"] = processID
	}

	// 进程状态过滤
	if processStatus := ctx.Query("process_status"); processStatus != "" {
		filters["process_status"] = processStatus
	}

	// 时间范围过滤
	if startTime := ctx.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filters["start_time"] = t
		}
	}

	if endTime := ctx.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filters["end_time"] = t
		}
	}

	// URL路径过滤
	if url := ctx.Query("url"); url != "" {
		filters["url"] = url
	}

	// 流式请求过滤
	if isStreaming := ctx.Query("is_streaming"); isStreaming != "" {
		filters["is_streaming"] = strings.ToLower(isStreaming) == "true"
	}

	// 流式状态过滤
	if streamingStatus := ctx.Query("streaming_status"); streamingStatus != "" {
		filters["streaming_status"] = streamingStatus
	}

	return filters
}

// calculatePagination 计算分页信息
func (c *Controller) calculatePagination(page, pageSize, total int) gin.H {
	// 处理pageSize为0的情况，避免除零错误
	if pageSize <= 0 {
		pageSize = 20 // 默认分页大小
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	return gin.H{
		"Page":       page,
		"PageSize":   pageSize,
		"Total":      total,
		"TotalPages": totalPages,
		"HasPrev":    page > 1,
		"HasNext":    page < totalPages,
		"PrevPage":   page - 1,
		"NextPage":   page + 1,
	}
}

// renderTemplate 渲染HTML模板
func (c *Controller) renderTemplate(ctx *gin.Context, templateName string, data gin.H) {
	// 设置默认数据
	if data == nil {
		data = gin.H{}
	}

	// 添加基础路径和标题
	data["BasePath"] = c.basePath
	if _, exists := data["Title"]; !exists {
		data["Title"] = c.config.Title
	}

	// 获取模板内容
	templateContent := getTemplateContent(templateName)
	if templateContent == "" {
		c.renderError(ctx, "模板不存在: "+templateName)
		return
	}

	// 创建模板函数映射
	funcMap := template.FuncMap{
		"lower": strings.ToLower,
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"add": func(a, b int) int {
			return a + b
		},
		"isJSON": func(s string) bool {
			// 检查字符串是否为空
			if s == "" {
				return false
			}

			// 去除首尾空白字符
			trimmed := strings.TrimSpace(s)

			// 检查是否以{或[开头，以}或]结尾
			if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
				(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {

				// 尝试解析JSON验证有效性
				var js interface{}
				return json.Unmarshal([]byte(trimmed), &js) == nil
			}

			return false
		},
	}

	// 解析模板
	tmpl, err := template.New(templateName).Funcs(funcMap).Parse(templateContent)
	if err != nil {
		c.renderError(ctx, "模板解析失败: "+err.Error())
		return
	}

	// 渲染模板
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(ctx.Writer, data)
	if err != nil {
		c.renderError(ctx, "模板渲染失败: "+err.Error())
		return
	}
}

// renderError 渲染错误页面
func (c *Controller) renderError(ctx *gin.Context, message string) {
	// 使用自定义模板渲染逻辑
	c.renderTemplate(ctx, "error.html", gin.H{
		"Title":    "错误",
		"Message":  message,
		"BasePath": c.basePath,
	})
}

// getTemplateContent 获取模板内容（内联模板）
func getTemplateContent(templateName string) string {
	switch templateName {
	case "index.html":
		return indexTemplate
	case "detail.html":
		return detailTemplate
	case "error.html":
		return errorTemplate
	default:
		return errorTemplate
	}
}

// RegisterRoutes 手动注册路由到指定的Gin引擎
// 用于在初始化时没有传入路由组的情况
func (c *Controller) RegisterRoutes(router *gin.Engine) {
	c.router = router
	c.registerRoutes(c.config.UseCDN)
}

// GetBasePath 获取基础路径
func (c *Controller) GetBasePath() string {
	return c.basePath
}

// SetBasePath 设置基础路径
func (c *Controller) SetBasePath(path string) {
	c.basePath = path
}

// GetDebugger 获取调试器实例
func (c *Controller) GetDebugger() *Debugger {
	return c.debugger
}

// calculateEntriesStorageSize 计算日志条目列表的存储大小
// 遍历所有日志条目，为每个条目计算并设置StorageSize字段
func (c *Controller) calculateEntriesStorageSize(entries []*LogEntry) {
	for _, entry := range entries {
		entry.StorageSize = entry.CalculateStorageSize()
	}
}
