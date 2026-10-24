package pushplus

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jcbowen/jcbaseGo"
)

// newTestServer 启动一个测试服务器，测试结束时自动关闭。
func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// TestNewDefaults 验证 New 会为渠道、模板、接口地址、超时填充默认值，并裁剪 token 空格。
func TestNewDefaults(t *testing.T) {
	c := New(jcbaseGo.PushPlusStruct{Token: " tk "})

	if c.token != "tk" {
		t.Errorf("token = %q, want %q", c.token, "tk")
	}
	if c.channel != ChannelWeChat {
		t.Errorf("channel = %q, want %q", c.channel, ChannelWeChat)
	}
	if c.template != TemplateHTML {
		t.Errorf("template = %q, want %q", c.template, TemplateHTML)
	}
	if c.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, DefaultBaseURL)
	}
	if c.http.Timeout != DefaultTimeout {
		t.Errorf("timeout = %v, want %v", c.http.Timeout, DefaultTimeout)
	}
}

// TestNewCustomValues 验证 New 保留调用方显式配置的超时、接口地址等值。
func TestNewCustomValues(t *testing.T) {
	c := New(jcbaseGo.PushPlusStruct{
		Token:    "tk",
		Topic:    "dev",
		Channel:  ChannelWebhook,
		Template: TemplateMarkdown,
		Timeout:  3 * time.Second,
		BaseURL:  "http://127.0.0.1:1/send",
	})

	if c.topic != "dev" || c.channel != ChannelWebhook || c.template != TemplateMarkdown {
		t.Errorf("自定义字段未保留: topic=%q channel=%q template=%q", c.topic, c.channel, c.template)
	}
	if c.http.Timeout != 3*time.Second {
		t.Errorf("timeout = %v, want 3s", c.http.Timeout)
	}
	if c.baseURL != "http://127.0.0.1:1/send" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
}

// TestSendMergesDefaults 验证发送时消息字段回落到客户端默认值，且请求方法与内容类型正确。
func TestSendMergesDefaults(t *testing.T) {
	got := make(chan payload, 1)
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		var p payload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		got <- p

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"msg":"请求成功","data":"msg-123"}`))
	})

	client := New(jcbaseGo.PushPlusStruct{
		Token:   "tk",
		Topic:   "dev",
		BaseURL: srv.URL,
	})
	resp, err := client.Send(Message{Title: " 部署完成 ", Content: "v1.0.0 已上线"})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	p := <-got
	if p.Token != "tk" {
		t.Errorf("token = %q, want tk", p.Token)
	}
	if p.Title != "部署完成" {
		t.Errorf("title = %q, want 部署完成", p.Title)
	}
	if p.Content != "v1.0.0 已上线" {
		t.Errorf("content = %q", p.Content)
	}
	if p.Channel != ChannelWeChat {
		t.Errorf("channel = %q, want %q", p.Channel, ChannelWeChat)
	}
	if p.Template != TemplateHTML {
		t.Errorf("template = %q, want %q", p.Template, TemplateHTML)
	}
	if p.Topic != "dev" {
		t.Errorf("topic = %q, want dev", p.Topic)
	}
	if resp.Code != http.StatusOK || resp.Data != "msg-123" {
		t.Errorf("response = %+v", resp)
	}
}

// TestSendOverrides 验证消息级字段可以覆盖客户端默认的群组、渠道与模板。
func TestSendOverrides(t *testing.T) {
	got := make(chan payload, 1)
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var p payload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		got <- p
		_, _ = w.Write([]byte(`{"code":200,"msg":"请求成功","data":"msg-456"}`))
	})

	client := New(jcbaseGo.PushPlusStruct{
		Token:    "tk",
		Topic:    "default-topic",
		Channel:  ChannelWeChat,
		Template: TemplateHTML,
		BaseURL:  srv.URL,
	})
	_, err := client.Send(Message{
		Content:     "告警内容",
		Topic:       "ops",
		Channel:     ChannelWebhook,
		Template:    TemplateMarkdown,
		Webhook:     "https://example.com/hook",
		CallbackURL: "https://example.com/callback",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	p := <-got
	if p.Topic != "ops" || p.Channel != ChannelWebhook || p.Template != TemplateMarkdown {
		t.Errorf("覆盖失败: topic=%q channel=%q template=%q", p.Topic, p.Channel, p.Template)
	}
	if p.Webhook != "https://example.com/hook" || p.CallbackURL != "https://example.com/callback" {
		t.Errorf("webhook=%q callbackUrl=%q", p.Webhook, p.CallbackURL)
	}
}

// TestSendAPIError 验证业务 code 非 200 时返回 APIError，同时保留响应内容。
func TestSendAPIError(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":903,"msg":"token无效","data":null}`))
	})

	client := New(jcbaseGo.PushPlusStruct{Token: "tk", BaseURL: srv.URL})
	resp, err := client.Send(Message{Content: "hi"})
	if err == nil {
		t.Fatal("Send() error = nil, want APIError")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Code != 903 || apiErr.Msg != "token无效" {
		t.Errorf("apiErr = %+v", apiErr)
	}
	if resp == nil || resp.Code != 903 || resp.Msg != "token无效" {
		t.Errorf("response = %+v, want code 903", resp)
	}
	if !strings.Contains(err.Error(), "903") {
		t.Errorf("error = %q, want contain 903", err.Error())
	}
}

// TestSendHTTPError 验证 HTTP 状态码非 2xx 时返回错误。
func TestSendHTTPError(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server down", http.StatusInternalServerError)
	})

	client := New(jcbaseGo.PushPlusStruct{Token: "tk", BaseURL: srv.URL})
	_, err := client.Send(Message{Content: "hi"})
	if err == nil {
		t.Fatal("Send() error = nil, want HTTP error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %q, want contain 500", err.Error())
	}
}

// TestSendInvalidJSON 验证响应不是合法 JSON 时返回解析错误。
func TestSendInvalidJSON(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not-json"))
	})

	client := New(jcbaseGo.PushPlusStruct{Token: "tk", BaseURL: srv.URL})
	_, err := client.Send(Message{Content: "hi"})
	if err == nil {
		t.Fatal("Send() error = nil, want parse error")
	}
	if !strings.Contains(err.Error(), "解析响应失败") {
		t.Errorf("error = %q, want 解析响应失败", err.Error())
	}
}

// TestSendEmptyToken 验证 token 为空时直接返回错误且不发起请求。
func TestSendEmptyToken(t *testing.T) {
	var called atomic.Bool
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		called.Store(true)
		_, _ = w.Write([]byte(`{"code":200,"msg":"请求成功"}`))
	})

	client := New(jcbaseGo.PushPlusStruct{BaseURL: srv.URL})
	_, err := client.Send(Message{Content: "hi"})
	if err == nil || !strings.Contains(err.Error(), "token") {
		t.Fatalf("error = %v, want token 不能为空", err)
	}
	if called.Load() {
		t.Error("token 为空时不应发起 HTTP 请求")
	}
}

// TestSendTimeout 验证请求超过客户端超时时间时返回错误。
func TestSendTimeout(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte(`{"code":200,"msg":"请求成功"}`))
	})

	client := New(jcbaseGo.PushPlusStruct{
		Token:   "tk",
		BaseURL: srv.URL,
		Timeout: 10 * time.Millisecond,
	})
	_, err := client.Send(Message{Content: "hi"})
	if err == nil {
		t.Fatal("Send() error = nil, want timeout error")
	}
}

// TestSendContextCanceled 验证上下文取消时立即返回取消错误。
func TestSendContextCanceled(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"msg":"请求成功"}`))
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := New(jcbaseGo.PushPlusStruct{Token: "tk", BaseURL: srv.URL})
	_, err := client.SendContext(ctx, Message{Content: "hi"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}
