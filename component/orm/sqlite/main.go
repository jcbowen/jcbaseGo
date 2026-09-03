// Package sqlite 提供 SQLite 数据库的 ORM 封装，基于 GORM，包含连接管理、表名处理与分页查询等辅助方法。
package sqlite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/debugger"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/orm"
	"github.com/jcbowen/jcbaseGo/component/orm/base"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Instance 表示 SQLite 连接实例，封装数据库配置、连接句柄、调试状态与错误收集。
type Instance struct {
	Conf           jcbaseGo.SqlLiteStruct
	Db             *gorm.DB
	debug          bool                     // 是否开启调试模式
	debuggerLogger debugger.LoggerInterface // debugger日志记录器（上下文感知模式下作为兜底）
	contextAware   bool                     // 是否启用上下文感知的SQL日志模式
	sqlLogOpts     []interface{}            // SQL日志可选参数，供后续按原配置重建
	Errors         []error
}

// New 创建一个 SQLite 实例并建立数据库连接；
// 可通过可选参数 `opts[0]` 指定配置别名以存入环境变量。
// 返回：
//   - *Instance: SQLite实例
//   - error: 初始化或连接失败时的错误信息
func New(Conf jcbaseGo.SqlLiteStruct, opts ...string) (*Instance, error) {
	i := &Instance{}

	alias := "main"
	if len(opts) > 0 && opts[0] != "" {
		alias = opts[0]
	}

	err := helper.CheckAndSetDefault(&Conf)
	if err != nil {
		return nil, err
	}

	// 获取dbFile的绝对路径
	fileNameFull, err := filepath.Abs(Conf.DbFile)
	if err != nil {
		return nil, err
	}

	// 检查目录是否存在，如果不存在则创建
	_, err = helper.NewFile(&helper.File{Path: fileNameFull}).DirExists(true)
	if err != nil {
		return nil, err
	}

	// 判断dbConfig是否为空
	if Conf.DbFile == "" {
		return i, errors.New("dbConfig is empty")
	}

	// 创建数据库连接
	db, err := gorm.Open(sqlite.Open(Conf.DbFile), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: Conf.DisableForeignKeyConstraintWhenMigrating,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   Conf.TablePrefix,   // 表名前缀，`User`表为`t_users`
			SingularTable: Conf.SingularTable, // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	})
	if err != nil {
		return nil, err
	}

	// 配置连接池参数，防止连接泄漏
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// 设置最大连接数（SQLite通常只需要少量连接）
	sqlDB.SetMaxOpenConns(1)
	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(1)
	// 设置连接最大生命周期（10分钟）
	sqlDB.SetConnMaxLifetime(10 * time.Minute)
	// 设置空闲连接超时时间（5分钟）
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	// 建连校验与轻量重试（SQLite 本地文件，一般即刻可用）
	{
		var lastErr error
		for i := 0; i < 2; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			lastErr = sqlDB.PingContext(ctx)
			cancel()
			if lastErr == nil {
				break
			}
			time.Sleep(time.Duration(100*(1<<i)) * time.Millisecond)
		}
		if lastErr != nil {
			return nil, lastErr
		}
	}

	i.Conf = Conf
	i.Db = db

	// 将配置信息储存到环境变量
	envStr := ""
	helper.Json(Conf).ToString(&envStr)
	err = os.Setenv("jc_sql_lite_"+alias, envStr)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// NewWithDebugger 创建SQLite实例并集成debugger日志记录
// 参数：
//   - conf: 数据库配置
//   - debuggerLogger: debugger日志记录器
//   - opts: 可选参数，第一个参数为配置别名
//
// 返回：
//   - *Instance: SQLite实例
func NewWithDebugger(conf jcbaseGo.SqlLiteStruct, debuggerLogger debugger.LoggerInterface, opts ...string) (*Instance, error) {
	instance, err := New(conf, opts...)
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
// 调试模式优先级：contextAware > 固定debuggerLogger > debug标志
// 这是一个只读方法，不会修改实例状态
func (c *Instance) GetDb() *gorm.DB {
	if c.Db == nil {
		log.Println("Database connection is nil")
		return nil
	}
	return c.applyDebugMode(c.Db)
}

// applyDebugMode 应用调试模式配置
func (c *Instance) applyDebugMode(db *gorm.DB) *gorm.DB {
	if db == nil {
		return nil
	}

	if c.contextAware {
		return c.applySQLLogging(db)
	}

	if c.debuggerLogger != nil {
		return c.applySQLLogging(db)
	}

	if c.debug {
		db = db.Debug()
	}

	return db
}

// applySQLLogging 按当前模式确保GORM实例已挂载SQL日志Logger
//
// 仅当Logger缺失或模式不匹配时才重新注入，避免每次 GetDb 都重复替换并重复打印启用日志。
func (c *Instance) applySQLLogging(db *gorm.DB) *gorm.DB {
	// 已挂载且模式一致时无需重建
	if gl, ok := db.Config.Logger.(*orm.GormDebuggerLogger); ok && gl.IsContextAware() == c.contextAware {
		return db
	}

	if c.contextAware {
		return orm.EnableContextAwareSQLLogging(db, c.debuggerLogger, c.sqlLogOpts...)
	}

	if c.debuggerLogger != nil && c.debuggerLogger.GetLevel() > debugger.LevelSilent {
		return orm.EnableSQLLogging(db, c.debuggerLogger, c.sqlLogOpts...)
	}

	return db
}

// SetDebuggerLogger 设置debugger日志记录器（固定Logger模式）
//
// 本方法为模式A入口：所有SQL日志固定写入传入的Logger。
// 每次调用都会替换GORM实例的Logger，高并发HTTP场景下并发调用会互相覆盖，
// 需要按请求归属SQL日志的项目请改用 EnableContextAwareSQLLogging。
//
// 参数：
//   - debuggerLogger: debugger日志记录器实例
func (i *Instance) SetDebuggerLogger(debuggerLogger debugger.LoggerInterface) {
	i.debuggerLogger = debuggerLogger
	i.contextAware = false
	i.sqlLogOpts = nil

	if i.Db != nil && debuggerLogger != nil {
		// 启用SQL日志记录
		i.Db = orm.EnableSQLLogging(i.Db, debuggerLogger)
	}
}

// GetDebuggerLogger 获取debugger日志记录器
func (c *Instance) GetDebuggerLogger() debugger.LoggerInterface {
	return c.debuggerLogger
}

// EnableSQLLogging 为当前数据库实例启用SQL日志记录（固定Logger模式）
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
		c.contextAware = false
		c.sqlLogOpts = opts
		c.Db = orm.EnableSQLLogging(c.Db, debuggerLogger, opts...)
	}
	return c
}

// EnableContextAwareSQLLogging 为当前数据库实例启用上下文感知的SQL日志记录（模式B）
//
// 本方法全局只需在初始化阶段调用一次：运行期每条SQL会依据 context.Context 中
// 携带的请求级Logger路由到对应请求的日志中，并发请求之间互不覆盖。
// 需要 debugger 中间件已开启，它会把请求级Logger自动注入 request 的 context。
//
// 参数：
//   - fallbackLogger: 兜底Logger，context 中无Logger时使用；可为 nil 表示丢弃
//     （建议传 dbg.GetMainLogger()，使定时任务等无请求上下文的SQL也能落盘）
//   - opts: 可选参数，语义同 EnableSQLLogging（日志级别、慢查询阈值）
//
// 返回：
//   - *Instance: 已启用上下文感知SQL日志的数据库实例
//
// 使用示例：
//
//	library.Sqlite.EnableContextAwareSQLLogging(dbg.GetMainLogger())
//	library.Sqlite.EnableContextAwareSQLLogging(dbg.GetMainLogger(), debugger.LevelInfo, 100*time.Millisecond)
func (c *Instance) EnableContextAwareSQLLogging(fallbackLogger debugger.LoggerInterface, opts ...interface{}) *Instance {
	c.debuggerLogger = fallbackLogger
	c.contextAware = true
	c.sqlLogOpts = opts

	if c.Db != nil {
		c.Db = orm.EnableContextAwareSQLLogging(c.Db, fallbackLogger, opts...)
	}

	return c
}

// IsContextAware 判断当前实例是否处于上下文感知的SQL日志模式
//
// 返回：
//   - bool: 上下文感知模式返回 true，固定Logger模式返回 false
func (c *Instance) IsContextAware() bool {
	return c.contextAware
}

// GetConf 返回当前实例的原始配置结构。
func (c *Instance) GetConf() interface{} {
	return c.Conf
}

// GetAllTableName 查询并返回当前数据库中的所有表名。
func (c *Instance) GetAllTableName() (tableNames []string, err error) {
	// 如果有错误，就不再执行
	if len(c.Errors) > 0 {
		return
	}

	err = c.Db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tableNames).Error
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
		b := &base.SqliteBaseModel{}
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
