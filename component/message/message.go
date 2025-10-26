package message

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Data 消息提示数据结构
type Data struct {
	Title           string // 标题
	Content         string // 内容
	Type            string // 类型：success, error, info
	Referer         string // 返回链接
	RedirectURL     string // 跳转链接
	AutoRedirect    bool   // 是否自动跳转
	AutoRedirectSet bool   // 是否显式设置自动跳转
}

// Render 渲染消息提示页面
// 可以直接在任意控制器中调用，不需要单独注册路由
func Render(c *gin.Context, title, content, msgType string, options ...func(*Data)) {
	data := &Data{
		Title:   title,
		Content: content,
		Type:    msgType,
	}

	// 应用可选参数
	for _, option := range options {
		option(data)
	}

	// 如果没有设置Referer，尝试从请求头获取
	if data.Referer == "" {
		data.Referer = c.GetHeader("Referer")
	}

	// 应用全局配置
	cfg := GetGlobalConfigManager().GetConfig()
	if data.Type == "" {
		data.Type = cfg.DefaultType
	}
	if cfg.TitlePrefix != "" || cfg.TitleSuffix != "" {
		data.Title = cfg.TitlePrefix + data.Title + cfg.TitleSuffix
	}
	if !data.AutoRedirectSet && cfg.AutoRedirect {
		data.AutoRedirect = true
	}
	// 如果启用自动跳转且未设置跳转链接但存在Referer，使用Referer作为跳转目标
	if data.AutoRedirect && data.RedirectURL == "" && data.Referer != "" {
		data.RedirectURL = data.Referer
	}

	// 使用默认渲染器渲染
	err := GetDefaultRenderer().Render(c, data)
	if err != nil {
		log.Println("渲染消息失败:", err)
		return
	}
}

// WithReferer 设置返回链接
func WithReferer(referer string) func(*Data) {
	return func(data *Data) {
		data.Referer = referer
	}
}

// WithRedirect 设置跳转链接
func WithRedirect(url string) func(*Data) {
	return func(data *Data) {
		data.RedirectURL = url
	}
}

// WithAutoRedirect 启用自动跳转
func WithAutoRedirect(enabled bool) func(*Data) {
	return func(data *Data) {
		data.AutoRedirect = enabled
		data.AutoRedirectSet = true
	}
}

// Success 成功消息快捷函数
func Success(c *gin.Context, title, content string, options ...func(*Data)) {
	Render(c, title, content, "success", options...)
}

// Error 错误消息快捷函数
func Error(c *gin.Context, title, content string, options ...func(*Data)) {
	Render(c, title, content, "error", options...)
}

// Info 信息消息快捷函数
func Info(c *gin.Context, title, content string, options ...func(*Data)) {
	Render(c, title, content, "info", options...)
}

// Warning 警告消息快捷函数
func Warning(c *gin.Context, title, content string, options ...func(*Data)) {
	Render(c, title, content, "warning", options...)
}

// APIResponseWithMessage API响应包含消息
func APIResponseWithMessage(c *gin.Context, code, msg string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  msg,
		"data": data,
	})
}

// Simple 简化版消息提示（最常用的场景）
func Simple(c *gin.Context, content string, isSuccess bool) {
	if isSuccess {
		Success(c, "操作成功", content, WithReferer(""), WithAutoRedirect(true))
	} else {
		Error(c, "操作失败", content, WithReferer(""), WithAutoRedirect(true))
	}
}

// DelayedMessage 延迟消息（用于需要等待的操作）
func DelayedMessage(c *gin.Context, title, content, msgType string, delay time.Duration, options ...func(*Data)) {
	// 这里可以实现延迟显示消息的逻辑
	// 目前简化实现，直接显示消息
	Render(c, title, content, msgType, options...)
}

// SetMessageTemplate 设置消息模板
func SetMessageTemplate(templateContent string) error {
	return GetDefaultRenderer().SetTemplate(templateContent)
}
