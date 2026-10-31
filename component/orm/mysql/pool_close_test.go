package mysql

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jcbowen/jcbaseGo"
)

// stubDriver 注册一个不产生真实网络连接的占位驱动，
// 仅用于构造合法的 *sql.DB 对象来验证 applyPoolConfig 的池参数设置逻辑。
type stubDriver struct{}

// Open 始终返回错误：本驱动只为拿到 *sql.DB 句柄，不建立真实连接。
// 参数：
//   - name: 数据源名称（未使用）
//
// 返回值：
//   - driver.Conn: 永远为 nil
//   - error: 始终返回占位错误
func (stubDriver) Open(name string) (driver.Conn, error) {
	return nil, errors.New("stub driver: no real connection")
}

func init() {
	// 测试文件只加载一次，init 中注册不会触发重复注册 panic
	sql.Register("jcbase_mysql_pool_test_stub", stubDriver{})
}

// newStubSQLDB 构造一个挂载 stub 驱动的 *sql.DB，用例结束时自动关闭。
// 参数：
//   - t: 测试对象
//
// 返回值：
//   - *sql.DB: 未建立真实连接的连接池句柄
func newStubSQLDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("jcbase_mysql_pool_test_stub", "")
	if err != nil {
		t.Fatalf("打开stub连接池失败: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestResolvePoolConfig_Defaults 校验未配置（0 值）时回落库内默认值，
// 保证默认行为与历史硬编码（100/10/5m/3m）逐项一致。
func TestResolvePoolConfig_Defaults(t *testing.T) {
	maxOpen, maxIdle, lifetime, idleTime := resolvePoolConfig(jcbaseGo.DbStruct{})

	if maxOpen != 100 {
		t.Errorf("maxOpen = %d, want 100", maxOpen)
	}
	if maxIdle != 10 {
		t.Errorf("maxIdle = %d, want 10", maxIdle)
	}
	if lifetime != 5*time.Minute {
		t.Errorf("lifetime = %v, want 5m", lifetime)
	}
	if idleTime != 3*time.Minute {
		t.Errorf("idleTime = %v, want 3m", idleTime)
	}
}

// TestResolvePoolConfig_Explicit 校验显式配置时按配置生效，不被默认值覆盖。
func TestResolvePoolConfig_Explicit(t *testing.T) {
	conf := jcbaseGo.DbStruct{
		MaxOpenConns:           20,
		MaxIdleConns:           5,
		ConnMaxLifetimeSeconds: 60,
		ConnMaxIdleTimeSeconds: 30,
	}
	maxOpen, maxIdle, lifetime, idleTime := resolvePoolConfig(conf)

	if maxOpen != 20 || maxIdle != 5 || lifetime != time.Minute || idleTime != 30*time.Second {
		t.Errorf("resolvePoolConfig() = (%d, %d, %v, %v), want (20, 5, 1m, 30s)", maxOpen, maxIdle, lifetime, idleTime)
	}
}

// TestResolvePoolConfig_Negative 校验负值与 0 同等回落默认，避免调用方误传负数产生异常配置。
func TestResolvePoolConfig_Negative(t *testing.T) {
	conf := jcbaseGo.DbStruct{
		MaxOpenConns:           -1,
		MaxIdleConns:           -1,
		ConnMaxLifetimeSeconds: -1,
		ConnMaxIdleTimeSeconds: -1,
	}
	maxOpen, maxIdle, lifetime, idleTime := resolvePoolConfig(conf)

	if maxOpen != 100 || maxIdle != 10 || lifetime != 5*time.Minute || idleTime != 3*time.Minute {
		t.Errorf("负值应回落默认, got (%d, %d, %v, %v)", maxOpen, maxIdle, lifetime, idleTime)
	}
}

// TestApplyPoolConfig_MaxOpenConnections 通过 Stats().MaxOpenConnections 断言
// applyPoolConfig 真正作用到了 *sql.DB 上（其余三项参数无读取接口，在纯函数用例中覆盖）。
func TestApplyPoolConfig_MaxOpenConnections(t *testing.T) {
	db := newStubSQLDB(t)

	applyPoolConfig(db, jcbaseGo.DbStruct{MaxOpenConns: 7})
	if got := db.Stats().MaxOpenConnections; got != 7 {
		t.Errorf("MaxOpenConnections = %d, want 7", got)
	}

	applyPoolConfig(db, jcbaseGo.DbStruct{})
	if got := db.Stats().MaxOpenConnections; got != 100 {
		t.Errorf("未配置时 MaxOpenConnections = %d, want 100", got)
	}
}

// testMysqlConf 构造真实 MySQL 的集成测试配置。
// 未设置 TEST_MYSQL_PASSWORD 时跳过用例，避免在无数据库环境（如 CI）中假绿。
func testMysqlConf(t *testing.T) jcbaseGo.DbStruct {
	t.Helper()
	password := os.Getenv("TEST_MYSQL_PASSWORD")
	if password == "" {
		t.Skip("未设置 TEST_MYSQL_PASSWORD，跳过需要真实 MySQL 的集成用例")
	}

	return jcbaseGo.DbStruct{
		Host:     getenvDefault("TEST_MYSQL_HOST", "127.0.0.1"),
		Port:     getenvDefault("TEST_MYSQL_PORT", "3306"),
		Dbname:   getenvDefault("TEST_MYSQL_DBNAME", "mysql"),
		Username: getenvDefault("TEST_MYSQL_USER", "root"),
		Password: password,
	}
}

// getenvDefault 读取环境变量，未设置时返回默认值。
func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// TestNew_PoolConfigFromDbStruct 校验 New 阶段 DbStruct 的池配置直接生效。
func TestNew_PoolConfigFromDbStruct(t *testing.T) {
	conf := testMysqlConf(t)
	conf.MaxOpenConns = 2

	inst, err := New(conf)
	if err != nil {
		t.Fatalf("New 失败: %v", err)
	}
	defer func() { _ = inst.Close() }()

	sqlDB, err := inst.Db.DB()
	if err != nil {
		t.Fatalf("获取底层 sql.DB 失败: %v", err)
	}
	if got := sqlDB.Stats().MaxOpenConnections; got != 2 {
		t.Errorf("MaxOpenConnections = %d, want 2", got)
	}
}

// TestClose_ReleasesConnections 校验 Close 后连接数归零、句柄不可再用，且重复调用安全。
func TestClose_ReleasesConnections(t *testing.T) {
	conf := testMysqlConf(t)

	inst, err := New(conf)
	if err != nil {
		t.Fatalf("New 失败: %v", err)
	}

	// 建立至少一条真实连接
	if err := inst.GetDb().Exec("SELECT 1").Error; err != nil {
		t.Fatalf("建立连接失败: %v", err)
	}

	sqlDB, err := inst.Db.DB()
	if err != nil {
		t.Fatalf("获取底层 sql.DB 失败: %v", err)
	}
	if sqlDB.Stats().OpenConnections < 1 {
		t.Fatal("前置条件失败：SELECT 1 后应至少存在 1 条连接")
	}

	if err := inst.Close(); err != nil {
		t.Fatalf("Close 失败: %v", err)
	}

	// 连接数归零（Close 关闭空闲连接；本用例无并发使用，归还后即全部关闭）
	if got := sqlDB.Stats().OpenConnections; got != 0 {
		t.Errorf("Close 后 OpenConnections = %d, want 0", got)
	}
	// 已关闭的池再执行 Ping 应报错（句柄不可复用）
	if err := sqlDB.Ping(); err == nil {
		t.Error("Close 后 Ping 应失败，实际成功")
	}
	// 重复调用 Close 应幂等
	if err := inst.Close(); err != nil {
		t.Errorf("重复 Close 应返回 nil, got %v", err)
	}
}

// TestClose_DisablesAutoReconnect 校验关闭后 GetDb 不再触发自动重连（不会把连接池重新建起来）。
func TestClose_DisablesAutoReconnect(t *testing.T) {
	conf := testMysqlConf(t)

	inst, err := New(conf)
	if err != nil {
		t.Fatalf("New 失败: %v", err)
	}

	if err := inst.Close(); err != nil {
		t.Fatalf("Close 失败: %v", err)
	}

	// EnableAutoReconnect 仍为默认开启，但显式关闭后不应重建
	if inst.GetDb() != nil {
		t.Error("Close 后 GetDb 应返回 nil")
	}
	if inst.GetDb() != nil {
		t.Error("Close 后再次 GetDb 仍应返回 nil（不应自动重连）")
	}
}

// TestReconnect_ClosesOldPool 校验重连成功后旧连接池已关闭，修复连接泄漏。
func TestReconnect_ClosesOldPool(t *testing.T) {
	conf := testMysqlConf(t)
	conf.MaxOpenConns = 3 // 显式自定义池配置，验证重连后不回退默认值

	inst, err := New(conf)
	if err != nil {
		t.Fatalf("New 失败: %v", err)
	}
	defer func() { _ = inst.Close() }()

	// 建立至少一条真实连接
	if err := inst.GetDb().Exec("SELECT 1").Error; err != nil {
		t.Fatalf("建立连接失败: %v", err)
	}

	oldDB, err := inst.Db.DB()
	if err != nil {
		t.Fatalf("获取旧 sql.DB 失败: %v", err)
	}

	// 直接触发一次重连（本测试与 reconnect 同包，可调用私有方法）
	if err := inst.reconnect(); err != nil {
		t.Fatalf("reconnect 失败: %v", err)
	}

	newDB, err := inst.Db.DB()
	if err != nil {
		t.Fatalf("获取新 sql.DB 失败: %v", err)
	}

	if oldDB == newDB {
		t.Fatal("重连后应替换为新的连接池，实际仍是同一个 *sql.DB")
	}
	// 旧池已关闭：连接数归零且 Ping 报错
	if got := oldDB.Stats().OpenConnections; got != 0 {
		t.Errorf("旧池 OpenConnections = %d, want 0（存在泄漏）", got)
	}
	if err := oldDB.Ping(); err == nil {
		t.Error("旧池 Ping 应失败（已关闭），实际成功")
	}
	// 新池按同一份配置生效（验证重连不会把自定义配置重置回默认）
	if got := newDB.Stats().MaxOpenConnections; got != conf.MaxOpenConns {
		t.Errorf("新池 MaxOpenConnections = %d, want %d（配置被重置回默认）", got, conf.MaxOpenConns)
	}
}
