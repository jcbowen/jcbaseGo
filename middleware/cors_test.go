package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// newCorsTestEngine 构建带 CORS 中间件与一个业务路由的测试引擎
// 业务 handler 返回 200，便于断言实际请求是否穿过中间件
func newCorsTestEngine(t *testing.T, handler gin.HandlerFunc) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(handler)
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	return r
}

// doCorsRequest 发起请求并返回响应记录
func doCorsRequest(t *testing.T, r *gin.Engine, method, origin string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, "/ping", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// assertHeader 断言响应头等于期望值
func assertHeader(t *testing.T, w *httptest.ResponseRecorder, key, want string) {
	t.Helper()

	if got := w.Header().Get(key); got != want {
		t.Errorf("响应头 %s = %q, 期望 %q", key, got, want)
	}
}

// TestCorsLegacyAnyOrigin 验证历史 Cors() 行为：全放行且回显 Origin、携带凭证
func TestCorsLegacyAnyOrigin(t *testing.T) {
	r := newCorsTestEngine(t, Base{}.Cors())

	// 带具体 Origin：回显并允许凭证
	w := doCorsRequest(t, r, http.MethodGet, "https://any.example.com")
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, 期望 %d", w.Code, http.StatusOK)
	}
	assertHeader(t, w, "Access-Control-Allow-Origin", "https://any.example.com")
	assertHeader(t, w, "Access-Control-Allow-Credentials", "true")

	// 无 Origin（同源请求）：回 *
	w = doCorsRequest(t, r, http.MethodGet, "")
	assertHeader(t, w, "Access-Control-Allow-Origin", "*")
	assertHeader(t, w, "Access-Control-Allow-Credentials", "")
}

// TestCorsPreflightShortCircuit 验证 OPTIONS 预检被短路为 204 且带授权头
func TestCorsPreflightShortCircuit(t *testing.T) {
	r := newCorsTestEngine(t, Base{}.Cors())

	w := doCorsRequest(t, r, http.MethodOptions, "https://any.example.com")
	if w.Code != http.StatusNoContent {
		t.Fatalf("状态码 = %d, 期望 %d", w.Code, http.StatusNoContent)
	}
	assertHeader(t, w, "Access-Control-Allow-Origin", "https://any.example.com")
	assertHeader(t, w, "Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

	// 反射客户端申请的请求头
	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "https://any.example.com")
	req.Header.Set("Access-Control-Request-Headers", "X-Custom-Header")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assertHeader(t, rec, "Access-Control-Allow-Headers", "X-Custom-Header, "+corsDefaultAllowHeaders)
}

// TestCorsWithConfigWhitelistHit 验证白名单命中的来源被授予 CORS 头与凭证
func TestCorsWithConfigWhitelistHit(t *testing.T) {
	r := newCorsTestEngine(t, Base{}.CorsWithConfig(CorsConfig{
		AllowOrigins:     []string{"https://www.example.com"},
		AllowCredentials: true,
	}))

	w := doCorsRequest(t, r, http.MethodGet, "https://www.example.com")
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, 期望 %d", w.Code, http.StatusOK)
	}
	assertHeader(t, w, "Access-Control-Allow-Origin", "https://www.example.com")
	assertHeader(t, w, "Access-Control-Allow-Credentials", "true")
}

// TestCorsWithConfigWhitelistMiss 验证白名单未命中的来源不下发任何 CORS 授权头，
// 请求仍正常穿过中间件由浏览器拦截
func TestCorsWithConfigWhitelistMiss(t *testing.T) {
	r := newCorsTestEngine(t, Base{}.CorsWithConfig(CorsConfig{
		AllowOrigins:     []string{"https://www.example.com"},
		AllowCredentials: true,
	}))

	// 精确白名单未命中
	w := doCorsRequest(t, r, http.MethodGet, "https://evil.example.org")
	if w.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, 期望 %d", w.Code, http.StatusOK)
	}
	assertHeader(t, w, "Access-Control-Allow-Origin", "")
	assertHeader(t, w, "Access-Control-Allow-Credentials", "")

	// 未命中来源的预检：204 但不带授权头
	w = doCorsRequest(t, r, http.MethodOptions, "https://evil.example.org")
	if w.Code != http.StatusNoContent {
		t.Fatalf("预检状态码 = %d, 期望 %d", w.Code, http.StatusNoContent)
	}
	assertHeader(t, w, "Access-Control-Allow-Origin", "")

	// 空白名单：任何来源都不授予
	r = newCorsTestEngine(t, Base{}.CorsWithConfig(CorsConfig{}))
	w = doCorsRequest(t, r, http.MethodGet, "https://www.example.com")
	assertHeader(t, w, "Access-Control-Allow-Origin", "")
}

// TestCorsWithConfigWildcard 验证 *.example.com 子域通配符的匹配边界
func TestCorsWithConfigWildcard(t *testing.T) {
	r := newCorsTestEngine(t, Base{}.CorsWithConfig(CorsConfig{
		AllowOrigins: []string{"*.example.com"},
	}))

	cases := []struct {
		origin string
		want   bool
	}{
		{"https://api.example.com", true},       // 子域命中
		{"http://example.com", true},            // 主域命中（scheme 不限）
		{"https://a.b.example.com", true},       // 多级子域命中
		{"https://api.example.com:8443", true},  // 带端口的子域命中
		{"https://evil.com", false},             // 无关域未命中
		{"https://notexample.com", false},       // 后缀撞车未命中
		{"https://api.example.org", false},      // TLD 不同未命中
		{"https://example.com.evil.com", false}, // 伪造后缀未命中
	}

	for _, tc := range cases {
		w := doCorsRequest(t, r, http.MethodGet, tc.origin)
		got := w.Header().Get("Access-Control-Allow-Origin") == tc.origin
		if got != tc.want {
			t.Errorf("origin %q 命中结果 = %v, 期望 %v", tc.origin, got, tc.want)
		}
	}
}

// TestCorsVaryAddNotOverwrite 验证 Vary 用 add 语义，不覆盖上游中间件已写入的值
func TestCorsVaryAddNotOverwrite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 模拟上游中间件（如压缩）已写入 Vary
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Add("Vary", "Accept-Encoding")
		c.Next()
	})
	r.Use(Base{}.CorsWithConfig(CorsConfig{AllowAllOrigins: true}))
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	w := doCorsRequest(t, r, http.MethodGet, "https://any.example.com")

	// 上游的 Vary 在前，不被覆盖；中间件自身再补 3 个
	vary := w.Header().Values("Vary")
	if len(vary) != 4 {
		t.Fatalf("Vary 头数量 = %d, 期望 4, 实际值 %v", len(vary), vary)
	}
	if vary[0] != "Accept-Encoding" {
		t.Errorf("Vary[0] = %q, 期望 %q", vary[0], "Accept-Encoding")
	}
	if vary[1] != "Origin" {
		t.Errorf("Vary[1] = %q, 期望 %q", vary[1], "Origin")
	}
}
