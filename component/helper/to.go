package helper

import (
	"encoding/json"
	"os"
	"strconv"
)

// ToString 将变量转为字符串
// 浮点型 3.0将会转换成字符串3, "3"
// 非数值或字符类型的变量将会被转换成JSON格式字符串
func ToString(value interface{}) (key string) {
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return
}

// ToBool 将变量转为bool类型
func ToBool(value interface{}) (b bool) {
	if value == nil {
		return
	}

	switch value.(type) {
	case bool:
		b = value.(bool)
	case string:
		b, _ = strconv.ParseBool(value.(string))
	case int:
		b = value.(int) > 0
	case int8:
		b = value.(int8) > 0
	case int16:
		b = value.(int16) > 0
	case int32:
		b = value.(int32) > 0
	case int64:
		b = value.(int64) > 0
	case uint:
		b = value.(uint) > 0
	case uint8:
		b = value.(uint8) > 0
	case uint16:
		b = value.(uint16) > 0
	case uint32:
		b = value.(uint32) > 0
	case uint64:
		b = value.(uint64) > 0
	case float32:
		b = value.(float32) > 0
	case float64:
		b = value.(float64) > 0
	}

	return
}

// ToFileMode 将变量转为os.FileMode类型
func ToFileMode(value interface{}) (mode os.FileMode) {
	if value == nil {
		return
	}

	switch value.(type) {
	case os.FileMode:
		mode = value.(os.FileMode)
	case string:
		m, _ := strconv.ParseUint(value.(string), 8, 32)
		mode = ToFileMode(m)
	case int:
		mode = os.FileMode(value.(int))
	case int8:
		mode = os.FileMode(value.(int8))
	case int16:
		mode = os.FileMode(value.(int16))
	case int32:
		mode = os.FileMode(value.(int32))
	case int64:
		mode = os.FileMode(value.(int64))
	case uint:
		mode = os.FileMode(value.(uint))
	case uint8:
		mode = os.FileMode(value.(uint8))
	case uint16:
		mode = os.FileMode(value.(uint16))
	case uint32:
		mode = os.FileMode(value.(uint32))
	case uint64:
		mode = os.FileMode(value.(uint64))
	}

	return
}

// ToArrByte 将变量转为[]byte类型
func ToArrByte(value interface{}) (arrByte []byte) {
	if value == nil {
		return
	}

	switch value.(type) {
	case []byte:
		arrByte = value.([]byte)
	case string:
		arrByte = []byte(value.(string))
	}

	return
}

// ToInt 将变量转为int类型
func ToInt(value interface{}) (i int) {
	if value == nil {
		return
	}

	switch value.(type) {
	case int:
		i = value.(int)
	case int8:
		i = int(value.(int8))
	case int16:
		i = int(value.(int16))
	case int32:
		i = int(value.(int32))
	case int64:
		i = int(value.(int64))
	case uint:
		i = int(value.(uint))
	case uint8:
		i = int(value.(uint8))
	case uint16:
		i = int(value.(uint16))
	case uint32:
		i = int(value.(uint32))
	case uint64:
		i = int(value.(uint64))
	case float32:
		i = int(value.(float32))
	case float64:
		i = int(value.(float64))
	case string:
		i, _ = strconv.Atoi(value.(string))
	}

	return
}

// Int int转换为*int类型
func Int(i int) *int {
	return &i
}
