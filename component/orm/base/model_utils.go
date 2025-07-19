package base

import (
	"reflect"
	"strings"
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
	}
	return ""
}

// setFieldIfExist 设置字段值（如果字段存在且可设置）
func setFieldIfExist(model interface{}, fieldName string, value string) {
	modelValue := reflect.ValueOf(model)

	if modelValue.Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
	}

	switch modelValue.Kind() {
	case reflect.Struct:
		field := modelValue.FieldByName(fieldName)
		if field.IsValid() && field.CanSet() && field.Kind() == reflect.String {
			field.SetString(value)
		}
	case reflect.Map:
		key := reflect.ValueOf(fieldName)
		if modelValue.MapIndex(key).IsValid() {
			modelValue.SetMapIndex(key, reflect.ValueOf(value))
		}
	default:
	}
}
