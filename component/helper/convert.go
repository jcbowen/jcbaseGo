package helper

import (
	"encoding/json"
	"os"
	"reflect"
	"strconv"
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
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case int, int8, int16, int32, uint, uint8, uint16, uint32:
		return strconv.Itoa(int(reflect.ValueOf(v).Int()))
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case string:
		return v
	case []byte:
		return string(v)
	default:
		newValue, err := json.Marshal(c.Value)
		if err != nil {
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
		b, _ := strconv.ParseBool(v)
		return b
	case int:
		return v > 0
	case int8:
		return v > 0
	case int16:
		return v > 0
	case int32:
		return v > 0
	case int64:
		return v > 0
	case uint:
		return v > 0
	case uint8:
		return v > 0
	case uint16:
		return v > 0
	case uint32:
		return v > 0
	case uint64:
		return v > 0
	case float32:
		return v > 0
	case float64:
		return v > 0
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
		m, _ := strconv.ParseUint(v, 8, 32)
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
		i, _ := strconv.Atoi(v)
		return i
	default:
		return 0
	}
}
