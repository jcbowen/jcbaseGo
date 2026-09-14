package attachment

import (
	"strings"
	"time"
)

// ToMedia 将附件相对路径补全为可访问的完整 URL。
// 参数：
//   - src: 附件源路径（可为相对路径、绝对路径或已带协议的 URL）
//   - args: 可选参数，args[0] 表示是否强制使用本地访问域名，args[1] 表示是否追加时间戳缓存参数
//
// 返回值：
//   - 访问链接，若 src 为空则返回空字符串
//
// 异常：
//   - 若 args 中类型不匹配，则使用默认值，避免因强制类型转换导致 panic
//
// 示例：
//
//	att.ToMedia("avatar.png") => "https://host/attachment/avatar.png"
func (a *Attachment) ToMedia(src string, args ...interface{}) string {
	isLocal := false // 是否为本地附件
	isCache := true  // 是否在尾部添加时间戳

	if len(args) > 0 {
		if v, ok := args[0].(bool); ok {
			isLocal = v
		}
	}
	if len(args) > 1 {
		if v, ok := args[1].(bool); ok {
			isCache = v
		}
	}

	if len(src) == 0 {
		return ""
	}
	if !isCache {
		src += "?v=" + time.Now().Format("20060102150405")
	}

	if strings.Index(src, "http://") == 0 || strings.Index(src, "https://") == 0 {
		return src
	} else if strings.Index(src, "//") == 0 {
		return "http:" + src
	}

	src = strings.TrimPrefix(src, "/")

	if isLocal || a.BaseConfig.StorageType == "local" {
		src = a.BaseConfig.LocalVisitDomain + a.BaseConfig.LocalDir + "/" + src
	} else {
		src = a.BaseConfig.VisitDomain + src
	}

	return src
}

// ToSource 将附件访问 URL 或带域名的路径还原为相对路径。
// 参数：
//   - src: 带附件域名的 URL 或路径
//   - args: 可选参数，args[0] 表示是否移除 URL 中的查询参数；默认 true
//
// 返回值：
//   - 去除域名后的相对路径，若 src 为空则返回空字符串
//
// 异常：
//   - 无异常；无法识别的 URL 会原样返回
//
// 示例：
//
//	att.ToSource("https://host/attachment/avatar.png") => "attachment/avatar.png"
//	att.ToSource("https://host/attachment/avatar.png?v=123") => "attachment/avatar.png"
//	att.ToSource("https://host/attachment/avatar.png?v=123", false) => "attachment/avatar.png?v=123"
func (a *Attachment) ToSource(src string, args ...bool) string {
	removeQuery := true
	if len(args) > 0 {
		removeQuery = args[0]
	}
	if src == "" {
		return ""
	}
	if removeQuery {
		src = stripURLQuery(src)
	}

	for _, domain := range a.buildDomainCandidates() {
		if strings.HasPrefix(src, domain) {
			return strings.TrimLeft(strings.TrimPrefix(src, domain), "/")
		}
	}

	if strings.HasPrefix(src, "/") {
		return strings.TrimLeft(src, "/")
	}
	return src
}

// stripURLQuery 去除 URL 中的查询字符串与锚点，保留有效路径。
// 参数：
//   - src: 待处理的 URL 或路径
//
// 返回值：
//   - 去除 ? 后参数及 # 后锚点后的路径
func stripURLQuery(src string) string {
	if idx := strings.IndexAny(src, "?#"); idx >= 0 {
		return src[:idx]
	}
	return src
}

// buildDomainCandidates 组装可用于剥离附件域名的候选域名集合。
// 返回值：
//   - 先返回远程访问域名，再返回本地访问域名，保证优先匹配更具体的访问域名
func (a *Attachment) buildDomainCandidates() []string {
	candidate := make([]string, 0, 4)
	if a != nil && a.BaseConfig != nil {
		if domain := strings.TrimSpace(a.BaseConfig.VisitDomain); domain != "" && domain != "/" {
			candidate = append(candidate, strings.TrimRight(domain, "/")+"/")
		}
		if domain := strings.TrimSpace(a.BaseConfig.LocalVisitDomain); domain != "" && domain != "/" {
			candidate = append(candidate, strings.TrimRight(domain, "/")+"/")
		}
	}
	return candidate
}
