package jcbaseGo

import (
	"github.com/jcbowen/jcbaseGo/component/helper"
	"gorm.io/gorm"
	"os"
	"time"
)

// SQLLiteModel gorm基础模型
type SQLLiteModel struct {
	Id        uint   `gorm:"column:id;type:INTEGER;primaryKey;autoIncrement" json:"id"`
	UpdatedAt string `gorm:"column:updated_at;type:STRING;default:NULL" json:"updated_at"`       // 更新时间
	CreatedAt string `gorm:"column:created_at;type:STRING;default:NULL" json:"created_at"`       // 创建时间
	DeletedAt string `gorm:"column:deleted_at;type:STRING;index;default:NULL" json:"deleted_at"` // 删除时间
}

func (m *SQLLiteModel) ConfigAlias() string {
	return "main"
}

func (m *SQLLiteModel) GetTableName(modelName string) string {
	var dbConfig SqlLiteStruct
	dbConfigStr := os.Getenv("jc_sql_lite_" + m.ConfigAlias())
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

func (m *SQLLiteModel) BeforeCreate(tx *gorm.DB) (err error) {
	strTime := time.Now().Format("2006-01-02 15:04:05")
	m.CreatedAt = strTime
	m.UpdatedAt = strTime
	return
}

func (m *SQLLiteModel) BeforeUpdate(tx *gorm.DB) (err error) {
	m.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
	return
}
