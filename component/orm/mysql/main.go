package mysql

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type AllTableName struct {
	TableName string `gorm:"table_name"`
}

type Instance struct {
	Dsn    string
	Conf   jcbaseGo.DbStruct
	Db     *gorm.DB
	debug  bool // 是否开启debug
	Errors []error
}

// GetDSN 拼接DataSourceName
func getDSN(dbConfig jcbaseGo.DbStruct) (dsn string) {
	// 拼接dsn
	dsn = "%s:%s@%s(%s:%s)/%s?charset=%s&parseTime=%s&loc=Local"
	dsn = fmt.Sprintf(dsn, dbConfig.Username, dbConfig.Password, dbConfig.Protocol, dbConfig.Host, dbConfig.Port, dbConfig.Dbname, dbConfig.Charset, dbConfig.ParseTime)

	return
}

// New 获取新的数据库连接
func New(dbConfig jcbaseGo.DbStruct) *Instance {
	context := &Instance{}

	err := helper.CheckAndSetDefault(&dbConfig)
	jcbaseGo.PanicIfError(err)

	// 判断dbConfig是否为空
	if dbConfig.Dbname == "" {
		context.AddError(errors.New("dbConfig is empty"))
		return context
	}

	dsn := getDSN(dbConfig)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   dbConfig.TablePrefix,   // 表名前缀，`User`表为`t_users`
			SingularTable: dbConfig.SingularTable, // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	})
	jcbaseGo.PanicIfError(err)

	context.Dsn = dsn
	context.Conf = dbConfig
	context.Db = db

	// 将配置信息储存到环境变量
	envStr := ""
	helper.Json(dbConfig).ToString(&envStr)
	err = os.Setenv("jc_mysql_"+dbConfig.Alias, envStr)
	jcbaseGo.PanicIfError(err)

	return context
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

// GetAllTableName 获取所有表名
func (c *Instance) GetAllTableName() (tableNames []AllTableName, err error) {
	// 如果有错误，就不再执行
	if len(c.Errors) > 0 {
		return
	}

	err = c.Db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema='" + c.Conf.Dbname + "' AND table_type='base table'").Scan(&tableNames).Error
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

func (c *Instance) AddError(err error) {
	if err != nil {
		c.Errors = append(c.Errors, err)
	}
}

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
