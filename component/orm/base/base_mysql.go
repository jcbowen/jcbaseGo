package base

import (
	"reflect"
)

// MysqlBaseModel gorm基础模型
// 继承 CommonBaseModel 的 BeforeCreate、BeforeUpdate 等方法，
// 默认配置别名为 "db"。
type MysqlBaseModel struct {
	CommonBaseModel
	//Id        uint   `gorm:"column:id;type:INT(11) UNSIGNED;primaryKey;autoIncrement" json:"id"`
	//UpdatedAt string `gorm:"column:updated_at;type:DATETIME;default:NULL;comment:更新时间" json:"updated_at"`
	//CreatedAt string `gorm:"column:created_at;type:DATETIME;default:NULL;comment:创建时间" json:"created_at"`
	//DeletedAt string `gorm:"column:deleted_at;type:DATETIME;index;default:NULL;comment:删除时间" json:"deleted_at"`
}

// ModelParse 解析模型信息（供 CRUD trait 使用）
// 复用 CommonBaseModel 的解析逻辑，默认优先读取 MySQL 配置。
//
// 参数：
//   - modelType: 具体模型的反射类型
//
// 返回值：
//   - tableName: 数据表名称
//   - fields: 模型字段列表
//   - softDeleteField: 软删除字段名
//   - softDeleteCondition: 软删除条件
func (b *MysqlBaseModel) ModelParse(modelType reflect.Type) (tableName string, fields []string, softDeleteField string, softDeleteCondition string) {
	return modelParse(b, modelType)
}

// dbConfigSources 返回 MySQL 配置读取顺序
// 为保持向后兼容，MySQL 模型只读取 MySQL 环境变量配置，不回退到 SQLite。
//
// 返回值：
//   - []string: 数据库类型标识列表
func (b *MysqlBaseModel) dbConfigSources() []string {
	return []string{"mysql"}
}
