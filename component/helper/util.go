package helper

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/jcbowen/jcbaseGo/component/validator"
)

// ----- map[string]interface{} 类型相关操作 -----/

type MapHelper struct {
	Data map[string]interface{}
	Keys []string
	Sort bool
}

func NewMap(mapData map[string]interface{}) *MapHelper {
	return &MapHelper{Data: mapData}
}

func (d *MapHelper) DoSort() *MapHelper {
	d.Sort = true
	return d
}

func (d *MapHelper) ArrayKeys() []string {
	if len(d.Data) == 0 {
		return d.Keys
	}

	for k := range d.Data {
		d.Keys = append(d.Keys, k)
	}

	if d.Sort {
		sort.Strings(d.Keys)
	}

	return d.Keys
}

func (d *MapHelper) ArrayValues() []interface{} {
	var values []interface{}

	if len(d.Data) == 0 {
		return values
	}

	if d.Sort {
		for _, k := range d.ArrayKeys() {
			values = append(values, d.Data[k])
		}
	} else {
		for _, v := range d.Data {
			values = append(values, v)
		}
	}

	return values
}

func (d *MapHelper) GetData() map[string]interface{} {
	if d.Sort {
		data := make(map[string]interface{})
		for _, k := range d.ArrayKeys() {
			data[k] = d.Data[k]
		}
		return data
	}

	return d.Data
}

// ExtractString 从多级map中提取字符串字段值
// - path: 字段路径，使用点号分隔，如 "data.member_data.gender_text"
// - defaultVal: 默认值，当路径不存在或类型不匹配时返回该默认值
// 返回：提取到的字符串值，如果路径不存在或类型不匹配返回空字符串
func (d *MapHelper) ExtractString(path string, defaultVal ...string) string {
	value := d.Extract(path)
	if value != nil {
		return fmt.Sprintf("%v", value)
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return ""
}

// ExtractInt 从多级map中提取整数字段值
// - path: 字段路径，使用点号分隔，如 "data.member_data.age"
// - defaultVal: 默认值，当路径不存在或类型不匹配时返回该默认值
// 返回：提取到的整数值，如果路径不存在或类型不匹配返回0
func (d *MapHelper) ExtractInt(path string, defaultVal ...int) int {
	value := d.Extract(path)
	if value != nil {
		return Convert{Value: value}.ToInt()
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return 0
}

// ExtractInt64 从多级map中提取64位整数字段值
// - path: 字段路径，使用点号分隔，如 "data.member_data.big_id"
// - defaultVal: 默认值，当路径不存在或类型不匹配时返回该默认值
// 返回：提取到的64位整数值，如果路径不存在或类型不匹配返回0
func (d *MapHelper) ExtractInt64(path string, defaultVal ...int64) int64 {
	value := d.Extract(path)
	if value != nil {
		return Convert{Value: value}.ToInt64()
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return 0
}

// ExtractFloat64 从多级map中提取浮点数字段值
// - path: 字段路径，使用点号分隔，如 "data.member_data.salary"
// - defaultVal: 默认值，当路径不存在或类型不匹配时返回该默认值
// 返回：提取到的浮点数值，如果路径不存在或类型不匹配返回0.0
func (d *MapHelper) ExtractFloat64(path string, defaultVal ...float64) float64 {
	value := d.Extract(path)
	if value != nil {
		return Convert{Value: value}.ToFloat64()
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return 0.0
}

// ExtractBool 从多级map中提取布尔字段值
// - path: 字段路径，使用点号分隔，如 "data.member_data.is_active"
// - defaultVal: 默认值，当路径不存在或类型不匹配时返回该默认值
// 返回：提取到的布尔值，如果路径不存在或类型不匹配返回false
func (d *MapHelper) ExtractBool(path string, defaultVal ...bool) bool {
	value := d.Extract(path)
	if value != nil {
		return Convert{Value: value}.ToBool()
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return false
}

// ExtractTime 从多级map中提取时间字段值
// - path: 字段路径，使用点号分隔，如 "data.member_data.created_at"
// - defaultVal: 默认值，当路径不存在或类型不匹配时返回该默认值
// 返回：提取到的时间值，如果路径不存在或类型不匹配返回零值时间
func (d *MapHelper) ExtractTime(path string, defaultVal ...time.Time) time.Time {
	value := d.Extract(path)
	if value != nil {
		return Convert{Value: value}.ToTime()
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return time.Time{}
}

// ExtractStringSlice 从多级map中提取字符串切片字段值
// - path: 字段路径，使用点号分隔，如 "data.member_data.tags"
// - defaultVal: 默认值，当路径不存在或类型不匹配时返回该默认值
// 返回：提取到的字符串切片，如果路径不存在或类型不匹配返回空切片
func (d *MapHelper) ExtractStringSlice(path string, defaultVal ...[]string) []string {
	value := d.Extract(path)
	if value != nil {
		switch v := value.(type) {
		case []string:
			return v
		case []interface{}:
			var result []string
			for _, item := range v {
				if item != nil {
					result = append(result, fmt.Sprintf("%v", item))
				}
			}
			return result
		case string:
			// 尝试解析逗号分隔的字符串
			if strings.Contains(v, ",") {
				return strings.Split(v, ",")
			}
			// 单个字符串转为切片
			return []string{v}
		default:
			// 其他类型转为单个元素的切片
			return []string{fmt.Sprintf("%v", v)}
		}
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return []string{}
}

// Extract 从多级map中提取任意类型的字段值
// - path: 字段路径，使用点号分隔，如 "data.member_data.age"
// - defaultVal: 默认值，当路径不存在或类型不匹配时返回该默认值
// 返回：提取到的原始值，如果路径不存在返回nil
func (d *MapHelper) Extract(path string, defaultVal ...interface{}) interface{} {
	if d.Data == nil {
		return nil
	}

	keys := strings.Split(path, ".")
	current := d.Data

	// 遍历路径中的每个键
	for i, key := range keys {
		// 如果是最后一个键，返回原始值
		if i == len(keys)-1 {
			if val, ok := current[key]; ok && val != nil {
				return val
			}
			if len(defaultVal) > 0 {
				return defaultVal[0]
			}
			return nil
		}

		// 如果不是最后一个键，继续深入
		if next, ok := current[key]; ok && next != nil {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				// 如果中间路径不是map，返回nil或默认值
				if len(defaultVal) > 0 {
					return defaultVal[0]
				}
				return nil
			}
		} else {
			// 如果键不存在，返回nil或默认值
			if len(defaultVal) > 0 {
				return defaultVal[0]
			}
			return nil
		}
	}

	return nil
}

// ----- []string 类型相关操作 -----/

type ArrStr struct {
	Arr  []string // 数组
	Sort bool     // 执行ArrayValue方法时是否排序
}

func SetArrStr(str []string) *ArrStr {
	return &ArrStr{Arr: str, Sort: false}
}

// DoSort 设置ArrayValue方法¬是否排序
func (a *ArrStr) DoSort() *ArrStr {
	a.Sort = true
	return a
}

func (a *ArrStr) ArrayValue() (value []string) {
	if len(a.Arr) == 0 {
		return
	}
	value = append(value, a.Arr...)
	if a.Sort {
		sort.Strings(value)
	}
	return
}

func (a *ArrStr) ArrayDiff(oArr ...[]string) (diff []string) {
	if len(a.Arr) == 0 {
		return
	}
	if len(a.Arr) > 0 && len(oArr) == 0 {
		diff = a.Arr
		return
	}
	for _, o := range oArr {
		for _, item := range a.Arr {
			if !InArray(item, o) {
				diff = append(diff, item)
			}
		}
	}
	return
}

func (a *ArrStr) ArrayIntersect(oArr ...[]string) (intersects []string) {
	if len(a.Arr) == 0 {
		return
	}
	if len(a.Arr) > 0 && len(oArr) == 0 {
		intersects = a.Arr
		return
	}
	var tmp = make(map[string]int, len(a.Arr))
	for _, v := range a.Arr {
		tmp[v] = 1
	}
	for _, param := range oArr {
		for _, arg := range param {
			if tmp[arg] != 0 {
				tmp[arg]++
			}
		}
	}
	for k, v := range tmp {
		if v > 1 {
			intersects = append(intersects, k)
		}
	}
	return
}

// StringStartWith 判断字符串是否以某个字符串开头
func StringStartWith(str, prefix string) bool {
	return strings.HasPrefix(str, prefix)
}

// StringEndWith 判断字符串是否以某个字符串结尾
func StringEndWith(str, suffix string) bool {
	return strings.HasSuffix(str, suffix)
}

// StrReplace 类似于php中的str_replace
func StrReplace(search interface{}, replace interface{}, subject interface{}, count int) (interface{}, error) {
	switch search.(type) {
	case string:
		switch replace.(type) {
		case string:
			switch subject.(type) {
			case string:
				return strings.Replace(subject.(string), search.(string), replace.(string), count), nil
			case []string:
				var slice []string
				for _, v := range subject.([]string) {
					slice = append(slice, strings.Replace(v, search.(string), replace.(string), count))
				}
				return slice, nil
			default:
				return nil, errors.New("invalid parameters")
			}
		default:
			return nil, errors.New("invalid parameters")
		}
	case []string:
		switch replace.(type) {
		case string:
			switch subject.(type) {
			case string:
				sub := subject.(string)

				for _, v := range search.([]string) {
					sub = strings.Replace(sub, v, replace.(string), count)
				}
				return sub, nil

			case []string:
				var slice []string
				for _, v := range subject.([]string) {
					sli, err := StrReplace(search, replace, v, count)
					if err != nil {
						return nil, err
					}
					slice = append(slice, sli.(string))
				}
				return slice, nil
			default:
				return nil, errors.New("invalid parameters")
			}
		case []string:
			switch subject.(type) {
			case string:
				rep := replace.([]string)
				sub := subject.(string)
				for i, s := range search.([]string) {
					if i < len(rep) {
						sub = strings.Replace(sub, s, rep[i], count)
					} else {
						sub = strings.Replace(sub, s, "", count)
					}
				}
				return sub, nil
			case []string:
				var slice []string
				for _, v := range subject.([]string) {
					sli, err := StrReplace(search, replace, v, count)
					if err != nil {
						return nil, err
					}
					slice = append(slice, sli.(string))
				}
				return slice, nil
			default:
				return nil, errors.New("invalid parameters")
			}
		default:
			return nil, errors.New("invalid parameters")
		}
	default:
		return nil, errors.New("invalid parameters")
	}
}

// InArray 检查某个值是否存在于切片或数组中
// val 是要检查的值
// array 是要检查的切片或数组
// exists 是返回的布尔值，表示 val 是否存在于 array 中；如果 array 不是切片或数组，返回 false
func InArray(val interface{}, array interface{}) (exists bool) {
	arr := reflect.ValueOf(array)

	// 确保 array 是一个切片或数组
	if arr.Kind() != reflect.Slice && arr.Kind() != reflect.Array {
		return false
	}

	// 遍历切片，检查 val 是否存在
	for i := 0; i < arr.Len(); i++ {
		if reflect.DeepEqual(val, arr.Index(i).Interface()) {
			return true
		}
	}

	return false
}

// StructToMap 通过reflect将结构体转换为map
func StructToMap(obj interface{}, useJsonTag bool) map[string]interface{} {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}

	objType := objValue.Type()

	result := make(map[string]interface{})
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		fieldValue := objValue.Field(i).Interface()
		fieldName := field.Name
		if useJsonTag {
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" {
				fieldName = jsonTag
			} else {
				log.Println("StructToMap: json tag not found in struct field:", field.Name)
				fieldName = strings.ToLower(fieldName)
			}
		}
		result[fieldName] = fieldValue
	}

	return result
}

// MapToStruct 通过reflect将map转换为结构体
// mapData 是要转换的 map，必须是 map[string]interface{} 类型或其兼容类型
// obj 是目标结构体指针
// 返回转换过程中发生的错误，成功时返回 nil
func MapToStruct(mapData interface{}, obj interface{}) error {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Ptr || objValue.IsNil() {
		return errors.New("MapToStruct: obj 必须是有效的结构体指针")
	}
	objValue = objValue.Elem()
	if objValue.Kind() != reflect.Struct {
		return errors.New("MapToStruct: obj 指向的值必须是结构体")
	}

	data, ok := mapData.(map[string]interface{})
	if !ok {
		return errors.New("MapToStruct: mapData 必须是 map[string]interface{} 类型")
	}

	for key, value := range data {
		field := objValue.FieldByName(key)
		if !field.IsValid() {
			// 如果结构体中不存在这个字段，则尝试匹配 JSON 标记
			fieldName := GetFieldNameByJSONTag(objValue.Type(), key)
			if fieldName == "" {
				log.Println("未找到对应的字段：", key)
				// 如果结构体中仍不存在这个字段，跳过
				continue
			}
			field = objValue.FieldByName(fieldName)
		}

		// 将 map 中的值转换为对应的类型，并设置到结构体字段中
		if !setFieldValue(field, value) {
			log.Println("值类型无法转换为字段类型：", key)
		}
	}

	return nil
}

// setFieldValue 将 map 中的值转换为对应的类型，并设置到结构体字段中（属于MapToStruct的递归调用）
func setFieldValue(field reflect.Value, value interface{}) bool {
	fieldValue := reflect.ValueOf(value)
	if !fieldValue.IsValid() {
		return false
	}

	if fieldValue.Type().ConvertibleTo(field.Type()) {
		convertedValue := fieldValue.Convert(field.Type())
		field.Set(convertedValue)
		return true
	}

	if field.Kind() == reflect.Struct && fieldValue.Kind() == reflect.Map {
		// 如果字段是结构体，并且值是一个 map，则递归调用 MapToStruct 函数
		data, ok := value.(map[string]interface{})
		if !ok {
			log.Println("递归设置结构体字段失败：value 不是 map[string]interface{} 类型")
			return false
		}
		if err := MapToStruct(data, field.Addr().Interface()); err != nil {
			log.Println("递归设置结构体字段失败：", err)
			return false
		}
		return true
	}

	return false
}

// CopyStruct 复制源结构体（src）的字段到目标结构体（dst）。
// 该函数执行的是深层复制，基于字段名和字段类型进行匹配。
// src 和 dst 都必须是指向结构体的指针。
//
// 注意事项：
// CopyStruct 函数用于将源结构体 (src) 的字段值深度复制到目标结构体 (dst) 中。
// 它通过反射机制实现，支持嵌套结构体、切片和映射的深层复制，并尝试进行灵活的类型转换。
//
// 注意事项:
// 1. 错误处理：如果 src 或 dst 不是指向结构体的指针，函数会返回错误。
// 2. 性能开销：反射操作通常比直接字段访问慢，频繁调用可能影响性能。
// 3. 深层复制：对于包含指针、切片、映射等复杂类型的字段，现在会进行深层复制。
// 4. 类型严格匹配：源字段和目标字段的类型现在支持可转换类型之间的灵活转换。
// 5. 未导出字段：无法设置来自不同包的未导出（小写字母开头）字段。
// 6. 字段名匹配：依赖于字段名完全匹配。
//
// 参数:
//
//	src: 源结构体（必须是指向结构体的指针）。
//	dst: 目标结构体（必须是指向结构体的指针）。
//
// 返回值:
//
//	error: 如果 src 或 dst 不是指向结构体的指针，或在复制过程中发生类型不匹配等错误，则返回错误；否则返回 nil。
//
// 示例:
//
//	type Source struct {
//	  Name string
//	  Age  int
//	}
//	type Destination struct {
//	  Name string
//	  Age  int
//	  City string
//	}
//	s := &Source{Name: "Alice", Age: 30}
//	d := &Destination{}
//	if err := CopyStruct(s, d); err != nil {
//	  log.Fatal(err)
//	}
//	// 此时 d 将为 {Name: "Alice", Age: 30, City: ""}
func CopyStruct(src, dst interface{}) error {
	return structDeepCopy(reflect.ValueOf(src), reflect.ValueOf(dst))
}

func structDeepCopy(src, dst reflect.Value) error {
	// Check if src and dst are pointers
	if src.Kind() != reflect.Ptr || dst.Kind() != reflect.Ptr {
		return fmt.Errorf("src and dst must be pointers")
	}

	// Get the element that the pointer points to
	src = src.Elem()
	dst = dst.Elem()

	// Check if the elements are structs
	if src.Kind() != reflect.Struct || dst.Kind() != reflect.Struct {
		return fmt.Errorf("src and dst must point to structs")
	}

	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		dstField := dst.FieldByName(src.Type().Field(i).Name)

		if dstField.IsValid() && dstField.CanSet() {
			srcFieldKind := srcField.Kind()
			dstFieldKind := dstField.Kind()

			if srcFieldKind == reflect.Struct && dstFieldKind == reflect.Struct {
				// Recursively deep copy nested structs
				if err := structDeepCopy(srcField.Addr(), dstField.Addr()); err != nil {
					return err
				}
			} else if srcFieldKind == reflect.Slice && dstFieldKind == reflect.Slice {
				// Deep copy slices
				if srcField.IsNil() {
					dstField.Set(reflect.Zero(dstField.Type()))
					continue
				}
				dstSlice := reflect.MakeSlice(dstField.Type(), srcField.Len(), srcField.Cap())
				for j := 0; j < srcField.Len(); j++ {
					srcElem := srcField.Index(j)
					dstElem := dstSlice.Index(j)

					if srcElem.Kind() == reflect.Ptr && !srcElem.IsNil() {
						newDstElemPtr := reflect.New(dstElem.Type().Elem())
						if err := structDeepCopy(srcElem, newDstElemPtr); err != nil {
							return err
						}
						dstElem.Set(newDstElemPtr)
					} else if srcElem.Kind() == reflect.Struct {
						// For struct elements in slice, create a new struct and deep copy
						newDstElem := reflect.New(dstElem.Type()).Elem()
						if err := structDeepCopy(srcElem.Addr(), newDstElem.Addr()); err != nil {
							return err
						}
						dstElem.Set(newDstElem)
					} else if srcElem.Type().AssignableTo(dstElem.Type()) {
						dstElem.Set(srcElem)
					} else {
						return fmt.Errorf("cannot assign slice element of type %s to %s", srcElem.Type(), dstElem.Type())
					}
				}
				dstField.Set(dstSlice)
			} else if srcFieldKind == reflect.Map && dstFieldKind == reflect.Map {
				// Deep copy maps
				if srcField.IsNil() {
					dstField.Set(reflect.Zero(dstField.Type()))
					continue
				}
				dstMap := reflect.MakeMap(dstField.Type())
				for _, key := range srcField.MapKeys() {
					srcMapValue := srcField.MapIndex(key)

					if srcMapValue.Kind() == reflect.Ptr && !srcMapValue.IsNil() {
						newDstMapValuePtr := reflect.New(dstField.Type().Elem().Elem())
						if err := structDeepCopy(srcMapValue, newDstMapValuePtr); err != nil {
							return err
						}
						dstMap.SetMapIndex(key, newDstMapValuePtr)
					} else if srcMapValue.Kind() == reflect.Struct {
						newDstMapValue := reflect.New(dstField.Type().Elem()).Elem()
						if err := structDeepCopy(srcMapValue.Addr(), newDstMapValue.Addr()); err != nil {
							return err
						}
						dstMap.SetMapIndex(key, newDstMapValue)
					} else if srcMapValue.Type().AssignableTo(dstField.Type().Elem()) {
						dstMap.SetMapIndex(key, srcMapValue)
					} else {
						return fmt.Errorf("cannot assign map value of type %s to %s", srcMapValue.Type(), dstField.Type().Elem())
					}
				}
				dstField.Set(dstMap)
			} else if srcField.Type().AssignableTo(dstField.Type()) {
				// Direct copy for assignable types
				dstField.Set(srcField)
			} else if srcField.Type().ConvertibleTo(dstField.Type()) {
				// Convert and copy for convertible types
				dstField.Set(srcField.Convert(dstField.Type()))
			} else {
				return fmt.Errorf("cannot assign field '%s' of type %s to type %s", src.Type().Field(i).Name, srcField.Type(), dstField.Type())
			}
		}
	}
	return nil
}

// StructMerge 函数将多个源结构体中的非零值合并到目标结构体中。
// 源结构体按照传入顺序反向合并，后面的源结构体会覆盖前面的。
// 目标结构体 (dst) 必须是指向结构体的指针，所有源结构体 (src) 必须是与目标结构体类型相同的指针。
//
// 参数:
//   - dst: 一个指向目标结构体的指针，非零值将合并到该结构体中。
//   - src: 一个变参，包含多个指向源结构体的指针，非零值将从这些源结构体中提取。
//
// 返回值:
//   - error: 如果 dst 不是指向结构体的指针，或任何 src 元素不是与 dst 类型相同的结构体指针，则返回错误。
func StructMerge(dst interface{}, src ...interface{}) error {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr || dstVal.Elem().Kind() != reflect.Struct {
		return errors.New("dst must be a pointer to a struct")
	}
	dstVal = dstVal.Elem()
	dstType := dstVal.Type()

	// 验证所有 src 元素是否都是指向与 dst 相同类型的结构体指针
	for _, s := range src {
		srcVal := reflect.ValueOf(s)
		if srcVal.Kind() != reflect.Ptr || srcVal.Elem().Kind() != reflect.Struct || srcVal.Elem().Type() != dstType {
			return errors.New("all src must be pointers to structs of the same type as dst")
		}
	}

	// 反向遍历源结构体数组，以确保后面的覆盖前面的
	for i := len(src) - 1; i >= 0; i-- {
		srcVal := reflect.ValueOf(src[i]).Elem()

		// 遍历源结构体的每个字段
		for j := 0; j < srcVal.NumField(); j++ {
			srcField := srcVal.Field(j)
			dstField := dstVal.FieldByName(srcVal.Type().Field(j).Name)

			// 检查目标结构体中是否有对应的字段
			if dstField.IsValid() && dstField.CanSet() {
				if srcField.Kind() == reflect.Struct && dstField.Kind() == reflect.Struct {
					// 递归处理嵌套结构体
					err := StructMerge(dstField.Addr().Interface(), srcField.Addr().Interface())
					if err != nil {
						return err
					}
				} else if srcField.Type() == dstField.Type() {
					// 检查源字段是否为零值
					zeroValue := reflect.Zero(srcField.Type()).Interface()
					if !reflect.DeepEqual(srcField.Interface(), zeroValue) {
						// 如果源字段不是零值，则将其设置到目标字段
						dstField.Set(srcField)
					}
				}
			}
		}
	}
	return nil
}

// GetFieldNameByJSONTag 根据 JSON 标记获取结构体字段名
func GetFieldNameByJSONTag(objType reflect.Type, jsonKey string) string {
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		tag := field.Tag.Get("json")
		if tag == jsonKey {
			return field.Name
		}
		// 支持逗号分隔的多个 JSON 名称
		tags := strings.Split(tag, ",")
		for _, t := range tags {
			if t == jsonKey {
				return field.Name
			}
		}
	}
	return ""
}

// CalculateAge 计算年龄的多功能方法
// 支持以下两种调用方式：
// 1. CalculateAge(year, month, day int) (int, error)
// 2. CalculateAge(dateString string) (int, error)
func CalculateAge(args ...interface{}) (int, error) {
	var year, month, day int

	switch len(args) {
	case 3:
		// 处理 year, month, day 的情况
		var ok bool
		if year, ok = args[0].(int); !ok {
			return 0, errors.New("年份参数类型必须为int")
		}
		if month, ok = args[1].(int); !ok {
			return 0, errors.New("月份参数类型必须为int")
		}
		if day, ok = args[2].(int); !ok {
			return 0, errors.New("日期参数类型必须为int")
		}
	case 1:
		// 处理 dateString 的情况
		dateString, ok := args[0].(string)
		if !ok {
			return 0, errors.New("单一参数类型必须为string")
		}
		// 支持多种日期格式
		formats := []string{"2006-01-02", "2006/01/02"}
		var date time.Time
		var err error
		for _, format := range formats {
			date, err = time.Parse(format, dateString)
			if err == nil {
				break
			}
		}
		if err != nil {
			return 0, errors.New("无效的日期格式")
		}
		year, month, day = date.Year(), int(date.Month()), date.Day()
	default:
		return 0, errors.New("无效的参数数量")
	}

	// 计算年龄
	now := time.Now()
	age := now.Year() - year
	if now.Month() < time.Month(month) || (now.Month() == time.Month(month) && now.Day() < day) {
		age--
	}

	return age, nil
}

// ParseIP 解析IP地址，输出是ipv4或ipv6
// 0: invalid ip
// 4: ipv4
// 6: ipv6
func ParseIP(s string) (net.IP, int) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, 0
	}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return ip, 4
		case ':':
			return ip, 6
		}
	}
	return nil, 0
}

// ParseChineseIDCard 解析中国大陆身份证号码，提取性别、年龄、生日、出生地等信息
func ParseChineseIDCard(idCard string) (gender string, age int, birthDay string, regionCode string, sequenceCode string, err error) {
	if !validator.IsChineseIDCard(idCard) {
		return "", 0, "", "", "", fmt.Errorf("无效的居民身份证")
	}

	var year, month, day int
	if len(idCard) == 15 {
		year, _ = strconv.Atoi("19" + idCard[6:8])
		month, _ = strconv.Atoi(idCard[8:10])
		day, _ = strconv.Atoi(idCard[10:12])
	} else if len(idCard) == 18 {
		year, _ = strconv.Atoi(idCard[6:10])
		month, _ = strconv.Atoi(idCard[10:12])
		day, _ = strconv.Atoi(idCard[12:14])
	}
	birthDay = fmt.Sprintf("%04d-%02d-%02d", year, month, day)

	// 计算年龄
	age, _ = CalculateAge(year, month, day)

	// 解析性别
	var genderCode int
	if len(idCard) == 15 {
		genderCode, _ = strconv.Atoi(string(idCard[14]))
		sequenceCode = idCard[12:15]
	} else if len(idCard) == 18 {
		genderCode, _ = strconv.Atoi(string(idCard[16]))
		sequenceCode = idCard[14:17]
	}
	if genderCode%2 == 0 {
		gender = "女"
	} else {
		gender = "男"
	}

	// 提取区域码
	regionCode = idCard[:6]

	return gender, age, birthDay, regionCode, sequenceCode, nil
}

// Base64Encode base64加密
func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// Base64Decode base64解密
func Base64Decode(str string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Random 生成随机字符
//
// 参数:
//   - length (必需): 随机字符长度。
//   - numericOnly (可选): 如果为 true，则只包含数字字符。
//
// 返回值:
//   - string: 生成的随机字符串。
//
// 示例:
//
//	randomString := Random(16)
//	randomString := Random(16, true)
func Random(length int, args ...bool) string {
	var charset string
	var numericOnly bool
	if len(args) > 0 {
		numericOnly = args[0]
	}
	if numericOnly {
		charset = "0123456789"
	} else {
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}

	seed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(seed)
	randomString := make([]byte, length)
	for i := range randomString {
		randomString[i] = charset[random.Intn(len(charset))]
	}
	return string(randomString)
}

// IsError 判断[]error是否存在错误
func IsError(errs []error) bool {
	for _, err := range errs {
		if err != nil {
			return true
		}
	}
	return false
}

// IsEmptyValue 检查值是否为空
func IsEmptyValue(val interface{}) bool {
	if val == nil {
		return true
	}

	value := reflect.ValueOf(val)
	switch value.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return value.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Interface, reflect.Ptr:
		if value.IsNil() {
			return true
		}
		return IsEmptyValue(value.Elem().Interface())
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if !IsEmptyValue(value.Field(i).Interface()) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(val, reflect.Zero(reflect.TypeOf(val)).Interface())
	}
}

// maxStructDefaultDepth 结构体默认值递归补充的最大层级
// 说明：
//   - 循环引用主要依靠指针地址登记表 visited 收敛（visited 为「当前递归路径」集合，进入登记、返回注销）；
//   - 存在不含指针的环（例如 map[string]interface{} 通过 interface 装下自身），visited 无法识别，
//     故仍需深度上限兜底；达到上限后不再深入，避免无限递归导致栈溢出；
//   - 深度语义统一为「容器嵌套层数」：每深入一层 struct / map 元素 / 指针指向的 struct 都会 +1；
//     interface 仅作拆包、不消耗深度，字段自身的 map 容器也不消耗深度（其元素才计入）；
//   - setDefaultForStruct 与 setDefaultForStructMap 两处入口都使用 depth >= max 校验：
//     后者的校验不可省略 —— map → interface → map 这类无指针环路完全不经过 setDefaultForStruct；
//   - 上限为容器嵌套层数，超出后静默停止深入（不报错、不 panic）。正常配置的嵌套深度远小于该值，
//     不要依赖「恰好多少层可用」这种由实现倒推得出的具体数字
const maxStructDefaultDepth = 32

// CheckAndSetDefault 检查结构体中的字段是否为空，如果为空则设置为默认值
// 函数名：CheckAndSetDefault
// 参数：i interface{} — 结构体或其指针，支持一层指针传入
// 返回值：error — 始终返回nil；本函数不抛出解析错误，保持兼容的静默行为
// 异常：不触发panic（除非外部传入不可寻址的值并强制Addr）
// 使用说明：
//   - 仅对以下类型在“空值”时设置默认：string、bool、int系、float32/float64、time.Duration
//   - struct、interface、ptr 三类字段会向下递归处理，递归过程中贯通传递深度与已访问指针集合
//   - map字段的值类型为自定义配置结构体时，会对每个元素递归补充默认值（支持结构体、结构体指针、interface承载的结构体）
//   - 切片字段的元素为自定义配置结构体时同样递归补充默认值（支持 []Struct、[]*Struct、[]interface{}、嵌套切片与 map[string][]Struct）
//   - 只有「配置结构体」才会被递归，判定标准为「在递归深度内，存在至少一个会被本工具作用的可导出字段」：
//     带 default 标签的字段、切片/映射字段（无标签时会被初始化为空容器）、可能装载配置结构体的 interface 字段；
//     因此 sync.Mutex、atomic.Value、url.URL、time.Time 这类无相关字段的第三方类型会被整体跳过，
//     既避免对不该复制的类型做「拷贝 → 写回」（例如 map[string]sync.Mutex），也省掉无收益的遍历
//   - nil 指针字段保持原样、不会被自动实例化（与 map 值为 nil 指针的处理一致）；
//     若需要指针字段也带上默认值，请先显式初始化：cfg.Inner = &InnerCfg{}
//   - 切片在递归范围内：[]SubConfig 的值元素会「拷贝 → 补充 → 写回」，[]*SubConfig 的非 nil 指针就地补充、
//     nil 指针保持原样，[]interface{} 按实际类型分派，[]T 的嵌套切片与 map[string][]SubConfig 逐层向内；
//     元素为基础类型的切片（[]string、[]int 等）不会被逐元素改写，仅空切片按 default 标签整体填充
//   - 递归受环路保护：含指针的环由 visited 路径集合收敛，不含指针的环（如 map 通过 interface 装下自身）由 maxStructDefaultDepth 兜底
//   - default标签为空时不会报错，数值/布尔解析失败会被忽略（保持原有逻辑）
//
// 行为变更提示（相比旧实现）：
//   - 旧实现对指针字段完全不处理，现在非 nil 的 *Struct 字段会被递归补充默认值；
//   - 旧实现对 interface 字段实际是空操作，现在会解引用后按实际类型（结构体 / 结构体指针 / 映射）递归补充；
//   - 旧实现对非空 map 的元素完全不处理，现在会对命中判定条件的 map 元素递归补充，并就地写回原 map
//     （调用方在别处共享的同一个 map 也会看到变化）。
//
// 使用示例：
//
//	type SubConfig struct {
//	    Host string `json:"host" default:"127.0.0.1"`
//	    Port int    `json:"port" default:"3306"`
//	}
//	type AppConfig struct {
//	    Name    string               `json:"name" default:"myapp"`
//	    Enabled bool                 `json:"enabled" default:"true"`
//	    Port    int                  `json:"port" default:"8080"`
//	    Timeout  time.Duration       `json:"timeout" default:"300ms"`
//	    Clusters map[string]SubConfig `json:"clusters"`
//	    Inner    *SubConfig          `json:"inner"`
//	}
//	cfg := &AppConfig{Clusters: map[string]SubConfig{"main": {}}, Inner: &SubConfig{}}
//	_ = helper.CheckAndSetDefault(cfg)
//	// cfg.Clusters["main"] == SubConfig{Host: "127.0.0.1", Port: 3306}
//	// cfg.Inner           == &SubConfig{Host: "127.0.0.1", Port: 3306}
//
// 常见问题：如果发现默认值赋值失败，但是又没有出现报错，可以看看是不是传递的指针的指针
func CheckAndSetDefault(i interface{}) error {
	// 获取结构体反射值
	val := reflect.ValueOf(i)

	// 如果传入的是指针类型，获取指向的结构体
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	return setDefaultForStruct(val, 0, make(map[uintptr]struct{}))
}

// isNestedStructKind 判断字段类型是否属于需要向下递归的「结构体载体」类型
// 函数名：isNestedStructKind
// 参数：kind reflect.Kind — 字段的反射种类
// 返回值：bool — 为 struct、interface 或 ptr 时返回true，表示需要递归尝试补充默认值
// 异常：不触发panic
// 使用说明：
// - struct：值类型结构体，直接递归其字段
// - interface：运行期可能装载结构体、结构体指针或映射，需要解引用后判定
// - ptr：可能指向配置结构体，也可能指向基础类型；本函数仅做粗筛，具体判定在 setDefaultForNestedField 内完成
// 使用示例：
//
//	if isNestedStructKind(field.Kind()) {
//	    if err := setDefaultForNestedField(field, depth, visited); err != nil {
//	        return err
//	    }
//	    continue
//	}
func isNestedStructKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Struct, reflect.Interface, reflect.Ptr:
		return true
	default:
		return false
	}
}

// setDefaultForNestedField 为结构体中的「结构体载体」字段递归补充默认值
// 函数名：setDefaultForNestedField
// 参数：
// - field reflect.Value — 结构体中可设置的字段值，其 Kind 为 Struct、Interface 或 Ptr
// - depth int — 当前所在结构体的递归深度
// - visited map[uintptr]struct{} — 已递归过的指针地址集合
// 返回值：error — 递归过程中返回的错误
// 异常：不触发panic；以下情况直接跳过，不做任何修改
// 使用说明：
// - struct 字段：类型命中配置结构体判定时递归补充其字段默认值；未命中的类型（time.Time、sync.Mutex、
// url.URL 等）整体跳过 —— 递归本来也不会改到它们，跳过可省掉无收益遍历
// - interface 字段：解引用装载的实际值后按实际类型分派，支持结构体、结构体指针、映射
// - ptr 字段：非 nil 且指向配置结构体时递归 Elem()；nil 指针保持原样不实例化
// - struct / ptr 字段递归时传递 depth+1，每次深入一层容器消耗一层深度；interface 仅拆包不消耗深度
//
// 设计约定（重要）：
// - nil 指针不会被自动实例化。nil 表达「字段不存在 / 未启用 / 未配置」，
// 自动 New 会把「不存在」变成「存在且带默认值」，改变调用方语义；
// 该约定与 setDefaultForMapValue 中 map 值为 nil 指针的处理保持一致。
// 若调用方需要指针字段也有默认值，请先自行显式初始化：cfg.Inner = &InnerCfg{}
//
// 使用示例：
//
//	if err := setDefaultForNestedField(field, depth, visited); err != nil {
//	    return err
//	}
func setDefaultForNestedField(field reflect.Value, depth int, visited map[uintptr]struct{}) error {
	switch field.Kind() {
	case reflect.Struct:
		// 非配置结构体整体跳过，避免对第三方类型做无意义的递归
		if !isConfigStructType(field.Type()) {
			return nil
		}
		return setDefaultForStruct(field, depth+1, visited)

	case reflect.Interface:
		return setDefaultForInterfaceField(field, depth, visited)

	case reflect.Ptr:
		return setDefaultForPtrField(field, depth, visited)

	default:
		return nil
	}
}

// setDefaultForInterfaceField 解引用 interface 字段装载的实际值后分派默认值补充
// 函数名：setDefaultForInterfaceField
// 参数：
// - field reflect.Value — 结构体中的 interface 字段值
// - depth int — 当前所在结构体的递归深度
// - visited map[uintptr]struct{} — 已递归过的指针地址集合
// 返回值：error — 递归过程中返回的错误
// 异常：不触发panic；nil interface 或不可设置的实际值直接跳过
// 使用说明：
//   - interface 承载的实际值在反射中不可直接 Set，但其中若为映射（引用类型）可直接修改；
//     若为值类型结构体，则无法就地修改，需要重新赋值回接口字段
//   - interface 本身只是包装层，不消耗深度：内部按实际类型分派时传入的仍是 depth，
//     由 struct / ptr / map 各自的实现决定 +1，保证与「同类型字段」路径的深度口径一致
//   - 修复要点：原实现遇到 interface 字段直接 continue，导致接口装载的 map 永远不会进入 map 分支
//
// 使用示例：
//
//	if err := setDefaultForInterfaceField(field, depth, visited); err != nil {
//	    return err
//	}
func setDefaultForInterfaceField(field reflect.Value, depth int, visited map[uintptr]struct{}) error {
	if field.IsNil() {
		return nil
	}

	actual := field.Elem()
	if !actual.IsValid() {
		return nil
	}

	// 映射、切片、结构体指针等引用类型可直接就地修改，无需写回接口字段
	switch actual.Kind() {
	case reflect.Map:
		return setDefaultForStructMap(actual, depth, visited)
	case reflect.Slice:
		return setDefaultForSliceField(actual, depth, visited)
	case reflect.Ptr:
		return setDefaultForPtrField(actual, depth, visited)
	}

	// 值类型结构体无法就地修改，需要通过 setDefaultForMapValue 的拷贝写回语义处理
	if !isConfigStructType(actual.Type()) {
		return nil
	}
	newValue, changed, err := setDefaultForMapValue(actual, depth, visited)
	if err != nil {
		return err
	}
	// 注意：不能对 interface 字段调用 Type().Elem()（interface 类型没有 Elem 方法，会 panic），
	// 直接校验新值能否赋值给接口字段本身即可
	if !changed || !newValue.IsValid() || !newValue.Type().AssignableTo(field.Type()) {
		return nil
	}
	field.Set(newValue)

	return nil
}

// setDefaultForPtrField 为指针字段指向的配置结构体递归补充默认值
// 函数名：setDefaultForPtrField
// 参数：
// - field reflect.Value — 结构体中的指针字段值，或接口解引用后的指针值
// - depth int — 当前所在结构体的递归深度
// - visited map[uintptr]struct{} — 当前递归路径上已访问的指针地址集合
// 返回值：error — 递归过程中返回的错误
// 异常：不触发panic
// 使用说明：
// - nil 指针保持原样，不自动实例化（见 setDefaultForNestedField 的设计约定）
// - 非配置结构体指针（如 *int）直接跳过
// - visited 语义为「当前递归路径」集合：进入时登记、函数返回时注销（defer delete），
// 因此只用于阻断真正的环，不会把「同一指针被多个字段/map 键引用」（DAG）误判为已处理；
// 默认值填充本身幂等，重复处理同一指针不会产生副作用
// - 深度已耗尽时直接返回且不登记地址：否则会在 visited 中残留一条「未真正处理过」的记录，
// 导致后续在更浅位置再次遇到该指针时被静默跳过、默认值漏填
// - 注意：本函数对非 nil 且指向配置结构体的指针会消耗一层深度（depth+1）
// 使用示例：
//
//	if err := setDefaultForPtrField(field, depth, visited); err != nil {
//	    return err
//	}
func setDefaultForPtrField(field reflect.Value, depth int, visited map[uintptr]struct{}) error {
	if field.IsNil() || !isConfigStructType(field.Type().Elem()) {
		return nil
	}

	// 深度耗尽时不要登记地址，避免污染后续同指针的浅层引用
	if depth >= maxStructDefaultDepth {
		return nil
	}

	addr := field.Pointer()
	if _, exists := visited[addr]; exists {
		return nil
	}
	visited[addr] = struct{}{}
	// 路径集合语义：本层处理完即注销，仅阻断真正的环；
	// 同一指针被多个字段/map键引用（DAG）时会各自处理一次，默认值填充幂等，不会产生副作用
	defer delete(visited, addr)

	return setDefaultForStruct(field.Elem(), depth+1, visited)
}

// setDefaultForStruct 对结构体反射值逐字段补充默认值，并支持嵌套结构递归
// 函数名：setDefaultForStruct
// 参数：
// - val reflect.Value — 已解引用到结构体的反射值；非结构体时直接返回nil
// - depth int — 当前递归深度（语义为容器嵌套层数），达到 maxStructDefaultDepth 后停止继续深入
// - visited map[uintptr]struct{} — 当前递归路径上已访问的指针地址集合（进入登记、返回注销）
// 返回值：error — 始终返回nil或递归过程中返回的错误
// 异常：不触发panic；不可设置的字段会被跳过
// 使用说明：
// 字段按以下三类分别处理：
//   - time.Time 字段：特判并交给 setTimeDefault 填充 default 标签（详见该函数说明）。
//     必须特判的原因：time.Time 的 Kind 为 Struct，若不特判会被 isNestedStructKind
//     的 Struct 分支拦截并转入「结构体向内递归」，导致其 default 标签被静默忽略
//   - struct / interface / ptr 字段：交给 setDefaultForNestedField 分派递归
//   - 其余字段（string / bool / int / float / time.Duration / slice / map）：
//     按「零值才填」的语义在此直接处理；其中 slice / map 的 default 标签
//     对空容器还意味着「初始化为空切片 / 空映射」
//
// 使用示例：
//
//	err := setDefaultForStruct(reflect.ValueOf(cfg).Elem(), 0, make(map[uintptr]struct{}))
func setDefaultForStruct(val reflect.Value, depth int, visited map[uintptr]struct{}) error {
	// 不是结构体的时候直接跳过处理
	if val.Kind() != reflect.Struct {
		// log.Printf("%s 不是结构体，直接跳过处理", val.String())
		return nil
	}

	// 超过递归深度上限，不再深入，避免自引用结构无限递归
	if depth >= maxStructDefaultDepth {
		return nil
	}

	// 遍历结构体字段
	for idx := 0; idx < val.NumField(); idx++ {
		field := val.Field(idx)
		fieldType := val.Type().Field(idx)

		// 忽略非导出字段
		if !field.CanSet() {
			continue
		}

		// time.Time 是值语义的时间类型，Kind 虽为 Struct，但语义同基础类型：
		// 需要作为「字段本身」填充默认值，而不是作为结构体向内递归。
		// 故必须在此特判并 continue，否则会被下面的 isNestedStructKind 的 Struct 分支拦截，
		// 导致 default 标签被静默忽略（这正是本次要修复的问题）
		if field.Type() == timeType {
			setTimeDefault(field, fieldType.Tag.Get("default"))
			continue
		}

		// 需要向下递归的字段类型统一在此处理：
		// - struct / interface / ptr 都可能是配置结构体的载体，交由 setDefaultForNestedField 分派
		if isNestedStructKind(field.Kind()) {
			if err := setDefaultForNestedField(field, depth, visited); err != nil {
				return err
			}
			continue
		}

		// 获取字段类型和默认值标签
		tag := fieldType.Tag.Get("default")
		fieldKind := field.Kind()

		// 字符串：空字符串设置为默认值，支持命名字符串类型
		if fieldKind == reflect.String && field.Len() == 0 {
			field.SetString(tag)
		}

		// 布尔：为 false 时，default=="true" 设置为 true
		if fieldKind == reflect.Bool && !field.Bool() {
			defaultVal := tag == "true"
			field.SetBool(defaultVal)
		}

		// 整型：为 0 时设置默认值，支持命名整型类型，解析失败忽略（默认0）
		if (fieldKind == reflect.Int || fieldKind == reflect.Int8 || fieldKind == reflect.Int16 || fieldKind == reflect.Int32 || fieldKind == reflect.Int64) && field.Int() == 0 {
			defaultVal, _ := strconv.ParseInt(tag, 10, 64)
			field.SetInt(defaultVal)
		}

		// 浮点：为 0.0 时设置默认值，解析失败忽略（默认0.0）
		if fieldKind == reflect.Float32 || fieldKind == reflect.Float64 {
			defaultVal, _ := strconv.ParseFloat(tag, 64)
			if field.Float() == 0 {
				field.SetFloat(defaultVal)
			}
		}

		// time.Duration：为 0 时设置默认值，解析失败忽略
		if field.Type() == reflect.TypeOf(time.Duration(0)) && field.Int() == 0 {
			duration, err := time.ParseDuration(tag)
			if err == nil {
				field.SetInt(int64(duration))
			}
		}

		// 切片：为空切片时按 default 标签整体填充；元素类型为配置结构体时递归补充每个元素的默认值
		if fieldKind == reflect.Slice {
			if field.Len() == 0 {
				if err := setDefaultValue(field, tag); err != nil {
					// 静默处理错误，保持与原有逻辑兼容
				}
			}
			if err := setDefaultForSliceField(field, depth, visited); err != nil {
				return err
			}
		}

		// 映射：为空映射时设置默认值；值类型为配置结构体时递归补充每个元素的默认值
		if fieldKind == reflect.Map {
			if field.Len() == 0 {
				if err := setDefaultValue(field, tag); err != nil {
					// 静默处理错误，保持与原有逻辑兼容
				}
			}
			if err := setDefaultForStructMap(field, depth, visited); err != nil {
				return err
			}
		}
	}

	return nil
}

// setTimeDefault 为 time.Time 字段填充 default 标签指定的默认值
// 函数名：setTimeDefault
// 参数：
// - field reflect.Value — 可设置的 time.Time 字段值（调用方需保证其类型为 time.Time）
// - tag string — `default` 标签文本，支持多种时间字符串格式与时间戳
// 返回值：无（本函数不返回错误，所有失败均静默处理）
// 异常：不触发panic
// 使用说明：
// - 仅当字段当前为零值（time.Time.IsZero）时才填充，已有非零值不会被覆盖，
// 与整型/浮点/字符串「零值才填」的语义保持一致
// - 标签为空时不做任何处理
// - 解析失败（非法格式字符串）时静默跳过、不修改字段，
// 与整型/浮点/time.Duration 的既有容错风格一致
// - 解析采用 Convert.ToTime，支持 RFC3339、「2006-01-02 15:04:05」、「2006-01-02」、
// 秒/毫秒/纳秒时间戳字符串、以及中文格式等（详见 Convert.ToTime）
//
// 实现要点：
// Convert.ToTime 在解析失败时返回零值 time.Time{}，无法与「成功解析到零值」区分，
// 因此这里用「解析结果为零值即视为失败」作为判据 —— 该判据不会误伤正常场景：
// 一个合法的默认值本就不应解析为零值（零值时间无业务意义），
// 且字段进入本函数前已确认为零值，即便判据有误也不会产生破坏性写入
//
// 使用示例：
//
//	type Cfg struct {
//	    StartAt time.Time `default:"2024-01-01 00:00:00"`
//	    EndAt   time.Time `default:"2024-12-31 23:59:59"`
//	}
//	cfg := &Cfg{}
//	_ = helper.CheckAndSetDefault(cfg)
//	// cfg.StartAt == 2024-01-01 00:00:00
func setTimeDefault(field reflect.Value, tag string) {
	if tag == "" || !field.IsZero() {
		return
	}

	parsed := Convert{Value: tag}.ToTime()
	// 解析失败返回零值，此时视为「无有效默认值」并跳过，避免写入零值污染语义
	if parsed.IsZero() {
		return
	}

	field.Set(reflect.ValueOf(parsed))
}

// setDefaultForStructMap 对映射中值为配置结构体的元素递归补充默认值
// 函数名：setDefaultForStructMap
// 参数：
// - field reflect.Value — 可设置的映射字段值
// - depth int — 当前递归深度（语义为容器嵌套层数）
// - visited map[uintptr]struct{} — 当前递归路径上已访问的指针地址集合
// 返回值：error — 递归补充默认值过程中返回的错误
// 异常：不触发panic；nil映射或值类型非配置结构体时直接返回nil
// 使用说明：
//   - 支持值类型为结构体、结构体指针、interface{}（运行时为结构体）的映射
//   - map value 在反射中不可寻址，故采用「拷贝到新实例 → 递归补充 → SetMapIndex 写回」的方式
//   - 深度上限校验必须保留在本入口：存在不含结构体与指针的环路
//     （map[string]interface{} 通过 interface 装下自身），该环路径为 map → interface → map，
//     完全不经过 setDefaultForStruct，若此处不设限将无法收敛；
//     校验条件与 setDefaultForStruct 保持一致，均为 depth >= max，两处各自针对不同入口，不存在重复扣减
//   - 含指针环路由 visited 路径集合收敛，无指针环路由本条深度上限兜底
//   - 注意：本层 map 的元素按 depth 处理，元素自身是容器（struct / 指针 / 内层 map）时才 +1
//
// 使用示例：
//
//	if err := setDefaultForStructMap(field, depth, visited); err != nil {
//	    return err
//	}
func setDefaultForStructMap(field reflect.Value, depth int, visited map[uintptr]struct{}) error {
	// 深度上限兜底：无指针环路（map -> interface -> map）不经过 setDefaultForStruct，必须在此拦截
	if depth >= maxStructDefaultDepth {
		return nil
	}

	if field.Kind() != reflect.Map || field.IsNil() || field.Len() == 0 {
		return nil
	}

	if !mapValueMayBeConfigStruct(field.Type().Elem()) {
		return nil
	}

	for _, key := range field.MapKeys() {
		newValue, changed, err := setDefaultForMapValue(field.MapIndex(key), depth, visited)
		if err != nil {
			return err
		}
		// 指针类型元素已就地修改，无需写回
		if !changed || !newValue.IsValid() {
			continue
		}
		if !newValue.Type().AssignableTo(field.Type().Elem()) {
			continue
		}
		field.SetMapIndex(key, newValue)
	}

	return nil
}

// setDefaultForMapValue 为映射中的单个元素递归补充默认值
// 函数名：setDefaultForMapValue
// 参数：
// - mapValue reflect.Value — 映射中的元素值，反射中不可寻址
// - depth int — 当前递归深度（语义为容器嵌套层数）
// - visited map[uintptr]struct{} — 当前递归路径上已访问的指针地址集合
// 返回值：
// - reflect.Value — 补充默认值后的元素值；需要写回映射时有效
// - bool — 是否需要调用方执行写回；结构体拷贝返回true，指针已就地修改返回false
// - error — 递归过程中返回的错误
// 异常：不触发panic；nil指针、非配置结构体元素均被跳过
// 使用说明：
// - 深度口径与「同类型的结构体字段」保持完全一致：
// 值结构体元素按 depth+1 处理（等价于 struct 字段），指针元素交给 setDefaultForPtrField(depth)
// 由其内部统一 +1（等价于 ptr 字段），内层 map 元素按 depth+1 处理（等价于嵌套 map）
// - interface 只是包装层，不消耗深度，直接按原 depth 解引用后继续分派
// 使用示例：
//
//	newValue, changed, err := setDefaultForMapValue(field.MapIndex(key), depth, visited)
//	if changed && err == nil {
//	    field.SetMapIndex(key, newValue)
//	}
func setDefaultForMapValue(mapValue reflect.Value, depth int, visited map[uintptr]struct{}) (reflect.Value, bool, error) {
	switch mapValue.Kind() {
	case reflect.Struct:
		if !isConfigStructType(mapValue.Type()) {
			return reflect.Value{}, false, nil
		}
		// map value 不可寻址，先拷贝到新实例再递归补充
		copied := reflect.New(mapValue.Type())
		copied.Elem().Set(mapValue)
		if err := setDefaultForStruct(copied.Elem(), depth+1, visited); err != nil {
			return reflect.Value{}, false, err
		}
		return copied.Elem(), true, nil

	case reflect.Ptr:
		// nil 指针保持原样，不自动实例化，避免改变调用方的 nil 语义
		// 复用 setDefaultForPtrField：内部已包含 nil 判定、配置结构体判定、深度消耗与指针地址环检测
		if err := setDefaultForPtrField(mapValue, depth, visited); err != nil {
			return reflect.Value{}, false, err
		}
		return reflect.Value{}, false, nil

	case reflect.Map:
		// 内层 map 本身是引用类型，可直接修改其中的元素，无需拷贝写回
		// depth+1 表示「进入下一层容器」，保证无指针环路（map -> interface -> map）能逐层消耗深度直至触发上限
		if err := setDefaultForStructMap(mapValue, depth+1, visited); err != nil {
			return reflect.Value{}, false, err
		}
		return reflect.Value{}, false, nil

	case reflect.Slice:
		// 映射的值类型为切片（如 map[string][]SubConfig）：切片本身是引用类型，可直接修改其元素
		// depth+1 表示「进入下一层容器」，与 map 分支口径一致
		if err := setDefaultForSliceField(mapValue, depth+1, visited); err != nil {
			return reflect.Value{}, false, err
		}
		return reflect.Value{}, false, nil

	case reflect.Interface:
		if mapValue.IsNil() {
			return reflect.Value{}, false, nil
		}
		// 接口只是包装层，本身不消耗深度；深度 +1 由解引用后的实际类型分支负责，
		// 避免「interface -> map -> interface」这类链路上同一层被重复计数
		return setDefaultForMapValue(mapValue.Elem(), depth, visited)

	default:
		return reflect.Value{}, false, nil
	}
}

// configStructTypeCache 缓存 isConfigStructType 的判定结果
// 说明：
// - 类型在运行期不变，判定结果可永久缓存；map[string]Config 这类大映射逐个元素判定时，
// 缓存可把「每个元素走一遍类型树」降为一次 map 查表
// - 全局共享，使用 sync.Map 保证并发调用 CheckAndSetDefault 时的安全
var configStructTypeCache sync.Map

// timeType 预缓存 time.Time 的反射类型
// 说明：setDefaultForStruct 需要按类型识别 time.Time 字段并单独填充其 default 标签，
// 该判定在热路径上反复执行，预先取出可避免每字段一次 reflect.TypeOf 调用
var timeType = reflect.TypeOf(time.Time{})

// isConfigStructType 判断类型是否为需要递归补充默认值的配置结构体
// 函数名：isConfigStructType
// 参数：typ reflect.Type — 待判断的类型
// 返回值：bool — 为配置结构体时返回true；基础类型、time.Time 及无 default 标签的第三方类型返回false
// 异常：不触发panic
// 使用说明：
// - 判定标准不是「是不是 struct」，而是「递归范围内是否存在带 default 标签的可导出字段」，
// 即 hasActionableField 的语义。只有这样的类型递归才有收益
// - 该判定同时把 http.Request、http.Response 这类「含 map / interface 字段但无 default 标签」的
// 第三方类型挡在门外，避免递归进入第三方对象、把其 nil map / nil slice 静默改写为空容器
// - sync.Mutex、atomic.Value、url.URL 等无标签类型同样整体跳过，可避免「拷贝 → 写回」
// 这类无意义且可能有隐患的操作（例如 map[string]sync.Mutex 不会被静默复制，
// 反射路径绕过了 go vet 的 copylocks 检查）
// - time.Time 是「值语义的时间类型」，需作为字段本身被填充（而非作为结构体向内递归），
// 故此处返回 false，由 setDefaultForStruct 的特判分支单独处理
// - 结果带类型级缓存，重复判定不会重复遍历类型树
// 使用示例：
//
//	if isConfigStructType(field.Type().Elem()) {
//	    // 对映射中的元素递归补充默认值
//	}
func isConfigStructType(typ reflect.Type) bool {
	if typ == nil || typ.Kind() != reflect.Struct {
		return false
	}

	// time.Time 不参与「结构体向内递归」，其 default 标签由 setDefaultForStruct 的特判分支处理
	if typ == reflect.TypeOf(time.Time{}) {
		return false
	}

	if cached, ok := configStructTypeCache.Load(typ); ok {
		return cached.(bool)
	}

	result := hasActionableField(typ, 0, make(map[reflect.Type]struct{}))
	configStructTypeCache.Store(typ, result)

	return result
}

// hasActionableField 判断结构体类型在递归范围内是否存在带 default 标签的可导出字段
// 函数名：hasActionableField
// 参数：
// - typ reflect.Type — 待判断的结构体类型
// - depth int — 当前类型递归深度，超过 maxStructDefaultDepth 后不再深入
// - visiting map[reflect.Type]struct{} — 当前类型递归路径上正在判定的类型，用于阻断类型层面的环
// 返回值：bool — 存在带 default 标签的字段时返回true
// 异常：不触发panic
// 使用说明：
// 判定标准收紧为「存在带 default 标签的可导出字段」：
//   - 只有 default 标签才是「本工具会改写该字段」的可靠信号；
//     切片/映射字段无标签时的「初始化为空容器」只作用于字段自身，
//     不会因递归而获益，故不再作为判定依据
//   - interface 字段同样不再作为判定依据：它虽可能在运行期装载配置结构体，
//     但真实配置结构体的宿主必然带有 default 标签，判定自会通过；
//     反之若宿主毫无标签，递归也无任何字段可填充，属行为中性
//
// 递归规则：struct 字段与「指向结构体的指针」字段继续向内判定；指针未指向结构体时不计入
// 环保护：类型进入 visiting 后若再次被访问，视为不可作用（返回false）以阻断类型层面的环
//
// 收紧的收益：避免把 http.Request / http.Response 这类含 map / interface 字段的第三方类型
// 误判为「配置结构体」，从而杜绝递归进入第三方对象、把其 nil map / nil slice
// 静默改写为空容器的副作用（该副作用由 setDefaultValue 的空标签分支触发）
//
// 注意：判定收紧只影响「无 default 标签」的类型；顶层直接调用 CheckAndSetDefault 时不经过本判定，
// 内部的 interface / map / ptr 字段递归能力不受影响
//
// 使用示例：
//
//	if hasActionableField(reflect.TypeOf(SubConfig{}), 0, make(map[reflect.Type]struct{})) {
//	    // 该类型值得递归
//	}
func hasActionableField(typ reflect.Type, depth int, visiting map[reflect.Type]struct{}) bool {
	if typ == nil || typ.Kind() != reflect.Struct {
		return false
	}

	if depth >= maxStructDefaultDepth {
		return false
	}

	if _, exists := visiting[typ]; exists {
		return false
	}
	visiting[typ] = struct{}{}
	defer delete(visiting, typ)

	for idx := 0; idx < typ.NumField(); idx++ {
		field := typ.Field(idx)

		// 非导出字段不会被本工具改写，直接跳过
		if field.PkgPath != "" {
			continue
		}

		// 带 default 标签的字段一定会被作用（含空标签：切片/映射会被初始化为空容器）
		if _, hasTag := field.Tag.Lookup("default"); hasTag {
			return true
		}

		switch field.Type.Kind() {
		case reflect.Struct:
			if hasActionableField(field.Type, depth+1, visiting) {
				return true
			}

		case reflect.Ptr:
			elem := field.Type.Elem()
			if elem.Kind() == reflect.Struct && hasActionableField(elem, depth+1, visiting) {
				return true
			}

		case reflect.Slice:
			// 切片元素可能是配置结构体（如 []SubConfig）：元素需要被递归补充时，
			// 宿主类型也应被判为配置结构体，否则整段切片能力会被上层的类型判定挡掉
			//
			// [环安全] 严禁在此分支经 sliceElemMayBeConfigStruct → isConfigStructType 判定：
			// isConfigStructType 是带缓存的独立入口，会新建 visiting 并把 depth 归零，
			// 自引用类型（T → []*T）会经该边界无限互递归直至栈溢出（进程级崩溃，无法 recover）。
			// hasActionableFieldInType 自身支持容器逐层剥壳并贯通 visiting / depth，直接调用即可。
			if hasActionableFieldInType(field.Type.Elem(), depth+1, visiting) {
				return true
			}
		}
	}

	return false
}

// hasActionableFieldInType 判断「非结构体类型」内部是否存在可作用的默认值字段
// 函数名：hasActionableFieldInType
// 参数：
// - typ reflect.Type — 待判断的类型，可能是结构体、结构体指针、切片、映射或基础类型
// - depth int — 当前类型递归深度，超过 maxStructDefaultDepth 后不再深入
// - visiting map[reflect.Type]struct{} — 当前类型递归路径上正在判定的类型，用于阻断类型层面的环
// 返回值：bool — 内部存在带 default 标签的字段时返回true
// 异常：不触发panic
// 使用说明：
// - 本函数是 hasActionableField 的「容器类型」补充入口：hasActionableField 只接受结构体，
// 切片 / 映射 / 结构体指针需要先剥掉容器包装才能继续向内判定
// - 逐层剥壳后交给 hasActionableField，由其在结构体层面做实际判定并维护 visiting 环保护
// - 基础类型、nil 类型返回 false
//
// 使用示例：
//
//	if hasActionableFieldInType(reflect.TypeOf([]SubConfig{}).Elem(), 0, make(map[reflect.Type]struct{})) {
//	    // 切片元素值得递归
//	}
func hasActionableFieldInType(typ reflect.Type, depth int, visiting map[reflect.Type]struct{}) bool {
	if typ == nil || depth >= maxStructDefaultDepth {
		return false
	}

	switch typ.Kind() {
	case reflect.Struct:
		return hasActionableField(typ, depth, visiting)
	case reflect.Ptr, reflect.Slice, reflect.Array, reflect.Map:
		return hasActionableFieldInType(typ.Elem(), depth+1, visiting)
	default:
		return false
	}
}

// setDefaultForSliceField 为切片字段逐元素递归补充默认值
// 函数名：setDefaultForSliceField
// 参数：
// - field reflect.Value — 可设置的切片字段值，或 map 元素 / interface 解引用后的切片值
// - depth int — 当前递归深度（语义为容器嵌套层数）
// - visited map[uintptr]struct{} — 当前递归路径上已访问的指针地址集合
// 返回值：error — 递归补充默认值过程中返回的错误
// 异常：不触发panic；nil切片、空切片或元素类型无收益时直接返回nil
// 使用说明：
// - 切片元素的处理障碍与 map 值一致：反射中「切片元素」不可寻址，
// 值类型结构体元素需「拷贝到新实例 → 递归补充 → Set 写回该下标」才能生效；
// 指针元素因指向内存可直接修改，就地递归即可（详见 setDefaultForSliceElement）
// - 深度上限校验必须保留在本入口：存在「切片通过 interface 装下自身」这类不含指针的环，
// 该环路径为 slice → interface → slice，完全不经过 setDefaultForStruct，
// 若此处不设限将无法收敛；校验条件与 setDefaultForStruct 保持一致
// - sliceElemMayBeConfigStruct 快速跳过无收益的元素类型（基础类型、无标签第三方类型），
// 避免对每个元素做一次无效分派
// - 注意：本层切片的元素按 depth 处理，元素自身是容器（struct / 指针 / 内层 map / 内层切片）时才 +1
//
// 使用示例：
//
//	if err := setDefaultForSliceField(field, depth, visited); err != nil {
//	    return err
//	}
func setDefaultForSliceField(field reflect.Value, depth int, visited map[uintptr]struct{}) error {
	// 深度上限兜底：切片 ↔ interface 构成的环不经过 setDefaultForStruct，必须在此拦截
	if depth >= maxStructDefaultDepth {
		return nil
	}

	if field.Kind() != reflect.Slice || field.Len() == 0 {
		return nil
	}

	if !sliceElemMayBeConfigStruct(field.Type().Elem()) {
		return nil
	}

	for idx := 0; idx < field.Len(); idx++ {
		newValue, changed, err := setDefaultForSliceElement(field.Index(idx), depth, visited)
		if err != nil {
			return err
		}
		// 指针类型元素已就地修改，无需写回
		if !changed || !newValue.IsValid() {
			continue
		}
		if !newValue.Type().AssignableTo(field.Type().Elem()) {
			continue
		}
		// 切片元素可寻址，直接 Set 写回即可（区别于 map 的 SetMapIndex）
		field.Index(idx).Set(newValue)
	}

	return nil
}

// setDefaultForSliceElement 为切片中的单个元素递归补充默认值
// 函数名：setDefaultForSliceElement
// 参数：
// - elemValue reflect.Value — 切片中的元素值（切片元素可寻址，但值类型结构体仍需拷贝路径处理）
// - depth int — 当前递归深度（语义为容器嵌套层数）
// - visited map[uintptr]struct{} — 当前递归路径上已访问的指针地址集合
// 返回值：
// - reflect.Value — 补充默认值后的元素值；需要调用方写回时有效
// - bool — 是否需要调用方执行写回；结构体拷贝返回true，指针已就地修改返回false
// - error — 递归过程中返回的错误
// 异常：不触发panic；nil指针、非配置结构体元素均被跳过
// 使用说明：
// - 深度口径与 setDefaultForMapValue 保持完全一致，保证「同一元素类型在不同容器下边界相同」：
// 值结构体元素按 depth+1 处理（等价于 struct 字段），指针元素交给 setDefaultForPtrField(depth)
// 由其内部统一 +1（等价于 ptr 字段），内层 map / 内层切片按 depth+1 处理
// - interface 只是包装层，不消耗深度，直接按原 depth 解引用后继续分派
// - 与 map 元素的差异：切片元素可寻址，值结构体元素理论上可直接递归，
// 但仍走「拷贝 → 递归 → 写回」的同一路径，以与 map 保持单一实现、语义完全对齐
//
// 使用示例：
//
//	newValue, changed, err := setDefaultForSliceElement(field.Index(i), depth, visited)
//	if changed && err == nil {
//	    field.Index(i).Set(newValue)
//	}
func setDefaultForSliceElement(elemValue reflect.Value, depth int, visited map[uintptr]struct{}) (reflect.Value, bool, error) {
	switch elemValue.Kind() {
	case reflect.Struct:
		if !isConfigStructType(elemValue.Type()) {
			return reflect.Value{}, false, nil
		}
		// 统一走拷贝路径，与 map 元素保持单一实现
		copied := reflect.New(elemValue.Type())
		copied.Elem().Set(elemValue)
		if err := setDefaultForStruct(copied.Elem(), depth+1, visited); err != nil {
			return reflect.Value{}, false, err
		}
		return copied.Elem(), true, nil

	case reflect.Ptr:
		// nil 指针保持原样，不自动实例化，避免改变调用方的 nil 语义（与 map 值指针一致）
		if err := setDefaultForPtrField(elemValue, depth, visited); err != nil {
			return reflect.Value{}, false, err
		}
		return reflect.Value{}, false, nil

	case reflect.Map:
		// 映射是引用类型，可直接修改其中的元素，无需写回
		if err := setDefaultForStructMap(elemValue, depth+1, visited); err != nil {
			return reflect.Value{}, false, err
		}
		return reflect.Value{}, false, nil

	case reflect.Slice:
		// 嵌套切片（如 [][]SubConfig）：向内一层继续分派，由 setDefaultForSliceField 自行消耗深度
		if err := setDefaultForSliceField(elemValue, depth+1, visited); err != nil {
			return reflect.Value{}, false, err
		}
		return reflect.Value{}, false, nil

	case reflect.Interface:
		if elemValue.IsNil() {
			return reflect.Value{}, false, nil
		}
		// 接口只是包装层，本身不消耗深度；深度 +1 由解引用后的实际类型分支负责
		return setDefaultForSliceElement(elemValue.Elem(), depth, visited)

	default:
		return reflect.Value{}, false, nil
	}
}

// sliceElemMayBeConfigStruct 判断切片的元素类型是否可能是配置结构体
// 函数名：sliceElemMayBeConfigStruct
// 参数：typ reflect.Type — 切片的元素类型，可能为结构体、结构体指针、interface{}、基础类型等
// 返回值：bool — 可能包含配置结构体时返回true；确定不可能时返回false，用于快速跳过无收益切片
// 异常：不触发panic
// 使用说明：
// - 结构体与其指针按 isConfigStructType 判定，未命中判定的第三方类型（sync.Mutex、atomic.Value 等）直接跳过，
// 避免对 []sync.Mutex 这类切片做无收益的「拷贝 → 递归 → 写回」
// - interface{} 静态无法判定，返回true，交由运行时的 setDefaultForSliceElement 按实际类型再次判定
// - 嵌套切片（[][]T）与嵌套映射（[]map[string]T）继续向内判定
// - 语义与 mapValueMayBeConfigStruct 对称，保证切片与映射的递归范围一致
//
// 使用示例：
//
//	if !sliceElemMayBeConfigStruct(field.Type().Elem()) {
//	    return nil
//	}
func sliceElemMayBeConfigStruct(typ reflect.Type) bool {
	if typ == nil {
		return false
	}

	switch typ.Kind() {
	case reflect.Struct:
		return isConfigStructType(typ)
	case reflect.Ptr:
		return isConfigStructType(typ.Elem())
	case reflect.Interface:
		// 接口的静态类型无法确定是否为结构体，需交由运行时判定
		return true
	case reflect.Map:
		// 元素为映射时继续向内判定（如 []map[string]SubConfig）
		return mapValueMayBeConfigStruct(typ)
	case reflect.Slice:
		// 元素为嵌套切片时继续向内判定（如 [][]SubConfig）
		return sliceElemMayBeConfigStruct(typ.Elem())
	default:
		return false
	}
}

// mapValueMayBeConfigStruct 判断映射的值类型是否可能是配置结构体
// 函数名：mapValueMayBeConfigStruct
// 参数：typ reflect.Type — 映射的值类型，可能为结构体、结构体指针、interface{}或基础类型
// 返回值：bool — 可能是配置结构体时返回true；确定不是时返回false，用于快速跳过无需处理的映射
// 异常：不触发panic
// 使用说明：
// - 结构体与其指针按 isConfigStructType 判定，未命中判定的第三方类型（sync.Mutex、atomic.Value 等）直接跳过
// - interface{} 静态无法判定，返回true，交由运行时的 setDefaultForMapValue 按实际类型再次判定
// - 嵌套映射与切片值类型继续向内判定（如 map[string]map[string]Config、map[string][]Config）
// 使用示例：
//
//	if !mapValueMayBeConfigStruct(field.Type().Elem()) {
//	    return nil
//	}
func mapValueMayBeConfigStruct(typ reflect.Type) bool {
	if typ == nil {
		return false
	}

	switch typ.Kind() {
	case reflect.Struct:
		return isConfigStructType(typ)
	case reflect.Ptr:
		return isConfigStructType(typ.Elem())
	case reflect.Interface:
		// 接口的静态类型无法确定是否为结构体，需交由运行时判定
		return true
	case reflect.Map:
		// 多层 map（如 map[string]map[string]Config）继续向内判定
		return mapValueMayBeConfigStruct(typ.Elem())
	case reflect.Slice:
		// 映射的值为切片（如 map[string][]Config）时，需按切片元素继续向内判定，
		// 否则该类型会在元素判定阶段被整段跳过，切片递归能力无法到达
		return sliceElemMayBeConfigStruct(typ.Elem())
	default:
		return false
	}
}

// CheckAndSetDefaultWithPreserveTag 按 `default` 标签设置默认值，同时保留带有 `preserve:"true"` 标签字段的原始值
// 函数名：CheckAndSetDefaultWithPreserveTag
// 参数：i interface{} — 结构体或其指针，支持多层指针；顶层为nil指针时直接返回
// 返回值：error — 当默认值解析失败、类型溢出或不支持的类型时返回错误
// 异常：不触发panic，本函数所有失败均以error返回
// 使用说明：
// - 为需要保留用户显式设置零值（如 bool=false）的字段添加 `preserve:"true"` 标签
// - 本函数先快照这些字段的原始值，再调用 CheckAndSetDefault 设置默认值，最后恢复快照值
// 使用示例：
//
//	type Config struct {
//	    EnableFeature bool `json:"enable_feature" default:"true" preserve:"true"`
//	}
//	cfg := &Config{}
//	_ = helper.CheckAndSetDefaultWithPreserveTag(cfg)
func CheckAndSetDefaultWithPreserveTag(i interface{}) error {
	v := reflect.ValueOf(i)
	if !v.IsValid() {
		return nil
	}

	// 解引用到结构体
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	// 快照需要保留的字段值（仅保留可判定为“显式提供”的值）
	snapshots := make(map[int]interface{})
	for idx := 0; idx < v.NumField(); idx++ {
		f := v.Field(idx)
		sf := t.Field(idx)
		if !f.CanSet() {
			continue
		}
		if sf.Tag.Get("preserve") == "true" {
			defTag, hasDef := sf.Tag.Lookup("default")
			if !hasDef {
				continue
			}
			defVal, ok := buildDefaultValueForType(f.Type(), defTag)
			if !ok {
				continue
			}

			if !equalToDefault(f, defVal) {
				snapshots[idx] = f.Interface()
			}
		}
	}

	// 设置默认值
	if err := CheckAndSetDefault(i); err != nil {
		return err
	}

	// 恢复快照值
	for idx, val := range snapshots {
		f := v.Field(idx)
		if !f.CanSet() {
			continue
		}
		if val == nil {
			f.Set(reflect.Zero(f.Type()))
			continue
		}
		rv := reflect.ValueOf(val)
		if rv.IsValid() && rv.Type().AssignableTo(f.Type()) {
			f.Set(rv)
			continue
		}
		if rv.IsValid() && rv.Type().ConvertibleTo(f.Type()) {
			f.Set(rv.Convert(f.Type()))
		}
	}

	return nil
}

// buildDefaultValueForType 根据字段类型与 `default` 标签构造用于比较的默认值
// 函数名：buildDefaultValueForType
// 参数：
// - typ reflect.Type — 字段的类型（可能为基本类型、指针、结构体等）
// - tag string — `default` 标签文本
// 返回值：
// - reflect.Value — 构造出的默认值（类型与字段匹配；指针字段返回指针）
// - bool — 构造是否成功（当类型不支持或解析失败时返回 false）
// 异常：不触发 panic；内部依赖的解析失败会通过返回 false 表示
// 使用示例：
//
//	def, ok := buildDefaultValueForType(field.Type(), sf.Tag.Get("default"))
//	if ok && !equalToDefault(field, def) { /* 记录快照 */ }
func buildDefaultValueForType(typ reflect.Type, tag string) (reflect.Value, bool) {
	if tag == "" {
		return reflect.Value{}, false
	}
	switch typ.Kind() {
	case reflect.Ptr:
		elem := typ.Elem()
		if elem.Kind() == reflect.Struct && elem != reflect.TypeOf(time.Time{}) {
			nv := reflect.New(elem)
			_ = CheckAndSetDefault(nv.Interface())
			return nv, true
		}
		nv := reflect.New(elem)
		if err := setDefaultValue(nv.Elem(), tag); err != nil {
			return reflect.Value{}, false
		}
		return nv, true
	case reflect.Struct:
		if typ == reflect.TypeOf(time.Time{}) {
			dv := reflect.New(typ).Elem()
			if err := setDefaultValue(dv, tag); err != nil {
				return reflect.Value{}, false
			}
			return dv, true
		}
		return reflect.Value{}, false
	default:
		dv := reflect.New(typ).Elem()
		if err := setDefaultValue(dv, tag); err != nil {
			return reflect.Value{}, false
		}
		return dv, true
	}
}

// equalToDefault 判断字段当前值是否等于构造出的默认值
// 函数名：equalToDefault
// 参数：
// - f reflect.Value — 字段当前值
// - def reflect.Value — 默认值（类型需与字段匹配；指针字段默认值应为指针）
// 返回值：
// - bool — 两值是否相等
// 异常：不触发 panic
// 使用示例：
//
//	if def.IsValid() && !equalToDefault(field, def) { /* 记录快照 */ }
func equalToDefault(f reflect.Value, def reflect.Value) bool {
	if !def.IsValid() {
		return false
	}
	switch f.Kind() {
	case reflect.Ptr:
		if f.IsNil() {
			return def.IsNil()
		}
		if def.IsNil() {
			return false
		}
		return reflect.DeepEqual(f.Elem().Interface(), def.Elem().Interface())
	default:
		return reflect.DeepEqual(f.Interface(), def.Interface())
	}
}

// setDefaultValue 根据字段类型解析并设置默认值
// 函数名：setDefaultValue
// 参数：
// - field reflect.Value — 可设置的字段值，可能为基本类型、切片、映射、结构体（time.Time）
// - tag string — `default` 标签文本，支持数字/布尔/字符串/JSON等表示
// 返回值：error — 当解析失败或发生溢出时返回错误
// 异常：不触发panic
// 说明：
// - 整数/无符号/浮点类型进行位宽溢出检查
// - time.Duration 支持形如 "300ms" 的时长字符串
// - 切片：支持 JSON 数组或逗号分隔；空标签初始化为空切片
// - 映射：仅支持 string 键；值类型按目标元素类型转换；空标签初始化空映射
// - time.Time：依赖 Convert.ToTime 支持多格式时间字符串
func setDefaultValue(field reflect.Value, tag string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(tag)
		return nil
	case reflect.Bool:
		if tag == "" {
			return nil
		}
		b, err := strconv.ParseBool(tag)
		if err != nil {
			return err
		}
		field.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			if tag == "" {
				return nil
			}
			d, err := time.ParseDuration(tag)
			if err != nil {
				return err
			}
			field.SetInt(int64(d))
			return nil
		}
		if tag == "" {
			return nil
		}
		x, err := strconv.ParseInt(tag, 10, 64)
		if err != nil {
			return err
		}
		if field.OverflowInt(x) {
			return fmt.Errorf("default value overflows %s", field.Type().String())
		}
		field.SetInt(x)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if tag == "" {
			return nil
		}
		x, err := strconv.ParseUint(tag, 10, 64)
		if err != nil {
			return err
		}
		if field.OverflowUint(x) {
			return fmt.Errorf("default value overflows %s", field.Type().String())
		}
		field.SetUint(x)
		return nil
	case reflect.Float32, reflect.Float64:
		if tag == "" {
			return nil
		}
		x, err := strconv.ParseFloat(tag, 64)
		if err != nil {
			return err
		}
		if field.OverflowFloat(x) {
			return fmt.Errorf("default value overflows %s", field.Type().String())
		}
		field.SetFloat(x)
		return nil
	case reflect.Slice:
		if tag == "" {
			field.Set(reflect.MakeSlice(field.Type(), 0, 0))
			return nil
		}
		arr := Convert{Value: tag}.ToSlice()
		if arr == nil {
			parts := strings.Split(tag, ",")
			tmp := make([]interface{}, 0, len(parts))
			for _, p := range parts {
				s := strings.TrimSpace(p)
				if s != "" {
					tmp = append(tmp, s)
				}
			}
			arr = tmp
		}
		if arr == nil {
			return fmt.Errorf("invalid slice default for %s", field.Type().String())
		}
		elemType := field.Type().Elem()
		newSlice := reflect.MakeSlice(field.Type(), 0, len(arr))
		for _, it := range arr {
			ev, ok := Convert{Value: it}.ToReflectValue(elemType)
			if !ok {
				return fmt.Errorf("invalid slice element for %s", field.Type().String())
			}
			newSlice = reflect.Append(newSlice, ev)
		}
		field.Set(newSlice)
		return nil
	case reflect.Map:
		if tag == "" {
			field.Set(reflect.MakeMap(field.Type()))
			return nil
		}
		if field.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("map key must be string for defaults: %s", field.Type().String())
		}
		m := Convert{Value: tag}.ToMap()
		if m == nil {
			return fmt.Errorf("invalid map default for %s", field.Type().String())
		}
		newMap := reflect.MakeMap(field.Type())
		valType := field.Type().Elem()
		for k, v := range m {
			ev, ok := Convert{Value: v}.ToReflectValue(valType)
			if !ok {
				return fmt.Errorf("invalid map value for %s", field.Type().String())
			}
			newMap.SetMapIndex(reflect.ValueOf(k), ev)
		}
		field.Set(newMap)
		return nil
	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			if tag == "" {
				return nil
			}
			t := Convert{Value: tag}.ToTime()
			field.Set(reflect.ValueOf(t))
			return nil
		}
		return nil
	default:
		return nil
	}
}

// CompareNumber 比较两个值，如果 a < b 返回 -1，如果 a == b 返回 0，如果 a > b 返回 1，如果错误则 panic
func CompareNumber(a, b interface{}) int {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	vaK := va.Kind()
	vbK := vb.Kind()

	if vaK != vbK {
		log.Panic(errors.New("比较的值应当是同一种类型"))
	}

	// 如果是字符串，就转换为数字
	if vaK == reflect.String {
		ai, ok := Convert{Value: va}.ToNumber()
		if !ok {
			log.Panic(errors.New("字符串转数值失败"))
		}
		bi, ok := Convert{Value: vb}.ToNumber()
		if !ok {
			log.Panic(errors.New("字符串转数值失败"))
		}
		va = reflect.ValueOf(ai)
		vb = reflect.ValueOf(bi)
	}

	switch va.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ai, bi := va.Int(), vb.Int()
		switch {
		case ai < bi:
			return -1
		case ai > bi:
			return 1
		default:
			return 0
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		au, bu := va.Uint(), vb.Uint()
		switch {
		case au < bu:
			return -1
		case au > bu:
			return 1
		default:
			return 0
		}
	case reflect.Float32, reflect.Float64:
		af, bf := va.Float(), vb.Float()
		switch {
		case af < bf:
			return -1
		case af > bf:
			return 1
		default:
			return 0
		}
	default:
		panic(errors.New("不支持的比较类型"))
	}
}

// Max 返回可变参数中最大的值
func Max(numbers ...interface{}) interface{} {
	if len(numbers) == 0 {
		panic("未提供任何数字")
	}
	maxValue := numbers[0]
	for _, num := range numbers[1:] {
		if CompareNumber(maxValue, num) < 0 {
			maxValue = num
		}
	}
	return maxValue
}

// Min 返回可变参数中最小的值
func Min(numbers ...interface{}) interface{} {
	if len(numbers) == 0 {
		panic("未提供任何数字")
	}
	minValue := numbers[0]
	for _, num := range numbers[1:] {
		if CompareNumber(minValue, num) > 0 {
			minValue = num
		}
	}
	return minValue
}

// TraceCaller 打印当前函数及其调用者的信息。
//
// 该函数会获取当前执行的函数名称和调用该函数的位置（文件和行号）。
// 如果无法获取当前函数或调用者的信息，将会打印错误信息并返回。
//
// 注意：
// - `runtime.Caller(1)` 获取的是TraceCaller的调用者信息。
// - `runtime.Caller(2)` 获取的是TraceCaller的调用者的调用者信息，即调用链的上一层。
//
// 输出格式为：
// Function: <当前函数名称> was called from <调用者函数名称>, file: <调用者文件路径>, line: <调用者行号>
func TraceCaller() {
	// 获取当前方法信息
	pcCurrent, _, _, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("Unable to get current function info")
		return
	}

	fnCurrent := runtime.FuncForPC(pcCurrent)

	// 获取调用者的信息
	pcCaller, file, line, ok := runtime.Caller(2)
	if !ok {
		fmt.Println("Unable to get caller info")
		return
	}

	fnCaller := runtime.FuncForPC(pcCaller)

	fmt.Printf("Function: %s was called from %s, file: %s, line: %d\n", fnCurrent.Name(), fnCaller.Name(), file, line)
}

// FindAvailablePort 查找可用端口
// 从指定端口开始，如果端口被占用则递增端口号，直到找到可用端口
// 未指定端口的情况下默认使用8080
func FindAvailablePort(startPort string) string {
	port, err := strconv.Atoi(startPort)
	if err != nil {
		log.Printf("无效的端口[:%s]，使用默认端口[:8080]", startPort)
		port = 8080
	}

	for {
		// 尝试监听端口
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			// 端口可用，关闭监听器并返回端口号
			err = listener.Close()
			if err != nil {
				log.Printf("关闭监听器失败: %v", err)
			}
			strPort := strconv.Itoa(port)
			log.Printf("确认端口[:%s]可用\n", strPort)
			return strPort
		}

		// 端口被占用，递增端口号
		log.Printf("端口 %d 已被占用，尝试端口 %d", port, port+1)
		port++
	}
}

// BuildYii2RedisCacheKey
// 生成一个规范化的缓存键，支持可选的前缀参数。
// 逻辑同yii2的yii\caching\Cache::buildKey
//
// 参数说明：
// - key: 需要规范化的缓存键字符串。
// - args: 可选参数，args[0] 为缓存键的前缀字符串。
//
// 处理逻辑：
//  1. 如果 key 只包含字母和数字，且长度不超过 32 个字符，
//     则直接返回前缀加 key。
//  2. 否则对 key 进行 MD5 哈希处理，
//     返回前缀加哈希字符串，保证缓存键长度和格式统一。
//
// 该方法适用于缓存键的标准化处理，避免因 key 格式差异导致缓存失效。
func BuildYii2RedisCacheKey(key string, args ...string) string {
	keyPrefix := ""
	if len(args) > 0 && args[0] != "" {
		keyPrefix = args[0]
	}

	// 判断 key 是否只包含字母和数字，且长度不超过 32
	isAlNum := true
	for _, r := range key {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			isAlNum = false
			break
		}
	}

	if isAlNum && len(key) <= 32 {
		return keyPrefix + key
	}

	// 否则对 key 进行 MD5 哈希处理，并返回带前缀的哈希值
	hash := md5.Sum([]byte(key))
	return keyPrefix + hex.EncodeToString(hash[:])
}

// CompareVersion 版本号比较（支持x.y.z，降级为数值分段比较，不足段补0；非法段按0处理）
// 参数：
//   - a: 版本号A
//   - b: 版本号B
//
// 返回：
//   - result: 比较结果（1 表示a>b；0表示相等；-1表示a<b）
//   - level: 差异级别（从1开始，表示第几级版本号差异；0表示无差异）
func CompareVersion(a string, b string) (result int, level int) {
	if a == b {
		return
	}

	as := splitVersion(a)
	bs := splitVersion(b)

	// 对齐长度到最大段数
	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}

	for i := 0; i < n; i++ {
		ai := 0
		bi := 0
		if i < len(as) {
			ai, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			bi, _ = strconv.Atoi(bs[i])
		}

		if ai != bi {
			level = i + 1

			if ai > bi {
				result = 1
			} else {
				result = -1
			}

			return
		}
	}

	// 如果所有段都相同，但长度不同，则认为是多出来的第一级版本号差异
	if len(as) != len(bs) {
		// 使用较短版本号长度+1作为差异级别（多出来的第一级）
		if len(as) > len(bs) {
			level = len(bs) + 1
			result = 1
		} else {
			level = len(as) + 1
			result = -1
		}

		return
	}

	return
}

// splitVersion 将版本号以点号拆分，清理空白
func splitVersion(v string) []string {
	v = strings.TrimSpace(v)
	if v == "" {
		return []string{"0"}
	}
	parts := strings.Split(v, ".")
	for i := range parts {
		p := strings.TrimSpace(parts[i])
		if p == "" {
			p = "0"
		}
		parts[i] = p
	}
	return parts
}
