package crud

import (
	"reflect"
	"strings"
	"sync"

	"github.com/jcbowen/jcbaseGo/component/helper"
)

// gormFieldTagMap 缓存模型字段到 gorm 标签的映射
// key 为数据库列名（与 ModelFields 中的字段名保持一致）
type gormFieldTagMap map[string]string

// gormTagMapCache 全局缓存模型类型到 gorm 标签映射
// 使用 sync.Map 保证并发安全，key 为 reflect.Type，value 为 gormFieldTagMap
var gormTagMapCache sync.Map

// buildGormTagMap 递归解析模型类型，建立列名 -> gorm 标签的映射
//
// 参数：
//   - modelType reflect.Type: 模型结构体类型
//
// 返回值：
//   - gormFieldTagMap: 列名到 gorm 标签的映射
func buildGormTagMap(modelType reflect.Type) gormFieldTagMap {
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		return gormFieldTagMap{}
	}

	m := make(gormFieldTagMap)
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// 递归解析匿名嵌入字段
		if field.Anonymous {
			for k, v := range buildGormTagMap(field.Type) {
				if _, exists := m[k]; !exists {
					m[k] = v
				}
			}
			continue
		}

		gormTag := field.Tag.Get("gorm")
		columnName := getColumnFromGormTag(gormTag)
		if columnName == "" {
			columnName = helper.NewStr(field.Name).ConvertCamelToSnake()
		}

		if _, exists := m[columnName]; !exists {
			m[columnName] = gormTag
		}
	}

	return m
}

// getGormTagByColumn 根据列名获取对应的 gorm 标签
// 优先从全局缓存读取，未命中时构建并缓存
//
// 参数：
//   - column string: 数据库列名
//
// 返回值：
//   - string: gorm 标签字符串，未找到返回空字符串
func (t *Trait) getGormTagByColumn(column string) string {
	if t.Model == nil {
		return ""
	}

	modelType := reflect.TypeOf(t.Model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 从缓存读取
	if cached, ok := gormTagMapCache.Load(modelType); ok {
		if m, ok := cached.(gormFieldTagMap); ok {
			return m[column]
		}
	}

	// 未命中则构建并写入缓存
	m := buildGormTagMap(modelType)
	gormTagMapCache.Store(modelType, m)
	return m[column]
}

// isGormUpdateIgnored 判断字段是否在更新操作中被 gorm 标签忽略
// 被忽略的标记包括：gorm:"-"、gorm:"-:all"、gorm:"-:update"
//
// 参数：
//   - column string: 数据库列名
//
// 返回值：
//   - bool: 是否在更新时被忽略
func (t *Trait) isGormUpdateIgnored(column string) bool {
	return isGormUpdateIgnoredTag(t.getGormTagByColumn(column))
}

// isGormUpdateIgnoredTag 判断 gorm 标签是否在更新操作中被忽略
//
// 参数：
//   - tag string: gorm 标签字符串
//
// 返回值：
//   - bool: 是否在更新时被忽略
func isGormUpdateIgnoredTag(tag string) bool {
	if tag == "-" {
		return true
	}
	for _, t := range strings.Split(tag, ";") {
		if t == "-" || t == "-:all" || t == "-:update" {
			return true
		}
	}
	return false
}

// getColumnFromGormTag 从 gorm 标签中提取列名
//
// 参数：
//   - tag string: gorm 标签字符串
//
// 返回值：
//   - string: 列名，未找到返回空字符串
func getColumnFromGormTag(tag string) string {
	for _, t := range strings.Split(tag, ";") {
		if strings.HasPrefix(t, "column:") {
			return strings.TrimPrefix(t, "column:")
		}
	}
	return ""
}

