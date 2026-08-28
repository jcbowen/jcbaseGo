package helper

import (
	"strings"
)

// MaskString 通用字符串脱敏。
// 保留头部和尾部指定 Unicode 字符数，中间用 maskChar 替换。
// 若字符串长度不足，则保留首尾各一位，中间替换；若长度小于等于 2，则直接返回原字符串。
//
// 参数:
//   - s: 原始字符串
//   - head: 头部保留字符数
//   - tail: 尾部保留字符数
//   - maskChar: 替换字符
//
// 返回值:
//   - string: 脱敏后的字符串
func MaskString(s string, head, tail int, maskChar rune) string {
	runes := []rune(s)
	length := len(runes)
	if length <= 2 || head < 0 || tail < 0 {
		return s
	}
	if head+tail >= length {
		head = 1
		tail = 1
	}
	maskLen := length - head - tail
	if maskLen <= 0 {
		return s
	}
	var b strings.Builder
	b.Grow(length)
	b.WriteString(string(runes[:head]))
	b.WriteString(strings.Repeat(string(maskChar), maskLen))
	b.WriteString(string(runes[length-tail:]))
	return b.String()
}

// MaskPhone 对手机号进行脱敏处理。
// 规则：保留前 3 位和后 4 位，中间 4 位用 * 替代，例如 138****1234。
// 若手机号长度不足 7 位，则保留首尾各一位，中间替换。
//
// 参数:
//   - phone: 原始手机号
//
// 返回值:
//   - string: 脱敏后的手机号
func MaskPhone(phone string) string {
	return MaskString(phone, 3, 4, '*')
}

// MaskIdCard 对身份证号进行脱敏处理。
// 规则：保留前 3 位和后 4 位，中间用 * 替代，例如 110***********1234。
// 若身份证号长度不足 7 位，则保留首尾各一位，中间替换。
//
// 参数:
//   - idCard: 原始身份证号
//
// 返回值:
//   - string: 脱敏后的身份证号
func MaskIdCard(idCard string) string {
	return MaskString(idCard, 3, 4, '*')
}
