package attachment

import (
	"net/url"
	"strings"
	"time"
)

// attachmentFileTypes 附件类型目录候选名，与 PHP 端对齐：File.php 上传目录为 {$type}s/（复数形式），
// removeMediaDomain 截取时同样使用 $item.'s'，故全部取复数。用于 ToSource 剥离域名后
// 定位类型目录起点、ToMedia 重组本地完整 URL；匹配时取所有候选名中出现最早的一个作为截取点。
var attachmentFileTypes = []string{"images", "thumbs", "voices", "videos", "audios", "offices", "zips"}

// ToMedia 将附件相对路径补全为可访问的完整 URL。
// 参数：
//   - src: 附件源路径（可为相对路径、绝对路径或已带协议的 URL）
//   - args: 可选参数，args[0] 表示是否强制使用本地访问域名；args[1] 表示是否启用缓存，传 false 时会在路径尾部追加 ?v= 时间戳以绕过缓存（与 PHP 端 $is_cache 语义一致，默认 true 不追加）
//
// 返回值：
//   - 访问链接，若 src 为空则返回空字符串
//
// 异常：
//   - 若 args 中类型不匹配，则使用默认值，避免因强制类型转换导致 panic
//
// 示例：
//
//	att.ToMedia("avatar.png") => "https://host/uploads/avatar.png" // 本地存储：域名 + LocalDir + 相对路径
//	att.ToMedia("images/avatar.png") => "https://cdn.host/images/avatar.png" // 远程存储：仅补全访问域名
//	att.ToMedia("static/js/app.js") => "https://host/static/js/app.js" // 含 static/ 的资源：从 static/ 处拼接本地域名
//	att.ToMedia("https://host/uploads/images/avatar.png") => "https://cdn.host/images/avatar.png" // 本地完整 URL 重组（远程存储时切换到远程域名）
func (a *Attachment) ToMedia(src string, args ...interface{}) string {
	isLocal := false // 是否强制使用本地附件域名
	isCache := true  // 是否启用缓存，false 时在路径尾部追加 ?v= 时间戳以绕过缓存

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

	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		isLocalURL := a.BaseConfig != nil && a.BaseConfig.LocalVisitDomain != "" &&
			strings.HasPrefix(src, a.BaseConfig.LocalVisitDomain) && !strings.Contains(src, "/static/")
		if !isLocalURL {
			return src
		}
		// 已是完整本地附件 URL：剥离域名与 LocalDir 前缀后重新拼装（对齐 PHP toMedia 的
		// startsWith(attachment_local) 分支，远程存储开启时可借此切换到远程访问域名）；
		// 未命中类型目录时视为非标准附件 URL，原样返回
		rel, ok := stripToLocalRelative(src)
		if !ok {
			return src
		}
		src = rel
	} else if strings.HasPrefix(src, "//") {
		return "http:" + src
	} else if idx := strings.Index(src, "static/"); idx >= 0 {
		// 相对路径含 static/ 时属于内部资源，从 static/ 处拼接本地资源域名（对齐 PHP toMedia 的 static/ 分支）
		return a.BaseConfig.LocalVisitDomain + src[idx:]
	}

	src = strings.TrimPrefix(src, "/")

	if isLocal || a.BaseConfig.StorageType == "local" {
		src = a.BaseConfig.LocalVisitDomain + a.BaseConfig.LocalDir + "/" + src
	} else {
		src = a.BaseConfig.VisitDomain + src
	}

	return src
}

// stripToLocalRelative 将完整附件 URL 剥离为从类型目录起的相对路径。
// 参数：
//   - src: 完整附件 URL
//
// 返回值：
//   - 相对路径（如 "images/2026/01/x.png"）与是否成功；未命中类型目录时返回原串与 false
func stripToLocalRelative(src string) (string, bool) {
	path := detPath(src)
	if path == "" {
		return src, false
	}
	if cut := indexOfFileType(path); cut >= 0 {
		return path[cut:], true
	}
	return src, false
}

// ToSource 将附件访问 URL 或带域名的路径还原为相对路径。
// 逻辑对齐 PHP 端 Util::removeMediaDomain：命中附件域名后，
// 普通附件从文件类型目录（images/videos/offices 等）处截取；
// 本地 /static/ 内部资源则保留 /static 起始的路径。
// 参数：
//   - src: 带附件域名的 URL 或路径
//   - args: 可选参数，args[0] 表示是否移除 URL 中的查询参数；默认 true
//
// 返回值：
//   - 剥离域名后的相对路径，若 src 为空则返回空字符串
//
// 异常：
//   - 无异常；无法识别的 URL 会原样返回
//
// 示例：
//
//	att.ToSource("https://host/attachment/avatar.png") => "attachment/avatar.png"
//	att.ToSource("https://host/attachment/avatar.png?v=123") => "attachment/avatar.png"
//	att.ToSource("https://host/attachment/avatar.png?v=123", false) => "attachment/avatar.png?v=123"
//	att.ToSource("https://localhost:8080/uploads/images/logo.png") => "images/logo.png" // 本地域名剥离 LocalDir 前缀
//	att.ToSource("https://localhost:8080/static/js/app.js") => "static/js/app.js" // 本地 /static/ 资源保留 static 起点
func (a *Attachment) ToSource(src string, args ...bool) string {
	removeQuery := true
	if len(args) > 0 {
		removeQuery = args[0]
	}
	if src == "" {
		return ""
	}

	// 记录原始查询串，removeQuery 为 false 时在还原后追加回路径尾部
	rawQuery := ""
	if idx := strings.Index(src, "?"); idx >= 0 {
		rawQuery = src[idx:]
	}
	if removeQuery {
		src = stripURLQuery(src)
	}

	remoteDomain := ""
	localDomain := ""
	if a != nil && a.BaseConfig != nil {
		remoteDomain = strings.TrimSpace(a.BaseConfig.VisitDomain)
		localDomain = strings.TrimSpace(a.BaseConfig.LocalVisitDomain)
	}
	if (localDomain == "" || localDomain == "/") && (remoteDomain == "" || remoteDomain == "/") {
		// 未配置附件域名时无法剥离，原样返回
		if strings.HasPrefix(src, "/") {
			return strings.TrimLeft(src, "/")
		}
		return src
	}

	// 仅当 URL 命中附件域名时才做剥离（对应 PHP 端的 startsWith 守卫）
	startsWithLocal := localDomain != "" && localDomain != "/" && strings.HasPrefix(src, localDomain)
	startsWithRemote := remoteDomain != "" && remoteDomain != "/" && strings.HasPrefix(src, remoteDomain)
	if !startsWithLocal && !startsWithRemote {
		if strings.HasPrefix(src, "/") {
			return strings.TrimLeft(src, "/")
		}
		return src
	}

	// 命中本地域名且路径含 /static/ 时属于内部资源，保留 /static 起始路径（对应 PHP else 分支）
	isStaticResource := startsWithLocal && !startsWithRemote && strings.Contains(src, "/static/")

	if base := detPath(src); base != "" {
		if isStaticResource {
			if idx := strings.Index(base, "/static"); idx >= 0 {
				return joinQuery(strings.TrimLeft(base[idx:], "/"), src, rawQuery, removeQuery)
			}
			return joinQuery(strings.TrimLeft(base, "/"), src, rawQuery, removeQuery)
		}
		// 默认定位到文件类型目录（对应 PHP 的 strpos($urls['path'], $type)），
		// 本地域名下的 LocalDir 前缀与远程域名后的 bucket/path 前缀均会被一并剥离
		if cut := indexOfFileType(base); cut >= 0 {
			return joinQuery(base[cut:], src, rawQuery, removeQuery)
		}
		// 未找到类型目录时兜底：detPath 已取到纯路径（域名部分自然去除），仅去掉开头的斜杠
		if startsWithRemote {
			return joinQuery(strings.TrimLeft(strings.TrimPrefix(base, strings.TrimRight(remoteDomain, "/")), "/"), src, rawQuery, removeQuery)
		}
		return joinQuery(strings.TrimLeft(strings.TrimPrefix(base, strings.TrimRight(localDomain, "/")), "/"), src, rawQuery, removeQuery)
	}

	if strings.HasPrefix(src, "/") {
		return strings.TrimLeft(src, "/")
	}
	return src
}

// detPath 提取 URL 的路径部分，对齐 PHP 端 parse_url($src)['path'] 的行为。
// 参数：
//   - src: 待处理的 URL 或路径
//
// 返回值：
//   - URL 的 path 部分；解析失败时返回空字符串
func detPath(src string) string {
	u, err := url.Parse(src)
	if err != nil {
		return ""
	}
	return u.Path
}

// joinQuery 在剥离域名后的相对路径尾部按需追加原始查询串。
// 参数：
//   - path: 剥离域名后的相对路径
//   - src: 本次参与剥离处理的 URL（removeQuery 为 false 时仍带查询串/锚点）
//   - rawQuery: 原始输入中含 ? 的尾部片段，为空则说明无查询串可恢复
//   - removeQuery: 是否移除查询参数
//
// 返回值：
//   - 保留查询参数时返回路径加原始尾部（含锚点），否则仅返回路径部分
func joinQuery(path, src, rawQuery string, removeQuery bool) string {
	if removeQuery || rawQuery == "" {
		return path
	}
	// stripURLQuery 未启用时 src 仍带查询串，锚点也一并保留
	if idx := strings.IndexAny(src, "?#"); idx >= 0 {
		return path + src[idx:]
	}
	return path
}

// indexOfFileType 定位路径中首个附件文件类型目录的起始下标。
// 参数：
//   - path: 待匹配的 URL 路径
//
// 返回值：
//   - 首个命中的类型目录在路径中的下标，未命中返回 -1
func indexOfFileType(path string) int {
	cut := -1
	for _, item := range attachmentFileTypes {
		if idx := strings.Index(path, item); idx >= 0 && (cut < 0 || idx < cut) {
			cut = idx
		}
	}
	return cut
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
