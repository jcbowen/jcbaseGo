package mysql

import (
	"errors"
	"fmt"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/helper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
)

type DB = gorm.DB

type AllTableName struct {
	TableName string `gorm:"table_name"`
}

type Context struct {
	Dsn    string
	Conf   jcbaseGo.DbStruct
	Db     *DB
	Errors []error
}

// GetDSN 拼接DataSourceName
func getDSN(dbConfig jcbaseGo.DbStruct) (dsn string) {
	err := helper.CheckAndSetDefault(&dbConfig)
	if err != nil {
		log.Fatalln(err)
		return ""
	}

	// 拼接dsn
	dsn = "%s:%s@%s(%s:%s)/%s?charset=%s&parseTime=True&loc=Local"
	dsn = fmt.Sprintf(dsn, dbConfig.Username, dbConfig.Password, dbConfig.Protocol, dbConfig.Host, dbConfig.Port, dbConfig.Dbname, dbConfig.Charset)

	return
}

// New 获取新的数据库连接
func New(dbConfig jcbaseGo.DbStruct) *Context {
	context := &Context{}

	// 判断dbConfig是否为空
	if dbConfig.Dbname == "" {
		context.AddError(errors.New("dbConfig is empty"))
		return context
	}

	dsn := getDSN(dbConfig)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   dbConfig.TablePrefix, // 表名前缀，`User`表为`t_users`
			SingularTable: true,                 // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	})

	context.Dsn = dsn
	context.Conf = dbConfig
	context.Db = db
	context.AddError(err)

	return context
}

func (c *Context) AddError(err error) {
	if err != nil {
		c.Errors = append(c.Errors, err)
	}
}

func (c *Context) Error() []error {
	// 过滤掉c.Errors中的nil
	var errs []error
	for _, err := range c.Errors {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (c *Context) GetDb() *DB {
	return c.Db
}

func (c *Context) GetAllTableName() (tableNames []AllTableName) {
	// 如果有错误，就不再执行
	if len(c.Errors) > 0 {
		return
	}

	c.Db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema='" + jcbaseGo.Config.Db.Dbname + "' AND table_type='base table'").Scan(&tableNames)
	return
}

// TableName 获取表名，
//
// param tableName string 表名
//
// param quotes bool 是否加上反单引号
func (c *Context) TableName(tableName *string, quotes ...bool) *Context {
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

// ----- 弃用 ----- /

var (
	// Db 数据库连接
	// Deprecated: As of jcbaseGo 0.3, this variable is no longer used
	Db *DB
	// conf 数据库配置
	// Deprecated: As of jcbaseGo 0.3, this variable is no longer used
	conf jcbaseGo.DbStruct
)

// Get 获取数据库连接
// Deprecated: As of jcbaseGo 0.3, this function simply calls New.GetDb
func Get() *DB {
	if Db == nil {
		var err error
		conf = jcbaseGo.Config.Get().Db
		Db = New(conf).GetDb()

		if err != nil {
			panic(err)
		}
	}

	return Db
}

// TableName 获取表名，
// Deprecated: As of jcbaseGo 0.3, this function simply calls New.TableName
func TableName(tableName string, quotes bool) string {
	tablePrefix := conf.TablePrefix
	// 如果已经有前缀了，就不再添加
	if len(tablePrefix) > 0 && helper.StringStartWith(tableName, tablePrefix) {
		tablePrefix = ""
	}

	if quotes {
		return "`" + tablePrefix + tableName + "`"
	} else {
		return tablePrefix + tableName
	}
}
