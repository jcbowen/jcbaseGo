package middleware

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// gpcProbe 承载一次请求的观测结果
type gpcProbe struct {
	data       map[string]any // GPC["data"]
	raw        string         // 业务 handler 读到的原始请求体
	rawLen     int
	handlerRan bool
	gpcSet     bool
	panicked   any
}

// runGPC 用 SetGPC 中间件处理一次 POST 请求，并在业务 handler 里重新读取原始请求体
// 参数：
//   - t: 测试上下文
//   - contentType: 请求的 Content-Type，空表示不设置
//   - body: 请求体内容
//   - chunked: 是否按 chunked 传输（ContentLength 置为 -1）
//
// 返回值：
//   - *gpcProbe: 观测结果；业务 handler 里没有 panic 时 panicked 为 nil
//
// 异常：捕获业务 handler 的 panic 而非让测试进程崩溃，便于断言 GPC 缺失场景
func runGPC(t *testing.T, contentType, body string, chunked bool) *gpcProbe {
	t.Helper()

	e := gin.New()
	e.Use(Base{}.SetGPC(e))

	probe := &gpcProbe{}
	e.POST("/x", func(c *gin.Context) {
		probe.handlerRan = true
		defer func() { probe.panicked = recover() }()

		value, ok := c.Get("GPC")
		probe.gpcSet = ok
		if partitions, isMap := value.(map[string]map[string]any); isMap {
			probe.data = partitions["data"]
		}

		data, err := c.GetRawData()
		if err != nil {
			t.Errorf("业务侧读取请求体失败: %v", err)
		}
		probe.rawLen = len(data)
		probe.raw = string(data)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if chunked {
		req.ContentLength = -1
		req.Header.Del("Content-Length")
	}
	e.ServeHTTP(httptest.NewRecorder(), req)

	return probe
}

// TestSetGPCBodyRemainsReadable 验证四类 Content-Type 下业务 handler 都能读到完整原始请求体，
// 且 GPC["data"] 的解析结果符合既有约定（[] 后缀保留数组，其余取首值）
func TestSetGPCBodyRemainsReadable(t *testing.T) {
	jsonBody := `{"name":"a","tag[]":["x","y"]}`
	formBody := "name=a&tag[]=x&tag[]=y"
	xmlBody := "<xml><ToUserName>gh</ToUserName><Content>hi</Content></xml>"

	cases := []struct {
		name        string
		contentType string
		body        string
		wantData    map[string]any
	}{
		// JSON 的键名原样保留，[] 约定只作用于 query 与表单类参数
		{"json", "application/json", jsonBody, map[string]any{"name": "a", "tag[]": []any{"x", "y"}}},
		{"form", "application/x-www-form-urlencoded", formBody, map[string]any{"name": "a", "tag": []string{"x", "y"}}},
		{"xml", "application/xml", xmlBody, map[string]any{"ToUserName": "gh", "Content": "hi"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			probe := runGPC(t, tc.contentType, tc.body, false)

			if probe.raw != tc.body {
				t.Errorf("业务侧读到的请求体 = %q, want %q", probe.raw, tc.body)
			}
			if !reflect.DeepEqual(probe.data, tc.wantData) {
				t.Errorf("GPC.data = %#v, want %#v", probe.data, tc.wantData)
			}
		})
	}
}

// TestSetGPCMultipartKeepsFormUsable 验证 multipart 解析后表单字段与文件仍可用
// multipart 不做原文回灌属既定设计：文件字节无法二次当流读，回灌要整块缓冲上传内容
func TestSetGPCMultipartKeepsFormUsable(t *testing.T) {
	e := gin.New()
	e.MaxMultipartMemory = 32 << 20
	e.Use(Base{}.SetGPC(e))

	var data map[string]any
	var postForm string
	var fileHeader *multipart.FileHeader
	e.POST("/x", func(c *gin.Context) {
		if partitions, ok := c.Get("GPC"); ok {
			if m, isMap := partitions.(map[string]map[string]any); isMap {
				data = m["data"]
			}
		}
		postForm = c.PostForm("name")
		fileHeader, _ = c.FormFile("upload")
		c.Status(http.StatusOK)
	})

	body := &strings.Builder{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("name", "a")
	part, err := writer.CreateFormFile("upload", "hello.txt")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := io.WriteString(part, "file-content"); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(body.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	e.ServeHTTP(httptest.NewRecorder(), req)

	if postForm != "a" {
		t.Errorf("业务侧 c.PostForm = %q, want %q", postForm, "a")
	}
	if fileHeader == nil || fileHeader.Filename != "hello.txt" {
		t.Errorf("业务侧 c.FormFile 拿不到文件: %+v", fileHeader)
	}
	if got, ok := data["name"].(string); !ok || got != "a" {
		t.Errorf("GPC.data[name] = %#v, want %q", data["name"], "a")
	}
}

// TestSetGPCChunkedBody 验证 chunked 传输（ContentLength == -1）也能正常解析
// 修复前 JSON/XML 分支因 ContentLength > 0 判断被整体跳过
func TestSetGPCChunkedBody(t *testing.T) {
	body := `{"name":"a"}`

	for _, contentType := range []string{"application/json", "application/xml"} {
		t.Run(contentType, func(t *testing.T) {
			probe := runGPC(t, contentType, body, true)

			if probe.raw != body {
				t.Errorf("业务侧读到的请求体 = %q, want %q", probe.raw, body)
			}
			if contentType == "application/json" && probe.data["name"] != "a" {
				t.Errorf("GPC.data = %#v, want name=a", probe.data)
			}
		})
	}
}

// TestSetGPCInvalidInputStillSetsContext 验证异常请求体下 GPC 依旧完整写入、业务 handler 正常执行
// 修复前 form/multipart 解析失败会跳过 c.Set("GPC")，下游 c.MustGet("GPC") 直接 panic
func TestSetGPCInvalidInputStillSetsContext(t *testing.T) {
	cases := []struct {
		name        string
		contentType string
		body        string
	}{
		{"json数组", "application/json", `[1,2,3]`},
		{"json标量", "application/json", `"hello"`},
		{"json非法", "application/json", `{bad`},
		{"form坏转义", "application/x-www-form-urlencoded", `a=%%zz`},
		{"multipart边界错", "multipart/form-data; boundary=none", `garbage`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			probe := runGPC(t, tc.contentType, tc.body, false)

			if !probe.handlerRan {
				t.Fatal("业务 handler 未执行")
			}
			if !probe.gpcSet {
				t.Fatal("GPC 未写入上下文")
			}
			if probe.panicked != nil {
				t.Fatalf("业务 handler panic: %v", probe.panicked)
			}
			if len(probe.data) != 0 {
				t.Errorf("GPC.data 应为空, got %#v", probe.data)
			}
		})
	}
}

// TestSetGPCFormKeepsGinBinding 验证 form 分支改为自行解析后，gin 的表单绑定仍能取到值
// 前提：请求体已回灌，binding 内部首次 ParseForm 才能读到字节
func TestSetGPCFormKeepsGinBinding(t *testing.T) {
	e := gin.New()
	e.Use(Base{}.SetGPC(e))

	var bound struct {
		Name string `form:"name"`
		Age  string `form:"age"`
	}
	var bindErr error
	e.POST("/x", func(c *gin.Context) {
		bindErr = c.ShouldBind(&bound)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader("name=a&age=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	e.ServeHTTP(httptest.NewRecorder(), req)

	if bindErr != nil {
		t.Fatalf("ShouldBind: %v", bindErr)
	}
	if bound.Name != "a" || bound.Age != "1" {
		t.Errorf("绑定结果 = %+v, want {Name:a Age:1}", bound)
	}
}

// TestSetGPCBodyOverLimit 验证超过上限的请求体只回灌不解析：GPC.data 为空，
// 但下游仍能读到完整字节，内存不会随请求体无限增长
func TestSetGPCBodyOverLimit(t *testing.T) {
	body := "name=" + strings.Repeat("x", int(gpcBodyParseLimit))

	probe := runGPC(t, "application/x-www-form-urlencoded", body, false)

	if len(probe.data) != 0 {
		t.Errorf("超限请求体不应被解析, GPC.data 长度 = %d", len(probe.data))
	}
	if probe.rawLen != len(body) {
		t.Errorf("下游读到的请求体长度 = %d, want %d", probe.rawLen, len(body))
	}
}
