package helper

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type MoneyHelper struct {
	Amount int64 `json:"amount" default:"0"` // 单位为厘，1000为1元
	err    error `json:"-"`                  // 存储链式调用中的错误，不进行 JSON 序列化
}

func Money(input interface{}) *MoneyHelper {
	return (&MoneyHelper{}).SetAmount(input)
}

// SetAmount 设置金额，接受不同格式的输入
func (m *MoneyHelper) SetAmount(value interface{}) *MoneyHelper {
	if m.err != nil {
		return m
	}
	switch v := value.(type) {
	case int64:
		m.Amount = v
	case float64:
		m.Amount = int64(v * 1000)
	case string:
		amountInYuan, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			m.err = errors.New("invalid amount")
			return m
		}
		m.Amount = int64(amountInYuan * 1000)
	default:
		m.err = errors.New("unsupported type")
	}
	return m
}

// Add 加法操作
func (m *MoneyHelper) Add(other *MoneyHelper) *MoneyHelper {
	if m.err != nil {
		return m
	}
	m.Amount += other.Amount
	return m
}

// Subtract 减法操作
func (m *MoneyHelper) Subtract(other *MoneyHelper) *MoneyHelper {
	if m.err != nil {
		return m
	}
	if m.Amount < other.Amount {
		m.err = errors.New("insufficient amount")
		return m
	}
	m.Amount -= other.Amount
	return m
}

// Multiply 乘法操作
func (m *MoneyHelper) Multiply(factor float64) *MoneyHelper {
	if m.err != nil {
		return m
	}
	m.Amount = int64(float64(m.Amount) * factor)
	return m
}

// Divide 除法操作
func (m *MoneyHelper) Divide(divisor float64) *MoneyHelper {
	if m.err != nil {
		return m
	}
	if divisor == 0 {
		m.err = errors.New("division by zero")
		return m
	}
	m.Amount = int64(float64(m.Amount) / divisor)
	return m
}

// FloatString 输出字符串金额，接受前缀和后缀
// 参数:
//
//	parts - 可变参数列表，第一个参数表示前缀，第二个参数表示后缀，如果省略则不添加
func (m *MoneyHelper) FloatString(parts ...string) string {
	prefix := ""
	suffix := ""
	if len(parts) > 0 {
		prefix = parts[0]
	}
	if len(parts) > 1 {
		suffix = parts[1]
	}
	amountInYuan := float64(m.Amount) / 1000.0
	return fmt.Sprintf("%s%.2f%s", prefix, amountInYuan, suffix)
}

// GreaterThan 比较大小
func (m *MoneyHelper) GreaterThan(other *MoneyHelper) bool {
	if m.err != nil {
		return false
	}
	return m.Amount > other.Amount
}

// LessThan 比较大小
func (m *MoneyHelper) LessThan(other *MoneyHelper) bool {
	if m.err != nil {
		return false
	}
	return m.Amount < other.Amount
}

// Equals 比较相等
func (m *MoneyHelper) Equals(other *MoneyHelper) bool {
	if m.err != nil {
		return false
	}
	return m.Amount == other.Amount
}

// GetError 返回当前的错误
func (m *MoneyHelper) GetError() error {
	return m.err
}
