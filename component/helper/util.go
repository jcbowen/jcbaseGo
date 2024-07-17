package helper

import (
	"encoding/base64"
	"errors"
	"log"
	"math/rand"
	"net"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
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

// SetMapStrInterface
// Deprecated: 请使用 NewMap
func SetMapStrInterface(data map[string]interface{}) *MapHelper {
	return NewMap(data)
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
	for _, v := range a.Arr {
		value = append(value, v)
	}
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

func InArray(val interface{}, array interface{}) (exists bool) {
	exists = false
	//index = -1
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				//index = i
				exists = true
				return
			}
		}
	}
	return
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
func MapToStruct(mapData interface{}, obj interface{}) {
	objValue := reflect.ValueOf(obj).Elem()

	for key, value := range mapData.(map[string]interface{}) {
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
		MapToStruct(value.(map[string]interface{}), field.Addr().Interface()) // 传递值的指针
		return true
	}

	return false
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

// GetHostInfo 从http.Request中获取hostInfo
func GetHostInfo(req *http.Request) string {
	hostInfo := ""

	// 判断是http还是https
	if req.TLS != nil || req.Header.Get("X-Scheme") == "https" {
		hostInfo = "https://"
	} else {
		hostInfo = "http://"
	}

	// 获取host
	if req.Header.Get("X-Forwarded-Host") != "" {
		hostInfo += req.Header.Get("X-Forwarded-Host")
	} else if req.Header.Get("X-Original-Host") != "" {
		hostInfo += req.Header.Get("X-Original-Host")
	} else if req.Header.Get("X-Host") != "" {
		hostInfo += req.Header.Get("X-Host")
	} else if req.Host != "" {
		hostInfo += req.Host
	} else {
		hostInfo += req.URL.Host
	}

	// 补充端口号
	if req.URL.Port() != "" {
		hostInfo += ":" + req.URL.Port()
	}

	// 判断hostInfo中是否有80或者443的端口号，如果有，应当移除
	if strings.Contains(hostInfo, ":80") || strings.Contains(hostInfo, ":443") {
		hostInfo = strings.ReplaceAll(hostInfo, ":80", "")
		hostInfo = strings.ReplaceAll(hostInfo, ":443", "")
	}

	return hostInfo
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
func Random(length int, numericOnly bool) string {
	var charset string
	if numericOnly {
		charset = "0123456789"
	} else {
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}

	rand.Seed(time.Now().UnixNano())
	randomString := make([]byte, length)
	for i := range randomString {
		randomString[i] = charset[rand.Intn(len(charset))]
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

// CheckAndSetDefault 检查结构体中的字段是否为空，如果为空则设置为默认值
func CheckAndSetDefault(i interface{}) error {
	// 获取结构体反射值
	val := reflect.ValueOf(i)

	// 如果传入的是指针类型，获取指向的结构体
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 不是结构体的时候直接跳过处理
	if val.Kind() != reflect.Struct {
		log.Printf("%s 不是结构体，直接跳过处理", val.String())
		return nil
	}

	// 遍历结构体字段
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		// 忽略非导出字段
		if !field.CanSet() {
			continue
		}

		// 如果字段是struct或interface，则递归检查
		if field.Kind() == reflect.Struct || field.Kind() == reflect.Interface {
			if err := CheckAndSetDefault(field.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		// 获取字段类型和默认值标签
		tag := fieldType.Tag.Get("default")
		fieldKind := field.Kind()

		// 如果字段为空字符串，则设置为默认值
		if fieldKind == reflect.String && field.Len() == 0 {
			field.SetString(tag)
		}

		// 如果字段是bool类型，则设置默认值
		if fieldKind == reflect.Bool && !field.Bool() {
			defaultVal := tag == "true"
			field.SetBool(defaultVal)
		}

		// 如果字段是int类型，则设置默认值
		if strings.HasPrefix(field.Type().String(), "int") && field.Int() == 0 {
			defaultVal, _ := strconv.ParseInt(tag, 10, 64)
			field.SetInt(defaultVal)
		}

		// 如果字段是float类型，则设置默认值
		if fieldKind == reflect.Float32 || fieldKind == reflect.Float64 {
			defaultVal, _ := strconv.ParseFloat(tag, 64)
			if field.Float() == 0 {
				field.SetFloat(defaultVal)
			}
		}
	}

	return nil
}

// ------------------------ 以下是弃用了的函数，将在后续版本中被移除 ------------------------ /

// Str2Int 字符串转数字
//
// Deprecated: As of jcbaseGo 0.2.1, this function simply calls ToInt.
func Str2Int(str string) int {
	return ToInt(str)
}
