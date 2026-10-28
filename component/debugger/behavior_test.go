package debugger

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// 本文件覆盖 J1–J3 引入的组件行为：清理口径与入参守卫、请求/响应体有界捕获、
// multipart 文件部件默认只存元数据、写库路径的主键 upsert 与上限均摊裁剪

// newCaptureDebugger 构造用于捕获行为验证的调试器（内存存储 + 默认掩码开启）
func newCaptureDebugger(t *testing.T, config *Config) (*Debugger, *MemoryStorage) {
	t.Helper()

	storage, err := NewMemoryStorage(100)
	assert.NoError(t, err)
	config.Storage = storage

	dbg, err := New(config)
	assert.NoError(t, err)
	return dbg, storage
}

// boolPtr 构造 *bool 三态配置项的测试值
func boolPtr(v bool) *bool {
	return &v
}

// TestIsCleanupBoundaryValid 清理时间点守卫：零值与未来时间都不允许清表
func TestIsCleanupBoundaryValid(t *testing.T) {
	assert.False(t, isCleanupBoundaryValid(time.Time{}), "零值时间会让 before 之前的判断变成全表命中")
	assert.False(t, isCleanupBoundaryValid(time.Now().Add(time.Hour)), "未来时间等同清空全部日志")
	assert.True(t, isCleanupBoundaryValid(time.Now().Add(-time.Hour)))
}

// TestMemoryStorageCleanupGuard 内存存储的过期清理：非法入参报错且不误删，合法入参只删过期项
func TestMemoryStorageCleanupGuard(t *testing.T) {
	storage, err := NewMemoryStorage(100)
	assert.NoError(t, err)

	now := time.Now()
	assert.NoError(t, storage.Save(&LogEntry{ID: "keep-1", Timestamp: now, Method: "GET", URL: "/keep"}))
	assert.NoError(t, storage.Save(&LogEntry{ID: "old-1", Timestamp: now.Add(-48 * time.Hour), Method: "GET", URL: "/old"}))

	assert.Error(t, storage.Cleanup(time.Time{}), "零值时间点必须拒绝")
	assert.Error(t, storage.Cleanup(now.Add(time.Hour)), "未来时间点必须拒绝")

	entries, total, err := storage.FindAll(1, 100, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, total, "非法清理入参不应删除任何记录")
	assert.Len(t, entries, 2)

	assert.NoError(t, storage.Cleanup(now.Add(-24*time.Hour)))
	_, total, err = storage.FindAll(1, 100, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, total, "只应清掉过期记录")
}

// TestNewNormalizesRetentionPeriod New 会把零值保留期归一化为默认 168h，
// 「不清理」的权威表达是调用方不启动清理协程，而不是留 0 给组件
func TestNewNormalizesRetentionPeriod(t *testing.T) {
	dbg, _ := newCaptureDebugger(t, &Config{Enabled: true, RetentionPeriod: 0})
	assert.Equal(t, 168*time.Hour, dbg.config.RetentionPeriod,
		"零值保留期应由 New 补为默认 7 天")
}

// TestStartPeriodicCleanupRetentionDisabled 保留期非正时不启动清理，绝不出现「清全表」。
// 这里绕过 New 的默认值归一化，直接构造该字段，验证方法自身的守卫
func TestStartPeriodicCleanupRetentionDisabled(t *testing.T) {
	dbg, storage := newCaptureDebugger(t, &Config{Enabled: true})
	dbg.config.RetentionPeriod = 0

	old := &LogEntry{ID: "old-1", Timestamp: time.Now().Add(-48 * time.Hour), Method: "GET", URL: "/old"}
	assert.NoError(t, storage.Save(old))

	stop := dbg.StartPeriodicCleanup()
	stop()
	stop() // stop 需可重复调用

	_, total, err := storage.FindAll(1, 100, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, total, "retention<=0 时不应启动任何清理")
}

// TestStartPeriodicCleanupRemovesExpired 保留期为正时启动即清一次并按间隔重复，stop 幂等
func TestStartPeriodicCleanupRemovesExpired(t *testing.T) {
	dbg, storage := newCaptureDebugger(t, &Config{
		Enabled:         true,
		RetentionPeriod: 24 * time.Hour,
		CleanupInterval: 20 * time.Millisecond,
	})

	assert.NoError(t, storage.Save(&LogEntry{ID: "fresh", Timestamp: time.Now(), Method: "GET", URL: "/fresh"}))
	assert.NoError(t, storage.Save(&LogEntry{ID: "expired", Timestamp: time.Now().Add(-48 * time.Hour), Method: "GET", URL: "/expired"}))

	stop := dbg.StartPeriodicCleanup()
	defer func() {
		stop()
		stop() // 幂等：重复调用不应 panic
	}()

	// 再次启动应复用同一个协程（幂等），而不是并发起两个清理循环
	again := dbg.StartPeriodicCleanup()
	again()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, total, err := storage.FindAll(1, 100, nil); err == nil && total == 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	entries, total, err := storage.FindAll(1, 100, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, total, "过期条目应被清理，未过期保留")
	if total == 1 {
		assert.Equal(t, "fresh", entries[0].ID)
	}
}

// TestTrimToRuneBoundary 按字节截断产生的残缺多字节序列应被回退掉
func TestTrimToRuneBoundary(t *testing.T) {
	// "调试" 每个字符 3 字节，切在第 4 字节处会留下半个 rune
	text := "调试日志"
	assert.Equal(t, text, trimToRuneBoundary(text), "完整文本不应被改动")
	assert.Equal(t, "", trimToRuneBoundary(text[:1]))
	assert.Equal(t, "", trimToRuneBoundary(text[:2]))
	assert.Equal(t, "调", trimToRuneBoundary(text[:3]))
	assert.Equal(t, "调", trimToRuneBoundary(text[:4]))

	capped := capBodyText(text, int64(len(text)), 4)
	assert.True(t, strings.HasPrefix(capped, "调"), "截断结果必须是合法 UTF-8 前缀: %q", capped)
	assert.Contains(t, capped, "[Body truncated:")
	assert.NotContains(t, capped, "\ufffd", "不得残留替换字符")
}

// TestBoundedRequestBodyCaptureKeepsHandlerBody 请求体超上限时只捕获上限内正文，且不影响业务处理
func TestBoundedRequestBodyCaptureKeepsHandlerBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dbg, storage := newCaptureDebugger(t, &Config{Enabled: true, SampleRate: 1.0, MaxBodySize: 1})

	var handlerRead int
	router := gin.New()
	router.Use(dbg.Middleware())
	router.POST("/api/echo", func(c *gin.Context) {
		data, err := c.GetRawData()
		assert.NoError(t, err)
		handlerRead = len(data)
		c.Status(http.StatusOK)
	})

	// 1KB 上限下发送 3KB 正文，尾部故意放未闭合的敏感字段，验证截断与掩码的先后顺序
	secret := strings.Repeat("S", 3000)
	body := `{"password":"` + secret
	req := httptest.NewRequest(http.MethodPost, "/api/echo", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, len(body), handlerRead, "业务侧必须仍能读到完整请求体")

	entries, _, err := storage.FindAll(1, 10, nil)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	recorded := entries[0].RequestBody
	assert.Contains(t, recorded, "[Body truncated:", "超上限正文应附截断说明")
	assert.LessOrEqual(t, len(recorded), 1024+len("\n[Body truncated: captured 0 bytes, max_body_size=1 KB]")+64,
		"入库正文不得突破捕获上限")
	assert.NotContains(t, recorded, strings.Repeat("S", 16), "被截半的凭据值也不应残留明文")
}

// TestMultipartFilePartsMetadataOnlyByDefault 默认只记文件部件元数据，显式关闭后才记内容
func TestMultipartFilePartsMetadataOnlyByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fileContent := "FILE-SECRET-BYTES-0123456789"
	newMultipartRequest := func(t *testing.T) (*http.Request, string) {
		t.Helper()
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)
		assert.NoError(t, writer.WriteField("password", "hunter2"))

		part, err := writer.CreateFormFile("avatar", "avatar.png")
		assert.NoError(t, err)
		_, err = io.WriteString(part, fileContent)
		assert.NoError(t, err)
		assert.NoError(t, writer.Close())

		req := httptest.NewRequest(http.MethodPost, "/api/upload", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		return req, writer.FormDataContentType()
	}

	runUpload := func(t *testing.T, config *Config) string {
		t.Helper()
		dbg, storage := newCaptureDebugger(t, config)

		router := gin.New()
		router.Use(dbg.Middleware())
		router.POST("/api/upload", func(c *gin.Context) {
			// 业务侧照常拿到完整文件内容
			file, err := c.FormFile("avatar")
			assert.NoError(t, err)
			opened, err := file.Open()
			assert.NoError(t, err)
			defer opened.Close()
			data, err := io.ReadAll(opened)
			assert.NoError(t, err)
			assert.Equal(t, fileContent, string(data), "中间件不得破坏上传文件")
			c.Status(http.StatusOK)
		})

		req, _ := newMultipartRequest(t)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		entries, _, err := storage.FindAll(1, 10, nil)
		assert.NoError(t, err)
		assert.Len(t, entries, 1)
		return entries[0].RequestBody
	}

	base := Config{Enabled: true, SampleRate: 1.0, EnableMultipartSupport: true}

	t.Run("默认只存元数据", func(t *testing.T) {
		recorded := runUpload(t, &base)
		assert.NotContains(t, recorded, fileContent, "文件字节不应进入调试日志")
		assert.Contains(t, recorded, "avatar.png", "文件名等元数据仍应可排查")
		assert.Contains(t, recorded, "(content not recorded)")
		assert.NotContains(t, recorded, "hunter2", "文本字段值仍需掩码")
	})

	t.Run("显式关闭后记录内容", func(t *testing.T) {
		recordFileContent := false
		config := base
		config.MultipartSkipFiles = &recordFileContent
		recorded := runUpload(t, &config)
		assert.Contains(t, recorded, fileContent, "显式关闭跳过开关后才记录文件内容")
	})
}

// TestDatabaseStorageSaveUpsertAndAmortizedTrim 主键 upsert 一次落库，上限裁剪按写入次数均摊
func TestDatabaseStorageSaveUpsertAndAmortizedTrim(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	const maxSize = 5
	storage, err := NewDatabaseStorage(db, maxSize)
	assert.NoError(t, err)

	// 同一 ID 二次保存应更新而非插入新行（进程条目补全结束时间走的就是这条路径）
	entry := &LogEntry{ID: "upsert-1", Timestamp: time.Now(), Method: "GET", URL: "/first", RecordType: "http"}
	assert.NoError(t, storage.Save(entry))
	entry.URL = "/second"
	entry.StatusCode = http.StatusCreated
	assert.NoError(t, storage.Save(entry))

	var rows int64
	assert.NoError(t, db.Table("debug_logs").Count(&rows).Error)
	assert.Equal(t, int64(1), rows, "同 ID 保存两次应只有一行")

	var stored LogEntryModel
	assert.NoError(t, db.Table("debug_logs").First(&stored).Error)
	assert.Equal(t, "/second", stored.URL)
	assert.Equal(t, http.StatusCreated, stored.StatusCode)

	// 未达检查频率时不裁剪（省去每条 COUNT(*)），达到频率后一次性删到上限。
	// 均摊计数从实例创建起累加，前置的两次 Save 已占用窗口，需按写入次数对齐断言
	const priorWrites = 2
	base := time.Now().Add(-time.Hour)
	for i := 0; i < sizeCheckEveryWrites-priorWrites-1; i++ {
		assert.NoError(t, storage.Save(&LogEntry{
			ID:         GenerateID(),
			Timestamp:  base.Add(time.Duration(i) * time.Second),
			Method:     "GET",
			URL:        "/bulk",
			RecordType: "http",
		}))
	}
	assert.NoError(t, db.Table("debug_logs").Count(&rows).Error)
	assert.Greater(t, rows, int64(maxSize), "均摊窗口内不做裁剪")

	assert.NoError(t, storage.Save(&LogEntry{ID: "last", Timestamp: time.Now(), Method: "GET", URL: "/last", RecordType: "http"}))
	assert.NoError(t, db.Table("debug_logs").Count(&rows).Error)
	assert.Equal(t, int64(maxSize), rows, "达到检查频率后应裁剪到上限")

	var remaining LogEntryModel
	assert.NoError(t, db.Table("debug_logs").Order("timestamp DESC").First(&remaining).Error)
	assert.Equal(t, "last", remaining.ID, "保留的应是最新记录")
}
