// Package mysql 提供 MySQL 数据库的 ORM 封装，基于 GORM，包含连接创建、连接池配置、表名处理与分页查询等辅助方法。
package mysql

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
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
}

// getDSN 拼接 Data Source Name（数据源）字符串，基于提供的 `DbStruct`。
func getDSN(dbConfig jcbaseGo.DbStruct) (dsn string) {
	// 拼接dsn
	dsn = "%s:%s@%s(%s:%s)/%s?charset=%s&parseTime=%s&loc=Local"
	dsn = fmt.Sprintf(dsn, dbConfig.Username, dbConfig.Password, dbConfig.Protocol, dbConfig.Host, dbConfig.Port, dbConfig.Dbname, dbConfig.Charset, dbConfig.ParseTime)

	return
}

// New 创建一个 MySQL 实例并建立数据库连接；
// 可通过可选参数 `opts[0]` 指定配置别名以存入环境变量。
func New(dbConfig jcbaseGo.DbStruct, opts ...string) *Instance {
	context := &Instance{}

	alias := "db"
	if len(opts) > 0 && opts[0] != "" {
		alias = opts[0]
	}

	err := helper.CheckAndSetDefault(&dbConfig)
	jcbaseGo.PanicIfError(err)

	// 判断dbConfig是否为空
	if dbConfig.Dbname == "" {
		context.AddError(errors.New("dbConfig is empty"))
		return context
	}

	dsn := getDSN(dbConfig)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: dbConfig.DisableForeignKeyConstraintWhenMigrating,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   dbConfig.TablePrefix,   // 表名前缀，`User`表为`t_users`
			SingularTable: dbConfig.SingularTable, // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	})
	jcbaseGo.PanicIfError(err)

	// 配置连接池参数，防止连接泄漏和长时间锁定
	sqlDB, err := db.DB()
	if err == nil {
		// 设置最大连接数
		sqlDB.SetMaxOpenConns(100)
		// 设置最大空闲连接数
		sqlDB.SetMaxIdleConns(10)
		// 设置连接最大生命周期（5分钟）
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
		// 设置空闲连接超时时间（3分钟）
		sqlDB.SetConnMaxIdleTime(3 * time.Minute)
	}

	context.Dsn = dsn
	context.Conf = dbConfig
	context.Db = db

	// 将配置信息储存到环境变量
	envStr := ""
	helper.Json(dbConfig).ToString(&envStr)
	err = os.Setenv("jc_mysql_"+alias, envStr)
	jcbaseGo.PanicIfError(err)

	return context
}

// NewWithDebugger 创建MySQL实例并集成debugger日志记录
// 参数：
//   - dbConfig: 数据库配置
//   - debuggerLogger: debugger日志记录器
//   - opts: 可选参数，第一个参数为配置别名
//
// 返回：
//   - *Instance: MySQL实例
func NewWithDebugger(dbConfig jcbaseGo.DbStruct, debuggerLogger debugger.LoggerInterface, opts ...string) *Instance {
	instance := New(dbConfig, opts...)
	if instance.Db != nil && debuggerLogger != nil {
		instance.SetDebuggerLogger(debuggerLogger)
	}
	return instance
}

// Debug 开启调试模式，后续通过 `GetDb()` 获取的 *gorm.DB 将启用 Debug。
func (c *Instance) Debug() *Instance {
	c.debug = true
	return c
}

// GetDb 返回当前数据库连接；若已开启调试模式返回 Debug 包装的连接。
func (c *Instance) GetDb() *gorm.DB {
	if c.Db == nil {
		log.Println("Database connection is nil")
		return nil
	}
	db := c.Db
	if c.debug {
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
