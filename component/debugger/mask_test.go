package debugger

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// newMaskTester 构造仅用于掩码逻辑校验的调试器实例
// 走 New 是为了让配置归一化（*bool 三态）与掩码键集合构建的真实路径参与测试
func newMaskTester(t *testing.T, config *Config) *Debugger {
	t.Helper()
	if config == nil {
		config = &Config{}
	}
	d, err := New(config)
	assert.NoError(t, err)
	return d
}

// TestMaskHeadersDefault 默认头名集合应被掩码，其余头保持不变
func TestMaskHeadersDefault(t *testing.T) {
	d := newMaskTester(t, nil)

	headers := map[string]string{
		"Authorization": "Bearer eyJhbGciOi",
		"Cookie":        "session_id=abc123",
		"Set-Cookie":    "token=xyz",
		"X-Api-Key":     "sk-123456",
		"Content-Type":  "application/json",
		"User-Agent":    "curl/8.0",
		"Empty-Header":  "",
	}

	masked := d.maskHeaders(headers)
	assert.Equal(t, maskPlaceholder, masked["Authorization"])
	assert.Equal(t, maskPlaceholder, masked["Cookie"])
	assert.Equal(t, maskPlaceholder, masked["Set-Cookie"])
	assert.Equal(t, maskPlaceholder, masked["X-Api-Key"])
	assert.Equal(t, "application/json", masked["Content-Type"], "非敏感头不应被改写")
	assert.Equal(t, "curl/8.0", masked["User-Agent"])
	assert.Empty(t, masked["Empty-Header"], "空值保持空值，不制造噪声")
}

// TestMaskHeadersExtraKeys 配置的 HeaderMaskKeys 追加到默认集合而非覆盖
func TestMaskHeadersExtraKeys(t *testing.T) {
	d := newMaskTester(t, &Config{HeaderMaskKeys: []string{" X-Custom-Auth "}})

	headers := map[string]string{"Authorization": "Bearer t", "x-custom-auth": "secret", "Accept": "*/*"}
	masked := d.maskHeaders(headers)
	assert.Equal(t, maskPlaceholder, masked["x-custom-auth"], "追加键（大小写不敏感）应命中")
	assert.Equal(t, maskPlaceholder, masked["Authorization"], "默认键不因追加配置而失效")
	assert.Equal(t, "*/*", masked["Accept"])
}

// TestMaskQueryParams 查询参数中的手机号与验证码应被掩码
func TestMaskQueryParams(t *testing.T) {
	d := newMaskTester(t, nil)

	params := d.maskQueryParams(map[string]string{
		"phone":         "13800138000",
		"sms_code":      "8463",
		"page":          "2",
		"PASSWORD":      "plain",
		"verification_": "",
	})
	assert.Equal(t, maskPlaceholder, params["phone"])
	assert.Equal(t, maskPlaceholder, params["sms_code"])
	assert.Equal(t, maskPlaceholder, params["PASSWORD"], "键名大小写不敏感")
	assert.Equal(t, "2", params["page"])
	assert.Empty(t, params["verification_"])
}

// TestMaskBodyJSON 请求体 JSON 掩码应保持结构与格式
func TestMaskBodyJSON(t *testing.T) {
	d := newMaskTester(t, nil)

	body := `{"phone":"13800138000","password":"p@ss","code":0,"data":{"access_token":"eyJhb","sms_code":8463,"name":"张三"}}`
	masked := d.maskBody(body, "application/json")

	assert.NotContains(t, masked, "13800138000")
	assert.NotContains(t, masked, "p@ss")
	assert.NotContains(t, masked, "eyJhb")
	assert.NotContains(t, masked, "8463")
	assert.Contains(t, masked, `"code":0`, "响应信封的业务状态码 code 不应被掩码")
	assert.Contains(t, masked, `"name":"张三"`)
	assert.True(t, strings.HasPrefix(masked, `{"phone":"***"`), "掩码只替换值，键与顺序保持原样: "+masked)
}

// TestMaskBodyJSONPrettyFormat 多行缩进 JSON 掩码后不应重排格式
func TestMaskBodyJSONPrettyFormat(t *testing.T) {
	d := newMaskTester(t, nil)

	body := "{\n  \"new_password\": \"hunter2\",\n  \"name\": \"jc\"\n}"
	masked := d.maskBody(body, "application/json")
	assert.Equal(t, "{\n  \"new_password\": \"***\",\n  \"name\": \"jc\"\n}", masked)
}

// TestMaskBodyJSONTruncatedTail 捕获窗口截断导致值未闭合时仍须掩码
func TestMaskBodyJSONTruncatedTail(t *testing.T) {
	d := newMaskTester(t, nil)

	body := `{"old_password":"abc123def456","phone":"138001` // 模拟截断：最后一个值未闭合
	masked := d.maskBody(body, "application/json")
	assert.NotContains(t, masked, "abc123def456")
	assert.NotContains(t, masked, "138001", "截断残留的凭据片段也应掩掉")
}

// TestMaskBodyForm form-urlencoded 掩码应保持原有顺序与不成对片段
func TestMaskBodyForm(t *testing.T) {
	d := newMaskTester(t, nil)

	body := "z=1&password=hunter2&phone=13800138000&flag=&name=jc%20bowen"
	masked := d.maskBody(body, "application/x-www-form-urlencoded")
	assert.Equal(t, "z=1&password=***&phone=***&flag=&name=jc%20bowen", masked)
}

// TestMaskBodyFormWithoutContentType 缺失 Content-Type 时按文本形态识别 form
func TestMaskBodyFormWithoutContentType(t *testing.T) {
	d := newMaskTester(t, nil)

	masked := d.maskBody("token=eyJhbGciOiJBUU5JTg&user=jc", "")
	assert.Equal(t, "token=***&user=jc", masked, "无类型但形似 form 的正文应走 form 掩码")
}

// TestMaskBodyMultipart multipart 摘要行中的字段值应被掩码，文件元数据保持不变
func TestMaskBodyMultipart(t *testing.T) {
	d := newMaskTester(t, nil)

	body := "------boundary\n" +
		"  [Field] Name: password, Size: 12 bytes, Content: hunter2\n" +
		"  [Field] Name: sms_code, Size: 4 bytes, Value: 8463\n" +
		"  [Field] Name: remark, Size: 5 bytes, Content: hello\n" +
		"  [File] Name: avatar.png, Field: file, Size: 20480 bytes, Type: image/png\n"
	masked := d.maskBody(body, "multipart/form-data; boundary=----boundary")

	assert.NotContains(t, masked, "hunter2")
	assert.NotContains(t, masked, "8463")
	assert.Contains(t, masked, "Content: hello", "非敏感字段值保持原样")
	assert.Contains(t, masked, "[File] Name: avatar.png", "文件元数据行不受影响")
	assert.Contains(t, masked, "Name: password", "字段名保留，便于确认提交了哪些字段")
}

// TestMaskBodyDisabled 显式关闭掩码后应保留原文（仅用于本地一次性排查）
func TestMaskBodyDisabled(t *testing.T) {
	disabled := false
	d := newMaskTester(t, &Config{BodyMaskEnabled: &disabled})

	body := `{"password":"hunter2"}`
	assert.Equal(t, body, d.maskBody(body, "application/json"))

	headers := d.maskHeaders(map[string]string{"Authorization": "Bearer t"})
	assert.Equal(t, "Bearer t", headers["Authorization"])

	params := d.maskQueryParams(map[string]string{"password": "hunter2"})
	assert.Equal(t, "hunter2", params["password"])
}

// TestMaskBodyBinarySafe 二进制与乱码正文不应被正则改写
func TestMaskBodyBinarySafe(t *testing.T) {
	d := newMaskTester(t, nil)

	body := "[Binary data: 1024 bytes, type: application/octet-stream]"
	assert.Equal(t, body, d.maskBody(body, "application/octet-stream"))
	assert.Equal(t, "", d.maskBody("", "application/json"), "空正文直接返回")
}

// TestMaskBodyExtraBodyKeys 配置的 BodyMaskKeys 追加默认字段集合
func TestMaskBodyExtraBodyKeys(t *testing.T) {
	d := newMaskTester(t, &Config{BodyMaskKeys: []string{"inner_secret"}})

	masked := d.maskBody(`{"inner_secret":"v1","keep":"v2"}`, "application/json")
	assert.Contains(t, masked, `"inner_secret":"***"`)
	assert.Contains(t, masked, `"keep":"v2"`)
}

// TestBuildMaskKeySetNormalization 掩码键集合应小写归一化并忽略空白项
func TestBuildMaskKeySetNormalization(t *testing.T) {
	set := buildMaskKeySet([]string{"A"}, []string{" B ", "", "  "})
	_, hasA := set["a"]
	_, hasB := set["b"]
	assert.True(t, hasA)
	assert.True(t, hasB)
	assert.Len(t, set, 2, "空白键名不应入集合")
}

// TestMaskThroughMiddlewareEndToEnd 走完整中间件链路验证落库条目不含凭据原文
// 覆盖请求头、查询参数、请求体、响应头、响应体五个入口，防止后续新增写入点绕过掩码
func TestMaskThroughMiddlewareEndToEnd(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storage, err := NewMemoryStorage(100)
	assert.NoError(t, err)

	dbg, err := New(&Config{Enabled: true, SampleRate: 1.0, MaxRecords: 100, Storage: storage})
	assert.NoError(t, err)

	router := gin.New()
	router.Use(dbg.Middleware())
	router.POST("/api/login", func(c *gin.Context) {
		c.Header("Set-Cookie", "session_id=sid-secret-value")
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{"access_token": "eyJhbGciOi.secret", "refresh_token": "rt-1234", "uid": 7},
		})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/login?phone=13800138000&sms_code=8463",
		strings.NewReader(`{"phone":"13800138000","password":"hunter2","remember":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer jwt-value")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	entries, _, err := storage.FindAll(1, 10, nil)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, maskPlaceholder, entry.RequestHeaders["Authorization"])
	assert.Equal(t, maskPlaceholder, entry.ResponseHeaders["Set-Cookie"])
	assert.Equal(t, maskPlaceholder, entry.QueryParams["phone"])
	assert.Equal(t, maskPlaceholder, entry.QueryParams["sms_code"])
	assert.NotContains(t, entry.RequestBody, "hunter2")
	assert.NotContains(t, entry.RequestBody, "13800138000")
	assert.Contains(t, entry.RequestBody, `"remember":true`, "非敏感字段应完整保留，日志仍可用于排查")
	assert.NotContains(t, entry.ResponseBody, "eyJhbGciOi")
	assert.NotContains(t, entry.ResponseBody, "rt-1234")
	assert.Contains(t, entry.ResponseBody, `"uid":7`, "业务数据不应被误掩")
	assert.Contains(t, entry.ResponseBody, `"code":0`, "响应信封状态码不应被误掩")
}

// TestMaskDisabledKeepsRawEndToEnd 显式关闭掩码后条目保留原文（本地一次性排查场景）
func TestMaskDisabledKeepsRawEndToEnd(t *testing.T) {
	gin.SetMode(gin.TestMode)

	disabled := false
	storage, err := NewMemoryStorage(100)
	assert.NoError(t, err)

	dbg, err := New(&Config{Enabled: true, SampleRate: 1.0, MaxRecords: 100, Storage: storage, BodyMaskEnabled: &disabled})
	assert.NoError(t, err)

	router := gin.New()
	router.Use(dbg.Middleware())
	router.POST("/api/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"access_token": "eyJhb"}})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"password":"hunter2"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	entries, _, err := storage.FindAll(1, 10, nil)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Contains(t, entries[0].RequestBody, "hunter2")
	assert.Contains(t, entries[0].ResponseBody, "eyJhb")
}
