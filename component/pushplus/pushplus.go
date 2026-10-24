package pushplus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jcbowen/jcbaseGo"
)

const (
	// DefaultBaseURL PushPlus 发送接口默认地址
	DefaultBaseURL = "https://www.pushplus.plus/send"
	// DefaultTimeout 默认请求超时时间
	DefaultTimeout = 10 * time.Second

	// ChannelWeChat 微信公众号渠道
	ChannelWeChat = "wechat"
	// ChannelWebhook webhook 渠道
	ChannelWebhook = "webhook"
	// ChannelCP 企业微信应用渠道
	ChannelCP = "cp"
	// ChannelMail 邮箱渠道
	ChannelMail = "mail"
	// ChannelSMS 短信渠道
	ChannelSMS = "sms"
	// ChannelVoice 语音渠道
	ChannelVoice = "voice"

	// TemplateHTML HTML 模板
	TemplateHTML = "html"
	// TemplateJSON JSON 模板
	TemplateJSON = "json"
	// TemplateMarkdown Markdown 模板
	TemplateMarkdown = "markdown"
	// TemplateTxt 纯文本模板
	TemplateTxt = "txt"
)

// maxResponseSize 响应体最大读取字节数，避免异常响应占用过多内存
const maxResponseSize = 1 << 20

// maxErrorSnippet 错误信息中携带的响应片段最大长度，避免日志被超大响应污染
const maxErrorSnippet = 512

// Message 单条推送消息，字段为空时回落到客户端默认值
type Message struct {
	Title       string // 消息标题
	Content     string // 消息正文，格式由 Template 决定
	Topic       string // 群组编码，为空时使用客户端默认群组
	Channel     string // 发送渠道，为空时使用客户端默认渠道
	Template    string // 消息模板，为空时使用客户端默认模板
	Webhook     string // webhook 渠道的接收地址
	CallbackURL string // 推送结果回调地址
}

// Response PushPlus 接口返回结构
type Response struct {
	Code int    `json:"code"` // 业务状态码，200 表示请求已受理
	Msg  string `json:"msg"`  // 业务提示信息
	Data string `json:"data"` // 消息流水号，可用于查询推送结果
}

// APIError PushPlus 返回业务错误（code 非 200）时携带的错误
type APIError struct {
	Code int    // 业务状态码
	Msg  string // 业务提示信息
}

// Error 实现 error 接口，返回包含业务码与提示信息的错误描述。
//
// 参数:
//   - 无
//
// 返回值:
//   - string: 形如 "pushplus: 接口返回失败 code=999 msg=token无效" 的错误描述
//
// 异常:
//   - 无
//
// 使用示例:
//
//	var apiErr *pushplus.APIError
//	if errors.As(err, &apiErr) {
//	    log.Printf("推送失败，业务码=%d", apiErr.Code)
//	}
func (e *APIError) Error() string {
	return fmt.Sprintf("pushplus: 接口返回失败 code=%d msg=%s", e.Code, e.Msg)
}

// Client PushPlus 推送客户端，创建后字段只读，可被多个 goroutine 并发调用。
type Client struct {
	token    string
	topic    string
	channel  string
	template string
	baseURL  string
	http     *http.Client
}

// New 创建 PushPlus 推送客户端。
//
// 参数:
//   - conf: 推送配置。Token 必填；Topic、Channel、Template、Timeout、BaseURL
//     为空时分别回落为无群组、wechat、html、10s、DefaultBaseURL。
//
// 返回值:
//   - *Client: 可复用的客户端实例，内部持有一个带超时的 http.Client。
//
// 异常:
//   - 不返回错误；Token 为空的问题在调用 Send / SendContext 时才会暴露。
//
// 使用示例:
//
//	client := pushplus.New(jcbaseGo.PushPlusStruct{
//	    Token: "your-token",
//	    Topic: "daily-report",
//	})
//	resp, err := client.Send(pushplus.Message{
//	    Title:   "部署完成",
//	    Content: "v1.0.0 已上线",
//	})
func New(conf jcbaseGo.PushPlusStruct) *Client {
	timeout := conf.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	baseURL := strings.TrimSpace(conf.BaseURL)
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	channel := strings.TrimSpace(conf.Channel)
	if channel == "" {
		channel = ChannelWeChat
	}

	template := strings.TrimSpace(conf.Template)
	if template == "" {
		template = TemplateHTML
	}

	return &Client{
		token:    strings.TrimSpace(conf.Token),
		topic:    strings.TrimSpace(conf.Topic),
		channel:  channel,
		template: template,
		baseURL:  baseURL,
		http:     &http.Client{Timeout: timeout},
	}
}

// payload /send 接口的请求体
type payload struct {
	Token       string `json:"token"`
	Title       string `json:"title,omitempty"`
	Content     string `json:"content"`
	Template    string `json:"template,omitempty"`
	Channel     string `json:"channel,omitempty"`
	Topic       string `json:"topic,omitempty"`
	Webhook     string `json:"webhook,omitempty"`
	CallbackURL string `json:"callbackUrl,omitempty"`
}

// Send 使用默认上下文发送一条推送消息。
//
// 参数:
//   - msg: 推送消息，为空字段回落到客户端默认值。
//
// 返回值:
//   - *Response: 接口返回的业务信息；业务失败时也会返回，便于读取 code 与 msg。
//   - error: 请求失败、响应解析失败或业务 code 非 200 时返回；业务失败可通过
//     errors.As 提取 *APIError。
//
// 异常:
//   - 不 panic；所有失败均以 error 返回。
//
// 使用示例:
//
//	resp, err := client.Send(pushplus.Message{
//	    Title:   "告警",
//	    Content: "磁盘使用率超过 90%",
//	})
//
// 注意:
//   - code=200 仅表示 PushPlus 已受理，实际投递是异步的，不代表已送达。
func (c *Client) Send(msg Message) (*Response, error) {
	return c.SendContext(context.Background(), msg)
}

// SendContext 在指定上下文中发送一条推送消息，可通过 ctx 取消请求。
//
// 参数:
//   - ctx: 请求上下文，不能为 nil。
//   - msg: 推送消息，为空字段回落到客户端默认值。
//
// 返回值:
//   - *Response: 接口返回的业务信息；业务失败时也会返回，便于读取 code 与 msg。
//   - error: 上下文取消、请求失败、响应解析失败或业务 code 非 200 时返回；
//     业务失败可通过 errors.As 提取 *APIError。
//
// 异常:
//   - 不 panic；ctx 为 nil 时由 net/http 触发 panic，调用方须自行保证非 nil。
//
// 使用示例:
//
//	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
//	defer cancel()
//	_, err := client.SendContext(ctx, pushplus.Message{Content: "任务完成"})
func (c *Client) SendContext(ctx context.Context, msg Message) (*Response, error) {
	if c.token == "" {
		return nil, fmt.Errorf("pushplus: token 不能为空")
	}

	body, err := json.Marshal(c.buildPayload(msg))
	if err != nil {
		return nil, fmt.Errorf("pushplus: 序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("pushplus: 创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pushplus: 请求发送失败: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("pushplus: 读取响应失败: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("pushplus: HTTP 状态码异常 %d: %s", resp.StatusCode, snippet(raw))
	}

	var result Response
	if err = json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("pushplus: 解析响应失败: %w, 响应: %s", err, snippet(raw))
	}
	if result.Code != http.StatusOK {
		return &result, &APIError{Code: result.Code, Msg: result.Msg}
	}

	return &result, nil
}

// buildPayload 合并消息字段与客户端默认值，生成接口请求体。
//
// 参数:
//   - msg: 推送消息。
//
// 返回值:
//   - payload: 已完成默认值回落的请求体。
//
// 异常:
//   - 无
//
// 使用示例:
//
//	p := client.buildPayload(pushplus.Message{Content: "hi"})
func (c *Client) buildPayload(msg Message) payload {
	return payload{
		Token:       c.token,
		Title:       strings.TrimSpace(msg.Title),
		Content:     msg.Content,
		Template:    orDefault(strings.TrimSpace(msg.Template), c.template),
		Channel:     orDefault(strings.TrimSpace(msg.Channel), c.channel),
		Topic:       orDefault(strings.TrimSpace(msg.Topic), c.topic),
		Webhook:     strings.TrimSpace(msg.Webhook),
		CallbackURL: strings.TrimSpace(msg.CallbackURL),
	}
}

// orDefault 当 value 为空时返回 fallback。
//
// 参数:
//   - value: 调用方传入的值
//   - fallback: 空值时使用的默认值
//
// 返回值:
//   - string: value 非空时返回 value，否则返回 fallback
//
// 异常:
//   - 无
//
// 使用示例:
//
//	channel := orDefault("", "wechat") // "wechat"
func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// snippet 截断响应内容，供错误信息使用。
//
// 参数:
//   - raw: 原始响应字节
//
// 返回值:
//   - string: 不超过 maxErrorSnippet 字节的响应文本，超长时追加 "..."
//
// 异常:
//   - 无
//
// 使用示例:
//
//	detail := snippet([]byte("<html>...</html>"))
func snippet(raw []byte) string {
	if len(raw) <= maxErrorSnippet {
		return string(raw)
	}
	return string(raw[:maxErrorSnippet]) + "..."
}
