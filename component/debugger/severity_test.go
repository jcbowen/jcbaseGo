package debugger

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestLogEntryUpdateSeverity 验证 UpdateSeverity 对派生严重度字段的统计逻辑。
//
// 覆盖场景：无日志的干净条目、含error/warn级Logger日志、Error字段非空、
// HTTP状态码>=500、进程记录（StatusCode为0）含error日志。
func TestLogEntryUpdateSeverity(t *testing.T) {
	tests := []struct {
		name           string
		entry          *LogEntry
		wantHasError   bool
		wantErrorCount int
		wantWarnCount  int
	}{
		{
			name: "干净条目不算错误",
			entry: &LogEntry{
				StatusCode: 200,
				LoggerLogs: []LoggerLog{
					{Level: LevelInfo, Message: "info日志"},
				},
			},
			wantHasError:   false,
			wantErrorCount: 0,
			wantWarnCount:  0,
		},
		{
			name: "统计error与warn级别日志数量",
			entry: &LogEntry{
				StatusCode: 200,
				LoggerLogs: []LoggerLog{
					{Level: LevelError, Message: "错误1"},
					{Level: LevelError, Message: "错误2"},
					{Level: LevelWarn, Message: "警告1"},
					{Level: LevelInfo, Message: "信息"},
				},
			},
			wantHasError:   true,
			wantErrorCount: 2,
			wantWarnCount:  1,
		},
		{
			name: "Error字段非空判定为错误",
			entry: &LogEntry{
				StatusCode: 200,
				Error:      "gin上下文错误",
			},
			wantHasError:   true,
			wantErrorCount: 0,
			wantWarnCount:  0,
		},
		{
			name: "HTTP状态码500判定为错误",
			entry: &LogEntry{
				StatusCode: 500,
			},
			wantHasError:   true,
			wantErrorCount: 0,
			wantWarnCount:  0,
		},
		{
			name: "进程记录completed但含error日志判定为错误",
			entry: &LogEntry{
				RecordType: "process",
				Status:     ProcessStatusCompleted,
				LoggerLogs: []LoggerLog{
					{Level: LevelError, Message: "进程内错误"},
				},
			},
			wantHasError:   true,
			wantErrorCount: 1,
			wantWarnCount:  0,
		},
		{
			name: "仅含warn日志不算错误但统计warn数量",
			entry: &LogEntry{
				StatusCode: 200,
				LoggerLogs: []LoggerLog{
					{Level: LevelWarn, Message: "警告"},
				},
			},
			wantHasError:   false,
			wantErrorCount: 0,
			wantWarnCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.entry.UpdateSeverity()
			assert.Equal(t, tt.wantHasError, tt.entry.HasError)
			assert.Equal(t, tt.wantErrorCount, tt.entry.ErrorCount)
			assert.Equal(t, tt.wantWarnCount, tt.entry.WarnCount)
		})
	}
}

// TestLogEntryComputeHasError 验证 ComputeHasError 对未回填严重度字段的历史数据能实时判定。
func TestLogEntryComputeHasError(t *testing.T) {
	// 历史数据：HasError未回填，但LoggerLogs含error级日志
	legacy := &LogEntry{
		StatusCode: 200,
		LoggerLogs: []LoggerLog{
			{Level: LevelError, Message: "历史错误日志"},
		},
	}
	assert.True(t, legacy.ComputeHasError(), "历史数据含error级日志应判定为错误")

	// 已回填的数据直接返回HasError
	backfilled := &LogEntry{HasError: true}
	assert.True(t, backfilled.ComputeHasError())

	// 状态码>=500的历史数据
	serverError := &LogEntry{StatusCode: 503}
	assert.True(t, serverError.ComputeHasError())

	// 干净条目
	clean := &LogEntry{StatusCode: 200}
	assert.False(t, clean.ComputeHasError())
}

// TestMemoryStorageHasErrorFilter 验证内存存储的 has_error 过滤基于派生严重度判定，
// 且对未统计严重度的历史条目同样生效。
func TestMemoryStorageHasErrorFilter(t *testing.T) {
	storage, err := NewMemoryStorage(100)
	assert.NoError(t, err)

	now := time.Now()

	// 干净条目
	clean := &LogEntry{
		ID:         "sev-mem-1",
		Timestamp:  now,
		RecordType: "http",
		Method:     "GET",
		URL:        "/api/clean",
		StatusCode: 200,
	}

	// 状态码500条目（保存前已统计）
	serverError := &LogEntry{
		ID:         "sev-mem-2",
		Timestamp:  now.Add(-time.Second),
		RecordType: "http",
		Method:     "GET",
		URL:        "/api/error",
		StatusCode: 500,
	}
	serverError.UpdateSeverity()

	// 历史进程条目：completed但含error日志，且未调用UpdateSeverity
	legacyProcess := &LogEntry{
		ID:         "sev-mem-3",
		Timestamp:  now.Add(-2 * time.Second),
		RecordType: "process",
		Status:     ProcessStatusCompleted,
		LoggerLogs: []LoggerLog{
			{Level: LevelError, Message: "历史进程错误"},
		},
	}

	for _, entry := range []*LogEntry{clean, serverError, legacyProcess} {
		assert.NoError(t, storage.Save(entry))
	}

	result, total, err := storage.FindAll(1, 10, map[string]interface{}{"has_error": true})
	assert.NoError(t, err)
	assert.Equal(t, 2, total, "含错误条目应为2条（500状态码+历史error日志进程）")
	assert.Len(t, result, 2)

	result, total, err = storage.FindAll(1, 10, map[string]interface{}{"has_error": false})
	assert.NoError(t, err)
	assert.Equal(t, 1, total, "无错误条目应为1条")
	assert.Len(t, result, 1)
	assert.Equal(t, "sev-mem-1", result[0].ID)
}

// TestDatabaseStorageHasErrorFilter 验证数据库存储的 has_error 过滤：
// 持久化的 has_error 列、error 列与状态码回退条件均能命中，
// 且严重度字段在 Save/FindByID 过程中正确持久化与读取。
func TestDatabaseStorageHasErrorFilter(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	storage, err := NewDatabaseStorage(db, 0)
	assert.NoError(t, err)

	now := time.Now()

	// 干净条目
	clean := &LogEntry{
		ID:         "sev-db-1",
		Timestamp:  now,
		Method:     "GET",
		URL:        "/api/clean",
		StatusCode: 200,
	}

	// 含error级Logger日志的进程条目（保存前已统计，has_error列=true）
	process := &LogEntry{
		ID:         "sev-db-2",
		Timestamp:  now.Add(-time.Second),
		RecordType: "process",
		Status:     ProcessStatusCompleted,
		LoggerLogs: []LoggerLog{
			{Level: LevelError, Message: "进程错误"},
			{Level: LevelWarn, Message: "进程警告"},
		},
	}
	process.UpdateSeverity()

	// 历史条目：仅error列非空（has_error列未回填）
	legacyError := &LogEntry{
		ID:         "sev-db-3",
		Timestamp:  now.Add(-2 * time.Second),
		Method:     "POST",
		URL:        "/api/legacy",
		StatusCode: 200,
		Error:      "历史错误信息",
	}

	// 历史条目：仅状态码500（has_error列未回填）
	legacy500 := &LogEntry{
		ID:         "sev-db-4",
		Timestamp:  now.Add(-3 * time.Second),
		Method:     "GET",
		URL:        "/api/legacy500",
		StatusCode: 500,
	}

	for _, entry := range []*LogEntry{clean, process, legacyError, legacy500} {
		assert.NoError(t, storage.Save(entry))
	}

	result, total, err := storage.FindAll(1, 10, map[string]interface{}{"has_error": true})
	assert.NoError(t, err)
	assert.Equal(t, 3, total, "含错误条目应为3条（进程error日志+error列+状态码500）")
	assert.Len(t, result, 3)

	result, total, err = storage.FindAll(1, 10, map[string]interface{}{"has_error": false})
	assert.NoError(t, err)
	assert.Equal(t, 1, total, "无错误条目应为1条")
	assert.Len(t, result, 1)
	assert.Equal(t, "sev-db-1", result[0].ID)

	// 验证严重度字段持久化往返
	entry, err := storage.FindByID("sev-db-2")
	assert.NoError(t, err)
	assert.True(t, entry.HasError, "has_error应持久化为true")
	assert.Equal(t, 1, entry.ErrorCount, "error_count应持久化")
	assert.Equal(t, 1, entry.WarnCount, "warn_count应持久化")
}

// TestDebuggerEndProcessAttachesLoggerLogs 验证通过 Debugger.EndProcess（仅凭processID）结束进程时，
// 注册表中的logger日志会被附带落库，且派生严重度随之统计。
// 覆盖场景：进程状态为completed但含error日志，列表应能识别为含错误。
func TestDebuggerEndProcessAttachesLoggerLogs(t *testing.T) {
	dbg, err := NewWithMemoryStorage(50)
	assert.NoError(t, err)

	logger := dbg.StartProcess("测试任务", "worker")
	logger.Error("处理失败", map[string]interface{}{"step": 4})
	logger.Warn("存在隐患")

	processID := logger.GetProcessID()

	// 进程运行中，GetProcessLogger 应返回注册表中的同一实例，保证日志汇聚
	same, err := dbg.GetProcessLogger(processID)
	assert.NoError(t, err)
	assert.Equal(t, logger.GetProcessID(), same.GetProcessID())

	assert.NoError(t, dbg.EndProcess(processID, ProcessStatusCompleted))

	entry, err := dbg.GetStorage().FindByID(processID)
	assert.NoError(t, err)
	assert.Equal(t, ProcessStatusCompleted, entry.Status)
	assert.Len(t, entry.LoggerLogs, 2, "Debugger.EndProcess应附带logger收集的日志")
	assert.True(t, entry.HasError, "completed但含error日志应判定为含错误")
	assert.Equal(t, 1, entry.ErrorCount)
	assert.Equal(t, 1, entry.WarnCount)

	// 进程已结束后不应再获取记录器
	ended, err := dbg.GetProcessLogger(processID)
	assert.Error(t, err)
	assert.Nil(t, ended)
}

// TestParseFiltersHasError 验证控制器将 has_error 查询参数解析为布尔类型过滤条件，
// 与存储层按bool匹配的约定保持一致。
func TestParseFiltersHasError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := &Controller{}

	newContext := func(query string) *gin.Context {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest("GET", "/list"+query, nil)
		return ctx
	}

	// has_error=true 解析为布尔true
	filters := c.parseFilters(newContext("?has_error=true"))
	v, ok := filters["has_error"].(bool)
	assert.True(t, ok, "has_error应为布尔类型")
	assert.True(t, v)

	// has_error=false 解析为布尔false
	filters = c.parseFilters(newContext("?has_error=false"))
	v, ok = filters["has_error"].(bool)
	assert.True(t, ok)
	assert.False(t, v)

	// 未传参数时不设置过滤条件
	filters = c.parseFilters(newContext(""))
	_, exists := filters["has_error"]
	assert.False(t, exists, "未传has_error时不应设置过滤条件")
}
