package mysql

import (
	"github.com/jcbowen/jcbaseGo"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
)

var Db *gorm.DB

func init() {
	dsn := jcbaseGo.Config.GetDSN()

	var err error
	Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   jcbaseGo.Config.Db.TablePrefix, // 表名前缀，`User`表为`t_users`
			SingularTable: true,                           // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	})

	if err != nil {
		log.Panic(err)
	}
}

func Check() (gormDB *gorm.DB) {
	gormDB = Db
	return
}

type AllTableName struct {
	TableName string `gorm:"table_name"`
}

// GetAllTableName 获取数据库中所有的表名
func GetAllTableName() (result []AllTableName) {
	Db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema='" + jcbaseGo.Config.Db.Dbname + "' AND table_type='base table'").Scan(&result)
	return
}
