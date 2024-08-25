package security

import (
	"github.com/jcbowen/jcbaseGo/component/helper"
	"html"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Input struct {
	Value        interface{}
	DefaultValue interface{}
}

var (
	htmlEntityRegex  = regexp.MustCompile(`&((#(\d{3,5}|x[a-fA-F0-9]{4}));)`)
	sqlRegex         = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|truncate|alter|exec|;|--)`)
	badStrRegex      = regexp.MustCompile(`\000|%00|%3C|%3E|<\?|<%|\{\?|{php|{if|{foreach|{for|\.\./`)
	controlCharRegex = regexp.MustCompile(`[\x00-\x1F\x7F]`)
	xssEventRegex    = regexp.MustCompile(`(?i)on[a-z]+\s*=\s*['"].*?['"]`)
	xssTagRegex      = regexp.MustCompile(`(?i)<(?:script|style|object|embed|link|meta|img|form|base|blink|xml|applet|bgsound|ilayer|layer|marquee|frameset|frame|iframe).*?>`)
	specialCharRegex = regexp.MustCompile(`&[a-zA-Z0-9#]{2,8};`)
	urlPattern       = regexp.MustCompile(`(href|src)=['"](.*?)['"]`)
	replacementStr   = ""
)

// Belong 检查值是否属于允许列表
func (s Input) Belong(allow []interface{}, strict bool) interface{} {
	if helper.IsEmptyValue(s.Value) {
		return s.DefaultValue
	}
	for _, v := range allow {
		if (strict && v == s.Value) || (!strict && helper.Convert{Value: v}.ToString() == helper.Convert{Value: s.Value}.ToString()) {
			return s.Value
		}
	}
	return s.DefaultValue
}

// Html 清理并返回HTML内容
func (s Input) Html() string {
	val, ok := s.getStringValue()
	if !ok || val == "" {
		if defVal, ok := s.DefaultValue.(string); ok {
			return defVal
		}
		return ""
	}

	// 先进行badStrReplace
	val = s.badStrReplace(val)

	// 然后进行XSS过滤
	val = s.removeXss(val)

	if val == "" && s.DefaultValue != "" {
		return s.DefaultValue.(string)
	}
	return val
}

// Sanitize 清理并返回值
func (s Input) Sanitize() interface{} {
	if helper.IsEmptyValue(s.Value) {
		return s.DefaultValue
	}

	valueType := reflect.TypeOf(s.Value)
	valueValue := reflect.ValueOf(s.Value)

	// 检查是否传入的是指针，并解引用
	if valueType.Kind() == reflect.Ptr {
		if valueValue.IsNil() {
			return s.DefaultValue
		}
		valueType = valueType.Elem()
		valueValue = valueValue.Elem()
	}

	switch valueType.Kind() {
	case reflect.String:
		return s.sanitizeString(valueValue.String())
	case reflect.Slice:
		return s.sanitizeSlice(valueValue)
	case reflect.Map:
		return s.sanitizeMap(valueValue)
	case reflect.Struct:
		return s.sanitizeStruct(valueValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Bool:
		return s.Value
	default:
		return s.DefaultValue
	}
}

// sanitizeString 清理字符串以防止 SQL 注入和 XSS 攻击
func (s Input) sanitizeString(str string) string {
	if str == "" {
		if defVal, ok := s.DefaultValue.(string); ok {
			return defVal
		}
		return ""
	}
	str = s.badStrReplace(str)
	str = htmlEntityRegex.ReplaceAllString(str, "&$1")
	str = sqlRegex.ReplaceAllString(str, "")
	str = html.EscapeString(str)
	return str
}

// badStrReplace 替换潜在的有害子字符串
func (s Input) badStrReplace(str string) string {
	return badStrRegex.ReplaceAllString(str, replacementStr)
}

// sanitizeSlice 清理切片
func (s Input) sanitizeSlice(sliceValue reflect.Value) []interface{} {
	sanitizedSlice := make([]interface{}, sliceValue.Len())

	for i := 0; i < sliceValue.Len(); i++ {
		sanitizedSlice[i] = Input{Value: sliceValue.Index(i).Interface(), DefaultValue: s.DefaultValue}.Sanitize()
	}

	return sanitizedSlice
}

// sanitizeMap 清理映射
func (s Input) sanitizeMap(mapValue reflect.Value) map[interface{}]interface{} {
	sanitizedMap := make(map[interface{}]interface{})

	for _, key := range mapValue.MapKeys() {
		sanitizedKey := Input{Value: key.Interface(), DefaultValue: s.DefaultValue}.Sanitize()
		sanitizedValue := Input{Value: mapValue.MapIndex(key).Interface(), DefaultValue: s.DefaultValue}.Sanitize()
		sanitizedMap[sanitizedKey] = sanitizedValue
	}

	return sanitizedMap
}

// sanitizeStruct 清理结构体
func (s Input) sanitizeStruct(value reflect.Value) interface{} {
	sanitizedStruct := reflect.New(value.Type()).Elem()

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := value.Type().Field(i)
		if field.CanInterface() {
			sanitizedField := Input{Value: field.Interface(), DefaultValue: reflect.Zero(fieldType.Type).Interface()}.Sanitize()
			sanitizedStruct.Field(i).Set(reflect.ValueOf(sanitizedField))
		}
	}

	return sanitizedStruct.Interface()
}

// removeXss 清理输入以防止 XSS 攻击
func (s Input) removeXss(val string) string {
	// 删除控制字符
	val = controlCharRegex.ReplaceAllString(val, "")

	// 处理特殊 HTML 实体字符
	val = specialCharRegex.ReplaceAllStringFunc(val, func(entity string) string {
		return html.UnescapeString(entity)
	})

	// 替换危险的事件属性和标签
	val = xssEventRegex.ReplaceAllString(val, "")
	val = xssTagRegex.ReplaceAllString(val, "")

	// 处理 URL
	matches := urlPattern.FindAllStringSubmatch(val, -1)
	var urlList []string
	for _, match := range matches {
		if match[2] != "" {
			urlList = append(urlList, match[2])
		}
	}

	var encodedUrlList []string
	for key, url := range urlList {
		placeholder := "jc_" + strconv.Itoa(key) + "_placeholder"
		val = strings.ReplaceAll(val, url, placeholder)
		encodedUrlList = append(encodedUrlList, url)
	}

	// 替换 URL 中的占位符
	for key, url := range encodedUrlList {
		placeholder := "jc_" + strconv.Itoa(key) + "_placeholder"
		val = strings.ReplaceAll(val, placeholder, url)
	}

	return val
}

// getStringValue 获取字符串值，如果传入的是指针，解引用
func (s Input) getStringValue() (string, bool) {
	valueType := reflect.TypeOf(s.Value)
	valueValue := reflect.ValueOf(s.Value)

	// 检查是否传入的是指针
	if valueType.Kind() == reflect.Ptr {
		if valueValue.IsNil() {
			return "", false
		}
		valueType = valueType.Elem()
		valueValue = valueValue.Elem()
	}

	if valueType.Kind() != reflect.String {
		return "", false
	}

	return valueValue.String(), true
}
