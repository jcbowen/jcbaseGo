package mysql

import (
	"fmt"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/helper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type DB = gorm.DB

var Db *DB

type AllTableName struct {
	TableName string `gorm:"table_name"`
}

func Get() *DB {
	if Db == nil {
		var err error
		Db, err = GetDb(jcbaseGo.Config.Get().Db)

		if err != nil {
			panic(err)
		}
	}

	return Db
}

// GetDSN 获取DataSourceName
func getDSN(dbConfig jcbaseGo.DbStruct) (dsn string) {
	dsn = "%s:%s@%s(%s:%s)/%s?charset=%s&parseTime=True&loc=Local"
	dsn = fmt.Sprintf(dsn, dbConfig.Username, dbConfig.Password, dbConfig.Protocol, dbConfig.Host, dbConfig.Port, dbConfig.Dbname, dbConfig.Charset)
	return
}

// GetDb 获取数据库连接
func GetDb(dbConfig jcbaseGo.DbStruct) (db *DB, err error) {
	dsn := getDSN(dbConfig)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   dbConfig.TablePrefix, // 表名前缀，`User`表为`t_users`
			SingularTable: true,                 // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	})

	return
}

// GetAllTableName 获取数据库中所有的表名
func GetAllTableName() (result []AllTableName) {
	Db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema='" + jcbaseGo.Config.Db.Dbname + "' AND table_type='base table'").Scan(&result)
	return
}

// TableName 获取表名，
func TableName(tableName string, quotes bool) string {
	tablePrefix := jcbaseGo.Config.Db.TablePrefix
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
