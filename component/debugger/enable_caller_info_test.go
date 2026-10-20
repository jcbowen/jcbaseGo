package debugger

// EnableCallerInfo 三态语义回归测试。
//
// 背景：该字段原为 bool + `default:"true"`，CheckAndSetDefault 的布尔默认值机制
// 会把显式设置的 false 强制改回 true（无法区分「未配置」与「显式关闭」），
// 导致「关闭调用位置信息」配置失效。
// 现改为 *bool：nil（JSON 未配置）默认开启；显式 false 生效关闭。本文件锁定该行为。

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// newCallerInfoTestDebugger 创建调试器并注册会写请求级日志的路由
func newCallerInfoTestDebugger(t *testing.T, cfg *Config) (*Debugger, *MemoryStorage, *gin.Engine) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	store, err := NewMemoryStorage(100)
	assert.NoError(t, err)
	cfg.Storage = store
	dbg, err := New(cfg)
	assert.NoError(t, err)

	r := gin.New()
	r.Use(dbg.Middleware())
	r.GET("/api/log", func(c *gin.Context) {
		GetLoggerFromContext(c).Info("测试调用位置信息")
		c.String(200, "ok")
	})
	return dbg, store, r
}

// findLogEntry 返回存储中 URL 匹配的首条记录
func findLogEntry(t *testing.T, store *MemoryStorage, url string) *LogEntry {
	t.Helper()
	records, _, err := store.FindAll(1, 100, nil)
	assert.NoError(t, err)
	for _, rec := range records {
		if rec.URL == url {
			return rec
		}
	}
	return nil
}

// TestEnableCallerInfo_TriStateViaJSON 通过 JSON 配置验证三态语义
func TestEnableCallerInfo_TriStateViaJSON(t *testing.T) {
	cases := []struct {
		name        string
		jsonConfig  string
		wantEnabled bool // New 解析后的 EnableCallerInfo 期望值
	}{
		{"未配置该键_默认开启", `{"enabled":true}`, true},
		{"显式true_保持开启", `{"enabled":true,"enable_caller_info":true}`, true},
		{"显式false_生效关闭", `{"enabled":true,"enable_caller_info":false}`, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{}
			assert.NoError(t, json.Unmarshal([]byte(tc.jsonConfig), cfg))

			dbg, _, _ := newCallerInfoTestDebugger(t, cfg)

			// New 之后指针必须非 nil（nil 已归一化），且值符合预期
			assert.NotNil(t, dbg.config.EnableCallerInfo)
			assert.Equal(t, tc.wantEnabled, *dbg.config.EnableCallerInfo)
		})
	}
}

// TestEnableCallerInfo_EffectOnLogs 验证开关对日志调用位置信息的实际效果
func TestEnableCallerInfo_EffectOnLogs(t *testing.T) {
	t.Run("默认开启_日志携带调用位置", func(t *testing.T) {
		dbg, store, r := newCallerInfoTestDebugger(t, &Config{Enabled: true, Level: LevelInfo})

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/api/log", nil))
		assert.Equal(t, 200, w.Code)

		entry := findLogEntry(t, store, "/api/log")
		if assert.NotNil(t, entry) && assert.NotEmpty(t, entry.LoggerLogs) {
			assert.NotEmpty(t, entry.LoggerLogs[0].FileName, "开启时日志应携带文件名")
			assert.NotZero(t, entry.LoggerLogs[0].Line, "开启时日志应携带行号")
		}
		_ = dbg
	})

	t.Run("显式false_日志无调用位置", func(t *testing.T) {
		callerInfo := false
		_, store, r := newCallerInfoTestDebugger(t, &Config{
			Enabled:          true,
			Level:            LevelInfo, // Info 级别日志才不会被默认 LevelWarn 过滤
			EnableCallerInfo: &callerInfo,
		})

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/api/log", nil))
		assert.Equal(t, 200, w.Code)

		entry := findLogEntry(t, store, "/api/log")
		if assert.NotNil(t, entry) && assert.NotEmpty(t, entry.LoggerLogs) {
			assert.Empty(t, entry.LoggerLogs[0].FileName, "显式关闭后日志不应携带文件名")
			assert.Zero(t, entry.LoggerLogs[0].Line, "显式关闭后日志不应携带行号")
		}
	})
}
