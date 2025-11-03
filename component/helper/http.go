package helper

import (
	"net/http"
	"strings"
)

// ExtractHeaders 提取HTTP头信息并转换为map[string]string格式
// 该方法将http.Header中的每个键值对转换为字符串，多个值用逗号分隔
//
// 参数:
//
//	header - HTTP请求头
//
// 返回值:
//
//	map[string]string - 转换后的头信息
func ExtractHeaders(header http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range header {
		result[key] = strings.Join(values, ", ")
	}
	return result
}
