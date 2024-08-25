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
)

// Sanitize 自动判断类型并进行通用清理
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
		return s.SanitizeString(valueValue.String())
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

// SanitizeString 清理字符串以防止 SQL 注入和 XSS 攻击
func (s Input) SanitizeString(str string, args ...any) string {
	var (
		defVal        string
		sanitizeTypes []string
		ok            bool
	)

	// 默认清理类型为 "badStr" 和 "htmlEntity"
	if len(args) == 0 {
		sanitizeTypes = []string{"badStr", "htmlEntity"}
	} else {
		// 类型安全检查，确保传入的参数是 []string
		if sanitizeTypes, ok = args[0].([]string); !ok {
			sanitizeTypes = []string{"badStr", "htmlEntity"}
		}
	}

	// 如果字符串为空，直接返回默认值或空字符串
	if str == "" {
		if defVal, ok = s.DefaultValue.(string); ok {
			return defVal
		}
		return ""
	}

	// 根据指定的清理类型进行处理
	for _, sanitizeType := range sanitizeTypes {
		switch sanitizeType {
		case "badStr":
			str = s.badStrReplace(str)
		case "htmlEntity":
			str = htmlEntityRegex.ReplaceAllString(str, "&$1")
		case "sql":
			str = sqlRegex.ReplaceAllString(str, "")
		case "xss":
			str = s.removeXss(str)
		}
	}

	// 再次检查清理后的字符串是否为空
	if str == "" {
		if defVal, ok = s.DefaultValue.(string); ok {
			return defVal
		}
		return ""
	}

	// 最后进行 HTML 转义
	str = html.EscapeString(str)

	return str
}

// Belong 检查值是否属于允许列表
func (s Input) Belong(allow interface{}, strict bool) interface{} {
	if helper.IsEmptyValue(s.Value) {
		return s.DefaultValue
	}

	allowValue := reflect.ValueOf(allow)
	if allowValue.Kind() != reflect.Slice {
		return s.DefaultValue
	}

	for i := 0; i < allowValue.Len(); i++ {
		item := allowValue.Index(i).Interface()
		if strict {
			if reflect.DeepEqual(item, s.Value) {
				return s.Value
			}
		} else {
			if (helper.Convert{Value: item}.ToString()) == (helper.Convert{Value: s.Value}.ToString()) {
				return s.Value
			}
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

// badStrReplace 替换潜在的有害子字符串
func (s Input) badStrReplace(str string) string {
	return badStrRegex.ReplaceAllString(str, "")
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
