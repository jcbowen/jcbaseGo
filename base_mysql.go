package jcbaseGo

import (
	"github.com/jcbowen/jcbaseGo/component/helper"
	"gorm.io/gorm"
	"os"
	"reflect"
	"strings"
	"time"
)

// MysqlBaseModel gorm基础模型
type MysqlBaseModel struct {
	//Id        uint   `gorm:"column:id;type:INT(11) UNSIGNED;primaryKey;autoIncrement" json:"id"`
	//UpdatedAt string `gorm:"column:updated_at;type:DATETIME;default:NULL;comment:更新时间" json:"updated_at"`
	//CreatedAt string `gorm:"column:created_at;type:DATETIME;default:NULL;comment:创建时间" json:"created_at"`
	//DeletedAt string `gorm:"column:deleted_at;type:DATETIME;index;default:NULL;comment:删除时间" json:"deleted_at"`
}

func (b *MysqlBaseModel) ConfigAlias() string {
	return "db"
}

func (b *MysqlBaseModel) ModelParse(modelType reflect.Type) (tableName string, fields []string) {
	// ----- 获取数据表名称 ----- /
	var dbConfig DbStruct
	dbConfigStr := os.Getenv("jc_mysql_" + b.ConfigAlias())
	helper.JsonString(dbConfigStr).ToStruct(&dbConfig)

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
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		gormTag := field.Tag.Get("gorm")
		columnName := getColumnFromTag(gormTag)
		if columnName != "" {
			fields = append(fields, columnName)
		} else if field.Name != "MysqlBaseModel" {
			// 如果没有定义gorm标签，则使用字段名称转换为下划线格式
			fields = append(fields, helper.NewStr(field.Name).ConvertCamelToSnake())
		}
	}

	return
}

func (b *MysqlBaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	strTime := time.Now().Format("2006-01-02 15:04:05")
	setFieldIfExist(tx.Statement.Dest, "CreatedAt", strTime)
	setFieldIfExist(tx.Statement.Dest, "UpdatedAt", strTime)
	return
}

func (b *MysqlBaseModel) BeforeUpdate(tx *gorm.DB) (err error) {
	setFieldIfExist(tx.Statement.Dest, "UpdatedAt", time.Now().Format("2006-01-02 15:04:05"))
	return
}

func getColumnFromTag(tag string) string {
	tags := strings.Split(tag, ";")
	for _, t := range tags {
		if strings.HasPrefix(t, "column:") {
			return strings.TrimPrefix(t, "column:")
		}
	}
	return ""
}

func setFieldIfExist(model interface{}, fieldName string, value string) {
	modelValue := reflect.ValueOf(model)

	if modelValue.Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
	}

	field := modelValue.FieldByName(fieldName)
	if field.IsValid() && field.CanSet() && field.Kind() == reflect.String {
		field.SetString(value)
	}
}
