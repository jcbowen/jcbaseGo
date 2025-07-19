package base

import (
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"gorm.io/gorm"
)

// MysqlBaseModel gorm基础模型
type MysqlBaseModel struct {
	//Id        uint   `gorm:"column:id;type:INT(11) UNSIGNED;primaryKey;autoIncrement" json:"id"`
	//UpdatedAt string `gorm:"column:updated_at;type:DATETIME;default:NULL;comment:更新时间" json:"updated_at"`
	//CreatedAt string `gorm:"column:created_at;type:DATETIME;default:NULL;comment:创建时间" json:"created_at"`
	//DeletedAt string `gorm:"column:deleted_at;type:DATETIME;index;default:NULL;comment:删除时间" json:"deleted_at"`
}

func (b *MysqlBaseModel) GetConfigAlias(model interface{}) string {
	if aliaser, ok := model.(interface{ ConfigAlias() string }); ok {
		return aliaser.ConfigAlias()
	}
	return "db"
}

// ModelParse 解析模型信息（供 CRUD trait 使用）
// 参数说明：
//   - modelType reflect.Type: 模型的反射类型
//
// 返回值：
//   - tableName string: 数据表名称
//   - fields []string: 字段列表
//   - softDeleteField string: 软删除字段名
//   - softDeleteCondition string: 软删除条件
func (b *MysqlBaseModel) ModelParse(modelType reflect.Type) (tableName string, fields []string, softDeleteField string, softDeleteCondition string) {
	// ----- 获取数据表名称 ----- /
	// 通过反射创建模型实例
	model := reflect.New(modelType).Interface()

	var dbConfig jcbaseGo.DbStruct
	dbConfigStr := os.Getenv("jc_mysql_" + b.GetConfigAlias(model))
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
	softDeleteField = ""            // 软删除字段名，默认为空
	softDeleteCondition = "IS NULL" // 默认软删除条件

	// 用于记录 deleted_at 字段信息（默认软删除字段）
	var deletedAtField string
	var deletedAtCondition string

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		gormTag := field.Tag.Get("gorm")
		columnName := getColumnFromTag(gormTag)

		if columnName != "" {
			fields = append(fields, columnName)

			// 检查是否有 soft_delete 标签，优先使用自定义软删除配置
			if softDeleteTag := getSoftDeleteFromTag(gormTag); softDeleteTag != "" {
				softDeleteField = columnName
				softDeleteCondition = softDeleteTag
			} else if columnName == "deleted_at" {
				// 记录 deleted_at 字段信息（系统默认的软删除字段）
				deletedAtField = columnName
				defaultValue := getDefaultFromTag(gormTag)
				if defaultValue == "0000-00-00 00:00:00" || strings.Contains(gormTag, "default:0000-00-00 00:00:00") {
					// 如果默认值是 0000-00-00 00:00:00，则用这个作为软删除判断条件
					deletedAtCondition = "= '0000-00-00 00:00:00'"
				} else {
					// 默认使用 IS NULL 作为软删除条件
					deletedAtCondition = "IS NULL"
				}
			}
		} else if field.Name != "MysqlBaseModel" {
			// 如果没有定义gorm标签，则使用字段名称转换为下划线格式
			fieldName := helper.NewStr(field.Name).ConvertCamelToSnake()
			fields = append(fields, fieldName)
		}
	}

	// 如果没有通过 soft_delete 标签自定义软删除字段，但存在 deleted_at 字段，则使用系统默认的软删除配置
	if softDeleteField == "" && deletedAtField != "" {
		softDeleteField = deletedAtField
		softDeleteCondition = deletedAtCondition
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
