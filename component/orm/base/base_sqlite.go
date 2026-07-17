package base

import (
	"reflect"
)

// SqliteBaseModel gorm基础模型
// 继承 CommonBaseModel 的 BeforeCreate、BeforeUpdate 等方法，
// 默认配置别名为 "main"。
type SqliteBaseModel struct {
	CommonBaseModel
	//Id        uint   `gorm:"column:id;type:INTEGER;primaryKey;autoIncrement" json:"id"`
	//UpdatedAt string `gorm:"column:updated_at;type:STRING;default:NULL" json:"updated_at"`
	//CreatedAt string `gorm:"column:created_at;type:STRING;default:NULL" json:"created_at"`
	//DeletedAt string `gorm:"column:deleted_at;type:STRING;index;default:NULL" json:"deleted_at"`
}

// GetConfigAlias 获取数据库配置别名
// SQLite 默认使用 "main" 作为配置别名。
//
// 参数：
//   - model: 模型实例
//
// 返回值：
//   - string: 配置别名
func (b *SqliteBaseModel) GetConfigAlias(model interface{}) string {
	if aliaser, ok := model.(interface{ ConfigAlias() string }); ok {
		return aliaser.ConfigAlias()
	}
	return "main"
}

// ModelParse 解析模型信息（供 CRUD trait 使用）
// 复用 CommonBaseModel 的解析逻辑，默认优先读取 SQLite 配置。
//
// 参数：
//   - modelType: 具体模型的反射类型
//
// 返回值：
//   - tableName: 数据表名称
//   - fields: 模型字段列表
//   - softDeleteField: 软删除字段名
//   - softDeleteCondition: 软删除条件
func (b *SqliteBaseModel) ModelParse(modelType reflect.Type) (tableName string, fields []string, softDeleteField string, softDeleteCondition string) {
	return modelParse(b, modelType)
}

// dbConfigSources 返回 SQLite 配置读取顺序
// 为保持向后兼容，SQLite 模型只读取 SQLite 环境变量配置，不回退到 MySQL。
//
// 返回值：
//   - []string: 数据库类型标识列表
func (b *SqliteBaseModel) dbConfigSources() []string {
	return []string{"sqlite"}
}
