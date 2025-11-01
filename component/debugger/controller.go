package debugger

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
)

// Controller 调试器控制器结构
type Controller struct {
	debugger *Debugger
	config   *ControllerConfig
	router   *gin.Engine
	basePath string // 基础路径，默认为 "/jcbase/debugger"
}

// ControllerConfig 控制器配置
type ControllerConfig struct {
	BasePath string `json:"base_path" default:"/jcbase/debugger"` // 基础路径，默认为 "/jcbase/debugger"
	Title    string `json:"title" default:"调试器"`                  // 页面标题，默认为 "调试器"
	PageSize int    `json:"page_size" default:"20"`               // 页面大小，默认为 20
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
		controller.registerRoutes()
	}

	return controller
}

// registerRoutes 注册调试器页面的路由
func (c *Controller) registerRoutes() {
	// 创建路由组
	routerGroup := c.router.Group("/" + c.basePath)

	// 重定向根目录到 /list
	routerGroup.GET("", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, helper.GetHostInfo(ctx.Request)+c.basePath+"/list")
	})

	// 调试器主页 - 显示日志列表
	routerGroup.GET("/list", c.indexHandler)

	// 日志详情页面
	routerGroup.GET("/detail/:id", c.detailHandler)

	// 搜索日志
	routerGroup.GET("/search", c.searchHandler)

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
}

// indexHandler 调试器主页处理器
func (c *Controller) indexHandler(ctx *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	// 获取过滤参数
	filters := c.parseFilters(ctx)

	// 获取日志列表
	entries, total, err := c.debugger.GetStorage().FindAll(page, pageSize, filters)
	if err != nil {
		c.renderError(ctx, "获取日志列表失败: "+err.Error())
		return
	}

	// 计算分页信息
	pagination := c.calculatePagination(page, pageSize, total)

	// 获取统计信息
	stats, _ := c.debugger.GetStorage().GetStats()

	// 计算每个日志条目的存储大小
	c.calculateEntriesStorageSize(entries)

	// 渲染页面
	c.renderTemplate(ctx, "index.html", gin.H{
		"Title":      "调试器 - 日志列表",
		"Entries":    entries,
		"Pagination": pagination,
		"Filters":    filters,
		"Stats":      stats,
		"BasePath":   c.basePath,
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
			// 日志不存在，返回404状态码
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
		// 日志不存在，返回404状态码
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

// searchHandler 搜索页面处理器
func (c *Controller) searchHandler(ctx *gin.Context) {
	keyword := ctx.Query("q")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	var entries []*LogEntry
	var total int
	var err error

	if keyword != "" {
		// 执行搜索
		entries, total, err = c.debugger.GetStorage().Search(keyword, page, pageSize)
	} else {
		// 没有关键词，显示所有日志
		entries, total, err = c.debugger.GetStorage().FindAll(page, pageSize, nil)
	}

	if err != nil {
		c.renderError(ctx, "搜索日志失败: "+err.Error())
		return
	}

	// 计算分页信息
	pagination := c.calculatePagination(page, pageSize, total)

	// 计算每个日志条目的存储大小
	c.calculateEntriesStorageSize(entries)

	// 渲染搜索页面
	c.renderTemplate(ctx, "search.html", gin.H{
		"Title":      "调试器 - 搜索",
		"Entries":    entries,
		"Pagination": pagination,
		"Keyword":    keyword,
		"BasePath":   c.basePath,
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
		// 如果错误消息包含"未找到"，则返回404状态码
		if strings.Contains(err.Error(), "未找到") {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		// 其他错误返回500状态码
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
func (c *Controller) parseFilters(ctx *gin.Context) map[string]interface{} {
	filters := make(map[string]interface{})

	// 方法过滤
	if method := ctx.Query("method"); method != "" {
		filters["method"] = method
	}

	// 状态码过滤
	if statusCode := ctx.Query("status_code"); statusCode != "" {
		if code, err := strconv.Atoi(statusCode); err == nil {
			filters["status_code"] = code
		}
	}

	// 时间范围过滤
	if startTime := ctx.Query("start_time"); startTime != "" {
		filters["start_time"] = startTime
	}

	if endTime := ctx.Query("end_time"); endTime != "" {
		filters["end_time"] = endTime
	}

	// URL路径过滤
	if url := ctx.Query("url"); url != "" {
		filters["url"] = url
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
	case "search.html":
		return searchTemplate
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
	c.registerRoutes()
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
