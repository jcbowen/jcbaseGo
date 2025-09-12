package base

import (
	"reflect"
	"strings"
	"time"
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

// SetFieldIfExist 设置字段值（如果字段存在且可设置）
// 支持字符串类型和时间类型字段的自动设置
// 参数说明：
//   - model interface{}: 要设置字段值的模型对象
//   - fieldName string: 字段名称
//   - value string: 字段值（字符串格式的时间）
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
			}
		}
	case reflect.Map:
		key := reflect.ValueOf(fieldName)
		if modelValue.MapIndex(key).IsValid() {
			modelValue.SetMapIndex(key, reflect.ValueOf(value))
		}
	default:
	}
}
