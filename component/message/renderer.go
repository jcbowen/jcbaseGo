package message

import (
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// Renderer 消息渲染器接口
type Renderer interface {
	Render(c *gin.Context, data *Data) error
	SetTemplate(tpl string) error
}

// DefaultMessageRenderer 默认消息渲染器实现
type DefaultMessageRenderer struct {
	template *template.Template
	mutex    sync.RWMutex
}

// NewDefaultMessageRenderer 创建默认消息渲染器
func NewDefaultMessageRenderer() *DefaultMessageRenderer {
	// 默认模板字符串
	defaultTemplate := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .message-container {
            max-width: 600px;
            margin: 50px auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            padding: 30px;
            text-align: center;
        }
        .icon {
            font-size: 48px;
            margin-bottom: 20px;
        }
        .success .icon { color: #07c160; }
        .error .icon { color: #fa5151; }
        .info .icon { color: #10aeff; }
        .warning .icon { color: #ff9d00; }
        .title {
            font-size: 24px;
            font-weight: bold;
            margin-bottom: 10px;
        }
        .content {
            font-size: 16px;
            color: #666;
            margin-bottom: 30px;
            line-height: 1.5;
        }
        .buttons {
            display: flex;
            gap: 10px;
            justify-content: center;
            flex-wrap: wrap;
        }
        .btn {
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            text-decoration: none;
            font-size: 14px;
            cursor: pointer;
            transition: background-color 0.3s;
        }
        .btn-primary {
            background-color: #07c160;
            color: white;
        }
        .btn-secondary {
            background-color: #10aeff;
            color: white;
        }
        .btn-default {
            background-color: #f0f0f0;
            color: #333;
        }
        .countdown {
            margin-top: 15px;
            font-size: 14px;
            color: #999;
        }
    </style>
</head>
<body>
    <div class="message-container {{.Type}}">
        <div class="icon">
            {{if eq .Type "success"}}✓{{else if eq .Type "error"}}✗{{else if eq .Type "warning"}}⚠{{else}}ℹ{{end}}
        </div>
        <div class="title">{{.Title}}</div>
        <div class="content">{{.Content}}</div>
        <div class="buttons">
            {{if .RedirectURL}}
                <a href="{{.RedirectURL}}" class="btn btn-primary">继续操作</a>
            {{end}}
            {{if .Referer}}
                <a href="{{.Referer}}" class="btn btn-secondary">返回上页</a>
            {{end}}
        </div>
        {{if and .AutoRedirect .RedirectURL}}
            <div class="countdown" id="countdown">5秒后自动跳转...</div>
        {{end}}
    </div>

    {{if and .AutoRedirect .RedirectURL}}
    <script>
        let countdown = 5;
        const countdownElement = document.getElementById('countdown');
        
        function updateCountdown() {
            countdown--;
            countdownElement.textContent = countdown + '秒后自动跳转...';
            
            if (countdown <= 0) {
                window.location.href = '{{.RedirectURL}}';
            } else {
                setTimeout(updateCountdown, 1000);
            }
        }
        
        setTimeout(updateCountdown, 1000);
    </script>
    {{end}}
</body>
</html>`

	renderer := &DefaultMessageRenderer{}
	tmpl, err := template.New("message").Parse(defaultTemplate)
	if err != nil {
		panic("创建默认消息模板失败: " + err.Error())
	}
	renderer.template = tmpl
	return renderer
}

// Render 渲染消息页面
func (r *DefaultMessageRenderer) Render(c *gin.Context, data *Data) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if r.template == nil {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Status(http.StatusInternalServerError)
		_, _ = c.Writer.WriteString("消息模板未初始化")
		return nil
	}

	// 设置Content-Type
	c.Header("Content-Type", "text/html; charset=utf-8")

	// 渲染模板
	err := r.template.Execute(c.Writer, data)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_, _ = c.Writer.WriteString("模板渲染失败: " + err.Error())
	}
	return err
}

// SetTemplate 设置自定义模板
func (r *DefaultMessageRenderer) SetTemplate(tpl string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	tmpl, err := template.New("message").Parse(tpl)
	if err != nil {
		return err
	}
	r.template = tmpl
	return nil
}

// GetTemplate 获取当前模板内容（用于调试）
func (r *DefaultMessageRenderer) GetTemplate() string {
	// 注意：这个方法主要用于调试，实际使用时模板内容可能无法直接获取
	return "模板内容已加载，可通过SetTemplate方法设置新模板"
}

// 全局默认渲染器实例
var defaultRenderer Renderer

func init() {
	defaultRenderer = NewDefaultMessageRenderer()
}

// GetDefaultRenderer 获取默认渲染器实例
func GetDefaultRenderer() Renderer {
	return defaultRenderer
}

// SetGlobalRenderer 设置全局渲染器
func SetGlobalRenderer(renderer Renderer) {
	defaultRenderer = renderer
}

// RenderMessageWithRenderer 使用指定渲染器渲染消息
func RenderMessageWithRenderer(c *gin.Context, title, content, msgType string, renderer Renderer, options ...func(*Data)) {
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

	// 使用指定的渲染器渲染
	err := renderer.Render(c, data)
	if err != nil {
		log.Println("渲染消息失败:", err)
		return
	}
}
