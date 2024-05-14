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

type AllTableName struct {
	TableName string `gorm:"table_name"`
}

type Helper struct {
	Dsn    string
	Conf   jcbaseGo.DbStruct
	Db     *gorm.DB
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
	dsn = "%s:%s@%s(%s:%s)/%s?charset=%s&parseTime=%s&loc=Local"
	dsn = fmt.Sprintf(dsn, dbConfig.Username, dbConfig.Password, dbConfig.Protocol, dbConfig.Host, dbConfig.Port, dbConfig.Dbname, dbConfig.Charset, dbConfig.ParseTime)

	return
}

// New 获取新的数据库连接
func New(dbConfig jcbaseGo.DbStruct) *Helper {
	context := &Helper{}

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

func (c *Helper) AddError(err error) {
	if err != nil {
		c.Errors = append(c.Errors, err)
	}
}

func (c *Helper) Error() []error {
	// 过滤掉c.Errors中的nil
	var errs []error
	for _, err := range c.Errors {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (c *Helper) GetDb() *gorm.DB {
	return c.Db
}

func (c *Helper) GetAllTableName() (tableNames []AllTableName) {
	// 如果有错误，就不再执行
	if len(c.Errors) > 0 {
		return
	}

	c.Db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema='" + c.Conf.Dbname + "' AND table_type='base table'").Scan(&tableNames)
	return
}

// TableName 获取表名，
//
// param tableName string 表名
//
// param quotes bool 是否加上反单引号
func (c *Helper) TableName(tableName *string, quotes ...bool) *Helper {
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

type Context = Helper
