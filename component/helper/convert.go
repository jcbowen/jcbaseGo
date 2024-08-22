package helper

import (
	"encoding/json"
	"log"
	"math"
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
		return 0
	}
	switch v := newValue.(type) {
	case int64:
		return float64(v)
	case uint64:
		return float64(v)
	case float64:
		return v
	default:
		return 0
	}
}
