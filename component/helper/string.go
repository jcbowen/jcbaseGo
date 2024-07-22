package helper

import (
	"encoding/base64"
	"html"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Str struct {
	String  string
	Float64 float64
	Float32 float32
	Int64   int64
	Int32   int32
	Int16   int16
	Int8    int8
	Int     int
	Uint64  uint64
	Uint32  uint32
	Uint16  uint16
	Uint8   uint8
	Uint    uint
}

func NewStr(str string) *Str {
	return &Str{
		String: str,
	}
}

// ByteLength 返回给定字符串的字节数
func (s *Str) ByteLength() int {
	return utf8.RuneCountInString(s.String)
}

// ByteSubstr 返回从 start 开始、长度为 length 的子字符串
func (s *Str) ByteSubstr(start int, length int) string {
	runes := []rune(s.String)
	if start >= len(runes) {
		return ""
	}
	end := start + length
	if length == -1 || end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}

// Truncate 将字符串截断为指定的字符数
func (s *Str) Truncate(length int, suffix string) string {
	if utf8.RuneCountInString(s.String) > length {
		return string([]rune(s.String)[:length]) + suffix
	}
	return s.String
}

// TruncateWords 将字符串截断为指定的单词数
func (s *Str) TruncateWords(count int, suffix string) string {
	words := regexp.MustCompile(`\s+`).Split(strings.TrimSpace(s.String), -1)
	if len(words) > count {
		return strings.Join(words[:count], " ") + suffix
	}
	return s.String
}

// StartsWith 检查字符串是否以指定的子字符串开头
func (s *Str) StartsWith(prefix string, caseSensitive bool) bool {
	str := s.String
	if !caseSensitive {
		str = strings.ToLower(s.String)
		prefix = strings.ToLower(prefix)
	}
	return strings.HasPrefix(str, prefix)
}

// EndsWith 检查字符串是否以指定的子字符串结尾
func (s *Str) EndsWith(suffix string, caseSensitive bool) bool {
	str := s.String
	if !caseSensitive {
		str = strings.ToLower(str)
		suffix = strings.ToLower(suffix)
	}
	return strings.HasSuffix(str, suffix)
}

// Explode 按分隔符分割字符串，选项修剪值并跳过空值
func (s *Str) Explode(delimiter string, trim bool, skipEmpty bool) []string {
	parts := strings.Split(s.String, delimiter)
	if trim {
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
	}
	if skipEmpty {
		var nonEmptyParts []string
		for _, part := range parts {
			if part != "" {
				nonEmptyParts = append(nonEmptyParts, part)
			}
		}
		return nonEmptyParts
	}
	return parts
}

// CountWords 计算字符串中的单词数
func (s *Str) CountWords() int {
	return len(regexp.MustCompile(`\s+`).Split(strings.TrimSpace(s.String), -1))
}

// Base64UrlEncode 对字符串进行 URL 和文件名安全的 Base64 编码
func (s *Str) Base64UrlEncode() string {
	return base64.URLEncoding.EncodeToString([]byte(s.String))
}

// Base64UrlDecode 解码 URL 和文件名安全的 Base64 编码字符串
func (s *Str) Base64UrlDecode() (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(s.String)
	return string(decoded), err
}

// MatchWildcard 检查字符串是否匹配给定的通配符模式
func MatchWildcard(pattern, s string, caseSensitive bool) bool {
	replacements := map[string]string{
		"\\*": ".*",
		"\\?": ".",
	}
	for old, newValue := range replacements {
		pattern = strings.ReplaceAll(pattern, old, newValue)
	}
	pattern = "^" + pattern + "$"
	if !caseSensitive {
		pattern = "(?i)" + pattern
	}
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// MbUcFirst 将字符串的第一个字符大写（支持多字节）
func (s *Str) MbUcFirst() string {
	if len(s.String) == 0 {
		return s.String
	}
	runes := []rune(s.String)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// MbUcWords 将字符串中的每个单词的第一个字符大写（支持多字节）
func (s *Str) MbUcWords() string {
	if len(s.String) == 0 {
		return s.String
	}
	words := strings.Fields(s.String)
	for i, word := range words {
		words[i] = NewStr(word).MbUcFirst()
	}
	return strings.Join(words, " ")
}

// ToLower converts the value to a lower case string
func (s *Str) ToLower() string {
	return strings.ToLower(s.String)
}

// ToUpper converts the value to an upper case string
func (s *Str) ToUpper() string {
	return strings.ToUpper(s.String)
}

// TrimSpace removes leading and trailing white spaces from the value
func (s *Str) TrimSpace() string {
	return strings.TrimSpace(s.String)
}

// Trim returns a slice of the string s with all leading and
// trailing Unicode code points contained in cutSet removed.
func (s *Str) Trim(cutSet string) string {
	return strings.Trim(s.String, cutSet)
}

// EscapeHTML escapes HTML characters in the value to prevent XSS
func (s *Str) EscapeHTML() string {
	return html.EscapeString(s.String)
}

// ConvertCamelToSnake 转换驼峰字符串为下划线字符串
func (s *Str) ConvertCamelToSnake() string {
	// 查找大写字母的位置
	var uppercasePattern = regexp.MustCompile(`([A-Z])`)
	// 将大写字母替换为下划线，后跟小写字母
	snake := uppercasePattern.ReplaceAllString(s.String, `_$1`)
	// 将整个字符串转换为小写
	snake = strings.ToLower(snake)
	// 删除前导下划线（如果存在）
	if strings.HasPrefix(snake, "_") {
		snake = snake[1:]
	}
	return snake
}

// 数值转换

// SetFloat64 设置float64值
func (s *Str) SetFloat64(number float64) *Str {
	s.Float64 = number
	return s
}

// FloatToString 安全地将浮点数转换为字符串
func (s *Str) FloatToString() *Str {
	s.String = strconv.FormatFloat(s.Float64, 'f', -1, 64)
	return s
}
