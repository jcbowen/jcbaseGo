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
	"net/http"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
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

// InArray 检查某个值是否存在于切片中
// val 是要检查的值
// array 是要检查的切片
// exists 是返回的布尔值，表示 val 是否存在于 array 中
func InArray(val interface{}, array interface{}) (exists bool) {
	arr := reflect.ValueOf(array)

	// 确保 array 是一个切片
	if arr.Kind() != reflect.Slice {
		panic("第二个参数必须是一个切片")
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

// CopyStruct 复制结构体，将 src 的值复制到 dst
//
// 用法：CopyStruct(&src, &dst)
//
// 注意：src、dst 必须是一个指针
func CopyStruct(src, dst interface{}) {
	srcVal := reflect.ValueOf(src).Elem()
	dstVal := reflect.ValueOf(dst).Elem()

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		dstField := dstVal.FieldByName(srcVal.Type().Field(i).Name)

		if dstField.IsValid() && dstField.CanSet() && srcField.Type() == dstField.Type() {
			dstField.Set(srcField)
		}
	}
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
		return value.IsNil() || IsEmptyValue(value.Elem().Interface())
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

// isEmptyStruct 检查是否是一个空结构体
func isEmptyStruct(value reflect.Value) bool {
	for i := 0; i < value.NumField(); i++ {
		if !IsEmptyValue(value.Field(i).Interface()) {
			return false
		}
	}
	return true
}

// CheckAndSetDefault 检查结构体中的字段是否为空，如果为空则设置为默认值
// 常见问题：如果发现默认值赋值失败，但是又没有出现报错，可以看看是不是传递的指针的指针
func CheckAndSetDefault(i interface{}) error {
	// 获取结构体反射值
	val := reflect.ValueOf(i)

	// 如果传入的是指针类型，获取指向的结构体
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 不是结构体的时候直接跳过处理
	if val.Kind() != reflect.Struct {
		// log.Printf("%s 不是结构体，直接跳过处理", val.String())
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
