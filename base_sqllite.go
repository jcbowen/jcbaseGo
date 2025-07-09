package jcbaseGo

import (
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/jcbowen/jcbaseGo/component/helper"
	"gorm.io/gorm"
)

// SQLLiteBaseModel gorm基础模型
type SQLLiteBaseModel struct {
	//Id        uint   `gorm:"column:id;type:INTEGER;primaryKey;autoIncrement" json:"id"`
	//UpdatedAt string `gorm:"column:updated_at;type:STRING;default:NULL" json:"updated_at"`
	//CreatedAt string `gorm:"column:created_at;type:STRING;default:NULL" json:"created_at"`
	//DeletedAt string `gorm:"column:deleted_at;type:STRING;index;default:NULL" json:"deleted_at"`
}

func (b *SQLLiteBaseModel) ConfigAlias() string {
	return "main"
}

func (b *SQLLiteBaseModel) ModelParse(modelType reflect.Type) (tableName string, fields []string, softDeleteCondition string) {
	// ----- 获取数据表名称 ----- /
	var dbConfig SqlLiteStruct
	dbConfigStr := os.Getenv("jc_sql_lite_" + b.ConfigAlias())
	helper.Json(dbConfigStr).ToStruct(&dbConfig)

	// 获取表前缀
	prefix := dbConfig.TablePrefix

	// 转换为小写字母并添加下划线
	convertModelName := helper.NewStr(modelType.Name()).ConvertCamelToSnake()

	if !dbConfig.SingularTable {
		convertModelName += "s"
	}

	tableName = prefix + convertModelName

	// ----- 获取数据表所有字段 ----- /
	fields = []string{}
	softDeleteCondition = "IS NULL" // 默认软删除条件

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		gormTag := field.Tag.Get("gorm")
		columnName := getColumnFromTag(gormTag)

		if columnName != "" {
			fields = append(fields, columnName)

			// 检查是否为 deleted_at 字段，解析其默认值
			if columnName == "deleted_at" {
				defaultValue := getDefaultFromTag(gormTag)
				if defaultValue == "0000-00-00 00:00:00" || strings.Contains(gormTag, "default:0000-00-00 00:00:00") {
					// 如果默认值是 0000-00-00 00:00:00，则用这个作为软删除判断条件
					softDeleteCondition = "= '0000-00-00 00:00:00'"
				}
			}
		} else if field.Name != "SQLLiteBaseModel" {
			// 如果没有定义gorm标签，则使用字段名称转换为下划线格式
			fieldName := helper.NewStr(field.Name).ConvertCamelToSnake()
			fields = append(fields, fieldName)
		}
	}

	return
}

func (b *SQLLiteBaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	strTime := time.Now().Format("2006-01-02 15:04:05")
	setFieldIfExist(tx.Statement.Dest, "CreatedAt", strTime)
	setFieldIfExist(tx.Statement.Dest, "UpdatedAt", strTime)
	return
}

func (b *SQLLiteBaseModel) BeforeUpdate(tx *gorm.DB) (err error) {
	setFieldIfExist(tx.Statement.Dest, "UpdatedAt", time.Now().Format("2006-01-02 15:04:05"))
	return
}
