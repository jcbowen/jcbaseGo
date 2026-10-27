package base

import (
	"os"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/helper"
)

// CommonBaseModel 跨数据库通用基础模型
// 不绑定具体数据库驱动，字段由业务模型自行定义。
// 提供 CRUD trait 所需的 ModelParse 方法以及 GORM 创建/更新钩子。
type CommonBaseModel struct{}

// modelParseHelper 定义 ModelParse 所需的依赖行为，
// 用于在 CommonBaseModel、MysqlBaseModel、SqliteBaseModel 之间复用统一的解析逻辑。
type modelParseHelper interface {
	GetConfigAlias(model interface{}) string
	dbConfigSources() []string
}

// ModelParse 解析模型信息（供 CRUD trait 使用）
// 参数：
//   - modelType: 具体模型的反射类型
//
// 返回值：
//   - tableName: 数据表名称，优先使用 TableName() 方法
//   - fields: 模型字段列表（包含匿名嵌入结构体的字段）
//   - softDeleteField: 软删除字段名
//   - softDeleteCondition: 软删除条件
func (b *CommonBaseModel) ModelParse(modelType reflect.Type) (tableName string, fields []string, softDeleteField string, softDeleteCondition string) {
	return modelParse(b, modelType)
}

// modelParse 统一的模型解析实现
//
// 参数：
//   - b: 提供配置别名与数据源顺序的辅助对象
//   - modelType: 具体模型的反射类型
//
// 返回值：
//   - tableName: 数据表名称
//   - fields: 模型字段列表
//   - softDeleteField: 软删除字段名
//   - softDeleteCondition: 软删除条件
func modelParse(b modelParseHelper, modelType reflect.Type) (tableName string, fields []string, softDeleteField string, softDeleteCondition string) {
	model := reflect.New(modelType).Interface()

	// 获取表名，优先使用 TableName() 方法
	if tn, ok := model.(interface{ TableName() string }); ok && tn.TableName() != "" {
		tableName = tn.TableName()
	} else {
		// 从环境变量读取数据库配置中的表前缀与单复数设置
		alias := b.GetConfigAlias(model)
		prefix, singularTable := getTableNameConfig(alias, b.dbConfigSources())

		// 模型自定义表前缀优先级高于数据库配置
		if pfxCtrl, ok := model.(interface{ TablePrefix() string }); ok {
			prefix = pfxCtrl.TablePrefix()
		}

		// 转换为小写字母并添加下划线
		convertModelName := helper.NewStr(modelType.Name()).ConvertCamelToSnake()

		if !singularTable {
			convertModelName += "s"
		}

		// 拼接数据表名称
		tableName = prefix + convertModelName
	}

	fields = []string{}
	fieldSet := make(map[string]struct{})
	softDeleteField = ""
	softDeleteCondition = "IS NULL"

	var deletedAtField string
	var deletedAtCondition string

	parseFields(modelType, &fields, fieldSet, &softDeleteField, &softDeleteCondition, &deletedAtField, &deletedAtCondition)

	if softDeleteField == "" && deletedAtField != "" {
		softDeleteField = deletedAtField
		softDeleteCondition = deletedAtCondition
	}

	return
}

// GetConfigAlias 获取数据库配置别名
// 若模型实现了 ConfigAlias() string 方法，则使用模型自定义的别名；否则返回默认别名 "db"。
//
// 参数：
//   - model: 模型实例
//
// 返回值：
//   - string: 配置别名，对应环境变量中的 jc_mysql_<alias> 或 jc_sql_lite_<alias>
func (b *CommonBaseModel) GetConfigAlias(model interface{}) string {
	if aliaser, ok := model.(interface{ ConfigAlias() string }); ok {
		return aliaser.ConfigAlias()
	}
	return "db"
}

// dbConfigSources 返回表名配置读取顺序
// 子类（如 SqliteBaseModel）可通过覆盖此方法调整配置优先级。
//
// 返回值：
//   - []string: 数据库类型标识列表，当前支持 "mysql" 和 "sqlite"
func (b *CommonBaseModel) dbConfigSources() []string {
	return []string{"mysql", "sqlite"}
}

// getTableNameConfig 从环境变量读取数据库配置的表前缀与单复数设置
// 按照 sources 返回的顺序依次尝试读取对应数据库配置。
//
// 参数：
//   - alias: 配置别名
//   - sources: 数据库类型标识列表
//
// 返回值：
//   - prefix: 表前缀
//   - singularTable: 是否使用单数表名
func getTableNameConfig(alias string, sources []string) (prefix string, singularTable bool) {
	for _, source := range sources {
		switch source {
		case "mysql":
			var mysqlConfig jcbaseGo.DbStruct
			if mysqlConfigStr := os.Getenv("jc_mysql_" + alias); mysqlConfigStr != "" {
				helper.Json(mysqlConfigStr).ToStruct(&mysqlConfig)
				return mysqlConfig.TablePrefix, mysqlConfig.SingularTable
			}
		case "sqlite":
			var sqliteConfig jcbaseGo.SqlLiteStruct
			if sqliteConfigStr := os.Getenv("jc_sql_lite_" + alias); sqliteConfigStr != "" {
				helper.Json(sqliteConfigStr).ToStruct(&sqliteConfig)
				return sqliteConfig.TablePrefix, sqliteConfig.SingularTable
			}
		}
	}

	return "", false
}

// parseFields 递归解析结构体字段
//
// 参数：
//   - modelType: 待解析的结构体类型
//   - fields: 字段名列表指针
//   - fieldSet: 已解析字段名集合，用于去重
//   - softDeleteField: 软删除字段名指针
//   - softDeleteCondition: 软删除条件指针
//   - deletedAtField: 默认删除字段名指针
//   - deletedAtCondition: 默认删除条件指针
func parseFields(
	modelType reflect.Type,
	fields *[]string,
	fieldSet map[string]struct{},
	softDeleteField, softDeleteCondition *string,
	deletedAtField, deletedAtCondition *string,
) {
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// 匿名嵌入字段递归解析其内部字段
		if field.Anonymous {
			parseFields(field.Type, fields, fieldSet, softDeleteField, softDeleteCondition, deletedAtField, deletedAtCondition)
			continue
		}

		gormTag := field.Tag.Get("gorm")

		// GORM 忽略字段（gorm:"-" 或 gorm:"-:all" 等）不应加入字段列表
		if isGormIgnoredField(gormTag) {
			continue
		}

		columnName := getColumnFromTag(gormTag)

		if columnName != "" {
			if _, exists := fieldSet[columnName]; !exists {
				fieldSet[columnName] = struct{}{}
				*fields = append(*fields, columnName)
			}

			if hasSoftDeleteTag(gormTag) {
				*softDeleteField = columnName
				*softDeleteCondition = parseSoftDeleteCondition(gormTag)
			} else if columnName == "deleted_at" {
				*deletedAtField = columnName
				// 对系统默认的 deleted_at 字段保持保守策略，避免破坏旧项目软删除行为
				*deletedAtCondition = parseDeletedAtCondition(gormTag)
			}
		} else if !isBaseModelField(field) {
			// 过滤基础模型自身字段，避免被当作普通字段
			columnName = helper.NewStr(field.Name).ConvertCamelToSnake()
			if _, exists := fieldSet[columnName]; !exists {
				fieldSet[columnName] = struct{}{}
				*fields = append(*fields, columnName)
			}
		}
	}
}

// isBaseModelField 判断字段类型是否为基础模型自身字段
// 通过类型比对而非字段名硬编码，避免后续新增基础模型时遗漏。
//
// 参数：
//   - field: 结构体字段
//
// 返回值：
//   - bool: 是否为基础模型字段
func isBaseModelField(field reflect.StructField) bool {
	fieldType := field.Type
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	switch fieldType {
	case reflect.TypeOf(CommonBaseModel{}),
		reflect.TypeOf(MysqlBaseModel{}),
		reflect.TypeOf(SqliteBaseModel{}):
		return true
	}
	return false
}

// isGormIgnoredField 判断字段是否被 GORM 完全忽略
// 仅把 gorm:"-" 和 gorm:"-:all" 视为完全忽略；
// gorm:"-:create"、gorm:"-:update"、gorm:"-:migrate" 等部分忽略标签
// 不影响字段在 CRUD 字段列表中的存在性，避免 update/set-value 等场景误删可用字段。
//
// 参数：
//   - tag: gorm 标签字符串
//
// 返回值：
//   - bool: 是否被完全忽略
func isGormIgnoredField(tag string) bool {
	if tag == "-" {
		return true
	}
	for _, t := range strings.Split(tag, ";") {
		if t == "-" || t == "-:all" {
			return true
		}
	}
	return false
}

// parseSoftDeleteCondition 根据 gorm 标签解析自定义软删除条件
// 解析规则：
//   - soft_delete:条件   直接使用指定条件
//   - soft_delete        或无值标记时，根据 default 标签推断；无 default 或 default 为 NULL / 空字符串 则返回 IS NULL
//
// 参数：
//   - tag: gorm 标签字符串
//
// 返回值：
//   - string: 软删除条件，不会返回空字符串
func parseSoftDeleteCondition(tag string) string {
	// 如果显式指定了 soft_delete 条件，直接使用
	if cond := getSoftDeleteFromTag(tag); cond != "" {
		return cond
	}

	defaultValue := getDefaultFromTag(tag)
	if defaultValue == "" || strings.EqualFold(defaultValue, "NULL") || defaultValue == "''" {
		return "IS NULL"
	}
	if defaultValue == "0000-00-00 00:00:00" {
		return "= '0000-00-00 00:00:00'"
	}
	// 对 SQL 字符串中的单引号进行转义，防止拼接条件时出现语法错误
	defaultValue = strings.ReplaceAll(defaultValue, "'", "''")
	return "= '" + defaultValue + "'"
}

// parseDeletedAtCondition 解析系统默认 deleted_at 字段的软删除条件
// 为保持向后兼容，仅对以下情况特殊处理：
//   - default 为 0000-00-00 00:00:00 时返回 = '0000-00-00 00:00:00'
//   - 其他情况统一返回 IS NULL
//
// 参数：
//   - tag: gorm 标签字符串
//
// 返回值：
//   - string: 软删除条件
func parseDeletedAtCondition(tag string) string {
	defaultValue := getDefaultFromTag(tag)
	if defaultValue == "0000-00-00 00:00:00" || strings.Contains(tag, "default:0000-00-00 00:00:00") {
		return "= '0000-00-00 00:00:00'"
	}
	return "IS NULL"
}

// BeforeCreate 创建前钩子
// GORM 在创建记录前自动调用。若模型中不存在对应字段，设置操作会被静默忽略。
// 仅当 CreatedAt/Created/UpdatedAt/Updated 字段为空值时才填充当前时间；
// 已显式指定时间的字段（如数据迁移、导入历史数据）不会被覆盖，
// 对齐 GORM 原生 autoCreateTime 语义。被 Omit 的字段同样跳过。
//
// 参数：
//   - tx *gorm.DB: 当前 GORM 事务上下文，包含 Statement.Dest 与 Statement.Selects
//
// 返回值：
//   - err error: 错误信息，当前实现始终返回 nil
//
// 异常：
//   - 无显式异常，内部通过反射安全设置字段值
//
// 使用示例：
//
//	type User struct {
//	    base.CommonBaseModel
//	    Username string `gorm:"column:username" json:"username"`
//	}
//	db.Create(&user) // CreatedAt/UpdatedAt 为空时自动填充；已指定则保留
func (b *CommonBaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if tx == nil || tx.Statement == nil || tx.Statement.Dest == nil {
		return
	}

	strTime := time.Now().Format("2006-01-02 15:04:05")
	SetTimeFieldDefaults(tx, strTime, "CreatedAt", "Created", "UpdatedAt", "Updated")
	EnsureSelects(tx, "CreatedAt", "Created", "UpdatedAt", "Updated")
	return
}

// BeforeUpdate 更新前钩子
// GORM 在更新记录前自动调用。若模型中不存在对应字段，设置操作会被静默忽略。
// 仅当 UpdatedAt/Updated 字段为空值时才填充当前时间；已显式指定时间的字段
// 不会被覆盖，对齐 GORM 原生 autoUpdateTime 语义。被 Omit 的字段同样跳过。
//
// 参数：
//   - tx *gorm.DB: 当前 GORM 事务上下文，包含 Statement.Dest 与 Statement.Selects
//
// 返回值：
//   - err error: 错误信息，当前实现始终返回 nil
//
// 异常：
//   - 无显式异常，内部通过反射安全设置字段值
//
// 使用示例：
//
//	type User struct {
//	    base.CommonBaseModel
//	    Username string `gorm:"column:username" json:"username"`
//	}
//	// UpdatedAt 为空时自动填充；map 中已带 updated_at 则保留
//	db.Model(&user).Updates(map[string]interface{}{"username": "new"})
func (b *CommonBaseModel) BeforeUpdate(tx *gorm.DB) (err error) {
	if tx == nil || tx.Statement == nil || tx.Statement.Dest == nil {
		return
	}

	strTime := time.Now().Format("2006-01-02 15:04:05")
	SetTimeFieldDefaults(tx, strTime, "UpdatedAt", "Updated")
	EnsureSelects(tx, "UpdatedAt", "Updated")
	return
}

// SetTimeFieldDefaults 通过 GORM 事务上下文为未设置的时间字段批量填充默认值
// 供基础模型钩子内部使用，也可在业务模型自定义 BeforeCreate/BeforeUpdate
// 钩子中复用（如为业务自有时间字段兜底）。
// 内部包含三层保护：
//  1. Schema 校验：字段不存在于模型中时直接跳过，避免向 map Dest 添加
//     GORM 无法解析的 key（未识别的 key 会被当作裸列名生成 SQL 而报错）；
//  2. Omit 检查：被 Omit 的字段（按字段名或列名匹配）不设置；
//  3. 列名解析：将字段名解析为数据库列名，作为 map 场景的备选 key
//     （map 更新时调用方通常使用列名）。
//
// 函数名：SetTimeFieldDefaults
// 参数：
//   - tx *gorm.DB: 当前 GORM 事务上下文（钩子回调参数）
//   - strTime: 默认时间字符串（格式 2006-01-02 15:04:05）
//   - fieldNames: 结构体字段名列表（如 "UpdatedAt"）
//
// 返回值：无
// 异常：无
// 使用示例：
//
//	// 业务模型自定义钩子中为自有时间字段兜底
//	func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
//	    if err = u.MysqlBaseModel.BeforeCreate(tx); err != nil {
//	        return err
//	    }
//	    base.SetTimeFieldDefaults(tx, time.Now().Format("2006-01-02 15:04:05"), "LoginAt")
//	    return nil
//	}
func SetTimeFieldDefaults(tx *gorm.DB, strTime string, fieldNames ...string) {
	for _, name := range fieldNames {
		dbName := lookupDBName(tx, name)
		if dbName == "" {
			continue
		}
		if isFieldOmitted(tx, name) {
			continue
		}
		SetFieldDefault(tx.Statement.Dest, name, strTime, dbName)
	}
}

// lookupDBName 从 GORM Schema 中查找字段名对应的数据库列名
// 查不到时返回空字符串，SetFieldDefault 会自动忽略空备选 key。
//
// 参数：
//   - tx *gorm.DB: 当前 GORM 事务上下文
//   - fieldName: 结构体字段名
//
// 返回值：
//   - string: 数据库列名，Schema 不可用或字段不存在时返回空字符串
func lookupDBName(tx *gorm.DB, fieldName string) string {
	if tx == nil || tx.Statement == nil || tx.Statement.Schema == nil {
		return ""
	}
	if f := tx.Statement.Schema.LookUpField(fieldName); f != nil {
		return f.DBName
	}
	return ""
}

// isFieldOmitted 判断字段是否被 Omit 排除（按字段名或列名匹配）
//
// 参数：
//   - tx *gorm.DB: 当前 GORM 事务上下文
//   - fieldName: 结构体字段名
//
// 返回值：
//   - bool: 字段名或其列名出现在 Statement.Omits 中时返回 true
func isFieldOmitted(tx *gorm.DB, fieldName string) bool {
	if tx == nil || tx.Statement == nil || len(tx.Statement.Omits) == 0 {
		return false
	}
	if helper.InArray(fieldName, tx.Statement.Omits) {
		return true
	}
	return helper.InArray(lookupDBName(tx, fieldName), tx.Statement.Omits)
}
