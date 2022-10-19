package mysql

import (
	"fmt"
	"github.com/jcbowen/jcbaseGo"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
)

var Db *gorm.DB

func init() {
	var err error
	Db, err = GetDb(jcbaseGo.Config.Get().Db)

	if err != nil {
		log.Panic(err)
	}
}

type AllTableName struct {
	TableName string `gorm:"table_name"`
}

// GetDSN 获取DataSourceName
func getDSN(dbConfig jcbaseGo.DbStruct) (dsn string) {
	dsn = "%s:%s@%s(%s:%s)/%s?charset=%s&parseTime=True&loc=Local"
	dsn = fmt.Sprintf(dsn, dbConfig.Username, dbConfig.Password, dbConfig.Protocol, dbConfig.Host, dbConfig.Port, dbConfig.Dbname, dbConfig.Charset)
	return
}

// GetDb 获取数据库连接
func GetDb(dbConfig jcbaseGo.DbStruct) (db *gorm.DB, err error) {
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
