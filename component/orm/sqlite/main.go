package sqlite

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Instance struct {
	Conf   jcbaseGo.SqlLiteStruct
	Db     *gorm.DB
	debug  bool // 是否开启调试模式
	Errors []error
}

// New 获取新的数据库连接
func New(Conf jcbaseGo.SqlLiteStruct, opts ...string) (i *Instance) {
	i = &Instance{}

	alias := "main"
	if len(opts) > 0 && opts[0] != "" {
		alias = opts[0]
	}

	err := helper.CheckAndSetDefault(&Conf)
	jcbaseGo.PanicIfError(err)

	// 获取dbFile的绝对路径
	fileNameFull, err := filepath.Abs(Conf.DbFile)
	jcbaseGo.PanicIfError(err)

	// 检查目录是否存在，如果不存在则创建
	_, err = helper.NewFile(&helper.File{Path: fileNameFull}).DirExists(true)
	jcbaseGo.PanicIfError(err)

	// 判断dbConfig是否为空
	if Conf.DbFile == "" {
		log.Panic(errors.New("dbConfig is empty"))
		return
	}

	// 创建数据库连接
	db, err := gorm.Open(sqlite.Open(Conf.DbFile), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: Conf.DisableForeignKeyConstraintWhenMigrating,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   Conf.TablePrefix,   // 表名前缀，`User`表为`t_users`
			SingularTable: Conf.SingularTable, // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	})
	jcbaseGo.PanicIfError(err)

	// 配置连接池参数，防止连接泄漏
	sqlDB, err := db.DB()
	if err == nil {
		// 设置最大连接数（SQLite通常只需要少量连接）
		sqlDB.SetMaxOpenConns(1)
		// 设置最大空闲连接数
		sqlDB.SetMaxIdleConns(1)
		// 设置连接最大生命周期（10分钟）
		sqlDB.SetConnMaxLifetime(10 * time.Minute)
		// 设置空闲连接超时时间（5分钟）
		sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	}

	i.Conf = Conf
	i.Db = db

	// 将配置信息储存到环境变量
	envStr := ""
	helper.Json(Conf).ToString(&envStr)
	err = os.Setenv("jc_sql_lite_"+alias, envStr)
	jcbaseGo.PanicIfError(err)

	return
}

// Debug 设置调试模式
func (c *Instance) Debug() *Instance {
	c.debug = true
	return c
}

// GetDb 获取db
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

// GetConf 获取配置信息
func (c *Instance) GetConf() interface{} {
	return c.Conf
}

// GetAllTableName 获取所有表名
func (c *Instance) GetAllTableName() (tableNames []string, err error) {
	// 如果有错误，就不再执行
	if len(c.Errors) > 0 {
		return
	}

	err = c.Db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tableNames).Error
	return
}

// TableName 获取表名，
// param tableName string 表名
// param quotes bool 是否加上反单引号
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

// AddError 添加错误到上下文
func (c *Instance) AddError(err error) {
	if err != nil {
		c.Errors = append(c.Errors, err)
	}
}

// Error 获取错误
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
