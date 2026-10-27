package base

import (
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/jcbowen/jcbaseGo/component/helper"
)

// getColumnFromTag 从gorm标签中获取列名
func getColumnFromTag(tag string) string {
	tags := strings.Split(tag, ";")
	for _, t := range tags {
		if strings.HasPrefix(t, "column:") {
			return strings.TrimPrefix(t, "column:")
		}
	}
	return ""
}

// getDefaultFromTag 从gorm标签中获取默认值
func getDefaultFromTag(tag string) string {
	tags := strings.Split(tag, ";")
	for _, t := range tags {
		if strings.HasPrefix(t, "default:") {
			return strings.TrimPrefix(t, "default:")
		}
	}
	return ""
}

// getSoftDeleteFromTag 从gorm标签中获取软删除条件
// 参数说明：
//   - tag string: gorm标签字符串
//
// 返回值：
//   - string: 软删除条件，如果没有找到soft_delete标签则返回空字符串
func getSoftDeleteFromTag(tag string) string {
	tags := strings.Split(tag, ";")
	for _, t := range tags {
		if strings.HasPrefix(t, "soft_delete:") {
			return strings.TrimPrefix(t, "soft_delete:")
		}
		// 支持无值标记 soft_delete
		if t == "soft_delete" {
			return ""
		}
	}
	return ""
}

// hasSoftDeleteTag 检查是否存在 soft_delete 标签（支持无值）
// 函数名：hasSoftDeleteTag
// 参数：
//   - tag string：gorm 标签字符串
//
// 返回值：
//   - bool：是否包含 soft_delete 标签
//
// 异常：无
// 使用示例：
//
//	has := hasSoftDeleteTag("column:deleted_at;soft_delete")
func hasSoftDeleteTag(tag string) bool {
	tags := strings.Split(tag, ";")
	for _, t := range tags {
		if t == "soft_delete" || strings.HasPrefix(t, "soft_delete:") {
			return true
		}
	}
	return false
}

// SetFieldIfExist 设置字段值（如果字段存在且可设置）
// 支持字符串类型和 time.Time 类型字段的自动设置
// 函数名：SetFieldIfExist
// 参数：
//   - model interface{}：模型对象或其指针
//   - fieldName string：结构体字段名（非列名）
//   - value string：用于设置的字符串值（时间格式：2006-01-02 15:04:05）
//
// 返回值：无
// 异常：无（内部忽略不可设或类型不匹配的字段）
// 使用示例：
//
//	SetFieldIfExist(&user, "UpdatedAt", "2025-01-01 00:00:00")
func SetFieldIfExist(model interface{}, fieldName string, value string) {
	modelValue := reflect.ValueOf(model)

	if modelValue.Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
	}

	switch modelValue.Kind() {
	case reflect.Struct:
		field := modelValue.FieldByName(fieldName)
		if field.IsValid() && field.CanSet() {
			// 根据字段类型设置相应的值
			switch field.Kind() {
			case reflect.String:
				// 字符串类型字段，直接设置
				field.SetString(value)
			case reflect.Struct:
				// 检查是否为 time.Time 类型
				if field.Type().String() == "time.Time" {
					// 解析时间字符串并设置
					if timeValue, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
						field.Set(reflect.ValueOf(timeValue))
					}
				}
			case reflect.Ptr:
				// 处理 *time.Time 等指针类型
				if field.Type().Elem().String() == "time.Time" {
					if timeValue, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
						// 如果当前为nil，分配新对象
						if field.IsNil() {
							newElem := reflect.New(field.Type().Elem())
							newElem.Elem().Set(reflect.ValueOf(timeValue))
							field.Set(newElem)
						} else {
							field.Elem().Set(reflect.ValueOf(timeValue))
						}
					}
				}
			}
		}
	case reflect.Map:
		key := reflect.ValueOf(fieldName)
		if modelValue.MapIndex(key).IsValid() {
			modelValue.SetMapIndex(key, reflect.ValueOf(value))
		}
	}
}

// SetFieldDefault 设置字段默认值（仅当字段未设置时写入）
// 与 SetFieldIfExist 的区别：字段已有值时跳过设置，保留调用方显式指定的时间，
// 对齐 GORM 原生 autoCreateTime/autoUpdateTime 与 default 标签的"零值才填充"语义。
// 两种目标类型的"未设置"判定与存在性语义：
//   - struct：字段必须存在（FieldByName 查不到则静默跳过），且值为零值
//     （string 为空 / time.Time 为零值 / 指针为 nil）时才写入；
//   - map：字段名与 aliasKeys 中任意 key 已存在即视为已设置（哪怕值为空字符串），
//     全部不存在时才以 fieldName 为 key 添加默认值。
//
// 函数名：SetFieldDefault
// 参数：
//   - model interface{}：模型对象指针或 map
//   - fieldName string：结构体字段名（非列名），同时作为 map 添加时的 key
//   - value string：默认值字符串（时间格式：2006-01-02 15:04:05）
//   - aliasKeys ...string：map 目标的备选 key（如数据库列名 updated_at）
//
// 返回值：无
// 异常：无（内部忽略不可设或类型不匹配的字段）
// 使用示例：
//
//	// 结构体：仅当 CreatedAt 为零值时才填充默认时间
//	SetFieldDefault(&user, "CreatedAt", "2026-01-01 00:00:00", "created_at")
//	// map：UpdatedAt 与 updated_at 均不存在时才添加默认时间
//	SetFieldDefault(dest, "UpdatedAt", "2026-01-01 00:00:00", "updated_at")
func SetFieldDefault(model interface{}, fieldName string, value string, aliasKeys ...string) {
	modelValue := reflect.ValueOf(model)

	if modelValue.Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
	}

	switch modelValue.Kind() {
	case reflect.Struct:
		field := modelValue.FieldByName(fieldName)
		if !field.IsValid() || !field.CanSet() {
			return
		}
		// 字段已有值时跳过，保留调用方显式指定的时间
		if !isZeroValue(field) {
			return
		}
		// 根据字段类型设置相应的值
		switch field.Kind() {
		case reflect.String:
			// 字符串类型字段，直接设置
			field.SetString(value)
		case reflect.Struct:
			// 检查是否为 time.Time 类型
			if field.Type().String() == "time.Time" {
				// 解析时间字符串并设置
				if timeValue, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
					field.Set(reflect.ValueOf(timeValue))
				}
			}
		case reflect.Ptr:
			// 处理 *time.Time 等指针类型，仅在 nil 时分配新对象
			if field.Type().Elem().String() == "time.Time" && field.IsNil() {
				if timeValue, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
					newElem := reflect.New(field.Type().Elem())
					newElem.Elem().Set(reflect.ValueOf(timeValue))
					field.Set(newElem)
				}
			}
		}
	case reflect.Map:
		// 仅支持 string key 的 map（与 GORM Updates 的 map 形态一致）
		if modelValue.Type().Key().Kind() != reflect.String {
			return
		}
		// key 已存在（字段名或列名形式）时跳过，尊重调用方显式指定的时间
		for _, key := range append([]string{fieldName}, aliasKeys...) {
			if key == "" {
				continue
			}
			if modelValue.MapIndex(reflect.ValueOf(key)).IsValid() {
				return
			}
		}
		modelValue.SetMapIndex(reflect.ValueOf(fieldName), reflect.ValueOf(value))
	}
}

// isZeroValue 判断反射值是否为空值
// 空值判定规则：字符串为空、time.Time 为零值、指针为 nil。
//
// 参数：
//   - v reflect.Value：待判断的反射值
//
// 返回值：
//   - bool：是否为空值
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Struct:
		return v.Type().String() == "time.Time" && v.Interface().(time.Time).IsZero()
	case reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// EnsureSelects 在使用 Select 限定插入/更新字段时，确保必要字段被包含
// 典型场景：在钩子中设置 CreatedAt/UpdatedAt 等字段，但外部使用了 Select 导致未持久化
// 函数名：EnsureSelects
// 参数：
//   - tx *gorm.DB：当前事务上下文
//   - fieldNames ...string：结构体字段名列表（如 "UpdatedAt"）
//
// 返回值：无（直接修改 tx.Statement.Selects）
// 异常：无（只在 Schema 可用时生效）
// 使用示例：
//
//	EnsureSelects(tx, "CreatedAt", "UpdatedAt")
func EnsureSelects(tx *gorm.DB, fieldNames ...string) {
	if tx == nil || tx.Statement == nil || tx.Statement.Schema == nil {
		return
	}
	if len(tx.Statement.Selects) == 0 {
		return
	}
	for _, sel := range tx.Statement.Selects {
		if sel == "*" {
			return
		}
	}
	for _, name := range fieldNames {
		f := tx.Statement.Schema.LookUpField(name)
		if f == nil {
			continue
		}
		dbn := f.DBName
		if dbn == "" {
			continue
		}
		if len(tx.Statement.Omits) > 0 {
			if helper.InArray(dbn, tx.Statement.Omits) || helper.InArray(name, tx.Statement.Omits) {
				continue
			}
		}
		if helper.InArray(dbn, tx.Statement.Selects) || helper.InArray(name, tx.Statement.Selects) {
			continue
		}
		if tx.Statement.Table != "" {
			tableCol := tx.Statement.Table + "." + dbn
			if helper.InArray(tableCol, tx.Statement.Selects) {
				continue
			}
		}
		tx.Statement.Selects = append(tx.Statement.Selects, dbn)
	}
}
