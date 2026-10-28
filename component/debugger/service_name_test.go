package debugger

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// newServiceNameTestDB 创建sqlite内存库并限制单连接
// :memory:模式下每个连接持有独立内存数据库，多连接会导致读写漂移（写入的表在另一连接中不可见），
// 因此必须限制MaxOpenConns=1，确保读写共享同一连接
// db: 配置好单连接的GORM数据库实例
func newServiceNameTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	sqlDB, err := db.DB()
	assert.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	return db
}

// newServiceNameTestDebugger 创建用于ServiceName测试的调试器实例
// debugger: 调试器实例
// err: 创建过程中的错误
func newServiceNameTestDebugger(t *testing.T, serviceName string, storage Storage) (*Debugger, error) {
	t.Helper()
	config := &Config{
		Enabled:     true,
		Level:       LevelSilent, // 静默日志，避免测试输出干扰
		ServiceName: serviceName,
		Storage:     storage,
	}
	return New(config)
}

// runMiddlewareRequest 构造Gin测试上下文并执行调试器中间件，模拟一次完整请求记录
// 返回测试响应记录器与Gin上下文
func runMiddlewareRequest(t *testing.T, dbg *Debugger, method, path string) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	req.RemoteAddr = "127.0.0.1:8080"

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	middlewareFunc := dbg.Middleware()
	middlewareFunc(c)
}

// TestServiceNameHTTPEntry 验证中间件记录的HTTP条目携带配置的ServiceName，
// 且未配置ServiceName时条目该字段为空（向后兼容）
func TestServiceNameHTTPEntry(t *testing.T) {
	memoryStorage, err := NewMemoryStorage(100)
	assert.NoError(t, err)

	dbg, err := newServiceNameTestDebugger(t, "gateway", memoryStorage)
	assert.NoError(t, err)

	runMiddlewareRequest(t, dbg, "GET", "/api/test")

	entries, total, err := memoryStorage.FindAll(1, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "gateway", entries[0].ServiceName)

	// 兼容性回归：未配置ServiceName时，条目字段为空串，行为与改造前一致
	memoryStorage2, err := NewMemoryStorage(100)
	assert.NoError(t, err)

	dbg2, err := newServiceNameTestDebugger(t, "", memoryStorage2)
	assert.NoError(t, err)

	runMiddlewareRequest(t, dbg2, "GET", "/api/test")

	entries2, total2, err := memoryStorage2.FindAll(1, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, total2)
	assert.Empty(t, entries2[0].ServiceName)
}

// TestServiceNameProcessEntry 验证进程记录条目（record_type=process）携带配置的ServiceName
func TestServiceNameProcessEntry(t *testing.T) {
	memoryStorage, err := NewMemoryStorage(100)
	assert.NoError(t, err)

	dbg, err := newServiceNameTestDebugger(t, "order-service", memoryStorage)
	assert.NoError(t, err)

	pl := NewProcessLogger(dbg, "数据同步任务", "batch")
	assert.NotNil(t, pl)

	entries, total, err := memoryStorage.FindAll(1, 10, map[string]interface{}{"record_type": "process"})
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "order-service", entries[0].ServiceName)
	assert.Equal(t, "process", entries[0].RecordType)
}

// TestMemoryStorageServiceName 验证内存存储的ServiceName写入、过滤与统计
func TestMemoryStorageServiceName(t *testing.T) {
	storage, err := NewMemoryStorage(100)
	assert.NoError(t, err)

	now := time.Now()
	entries := []*LogEntry{
		{ID: "mem-svc-1", Timestamp: now, RecordType: "http", ServiceName: "gateway"},
		{ID: "mem-svc-2", Timestamp: now.Add(-time.Second), RecordType: "http", ServiceName: "order"},
		{ID: "mem-svc-3", Timestamp: now.Add(-2 * time.Second), RecordType: "http"}, // 无服务标识
	}
	for _, entry := range entries {
		assert.NoError(t, storage.Save(entry))
	}

	// 按服务过滤：精确匹配
	result, total, err := storage.FindAll(1, 10, map[string]interface{}{"service_name": "gateway"})
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "mem-svc-1", result[0].ID)

	// 搜索接口同样支持service_name过滤
	searchResult, searchTotal, err := storage.Search("test", 1, 10, map[string]interface{}{"service_name": "order"})
	assert.NoError(t, err)
	assert.Equal(t, 0, searchTotal)
	assert.Empty(t, searchResult)

	// 服务统计：含空标识计数，调用方（controller）负责过滤空值
	services, err := storage.GetServices()
	assert.NoError(t, err)
	assert.Equal(t, 1, services["gateway"])
	assert.Equal(t, 1, services["order"])
	assert.Equal(t, 1, services[""])
}

// TestFileStorageServiceName 验证文件存储的ServiceName写入、过滤与统计
// FileStorage走JSON序列化，ServiceName字段随结构体完整保留；
// 旧文件无该字段时反序列化为空串，天然兼容
func TestFileStorageServiceName(t *testing.T) {
	storage, err := NewFileStorage(t.TempDir(), 0)
	assert.NoError(t, err)

	now := time.Now()
	entries := []*LogEntry{
		{ID: "file-svc-1", Timestamp: now, RecordType: "http", ServiceName: "gateway"},
		{ID: "file-svc-2", Timestamp: now.Add(-time.Second), RecordType: "http", ServiceName: "order"},
		{ID: "file-svc-3", Timestamp: now.Add(-2 * time.Second), RecordType: "http"}, // 无服务标识
	}
	for _, entry := range entries {
		assert.NoError(t, storage.Save(entry))
	}

	// 按服务过滤：精确匹配
	result, total, err := storage.FindAll(1, 10, map[string]interface{}{"service_name": "gateway"})
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "file-svc-1", result[0].ID)

	// 服务统计
	services, err := storage.GetServices()
	assert.NoError(t, err)
	assert.Equal(t, 1, services["gateway"])
	assert.Equal(t, 1, services["order"])
	assert.Equal(t, 1, services[""])
}

// TestDatabaseStorageServiceName 验证数据库存储的ServiceName写入、过滤与统计
// 使用sqlite内存库，AutoMigrate自动创建含service_name列的表
func TestDatabaseStorageServiceName(t *testing.T) {
	db := newServiceNameTestDB(t)

	storage, err := NewDatabaseStorage(db, 0)
	assert.NoError(t, err)

	now := time.Now()
	entries := []*LogEntry{
		{ID: "db-svc-1", Timestamp: now, RecordType: "http", ServiceName: "gateway"},
		{ID: "db-svc-2", Timestamp: now.Add(-time.Second), RecordType: "http", ServiceName: "order"},
		{ID: "db-svc-3", Timestamp: now.Add(-2 * time.Second), RecordType: "http"}, // 无服务标识
	}
	for _, entry := range entries {
		assert.NoError(t, storage.Save(entry))
	}

	// 按服务过滤：精确匹配
	result, total, err := storage.FindAll(1, 10, map[string]interface{}{"service_name": "gateway"})
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "db-svc-1", result[0].ID)

	// 服务统计（GROUP BY service_name）
	services, err := storage.GetServices()
	assert.NoError(t, err)
	assert.Equal(t, 1, services["gateway"])
	assert.Equal(t, 1, services["order"])
	assert.Equal(t, 1, services[""])
}

// TestServiceNameSharedDatabaseTwoDebuggers 验收测试：
// 两个不同ServiceName的Debugger实例共用同一个DatabaseStorage（同一张表），
// 分别通过中间件写入HTTP记录后，可按服务区分并过滤
func TestServiceNameSharedDatabaseTwoDebuggers(t *testing.T) {
	db := newServiceNameTestDB(t)

	sharedStorage, err := NewDatabaseStorage(db, 0)
	assert.NoError(t, err)

	dbgA, err := newServiceNameTestDebugger(t, "gateway", sharedStorage)
	assert.NoError(t, err)

	dbgB, err := newServiceNameTestDebugger(t, "order-service", sharedStorage)
	assert.NoError(t, err)

	// 两个实例分别模拟一次请求
	runMiddlewareRequest(t, dbgA, "GET", "/api/gateway-test")
	runMiddlewareRequest(t, dbgB, "POST", "/api/order-test")

	// 全量：两条记录都能查到
	_, totalAll, err := sharedStorage.FindAll(1, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, totalAll)

	// 按服务A过滤：仅返回实例A的记录
	entriesA, totalA, err := sharedStorage.FindAll(1, 10, map[string]interface{}{"service_name": "gateway"})
	assert.NoError(t, err)
	assert.Equal(t, 1, totalA)
	assert.Equal(t, "gateway", entriesA[0].ServiceName)
	assert.Equal(t, "/api/gateway-test", entriesA[0].URL)

	// 按服务B过滤：仅返回实例B的记录
	entriesB, totalB, err := sharedStorage.FindAll(1, 10, map[string]interface{}{"service_name": "order-service"})
	assert.NoError(t, err)
	assert.Equal(t, 1, totalB)
	assert.Equal(t, "order-service", entriesB[0].ServiceName)
	assert.Equal(t, "/api/order-test", entriesB[0].URL)

	// 服务统计可区分两个服务
	services, err := sharedStorage.GetServices()
	assert.NoError(t, err)
	assert.Equal(t, 1, services["gateway"])
	assert.Equal(t, 1, services["order-service"])
}

// TestServiceNameNoConfigCompatibility 验证不配置ServiceName时，各存储过滤行为与改造前一致：
// 现有过滤条件（record_type/method等）不受新字段影响
func TestServiceNameNoConfigCompatibility(t *testing.T) {
	memoryStorage, err := NewMemoryStorage(100)
	assert.NoError(t, err)

	dbg, err := newServiceNameTestDebugger(t, "", memoryStorage)
	assert.NoError(t, err)

	runMiddlewareRequest(t, dbg, "GET", "/api/legacy-test")

	// 不带service_name过滤：行为与改造前完全一致
	entries, total, err := memoryStorage.FindAll(1, 10, map[string]interface{}{"record_type": "http"})
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Empty(t, entries[0].ServiceName)
	assert.Equal(t, "/api/legacy-test", entries[0].URL)
}
