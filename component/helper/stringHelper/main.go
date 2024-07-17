package stringHelper

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ByteLength 返回给定字符串的字节数
func ByteLength(s string) int {
	return utf8.RuneCountInString(s)
}

// ByteSubstr 返回从 start 开始、长度为 length 的子字符串
func ByteSubstr(s string, start int, length int) string {
	runes := []rune(s)
	if start >= len(runes) {
		return ""
	}
	end := start + length
	if length == -1 || end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}

// Basename 返回路径的最后一个名字组件
func Basename(path, suffix string) string {
	path = strings.TrimSuffix(path, suffix)
	path = strings.ReplaceAll(path, "\\", "/")
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

// Dirname 返回父目录的路径
func Dirname(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.TrimSuffix(path, "/")
	pos := strings.LastIndex(path, "/")
	if pos == -1 {
		return ""
	}
	return path[:pos]
}

// Truncate 将字符串截断为指定的字符数
func Truncate(s string, length int, suffix string) string {
	if utf8.RuneCountInString(s) > length {
		return string([]rune(s)[:length]) + suffix
	}
	return s
}

// TruncateWords 将字符串截断为指定的单词数
func TruncateWords(s string, count int, suffix string) string {
	words := regexp.MustCompile(`\s+`).Split(strings.TrimSpace(s), -1)
	if len(words) > count {
		return strings.Join(words[:count], " ") + suffix
	}
	return s
}

// StartsWith 检查字符串是否以指定的子字符串开头
func StartsWith(s, prefix string, caseSensitive bool) bool {
	if !caseSensitive {
		s = strings.ToLower(s)
		prefix = strings.ToLower(prefix)
	}
	return strings.HasPrefix(s, prefix)
}

// EndsWith 检查字符串是否以指定的子字符串结尾
func EndsWith(s, suffix string, caseSensitive bool) bool {
	if !caseSensitive {
		s = strings.ToLower(s)
		suffix = strings.ToLower(suffix)
	}
	return strings.HasSuffix(s, suffix)
}

// Explode 按分隔符分割字符串，选项修剪值并跳过空值
func Explode(s, delimiter string, trim bool, skipEmpty bool) []string {
	parts := strings.Split(s, delimiter)
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
func CountWords(s string) int {
	return len(regexp.MustCompile(`\s+`).Split(strings.TrimSpace(s), -1))
}

// NormalizeNumber 将数字字符串中的逗号替换为点
func NormalizeNumber(value interface{}) string {
	val := fmt.Sprintf("%v", value)
	return strings.ReplaceAll(val, ",", ".")
}

// Base64UrlEncode 对字符串进行 URL 和文件名安全的 Base64 编码
func Base64UrlEncode(input string) string {
	return base64.URLEncoding.EncodeToString([]byte(input))
}

// Base64UrlDecode 解码 URL 和文件名安全的 Base64 编码字符串
func Base64UrlDecode(input string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(input)
	return string(decoded), err
}

// FloatToString 安全地将浮点数转换为字符串
func FloatToString(number float64) string {
	return strconv.FormatFloat(number, 'f', -1, 64)
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

// MbUcfirst 将字符串的第一个字符大写（支持多字节）
func MbUcfirst(s, encoding string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// MbUcwords 将字符串中的每个单词的第一个字符大写（支持多字节）
func MbUcwords(s, encoding string) string {
	if len(s) == 0 {
		return s
	}
	words := strings.Fields(s)
	for i, word := range words {
		words[i] = MbUcfirst(word, encoding)
	}
	return strings.Join(words, " ")
}
