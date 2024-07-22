package jcbaseGo

import (
	"github.com/jcbowen/jcbaseGo/component/helper"
	"gorm.io/gorm"
	"os"
	"time"
)

// MysqlModel gorm基础模型
type MysqlModel struct {
	Id        uint   `gorm:"column:id;type:INT(11) UNSIGNED;primaryKey;autoIncrement" json:"id"`
	UpdatedAt string `gorm:"column:updated_at;type:DATETIME;default:NULL;comment:更新时间" json:"updated_at"`       // 更新时间
	CreatedAt string `gorm:"column:created_at;type:DATETIME;default:NULL;comment:创建时间" json:"created_at"`       // 创建时间
	DeletedAt string `gorm:"column:deleted_at;type:DATETIME;index;default:NULL;comment:删除时间" json:"deleted_at"` // 删除时间
}

func (m *MysqlModel) ConfigAlias() string {
	return "db"
}

func (m *MysqlModel) GetTableName(modelName string) string {
	var dbConfig DbStruct
	dbConfigStr := os.Getenv("jc_mysql_" + m.ConfigAlias())
	helper.JsonString(dbConfigStr).ToStruct(&dbConfig)

	// 获取表前缀
	prefix := dbConfig.TablePrefix

	// 转换为小写字母并添加下划线
	tableName := helper.NewStr(modelName).ConvertCamelToSnake()

	if !dbConfig.SingularTable {
		tableName += "s"
	}

	return prefix + tableName
}

func (m *MysqlModel) BeforeCreate(tx *gorm.DB) (err error) {
	strTime := time.Now().Format("2006-01-02 15:04:05")
	m.CreatedAt = strTime
	m.UpdatedAt = strTime
	return
}

func (m *MysqlModel) BeforeUpdate(tx *gorm.DB) (err error) {
	m.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
	return
}
