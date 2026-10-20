package debugger

// AutoSkipStatic 三态语义回归测试。
//
// 背景：该字段原为 bool + `default:"true"`，CheckAndSetDefault 的布尔默认值机制
// 会把显式设置的 false 强制改回 true（无法区分「未配置」与「显式关闭」），
// 导致「关闭自动跳过静态资源」配置失效。
// 现改为 *bool：nil（JSON 未配置）默认开启；显式 false 生效关闭。本文件锁定该行为。

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// newAutoSkipTestDebugger 创建调试器并注册静态与API路由，返回调试器与内存存储
func newAutoSkipTestDebugger(t *testing.T, cfg *Config) (*Debugger, *MemoryStorage, *gin.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	store, err := NewMemoryStorage(100)
	assert.NoError(t, err)
	cfg.Storage = store
	dbg, err := New(cfg)
	assert.NoError(t, err)

	r := gin.New()
	r.Use(dbg.Middleware())
	r.GET("/api/test", func(c *gin.Context) { c.String(200, "ok") })
	r.GET("/static/*filepath", func(c *gin.Context) { c.String(200, "body{}") })
	return dbg, store, r
}

// countRecordedURLs 返回存储中匹配指定URL的记录数
func countRecordedURLs(t *testing.T, store *MemoryStorage, url string) int {
	t.Helper()
	records, _, err := store.FindAll(1, 100, nil)
	assert.NoError(t, err)
	count := 0
	for _, rec := range records {
		if rec.URL == url {
			count++
		}
	}
	return count
}

// TestAutoSkipStatic_TriStateViaJSON 通过 JSON 配置验证三态语义
func TestAutoSkipStatic_TriStateViaJSON(t *testing.T) {
	cases := []struct {
		name        string
		jsonConfig  string
		wantEnabled bool // New 解析后的 AutoSkipStatic 期望值
	}{
		{"未配置该键_默认开启", `{"enabled":true}`, true},
		{"显式true_保持开启", `{"enabled":true,"auto_skip_static":true}`, true},
		{"显式false_生效关闭", `{"enabled":true,"auto_skip_static":false}`, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{}
			assert.NoError(t, json.Unmarshal([]byte(tc.jsonConfig), cfg))

			dbg, store, r := newAutoSkipTestDebugger(t, cfg)

			// New 之后指针必须非 nil（nil 已归一化），且值符合预期
			assert.NotNil(t, dbg.config.AutoSkipStatic)
			assert.Equal(t, tc.wantEnabled, *dbg.config.AutoSkipStatic)

			// 发起静态请求与API请求
			w1 := httptest.NewRecorder()
			r.ServeHTTP(w1, httptest.NewRequest("GET", "/static/app.css", nil))
			assert.Equal(t, 200, w1.Code)
			w2 := httptest.NewRecorder()
			r.ServeHTTP(w2, httptest.NewRequest("GET", "/api/test", nil))
			assert.Equal(t, 200, w2.Code)

			// 开启时静态请求被跳过，关闭时静态请求被记录；API 请求始终记录
			assert.Equal(t, 1, countRecordedURLs(t, store, "/api/test"), "/api/test 应始终被记录")
			wantStatic := 0
			if !tc.wantEnabled {
				wantStatic = 1
			}
			assert.Equal(t, wantStatic, countRecordedURLs(t, store, "/static/app.css"),
				"AutoSkipStatic=%v 时静态请求的记录数不符", tc.wantEnabled)
		})
	}
}

// TestAutoSkipStatic_ExplicitFalseInGo Go 结构体显式 false 同样生效
func TestAutoSkipStatic_ExplicitFalseInGo(t *testing.T) {
	autoSkip := false
	dbg, store, r := newAutoSkipTestDebugger(t, &Config{
		Enabled:        true,
		AutoSkipStatic: &autoSkip,
	})

	assert.NotNil(t, dbg.config.AutoSkipStatic)
	assert.False(t, *dbg.config.AutoSkipStatic)

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest("GET", "/static/app.css", nil))
	assert.Equal(t, 1, countRecordedURLs(t, store, "/static/app.css"),
		"显式关闭后静态请求应被记录")
}
