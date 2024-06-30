package sqlLite

import (
	"errors"
	"fmt"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/helper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
	"path/filepath"
)

type Instance struct {
	Conf   jcbaseGo.SqlLiteStruct
	Db     *gorm.DB
	Errors []error
}

// New 获取新的数据库连接
func New(Conf jcbaseGo.SqlLiteStruct) (i *Instance) {
	i = &Instance{}

	err := helper.CheckAndSetDefault(&Conf)
	if err != nil {
		i.AddError(err)
		return
	}

	// 获取dbFile的绝对路径
	fileNameFull, err := filepath.Abs(Conf.DbFile)
	if err != nil {
		log.Panic(err)
	}

	// 检查目录是否存在，如果不存在则创建
	_, err = helper.DirExists(fileNameFull, true, 0755)
	if err != nil {
		i.AddError(err)
		return
	}

	// 判断dbConfig是否为空
	if Conf.DbFile == "" {
		i.AddError(errors.New("dbConfig is empty"))
		return
	}

	// 创建数据库连接
	db, err := gorm.Open(sqlite.Open(Conf.DbFile), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   Conf.TablePrefix,             // 表名前缀，`User`表为`t_users`
			SingularTable: Conf.SingularTable == "true", // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	})

	i.Conf = Conf
	i.Db = db
	i.AddError(err)

	return
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

// GetDb 获取db
func (c *Instance) GetDb() *gorm.DB {
	return c.Db
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
