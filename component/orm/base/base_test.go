package base

import (
	"reflect"
	"testing"

	"gorm.io/gorm"
)

// 用于验证递归解析的内嵌结构体
type AuditInfo struct {
	CreatedBy uint `gorm:"column:created_by" json:"created_by"`
	UpdatedBy uint `gorm:"column:updated_by" json:"updated_by"`
}

// CommonModelUser 直接嵌入 CommonBaseModel 并嵌套 AuditInfo
type CommonModelUser struct {
	CommonBaseModel
	AuditInfo
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	Username  string `gorm:"column:username" json:"username"`
	DeletedAt string `gorm:"column:deleted_at;default:NULL" json:"deleted_at"`
}

// MysqlModelUser 嵌入 MysqlBaseModel 并嵌套 AuditInfo
type MysqlModelUser struct {
	MysqlBaseModel
	AuditInfo
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	Username  string `gorm:"column:username" json:"username"`
	DeletedAt string `gorm:"column:deleted_at;default:NULL" json:"deleted_at"`
}

// SqliteModelUser 嵌入 SqliteBaseModel
type SqliteModelUser struct {
	SqliteBaseModel
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	Username  string `gorm:"column:username" json:"username"`
	DeletedAt string `gorm:"column:deleted_at;default:NULL" json:"deleted_at"`
}

// SoftDeleteEmptyUser 用于验证 soft_delete: 空值
type SoftDeleteEmptyUser struct {
	CommonBaseModel
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	IsDeleted string `gorm:"column:is_deleted;soft_delete:" json:"is_deleted"`
}

// SoftDeleteDefaultEmptyUser 用于验证 default:” 时软删除条件回退到 IS NULL
type SoftDeleteDefaultEmptyUser struct {
	CommonBaseModel
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	DeletedAt string `gorm:"column:deleted_at;default:''" json:"deleted_at"`
}

// SoftDeleteQuoteDefaultUser 用于验证 default 值含单引号时被正确转义
type SoftDeleteQuoteDefaultUser struct {
	CommonBaseModel
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	IsDeleted string `gorm:"column:is_deleted;default:O'Brien;soft_delete" json:"is_deleted"`
}

// DeletedAtCustomDefaultUser 用于验证 deleted_at 字段非标准 default 时保持 IS NULL
type DeletedAtCustomDefaultUser struct {
	CommonBaseModel
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	DeletedAt string `gorm:"column:deleted_at;default:0" json:"deleted_at"`
}

// DeletedAtZeroDefaultUser 用于验证 deleted_at 字段 default 为 0000-00-00 00:00:00 时的特殊条件
type DeletedAtZeroDefaultUser struct {
	CommonBaseModel
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	DeletedAt string `gorm:"column:deleted_at;default:0000-00-00 00:00:00" json:"deleted_at"`
}

type SoftDeleteEmptyUserMySQL struct {
	MysqlBaseModel
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	IsDeleted string `gorm:"column:is_deleted;soft_delete:" json:"is_deleted"`
}

// GormIgnoreUser 用于验证 gorm:"-" 字段被正确忽略
type GormIgnoreUser struct {
	CommonBaseModel
	ID       uint   `gorm:"column:id;primaryKey" json:"id"`
	Username string `gorm:"column:username" json:"username"`
	Password string `gorm:"-" json:"-"`
}

// GormIgnoreAllUser 用于验证 gorm:"-:all" 字段被正确忽略
type GormIgnoreAllUser struct {
	CommonBaseModel
	ID       uint   `gorm:"column:id;primaryKey" json:"id"`
	Username string `gorm:"column:username" json:"username"`
	Password string `gorm:"-:all" json:"-"`
}

// GormIgnoreCreateUser 用于验证 gorm:"-:create" 字段不应被完全忽略
type GormIgnoreCreateUser struct {
	CommonBaseModel
	ID       uint   `gorm:"column:id;primaryKey" json:"id"`
	Username string `gorm:"column:username" json:"username"`
	Temp     string `gorm:"column:temp;-:create" json:"temp"`
}

// GormIgnoreUpdateUser 用于验证 gorm:"-:update" 字段不应被完全忽略
type GormIgnoreUpdateUser struct {
	CommonBaseModel
	ID       uint   `gorm:"column:id;primaryKey" json:"id"`
	Username string `gorm:"column:username" json:"username"`
	Temp     string `gorm:"column:temp;-:update" json:"temp"`
}

// DuplicateFieldUser 用于验证多层嵌入出现同名字段时不会重复解析
type DuplicateFieldBase struct {
	ID uint `gorm:"column:id" json:"id"`
}

type DuplicateFieldUser struct {
	CommonBaseModel
	DuplicateFieldBase
	ID       uint   `gorm:"column:id;primaryKey" json:"id"`
	Username string `gorm:"column:username" json:"username"`
}

// containsField 检查字段列表是否包含指定字段
func containsField(fields []string, field string) bool {
	for _, f := range fields {
		if f == field {
			return true
		}
	}
	return false
}

// TestCommonBaseModelRecursiveParse 验证 CommonBaseModel 能递归解析内嵌字段
func TestCommonBaseModelRecursiveParse(t *testing.T) {
	t.Setenv("jc_mysql_db", `{"TablePrefix":"t_","SingularTable":false}`)

	b := &CommonBaseModel{}
	_, fields, _, _ := b.ModelParse(reflect.TypeOf(CommonModelUser{}))

	expected := []string{"created_by", "updated_by", "id", "username", "deleted_at"}
	for _, f := range expected {
		if !containsField(fields, f) {
			t.Errorf("CommonBaseModel 应解析出字段 %s，实际得到 %v", f, fields)
		}
	}
}

// TestMysqlBaseModelRecursiveParse 验证 MysqlBaseModel 是否遗漏递归解析
func TestMysqlBaseModelRecursiveParse(t *testing.T) {
	t.Setenv("jc_mysql_db", `{"TablePrefix":"t_","SingularTable":false}`)

	b := &MysqlBaseModel{}
	_, fields, _, _ := b.ModelParse(reflect.TypeOf(MysqlModelUser{}))

	// 期望包含 AuditInfo 中的 created_by / updated_by
	for _, f := range []string{"created_by", "updated_by"} {
		if !containsField(fields, f) {
			t.Errorf("MysqlBaseModel 应递归解析出内嵌字段 %s，实际得到 %v", f, fields)
		}
	}
}

// TestGormIgnoreTag 验证 gorm:"-" 字段不会被解析为模型字段
func TestGormIgnoreTag(t *testing.T) {
	t.Setenv("jc_mysql_db", `{"TablePrefix":"t_","SingularTable":false}`)

	b := &CommonBaseModel{}
	_, fields, _, _ := b.ModelParse(reflect.TypeOf(GormIgnoreUser{}))

	if containsField(fields, "password") {
		t.Errorf("gorm:\"-\" 字段不应被解析，实际得到 %v", fields)
	}

	expected := []string{"id", "username"}
	for _, f := range expected {
		if !containsField(fields, f) {
			t.Errorf("应解析出字段 %s，实际得到 %v", f, fields)
		}
	}
}

// TestGormIgnoreAllTag 验证 gorm:"-:all" 字段不会被解析为模型字段
func TestGormIgnoreAllTag(t *testing.T) {
	t.Setenv("jc_mysql_db", `{"TablePrefix":"t_","SingularTable":false}`)

	b := &CommonBaseModel{}
	_, fields, _, _ := b.ModelParse(reflect.TypeOf(GormIgnoreAllUser{}))

	if containsField(fields, "password") {
		t.Errorf("gorm:\"-:all\" 字段不应被解析，实际得到 %v", fields)
	}

	expected := []string{"id", "username"}
	for _, f := range expected {
		if !containsField(fields, f) {
			t.Errorf("应解析出字段 %s，实际得到 %v", f, fields)
		}
	}
}

// TestDuplicateFieldDedup 验证多层嵌入出现同名字段时去重
func TestDuplicateFieldDedup(t *testing.T) {
	t.Setenv("jc_mysql_db", `{"TablePrefix":"t_","SingularTable":false}`)

	b := &CommonBaseModel{}
	_, fields, _, _ := b.ModelParse(reflect.TypeOf(DuplicateFieldUser{}))

	idCount := 0
	for _, f := range fields {
		if f == "id" {
			idCount++
		}
	}
	if idCount != 1 {
		t.Errorf("同名字段 id 应只出现一次，实际出现 %d 次，字段列表 %v", idCount, fields)
	}
}

// TestSoftDeleteEmptyValue 验证 soft_delete: 空值在 CommonBaseModel 中是否返回空条件
func TestSoftDeleteEmptyValue(t *testing.T) {
	t.Setenv("jc_mysql_db", `{"TablePrefix":"t_","SingularTable":false}`)

	b := &CommonBaseModel{}
	_, _, _, condition := b.ModelParse(reflect.TypeOf(SoftDeleteEmptyUser{}))
	if condition == "" {
		t.Errorf("CommonBaseModel 对 soft_delete: 空值不应返回空条件")
	}

	// 与 MysqlBaseModel 行为保持一致：应回退到 IS NULL
	bm := &MysqlBaseModel{}
	_, _, _, conditionMySQL := bm.ModelParse(reflect.TypeOf(SoftDeleteEmptyUserMySQL{}))
	if condition != conditionMySQL {
		t.Errorf("CommonBaseModel(%s) 与 MysqlBaseModel(%s) 对 soft_delete: 空值解析不一致", condition, conditionMySQL)
	}
}

// TestSoftDeleteDefaultEmptyValue 验证 default:” 时软删除条件回退到 IS NULL
func TestSoftDeleteDefaultEmptyValue(t *testing.T) {
	t.Setenv("jc_mysql_db", `{"TablePrefix":"t_","SingularTable":false}`)

	b := &CommonBaseModel{}
	_, _, field, condition := b.ModelParse(reflect.TypeOf(SoftDeleteDefaultEmptyUser{}))
	if field != "deleted_at" {
		t.Errorf("应识别 deleted_at 为软删除字段，实际得到 %s", field)
	}
	if condition != "IS NULL" {
		t.Errorf("default:'' 时应回退到 IS NULL，实际得到 %s", condition)
	}
}

// TestSqliteConfigSourcePriority 验证 SqliteBaseModel 优先使用 SQLite 配置
func TestSqliteConfigSourcePriority(t *testing.T) {
	t.Setenv("jc_sql_lite_main", `{"TablePrefix":"sq_","SingularTable":false}`)

	b := &SqliteBaseModel{}
	tableName, _, _, _ := b.ModelParse(reflect.TypeOf(SqliteModelUser{}))
	if tableName != "sq_sqlite_model_users" {
		t.Errorf("SqliteBaseModel 应使用 SQLite 配置生成表名 sq_sqlite_model_users，实际得到 %s", tableName)
	}
}

// TestCommonBaseModelFallbackToSqlite 验证 CommonBaseModel 在 MySQL 配置缺失时回退到 SQLite 配置
func TestCommonBaseModelFallbackToSqlite(t *testing.T) {
	t.Setenv("jc_sql_lite_db", `{"TablePrefix":"sp_","SingularTable":false}`)

	b := &CommonBaseModel{}
	tableName, _, _, _ := b.ModelParse(reflect.TypeOf(CommonModelUser{}))
	if tableName != "sp_common_model_users" {
		t.Errorf("CommonBaseModel 在 MySQL 配置缺失时应回退到 SQLite 配置生成表名 sp_common_model_users，实际得到 %s", tableName)
	}
}

// TestMysqlModelNoFallbackToSqlite 验证 MysqlBaseModel 不回退到 SQLite 配置，保持向后兼容
func TestMysqlModelNoFallbackToSqlite(t *testing.T) {
	t.Setenv("jc_sql_lite_db", `{"TablePrefix":"sp_","SingularTable":false}`)

	b := &MysqlBaseModel{}
	tableName, _, _, _ := b.ModelParse(reflect.TypeOf(MysqlModelUser{}))
	if tableName != "mysql_model_users" {
		t.Errorf("MysqlBaseModel 不应回退到 SQLite 配置，实际得到 %s", tableName)
	}
}

// TestSqliteModelNoFallbackToMysql 验证 SqliteBaseModel 不回退到 MySQL 配置，保持向后兼容
func TestSqliteModelNoFallbackToMysql(t *testing.T) {
	t.Setenv("jc_mysql_main", `{"TablePrefix":"mp_","SingularTable":false}`)

	b := &SqliteBaseModel{}
	tableName, _, _, _ := b.ModelParse(reflect.TypeOf(SqliteModelUser{}))
	if tableName != "sqlite_model_users" {
		t.Errorf("SqliteBaseModel 不应回退到 MySQL 配置，实际得到 %s", tableName)
	}
}

// TestCommonBaseModelFieldFilter 验证 CommonBaseModel 不会把自身当作字段
func TestCommonBaseModelFieldFilter(t *testing.T) {
	t.Setenv("jc_mysql_db", `{"TablePrefix":"t_","SingularTable":false}`)

	b := &CommonBaseModel{}
	_, fields, _, _ := b.ModelParse(reflect.TypeOf(CommonModelUser{}))
	for _, f := range fields {
		if f == "common_base_model" {
			t.Errorf("CommonBaseModel 不应把自身解析为字段")
		}
	}
}

// TestSoftDeleteQuoteDefaultValue 验证 default 值含单引号时被正确转义
func TestSoftDeleteQuoteDefaultValue(t *testing.T) {
	t.Setenv("jc_mysql_db", `{"TablePrefix":"t_","SingularTable":false}`)

	b := &CommonBaseModel{}
	_, _, field, condition := b.ModelParse(reflect.TypeOf(SoftDeleteQuoteDefaultUser{}))
	if field != "is_deleted" {
		t.Errorf("应识别 is_deleted 为软删除字段，实际得到 %s", field)
	}
	expected := "= 'O''Brien'"
	if condition != expected {
		t.Errorf("default 含单引号时应转义为 %s，实际得到 %s", expected, condition)
	}
}

// TestDeletedAtConservativeCondition 验证 deleted_at 字段非标准 default 时保持 IS NULL
func TestDeletedAtConservativeCondition(t *testing.T) {
	t.Setenv("jc_mysql_db", `{"TablePrefix":"t_","SingularTable":false}`)

	b := &CommonBaseModel{}
	_, _, field, condition := b.ModelParse(reflect.TypeOf(DeletedAtCustomDefaultUser{}))
	if field != "deleted_at" {
		t.Errorf("应识别 deleted_at 为软删除字段，实际得到 %s", field)
	}
	if condition != "IS NULL" {
		t.Errorf("deleted_at 非标准 default 时应回退到 IS NULL，实际得到 %s", condition)
	}
}

// TestDeletedAtZeroDefaultCondition 验证 deleted_at 字段 default 为 0000-00-00 00:00:00 时的特殊条件
func TestDeletedAtZeroDefaultCondition(t *testing.T) {
	t.Setenv("jc_mysql_db", `{"TablePrefix":"t_","SingularTable":false}`)

	b := &CommonBaseModel{}
	_, _, field, condition := b.ModelParse(reflect.TypeOf(DeletedAtZeroDefaultUser{}))
	if field != "deleted_at" {
		t.Errorf("应识别 deleted_at 为软删除字段，实际得到 %s", field)
	}
	expected := "= '0000-00-00 00:00:00'"
	if condition != expected {
		t.Errorf("deleted_at default 为 0000-00-00 00:00:00 时应生成 %s，实际得到 %s", expected, condition)
	}
}

// TestGormIgnoreCreateTag 验证 gorm:"-:create" 字段不应被完全忽略
func TestGormIgnoreCreateTag(t *testing.T) {
	t.Setenv("jc_mysql_db", `{"TablePrefix":"t_","SingularTable":false}`)

	b := &CommonBaseModel{}
	_, fields, _, _ := b.ModelParse(reflect.TypeOf(GormIgnoreCreateUser{}))

	if !containsField(fields, "temp") {
		t.Errorf("gorm:\"-:create\" 字段不应被完全忽略，实际字段列表 %v", fields)
	}

	expected := []string{"id", "username", "temp"}
	for _, f := range expected {
		if !containsField(fields, f) {
			t.Errorf("应解析出字段 %s，实际得到 %v", f, fields)
		}
	}
}

// TestGormIgnoreUpdateTag 验证 gorm:"-:update" 字段不应被完全忽略
func TestGormIgnoreUpdateTag(t *testing.T) {
	t.Setenv("jc_mysql_db", `{"TablePrefix":"t_","SingularTable":false}`)

	b := &CommonBaseModel{}
	_, fields, _, _ := b.ModelParse(reflect.TypeOf(GormIgnoreUpdateUser{}))

	if !containsField(fields, "temp") {
		t.Errorf("gorm:\"-:update\" 字段不应被完全忽略，实际字段列表 %v", fields)
	}

	expected := []string{"id", "username", "temp"}
	for _, f := range expected {
		if !containsField(fields, f) {
			t.Errorf("应解析出字段 %s，实际得到 %v", f, fields)
		}
	}
}

// TestBeforeCreateWithNilDest 验证 Statement.Dest 为 nil 时不会 panic
func TestBeforeCreateWithNilDest(t *testing.T) {
	b := &CommonBaseModel{}

	// tx 为 nil
	if err := b.BeforeCreate(nil); err != nil {
		t.Errorf("tx 为 nil 时应返回 nil，实际 %v", err)
	}

	// Statement 为 nil
	if err := b.BeforeCreate(&gorm.DB{}); err != nil {
		t.Errorf("Statement 为 nil 时应返回 nil，实际 %v", err)
	}

	// Dest 为 nil
	if err := b.BeforeCreate(&gorm.DB{Statement: &gorm.Statement{}}); err != nil {
		t.Errorf("Dest 为 nil 时应返回 nil，实际 %v", err)
	}
}

// TestBeforeUpdateWithNilDest 验证 Statement.Dest 为 nil 时不会 panic
func TestBeforeUpdateWithNilDest(t *testing.T) {
	b := &CommonBaseModel{}

	// tx 为 nil
	if err := b.BeforeUpdate(nil); err != nil {
		t.Errorf("tx 为 nil 时应返回 nil，实际 %v", err)
	}

	// Statement 为 nil
	if err := b.BeforeUpdate(&gorm.DB{}); err != nil {
		t.Errorf("Statement 为 nil 时应返回 nil，实际 %v", err)
	}

	// Dest 为 nil
	if err := b.BeforeUpdate(&gorm.DB{Statement: &gorm.Statement{}}); err != nil {
		t.Errorf("Dest 为 nil 时应返回 nil，实际 %v", err)
	}
}
