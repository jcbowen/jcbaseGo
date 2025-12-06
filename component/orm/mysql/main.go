// Package mysql 提供 MySQL 数据库的 ORM 封装，基于 GORM，包含连接创建、连接池配置、表名处理与分页查询等辅助方法。
package mysql

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/debugger"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/orm"
	"github.com/jcbowen/jcbaseGo/component/orm/base"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// AllTableName 用于映射 information_schema 查询结果的表名字段。
type AllTableName struct {
	TableName string `gorm:"table_name"`
}

// Instance 表示 MySQL 连接实例，封装 DSN、数据库配置、连接句柄、调试状态与错误收集。
type Instance struct {
	Dsn            string
	Conf           jcbaseGo.DbStruct
	Db             *gorm.DB
	debug          bool                     // 是否开启debug
	debuggerLogger debugger.LoggerInterface // debugger日志记录器
	Errors         []error

	// 连接重连配置
	reconnectConfig ReconnectConfig

	// 连接状态管理
	lastValidCheck time.Time     // 最后一次有效性检查时间
	validityCache  time.Duration // 有效性缓存时间，默认30秒
	mu             sync.RWMutex  // 读写锁，保护并发访问
}

// ReconnectConfig 定义连接重连配置
// 用于控制连接失效时的重连行为
type ReconnectConfig struct {
	MaxRetries          int           // 最大重试次数，默认3次
	RetryInterval       time.Duration // 重试间隔，默认1秒
	PingTimeout         time.Duration // Ping测试超时时间，默认2秒
	EnableAutoReconnect bool          // 是否启用自动重连，默认true
}

// DefaultReconnectConfig 返回默认的重连配置
func DefaultReconnectConfig() ReconnectConfig {
	return ReconnectConfig{
		MaxRetries:          3,
		RetryInterval:       time.Second,
		PingTimeout:         2 * time.Second,
		EnableAutoReconnect: true,
	}
}

// getDSN 拼接 Data Source Name（数据源）字符串，基于提供的 `DbStruct`。
func getDSN(dbConfig jcbaseGo.DbStruct) (dsn string) {
	// 拼接dsn
	parseTime := strings.ToLower(dbConfig.ParseTime)
	if parseTime != "false" {
		parseTime = "true"
	}
	dsn = "%s:%s@%s(%s:%s)/%s?charset=%s&parseTime=%s&loc=Local"
	dsn = fmt.Sprintf(dsn, dbConfig.Username, dbConfig.Password, dbConfig.Protocol, dbConfig.Host, dbConfig.Port, dbConfig.Dbname, dbConfig.Charset, parseTime)

	return
}

// New 创建一个 MySQL 实例并建立数据库连接；
// 可通过可选参数 `opts[0]` 指定配置别名以存入环境变量。
func New(dbConfig jcbaseGo.DbStruct, opts ...string) (*Instance, error) {
	inst := &Instance{
		reconnectConfig: DefaultReconnectConfig(),
		validityCache:   30 * time.Second, // 默认30秒缓存
		lastValidCheck:  time.Now(),       // 初始化最后检查时间
	}

	alias := "db"
	if len(opts) > 0 && opts[0] != "" {
		alias = opts[0]
	}

	err := helper.CheckAndSetDefault(&dbConfig)
	if err != nil {
		return nil, err
	}

	// 判断dbConfig是否为空
	if dbConfig.Dbname == "" {
		inst.AddError(errors.New("dbConfig is empty"))
		return inst, errors.New("dbConfig is empty")
	}

	dsn := getDSN(dbConfig)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: dbConfig.DisableForeignKeyConstraintWhenMigrating,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   dbConfig.TablePrefix,   // 表名前缀，`User`表为`t_users`
			SingularTable: dbConfig.SingularTable, // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	})
	if err != nil {
		return nil, err
	}

	// 配置连接池参数，防止连接泄漏和长时间锁定
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// 设置最大连接数
	sqlDB.SetMaxOpenConns(100)
	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(10)
	// 设置连接最大生命周期（5分钟）
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	// 设置空闲连接超时时间（3分钟）
	sqlDB.SetConnMaxIdleTime(3 * time.Minute)

	// 建连校验与重试
	{
		var lastErr error
		for i := 0; i < 3; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			lastErr = sqlDB.PingContext(ctx)
			cancel()
			if lastErr == nil {
				break
			}
			time.Sleep(time.Duration(200*(1<<i)) * time.Millisecond)
		}
		if lastErr != nil {
			return nil, lastErr
		}
	}

	inst.Dsn = dsn
	inst.Conf = dbConfig
	inst.Db = db

	// 将配置信息储存到环境变量
	envStr := ""
	helper.Json(dbConfig).ToString(&envStr)
	err = os.Setenv("jc_mysql_"+alias, envStr)
	if err != nil {
		return nil, err
	}

	return inst, nil
}

// NewWithDebugger 创建MySQL实例并集成debugger日志记录
// 参数：
//   - dbConfig: 数据库配置
//   - debuggerLogger: debugger日志记录器
//   - opts: 可选参数，第一个参数为配置别名
//
// 返回：
//   - *Instance: MySQL实例
func NewWithDebugger(dbConfig jcbaseGo.DbStruct, debuggerLogger debugger.LoggerInterface, opts ...string) (*Instance, error) {
	instance, err := New(dbConfig, opts...)
	if err != nil {
		return nil, err
	}
	if instance.Db != nil && debuggerLogger != nil {
		instance.SetDebuggerLogger(debuggerLogger)
	}
	return instance, nil
}

// Debug 开启调试模式，后续通过 `GetDb()` 获取的 *gorm.DB 将启用 Debug。
func (c *Instance) Debug() *Instance {
	c.debug = true
	return c
}

// GetDb 返回当前数据库连接；若已开启调试模式返回 Debug 包装的连接。
// 调试模式优先级：debuggerLogger > debug标志
// 注意：此方法会验证连接有效性并在需要时尝试重连，支持并发安全
func (c *Instance) GetDb() *gorm.DB {
	// 获取读锁，允许多个读操作并发
	c.mu.RLock()

	// 检查连接是否为 nil
	if c.Db == nil {
		c.mu.RUnlock() // 释放读锁

		// 尝试重新连接（需要写锁）
		if !c.reconnectConfig.EnableAutoReconnect {
			log.Println("Database connection is nil and auto reconnect is disabled")
			return nil
		}

		return c.tryReconnect()
	}

	// 检查是否需要验证连接有效性（基于缓存时间）
	needCheck := time.Since(c.lastValidCheck) > c.validityCache

	// 如果不需要检查，直接返回当前连接
	if !needCheck {
		db := c.applyDebugMode(c.Db)
		c.mu.RUnlock()
		return db
	}

	c.mu.RUnlock() // 释放读锁，准备进行写操作

	// 验证连接有效性并处理重连
	return c.validateAndReconnect()
}

// tryReconnect 尝试重新连接（带写锁保护）
func (c *Instance) tryReconnect() *gorm.DB {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查，防止多个goroutine同时重连
	if c.Db != nil {
		return c.applyDebugMode(c.Db)
	}

	log.Println("Database connection is nil, attempting to reconnect...")

	if err := c.reconnect(); err != nil {
		log.Printf("Failed to reconnect: %v", err)
		return nil
	}

	c.lastValidCheck = time.Now()
	return c.applyDebugMode(c.Db)
}

// validateAndReconnect 验证连接有效性并处理重连（带写锁保护）
func (c *Instance) validateAndReconnect() *gorm.DB {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查，防止其他goroutine已经处理过
	if time.Since(c.lastValidCheck) <= c.validityCache {
		return c.applyDebugMode(c.Db)
	}

	// 验证连接有效性
	if c.isConnectionValid() {
		c.lastValidCheck = time.Now()
		return c.applyDebugMode(c.Db)
	}

	log.Println("Database connection is invalid, attempting to reconnect...")

	if err := c.reconnect(); err != nil {
		log.Printf("Failed to reconnect: %v", err)
		return nil
	}

	c.lastValidCheck = time.Now()
	return c.applyDebugMode(c.Db)
}

// applyDebugMode 应用调试模式配置
func (c *Instance) applyDebugMode(db *gorm.DB) *gorm.DB {
	if db == nil {
		return nil
	}

	// 调试模式优先级：debuggerLogger > debug标志
	if c.debuggerLogger != nil {
		// 如果设置了debuggerLogger，优先使用debuggerLogger的配置
		// 只有在debuggerLogger级别不是静默模式时就启用SQL日志记录
		if c.debuggerLogger.GetLevel() > debugger.LevelSilent {
			// 已经通过SetDebuggerLogger或EnableSQLLogging配置过，直接返回
			return db
		}
		// 如果debuggerLogger级别为静默模式，返回原始连接
		return db
	} else if c.debug {
		// 如果没有设置debuggerLogger但开启了debug标志，使用GORM的Debug模式
		db = db.Debug()
	}

	return db
}

// SetDebuggerLogger 设置debugger日志记录器
// 参数：
//   - debuggerLogger: debugger日志记录器实例
func (c *Instance) SetDebuggerLogger(debuggerLogger debugger.LoggerInterface) {
	c.debuggerLogger = debuggerLogger
	if c.Db != nil && debuggerLogger != nil {
		// 启用SQL日志记录
		c.Db = orm.EnableSQLLogging(c.Db, debuggerLogger)
	}
}

// GetDebuggerLogger 获取debugger日志记录器
func (c *Instance) GetDebuggerLogger() debugger.LoggerInterface {
	return c.debuggerLogger
}

// isConnectionValid 检查数据库连接是否有效
// 通过执行 Ping 操作验证连接状态
func (c *Instance) isConnectionValid() bool {
	if c.Db == nil {
		return false
	}

	sqlDB, err := c.Db.DB()
	if err != nil {
		log.Printf("Failed to get underlying SQL DB: %v", err)
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.reconnectConfig.PingTimeout)
	defer cancel()

	err = sqlDB.PingContext(ctx)
	if err != nil {
		log.Printf("Database connection is invalid: %v", err)
		return false
	}

	return true
}

// reconnect 尝试重新建立数据库连接
// 根据重连配置进行多次重试
func (c *Instance) reconnect() error {
	if !c.reconnectConfig.EnableAutoReconnect {
		return errors.New("auto reconnect is disabled")
	}

	var lastErr error
	for i := 0; i < c.reconnectConfig.MaxRetries; i++ {
		log.Printf("Attempting to reconnect to database (attempt %d/%d)", i+1, c.reconnectConfig.MaxRetries)

		// 创建新的数据库连接
		db, err := gorm.Open(mysql.Open(c.Dsn), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: c.Conf.DisableForeignKeyConstraintWhenMigrating,
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   c.Conf.TablePrefix,
				SingularTable: c.Conf.SingularTable,
			},
		})
		if err != nil {
			lastErr = err
			log.Printf("Reconnection attempt %d failed: %v", i+1, err)

			if i < c.reconnectConfig.MaxRetries-1 {
				time.Sleep(c.reconnectConfig.RetryInterval)
			}
			continue
		}

		// 配置连接池
		sqlDB, err := db.DB()
		if err != nil {
			lastErr = err
			log.Printf("Failed to configure connection pool: %v", err)
			continue
		}

		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
		sqlDB.SetConnMaxIdleTime(3 * time.Minute)

		// 验证连接
		ctx, cancel := context.WithTimeout(context.Background(), c.reconnectConfig.PingTimeout)
		err = sqlDB.PingContext(ctx)
		cancel()

		if err != nil {
			lastErr = err
			log.Printf("Reconnected but ping failed: %v", err)
			continue
		}

		// 恢复调试模式配置
		if c.debuggerLogger != nil {
			db = orm.EnableSQLLogging(db, c.debuggerLogger)
		} else if c.debug {
			db = db.Debug()
		}

		c.Db = db
		log.Printf("Database reconnection successful")
		return nil
	}

	return fmt.Errorf("failed to reconnect after %d attempts: %v", c.reconnectConfig.MaxRetries, lastErr)
}

// EnableSQLLogging 为当前数据库实例启用SQL日志记录
// 参数：
//   - debuggerLogger: debugger日志记录器
//   - logLevel: GORM日志级别（可选，默认logger.Info）
//   - slowThreshold: 慢查询阈值（可选，默认200ms）
//
// 返回：
//   - *Instance: 支持SQL日志记录的数据库实例
func (c *Instance) EnableSQLLogging(debuggerLogger debugger.LoggerInterface, opts ...interface{}) *Instance {
	if c.Db != nil && debuggerLogger != nil {
		c.debuggerLogger = debuggerLogger
		c.Db = orm.EnableSQLLogging(c.Db, debuggerLogger, opts...)
	}
	return c
}

// GetConf 返回当前实例的原始配置结构。
func (c *Instance) GetConf() interface{} {
	return c.Conf
}

// GetAllTableName 查询并返回当前数据库中的所有表名（过滤视图与系统表）。
func (c *Instance) GetAllTableName() (tableNames []AllTableName, err error) {
	// 如果有错误，就不再执行
	if len(c.Errors) > 0 {
		return
	}

	err = c.Db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema='" + c.Conf.Dbname + "' AND table_type='base table'").Scan(&tableNames).Error
	return
}

// TableName 根据配置的表前缀拼接并可选包裹反引号，返回处理后的表名。
// 参数：
//   - tableName: 传入待处理的表名指针
//   - quotes: 可选，是否为表名添加反引号
func (c *Instance) TableName(tableName *string, quotes ...bool) *Instance {
	// 如果有错误，就不再执行
	if len(c.Errors) > 0 {
		return c
	}

	tablePrefix := c.Conf.TablePrefix
	// 如果已经有前缀了，就不再添加
	if len(tablePrefix) > 0 && helper.StringStartWith(*tableName, tablePrefix) {
		tablePrefix = ""
	}

	if len(quotes) > 0 && quotes[0] {
		*tableName = fmt.Sprintf("`%s%s`", tablePrefix, *tableName)
	} else {
		*tableName = fmt.Sprintf("%s%s", tablePrefix, *tableName)
	}

	return c
}

// AddError 将错误追加到实例的错误收集切片中（忽略 nil）。
func (c *Instance) AddError(err error) {
	if err != nil {
		c.Errors = append(c.Errors, err)
	}
}

// Error 返回已收集的非 nil 错误列表。
func (c *Instance) Error() []error {
	// 过滤掉c.Errors中的nil
	var errs []error
	for _, err := range c.Errors {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// FindPageOptions 定义分页查询选项，语义与 CRUD list 保持一致；
type FindPageOptions struct {
	// 查询配置
	Page        int  `default:"1"`     // 页码，默认 1
	PageSize    int  `default:"10"`    // 分页大小，默认 10，最大 1000
	ShowDeleted bool `default:"false"` // 是否显示软删除数据，默认 false

	// 模型配置
	PkId            string `default:"id"` // 主键字段名，默认 "id"
	ModelTableAlias string `default:""`   // 模型表别名，默认为表名

	// 回调
	ListQuery  func(*gorm.DB) *gorm.DB                   // 可选，自定义查询回调
	ListSelect func(*gorm.DB) *gorm.DB                   // 可选，自定义查询字段回调
	ListOrder  func() interface{}                        // 可选，自定义排序回调
	ListEach   func(interface{}) interface{}             // 可选，自定义遍历回调
	ListReturn func(jcbaseGo.ListData) jcbaseGo.ListData // 可选，自定义返回回调
}

// FindForPage 按分页选项查询并返回列表数据；
func (c *Instance) FindForPage(model interface{}, options *FindPageOptions) (jcbaseGo.ListData, error) {
	if model == nil {
		return jcbaseGo.ListData{}, fmt.Errorf("FindForPage: model 不能为空")
	}

	// 默认参数
	if options == nil {
		options = &FindPageOptions{}
	}
	if err := helper.CheckAndSetDefault(options); err != nil {
		jcbaseGo.PanicIfError(err)
	}

	page := options.Page
	if page < 1 {
		page = 1
	}
	pageSize := options.PageSize
	if pageSize < 1 {
		pageSize = 10
	} else if pageSize > 1000 {
		pageSize = 1000
	}

	// 解析模型类型
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 解析表与软删除配置
	var (
		fields                                          []string
		tableName, softDeleteField, softDeleteCondition string
	)
	if modelParseProvider, ok := model.(interface {
		ModelParse(modelType reflect.Type) (tableName string, fields []string, softDeleteField string, softDeleteCondition string)
	}); ok {
		tableName, fields, softDeleteField, softDeleteCondition = modelParseProvider.ModelParse(modelType)
	} else {
		// panic("FindForPage: model 未实现 ModelParse 方法，或返回空表名")
		b := &base.MysqlBaseModel{}
		tableName, fields, softDeleteField, softDeleteCondition = b.ModelParse(modelType)
	}

	// 表别名（以空格开头），方便补充到查询条件时表后
	var tableAliasStartWidthSpace string
	if options.ModelTableAlias != "" {
		tableAliasStartWidthSpace = " " + options.ModelTableAlias
	}

	// 表别名（以点号结尾），方便补充到字段前
	tableAliasEndWidthPoint := tableName
	if options.ModelTableAlias != "" {
		tableAliasEndWidthPoint = options.ModelTableAlias
	}
	tableAliasEndWidthPoint += "."

	// 初始查询
	query := c.GetDb().Table(tableName + tableAliasStartWidthSpace)

	// 软删除条件
	if !options.ShowDeleted && softDeleteField != "" && helper.InArray(softDeleteField, fields) {
		query = query.Where(tableAliasEndWidthPoint + softDeleteField + " " + softDeleteCondition)
	}

	// 自定义查询回调
	if options.ListQuery != nil {
		query = options.ListQuery(query)
	}

	// 统计总数
	total := int64(0)
	if err := query.
		Model(reflect.New(modelType).Interface()).
		Count(&total).Error; err != nil {
		return jcbaseGo.ListData{}, err
	}

	// 自定义查询
	if options.ListSelect != nil {
		query = options.ListSelect(query)
	} else {
		query = query.Select(tableAliasEndWidthPoint + "*")
	}

	// 自定义排序
	var order any
	if options.ListOrder != nil {
		order = options.ListOrder()
	} else {
		order = tableAliasEndWidthPoint + options.PkId + " DESC"
	}

	// 查询结果
	var resultList []interface{}

	// 结果结构体类型
	resultType := modelType
	if options.ListReturn != nil {
		resultStructType := reflect.TypeOf(options.ListReturn)
		if resultStructType.Kind() == reflect.Ptr {
			resultStructType = resultStructType.Elem()
		}
		if resultStructType.Kind() == reflect.Struct {
			resultType = resultStructType
		}
	}

	// 返回结构体
	sliceType := reflect.SliceOf(resultType)
	resultsPtr := reflect.New(sliceType)
	if err := query.Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(resultsPtr.Interface()).Error; err != nil {
		return jcbaseGo.ListData{}, err
	}
	resultsVal := resultsPtr.Elem()
	resultList = make([]interface{}, resultsVal.Len())
	for i := 0; i < resultsVal.Len(); i++ {
		if options.ListEach != nil {
			itemPtr := resultsVal.Index(i).Addr().Interface()
			out := options.ListEach(itemPtr)
			if out != nil && reflect.TypeOf(out).Kind() == reflect.Ptr {
				resultList[i] = reflect.ValueOf(out).Elem().Interface()
			} else {
				resultList[i] = out
			}
		} else {
			resultList[i] = resultsVal.Index(i).Interface()
		}
	}

	listData := jcbaseGo.ListData{
		List:     resultList,
		Total:    int(total),
		Page:     page,
		PageSize: pageSize,
	}

	if options.ListReturn != nil {
		listData = options.ListReturn(listData)
	}

	return listData, nil
}

// SetReconnectConfig 设置连接重连配置
// 参数：
//   - config: 重连配置参数
func (c *Instance) SetReconnectConfig(config ReconnectConfig) {
	c.reconnectConfig = config
}

// GetReconnectConfig 获取当前的重连配置
func (c *Instance) GetReconnectConfig() ReconnectConfig {
	return c.reconnectConfig
}

// EnableAutoReconnect 启用自动重连功能
func (c *Instance) EnableAutoReconnect() *Instance {
	c.reconnectConfig.EnableAutoReconnect = true
	return c
}

// DisableAutoReconnect 禁用自动重连功能
func (c *Instance) DisableAutoReconnect() *Instance {
	c.reconnectConfig.EnableAutoReconnect = false
	return c
}

// SetMaxRetries 设置最大重试次数
func (c *Instance) SetMaxRetries(maxRetries int) *Instance {
	if maxRetries < 1 {
		maxRetries = 1
	}
	c.reconnectConfig.MaxRetries = maxRetries
	return c
}

// SetRetryInterval 设置重试间隔
func (c *Instance) SetRetryInterval(interval time.Duration) *Instance {
	if interval < time.Second {
		interval = time.Second
	}
	c.reconnectConfig.RetryInterval = interval
	return c
}

// SetPingTimeout 设置Ping测试超时时间
func (c *Instance) SetPingTimeout(timeout time.Duration) *Instance {
	if timeout < time.Second {
		timeout = time.Second
	}
	c.reconnectConfig.PingTimeout = timeout
	return c
}
