# Message 组件

基于 Gin 框架的消息提示组件，提供统一的消息渲染、配置管理和模板自定义功能。

## 概述

Message 组件为 Web 应用程序提供了强大的消息提示功能，具有以下特点：

- **多种消息类型**：支持成功、错误、信息和警告四种消息类型
- **灵活配置**：支持全局配置管理和运行时配置更新
- **自定义模板**：提供默认模板，支持完全自定义模板
- **自动跳转**：支持自动跳转和倒计时功能
- **API 集成**：提供 API 响应格式支持

## 功能特性

### 消息类型支持
- **成功消息**：操作成功提示
- **错误消息**：操作失败提示  
- **信息消息**：普通信息提示
- **警告消息**：警告信息提示

### 配置管理
- **全局配置**：支持全局配置管理
- **配置文件**：支持 JSON 配置文件
- **运行时配置**：支持运行时动态更新配置
- **配置持久化**：配置自动保存到文件

### 模板系统
- **默认模板**：提供美观的默认模板
- **自定义模板**：支持完全自定义 HTML 模板
- **模板热更新**：支持运行时模板更新
- **响应式设计**：模板支持移动端适配

### 跳转功能
- **自动跳转**：支持自动跳转到指定页面
- **倒计时显示**：显示剩余跳转时间
- **返回链接**：自动获取和设置返回链接
- **多按钮支持**：支持多个操作按钮

## 安装指南

### 依赖要求

- Go 1.16+
- Gin Web 框架

### 安装依赖

```bash
go get -u github.com/gin-gonic/gin
```

## 使用示例

### 基本使用

#### 快速显示消息

```go
import "github.com/jcbowen/jcbaseGo/component/message"

// 在 Gin 路由处理函数中使用
func handleOperation(c *gin.Context) {
    // 操作成功
    message.Success(c, "操作成功", "您的操作已成功完成")
    
    // 操作失败
    message.Error(c, "操作失败", "请检查输入信息")
    
    // 信息提示
    message.Info(c, "系统提示", "系统将在5分钟后维护")
    
    // 警告提示
    message.Warning(c, "警告", "此操作不可撤销")
}
```

#### 带选项的消息

```go
func handleFormSubmit(c *gin.Context) {
    // 成功消息，带跳转链接
    message.Success(c, "提交成功", "表单已成功提交", 
        message.WithRedirect("/dashboard"),
        message.WithAutoRedirect(true),
    )
    
    // 错误消息，带返回链接
    message.Error(c, "提交失败", "请检查表单数据",
        message.WithReferer("/form"),
        message.WithAutoRedirect(false),
    )
}
```

### 配置管理

#### 全局配置

```go
import "github.com/jcbowen/jcbaseGo/component/message"

// 配置消息系统
message.ConfigureMessage(
    message.WithDefaultMessageType("success"),
    message.WithMessageAutoRedirect(true),
    message.WithMessageTitlePrefix("【系统】"),
    message.WithMessageTitleSuffix(" - 提示"),
)

// 设置自定义模板
message.ConfigureMessage(
    message.WithTemplateContent(`
        <!DOCTYPE html>
        <html>
        <head><title>{{.Title}}</title></head>
        <body>
            <h1>{{.Title}}</h1>
            <p>{{.Content}}</p>
        </body>
        </html>
    `),
)

// 从文件加载模板
message.ConfigureMessage(
    message.WithTemplatePath("./templates/message.html"),
)
```

#### 自定义配置文件路径

```go
// 设置自定义配置文件路径
message.SetGlobalConfigPath("./config/custom_message.json")
```

### 高级用法

#### API 响应集成

```go
// API 响应格式
func apiHandler(c *gin.Context) {
    // 处理业务逻辑
    result, err := processRequest(c)
    
    if err != nil {
        // 返回错误响应
        message.APIResponseWithMessage(c, "ERROR", "操作失败", nil)
        return
    }
    
    // 返回成功响应
    message.APIResponseWithMessage(c, "SUCCESS", "操作成功", result)
}
```

#### 简化消息

```go
// 简化版消息（最常用场景）
func simpleHandler(c *gin.Context) {
    // 操作成功
    message.Simple(c, "操作成功", true)
    
    // 操作失败
    message.Simple(c, "操作失败", false)
}
```

#### 延迟消息

```go
// 延迟显示消息（用于需要等待的操作）
func delayedHandler(c *gin.Context) {
    message.DelayedMessage(c, "处理中", "请稍候...", "info", 
        3*time.Second,
        message.WithRedirect("/result"),
    )
}
```

### 自定义渲染器

```go
import "github.com/jcbowen/jcbaseGo/component/message"

// 创建自定义渲染器
type CustomRenderer struct {
    // 实现 Renderer 接口
}

func (r *CustomRenderer) Render(c *gin.Context, data *message.Data) error {
    // 自定义渲染逻辑
    return nil
}

func (r *CustomRenderer) SetTemplate(tpl string) error {
    // 自定义模板设置逻辑
    return nil
}

// 使用自定义渲染器
customRenderer := &CustomRenderer{}
message.RenderMessageWithRenderer(c, "标题", "内容", "success", customRenderer)

// 设置全局自定义渲染器
message.SetGlobalRenderer(customRenderer)
```

## 详细功能说明

### 消息数据结构

```go
type Data struct {
    Title           string // 标题
    Content         string // 内容
    Type            string // 类型：success, error, info, warning
    Referer         string // 返回链接
    RedirectURL     string // 跳转链接
    AutoRedirect    bool   // 是否自动跳转
    AutoRedirectSet bool   // 是否显式设置自动跳转
}
```

### 配置结构

```go
type Config struct {
    TemplatePath    string `json:"template_path"`    // 模板文件路径
    TemplateContent string `json:"template_content"` // 模板内容
    AutoRedirect    bool   `json:"auto_redirect"`    // 自动跳转
    DefaultType     string `json:"default_type"`     // 默认消息类型
    TitlePrefix     string `json:"title_prefix"`    // 标题前缀
    TitleSuffix     string `json:"title_suffix"`    // 标题后缀
}
```

### 默认模板

组件提供美观的默认模板，包含：

- **响应式设计**：支持各种屏幕尺寸
- **图标显示**：不同类型消息显示不同图标
- **按钮布局**：智能按钮显示逻辑
- **倒计时功能**：自动跳转倒计时显示
- **现代化样式**：使用现代 CSS 设计

### 配置选项函数

- `WithTemplateContent(content string)` - 设置模板内容
- `WithTemplatePath(path string)` - 设置模板文件路径
- `WithMessageAutoRedirect(auto bool)` - 设置自动跳转
- `WithDefaultMessageType(msgType string)` - 设置默认消息类型
- `WithMessageTitlePrefix(prefix string)` - 设置标题前缀
- `WithMessageTitleSuffix(suffix string)` - 设置标题后缀

### 消息选项函数

- `WithReferer(referer string)` - 设置返回链接
- `WithRedirect(url string)` - 设置跳转链接
- `WithAutoRedirect(enabled bool)` - 启用自动跳转

## 高级用法

### 多语言支持

```go
// 根据语言环境显示不同消息
func localizedHandler(c *gin.Context) {
    lang := c.GetHeader("Accept-Language")
    
    var title, content string
    switch lang {
    case "en":
        title = "Success"
        content = "Operation completed successfully"
    case "zh":
        title = "操作成功"
        content = "操作已成功完成"
    default:
        title = "Success"
        content = "Operation completed"
    }
    
    message.Success(c, title, content)
}
```

### 主题定制

```go
// 根据主题使用不同模板
func themedHandler(c *gin.Context) {
    theme := c.Query("theme")
    
    var templateContent string
    switch theme {
    case "dark":
        templateContent = darkThemeTemplate
    case "light":
        templateContent = lightThemeTemplate
    default:
        templateContent = defaultTemplate
    }
    
    message.SetMessageTemplate(templateContent)
    message.Success(c, "操作成功", "主题已切换")
}
```

### 消息队列

```go
// 实现消息队列功能
func messageQueueHandler(c *gin.Context) {
    // 存储消息到 Session 或 Cookie
    // 下次页面加载时显示
    
    message.Success(c, "操作成功", "消息已加入队列")
}
```

## 性能优化建议

### 模板缓存

- 模板在内存中缓存，避免重复解析
- 支持模板热更新，无需重启服务
- 合理使用模板复用

### 配置优化

- 配置文件自动缓存
- 配置变更时自动重新加载
- 避免频繁的配置更新

### 内存管理

- 消息数据轻量级
- 模板内容合理控制大小
- 及时清理不必要的模板缓存

## 安全考虑

### XSS 防护

- 使用 Go 模板自动转义 HTML
- 避免直接输出用户输入内容
- 对动态内容进行适当过滤

### CSRF 防护

- 配合 Gin 的 CSRF 中间件使用
- 对重要操作进行二次确认
- 使用安全的跳转链接

### 内容安全

- 验证跳转链接的合法性
- 限制自动跳转的目标域名
- 对用户输入进行适当验证

## API 参考

### 主要函数

#### 消息显示函数

- `Render(c *gin.Context, title, content, msgType string, options ...func(*Data))` - 渲染消息
- `Success(c *gin.Context, title, content string, options ...func(*Data))` - 成功消息
- `Error(c *gin.Context, title, content string, options ...func(*Data))` - 错误消息
- `Info(c *gin.Context, title, content string, options ...func(*Data))` - 信息消息
- `Warning(c *gin.Context, title, content string, options ...func(*Data))` - 警告消息
- `Simple(c *gin.Context, content string, isSuccess bool)` - 简化消息
- `DelayedMessage(c *gin.Context, title, content, msgType string, delay time.Duration, options ...func(*Data))` - 延迟消息

#### API 响应函数

- `APIResponseWithMessage(c *gin.Context, code, msg string, data interface{})` - API 响应

#### 配置管理函数

- `ConfigureMessage(options ...func(*Config))` - 配置消息系统
- `SetGlobalConfigPath(path string)` - 设置全局配置路径
- `SetMessageTemplate(templateContent string) error` - 设置消息模板

#### 渲染器函数

- `RenderMessageWithRenderer(c *gin.Context, title, content, msgType string, renderer Renderer, options ...func(*Data))` - 使用指定渲染器
- `SetGlobalRenderer(renderer Renderer)` - 设置全局渲染器
- `GetDefaultRenderer() Renderer` - 获取默认渲染器

### 配置选项函数

- `WithTemplateContent(content string) func(*Config)`
- `WithTemplatePath(path string) func(*Config)`
- `WithMessageAutoRedirect(auto bool) func(*Config)`
- `WithDefaultMessageType(msgType string) func(*Config)`
- `WithMessageTitlePrefix(prefix string) func(*Config)`
- `WithMessageTitleSuffix(suffix string) func(*Config)`

### 消息选项函数

- `WithReferer(referer string) func(*Data)`
- `WithRedirect(url string) func(*Data)`
- `WithAutoRedirect(enabled bool) func(*Data)`

## 错误处理

组件提供完善的错误处理机制：

```go
// 模板设置错误处理
err := message.SetMessageTemplate(templateContent)
if err != nil {
    log.Printf("设置模板失败: %v", err)
    // 使用默认模板继续
}

// 渲染错误处理
// 渲染失败时会自动记录日志并返回错误页面
```

## 常见问题

### 模板不生效

检查模板语法是否正确，确保模板内容符合 Go template 规范。

### 配置不保存

确保配置文件路径有写权限，检查文件系统权限设置。

### 自动跳转不工作

检查是否设置了正确的跳转链接，确认自动跳转功能已启用。

### 消息显示异常

检查 Gin 上下文是否正确传递，确认响应头没有被其他中间件修改。

## 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个组件。

## 许可证

MIT License