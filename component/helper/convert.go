package helper

import (
	"encoding/json"
	"log"
	"math"
	"os"
	"reflect"
	"strconv"
	"time"
)

type Convert struct {
	Value interface{}
}

// ToString 将变量转为字符串
// 浮点型 3.0将会转换成字符串3, "3"
// 非数值或字符类型的变量将会被转换成JSON格式字符串
func (c Convert) ToString() string {
	if c.Value == nil {
		return ""
	}

	switch v := c.Value.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case int8, int16, int32:
		return strconv.Itoa(int(reflect.ValueOf(v).Int()))
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8, uint16, uint32:
		return strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case time.Time:
		return v.Format(time.DateTime)
	case []byte:
		return string(v)
	default:
		// 将值序列化为 JSON
		newValue, err := json.Marshal(c.Value)
		if err != nil {
			log.Println("Error marshaling value to JSON:", err)
			return ""
		}
		return string(newValue)
	}
}

// ToBool 将变量转为bool类型
func (c Convert) ToBool() bool {
	if c.Value == nil {
		return false
	}

	switch v := c.Value.(type) {
	case bool:
		return v
	case string:
		if v == "" {
			return false
		}
		b, err := strconv.ParseBool(v)
		if err != nil {
			log.Println("Error parsing bool from string:", err)
		}
		return err == nil && b
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() > 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() > 0
	case float32, float64:
		return reflect.ValueOf(v).Float() > 0
	default:
		return false
	}
}

// ToFileMode 将变量转为os.FileMode类型
func (c Convert) ToFileMode() os.FileMode {
	if c.Value == nil {
		return 0
	}

	switch v := c.Value.(type) {
	case os.FileMode:
		return v
	case string:
		m, err := strconv.ParseUint(v, 8, 32)
		if err != nil {
			log.Println("Error parsing FileMode from string:", err)
			return 0
		}
		return os.FileMode(m)
	case int, int8, int16, int32, int64:
		return os.FileMode(reflect.ValueOf(v).Int())
	case uint, uint8, uint16, uint32, uint64:
		return os.FileMode(reflect.ValueOf(v).Uint())
	default:
		return 0
	}
}

// ToArrByte 将变量转为[]byte类型
func (c Convert) ToArrByte() []byte {
	if c.Value == nil {
		return nil
	}

	switch v := c.Value.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	default:
		return nil
	}
}

// ToInt 将变量转为int类型
func (c Convert) ToInt() int {
	if c.Value == nil {
		return 0
	}

	switch v := c.Value.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			log.Println("Error parsing int from string:", err)
			return 0
		}
		return i
	default:
		return 0
	}
}

// ToInt64 将变量转为int64类型
func (c Convert) ToInt64() int64 {
	if c.Value == nil {
		return 0
	}

	switch v := c.Value.(type) {
	case int64:
		return v
	case int, int8, int16, int32:
		return int64(reflect.ValueOf(v).Int())
	case uint, uint8, uint16, uint32, uint64:
		u := reflect.ValueOf(v).Uint()
		if u > uint64(math.MaxInt64) {
			return 0
		}
		return int64(u)
	case float32, float64:
		f := reflect.ValueOf(v).Float()
		if f > float64(math.MaxInt64) || f < float64(math.MinInt64) {
			return 0
		}
		return int64(f)
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Println("Error parsing int64 from string:", err)
			return 0
		}
		return i
	default:
		return 0
	}
}

// ToInt8 将变量转为int8类型
func (c Convert) ToInt8() int8 {
	if c.Value == nil {
		return 0
	}

	switch v := c.Value.(type) {
	case int8:
		return v
	case int, int16, int32, int64:
		i := reflect.ValueOf(v).Int()
		if i < int64(math.MinInt8) || i > int64(math.MaxInt8) {
			return 0
		}
		return int8(i)
	case uint, uint8, uint16, uint32, uint64:
		u := reflect.ValueOf(v).Uint()
		if u > uint64(math.MaxInt8) {
			return 0
		}
		return int8(u)
	case float32, float64:
		f := reflect.ValueOf(v).Float()
		if f > float64(math.MaxInt8) || f < float64(math.MinInt8) {
			return 0
		}
		return int8(f)
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Println("Error parsing int8 from string:", err)
			return 0
		}
		if i < int64(math.MinInt8) || i > int64(math.MaxInt8) {
			return 0
		}
		return int8(i)
	default:
		return 0
	}
}

// ToUint 将变量转为uint类型
func (c Convert) ToUint() uint {
	newValue, ok := c.ToNumber()
	if !ok {
		return 0
	}
	switch v := newValue.(type) {
	case int64:
		if v < 0 {
			return 0
		}
		return uint(v)
	case uint64:
		return uint(v)
	case float64:
		if v < 0 {
			return 0
		}
		return uint(v)
	default:
		return 0
	}
}

// ToUint64 将变量转为uint64类型
func (c Convert) ToUint64() uint64 {
	if c.Value == nil {
		return 0
	}

	switch v := c.Value.(type) {
	case uint64:
		return v
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(v).Int()
		if i < 0 {
			return 0
		}
		return uint64(i)
	case uint, uint8, uint16, uint32:
		return uint64(reflect.ValueOf(v).Uint())
	case float32, float64:
		f := reflect.ValueOf(v).Float()
		if f < 0 {
			return 0
		}
		return uint64(f)
	case string:
		u, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			log.Println("Error parsing uint64 from string:", err)
			return 0
		}
		return u
	default:
		return 0
	}
}

// ToNumber 将字符串变量转为数字类型
func (c Convert) ToNumber() (interface{}, bool) {
	if c.Value == nil {
		return int(0), false
	}

	switch v := c.Value.(type) {
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i, true
		}
		if u, err := strconv.ParseUint(v, 10, 64); err == nil {
			return u, true
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int(), true
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint(), true
	case float32, float64:
		return reflect.ValueOf(v).Float(), true
	}
	return int(0), false
}

// ToFloat64 将变量转为float64
func (c Convert) ToFloat64() float64 {
	newValue, ok := c.ToNumber()
	if !ok {
		return 0.0
	}
	switch v := newValue.(type) {
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case uint64:
		return float64(v)
	case float64:
		return v
	case float32:
		return float64(v)
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			log.Println("Error parsing float64 from string:", err)
			return 0.0
		}
		return f
	default:
		return 0.0
	}
}

// ToTime 将变量转为time.Time类型
func (c Convert) ToTime() time.Time {
	if c.Value == nil {
		return time.Time{}
	}

	switch v := c.Value.(type) {
	case time.Time:
		return v
	case string:
		// 尝试常见的时间格式
		formats := []string{
			// RFC 标准格式
			time.RFC3339,
			time.RFC3339Nano,
			time.RFC1123,
			time.RFC1123Z,
			time.RFC822,
			time.RFC822Z,
			time.RFC850,

			// 数据库常用格式
			"2006-01-02 15:04:05",
			"2006-01-02 15:04:05.999999999",
			"2006-01-02 15:04:05.999999",
			"2006-01-02 15:04:05.999",
			"2006-01-02 15:04:05.99",
			"2006-01-02 15:04:05.9",

			// 日期格式
			"2006-01-02",
			"2006/01/02",
			"2006.01.02",
			"20060102",

			// 带时区的格式
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02T15:04:05.999999999Z07:00",
			"2006-01-02 15:04:05 Z07:00",
			"2006-01-02 15:04:05 MST",

			// 时间格式
			"15:04:05",
			"15:04",

			// 斜杠分隔格式
			"2006/01/02 15:04:05",
			"2006/01/02 15:04:05.999",
			"2006/01/02 15:04",

			// 点号分隔格式
			"2006.01.02 15:04:05",
			"2006.01.02 15:04",

			// 中文格式
			"2006年01月02日 15:04:05",
			"2006年01月02日",

			// 日志常用格式
			"Jan _2 15:04:05",
			"Jan _2 15:04:05.000",
			"Jan _2 15:04:05.000000",
			"02/Jan/2006:15:04:05 -0700",

			// 美国格式
			"01/02/2006 15:04:05",
			"01/02/2006 3:04:05 PM",
			"01/02/2006",

			// 12小时制格式
			"2006-01-02 3:04:05 PM",
			"2006-01-02 3:04 PM",
			"2006/01/02 3:04:05 PM",
		}

		// 首先尝试解析为时间戳（毫秒、秒、纳秒）
		if timestamp, err := strconv.ParseInt(v, 10, 64); err == nil {
			// 判断时间戳长度来确定单位
			if len(v) == 13 { // 毫秒时间戳
				return time.Unix(timestamp/1000, (timestamp%1000)*1e6).UTC()
			} else if len(v) == 10 { // 秒时间戳
				return time.Unix(timestamp, 0).UTC()
			} else if len(v) == 19 { // 纳秒时间戳
				return time.Unix(0, timestamp).UTC()
			} else {
				// 默认按秒处理
				return time.Unix(timestamp, 0).UTC()
			}
		}

		// 尝试各种时间格式
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t
			}
		}

		// 尝试使用time.ParseInLocation（本地时间解析）
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", v, time.Local); err == nil {
			return t
		}
		if t, err := time.ParseInLocation("2006-01-02", v, time.Local); err == nil {
			return t
		}

	case int64:
		return time.Unix(v, 0).UTC()
	case int:
		return time.Unix(int64(v), 0).UTC()
	case float64:
		// 支持浮点数时间戳（可能包含小数部分）
		seconds := int64(v)
		nanoseconds := int64((v - float64(seconds)) * 1e9)
		return time.Unix(seconds, nanoseconds).UTC()
	case float32:
		seconds := int64(v)
		nanoseconds := int64((float64(v) - float64(seconds)) * 1e9)
		return time.Unix(seconds, nanoseconds).UTC()
	}
	return time.Time{}
}

// ToMap 将变量转为map[string]interface{}类型
// 支持从JSON字符串、结构体、map等类型转换
func (c Convert) ToMap() map[string]interface{} {
	if c.Value == nil {
		return nil
	}

	switch v := c.Value.(type) {
	case map[string]interface{}:
		return v
	case string:
		// 尝试解析JSON字符串
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(v), &result); err == nil {
			return result
		}
		// 如果不是有效的JSON，返回空map
		return nil
	default:
		// 尝试通过反射将结构体转换为map
		val := reflect.ValueOf(c.Value)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() == reflect.Struct {
			result := make(map[string]interface{})
			valType := val.Type()

			for i := 0; i < valType.NumField(); i++ {
				field := valType.Field(i)
				fieldValue := val.Field(i)

				// 跳过不可导出的字段
				if !fieldValue.CanInterface() {
					continue
				}

				// 使用JSON标签作为key，如果没有则使用字段名
				fieldName := field.Name
				if jsonTag := field.Tag.Get("json"); jsonTag != "" {
					fieldName = jsonTag
				}

				result[fieldName] = fieldValue.Interface()
			}
			return result
		}

		// 其他类型无法转换为map
		return nil
	}
}

// ToMapString 将变量转为map[string]string类型
// 支持从JSON字符串、结构体、map等类型转换，所有值都会被转换为字符串
func (c Convert) ToMapString() map[string]string {
	if c.Value == nil {
		return nil
	}

	switch v := c.Value.(type) {
	case map[string]string:
		return v
	case map[string]interface{}:
		result := make(map[string]string)
		for key, value := range v {
			result[key] = Convert{Value: value}.ToString()
		}
		return result
	case string:
		// 尝试解析JSON字符串
		var temp map[string]interface{}
		if err := json.Unmarshal([]byte(v), &temp); err == nil {
			result := make(map[string]string)
			for key, value := range temp {
				result[key] = Convert{Value: value}.ToString()
			}
			return result
		}
		// 如果不是有效的JSON，返回空map
		return make(map[string]string)
	default:
		// 尝试通过反射将结构体转换为map[string]string
		val := reflect.ValueOf(c.Value)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() == reflect.Struct {
			result := make(map[string]string)
			valType := val.Type()

			for i := 0; i < valType.NumField(); i++ {
				field := valType.Field(i)
				fieldValue := val.Field(i)

				// 跳过不可导出的字段
				if !fieldValue.CanInterface() {
					continue
				}

				// 使用JSON标签作为key，如果没有则使用字段名
				fieldName := field.Name
				if jsonTag := field.Tag.Get("json"); jsonTag != "" {
					fieldName = jsonTag
				}

				result[fieldName] = Convert{Value: fieldValue.Interface()}.ToString()
			}
			return result
		}

		// 其他类型无法转换为map，返回空map
		return make(map[string]string)
	}
}

// ToSlice 将变量转为[]interface{}类型
// 支持从数组、切片、JSON字符串等类型转换
func (c Convert) ToSlice() []interface{} {
	if c.Value == nil {
		return nil
	}

	switch v := c.Value.(type) {
	case []interface{}:
		return v
	case string:
		// 尝试解析JSON数组字符串
		var result []interface{}
		if err := json.Unmarshal([]byte(v), &result); err == nil {
			return result
		}
		// 如果不是有效的JSON数组，返回nil
		return nil
	default:
		// 通过反射处理数组和切片
		val := reflect.ValueOf(c.Value)
		if val.Kind() != reflect.Array && val.Kind() != reflect.Slice {
			return nil
		}

		result := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			result[i] = val.Index(i).Interface()
		}
		return result
	}
}

// ToDuration 将变量转为time.Duration类型
// 支持从字符串、整数、浮点数等类型转换
func (c Convert) ToDuration() time.Duration {
	if c.Value == nil {
		return 0
	}

	switch v := c.Value.(type) {
	case time.Duration:
		return v
	case string:
		// 尝试解析为duration字符串
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}

		// 尝试解析为秒数
		if seconds, err := strconv.ParseFloat(v, 64); err == nil {
			return time.Duration(seconds * float64(time.Second))
		}

		return 0
	case int, int8, int16, int32, int64:
		return time.Duration(reflect.ValueOf(v).Int()) * time.Second
	case uint, uint8, uint16, uint32, uint64:
		return time.Duration(reflect.ValueOf(v).Uint()) * time.Second
	case float32, float64:
		return time.Duration(reflect.ValueOf(v).Float() * float64(time.Second))
	default:
		return 0
	}
}

// ToInterface 将变量转为interface{}类型，主要用于类型断言前的准备
// 此方法主要用于保持类型一致性，实际返回原始值
func (c Convert) ToInterface() interface{} {
	return c.Value
}
