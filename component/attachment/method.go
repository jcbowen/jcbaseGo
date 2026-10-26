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

// defaultLocalDir 本地附件目录缺省名，与 jcbaseGo.AttachmentStruct.LocalDir 的 default 标签一致
const defaultLocalDir = "attachment"

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
			strings.HasPrefix(src, a.BaseConfig.LocalVisitDomain) && indexOfStaticSegment(detPath(src)) < 0
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
	} else if idx := indexOfStaticSegment(src); idx >= 0 {
		// 相对路径含独立 static 目录段时属于内部资源，从 static 处拼接本地资源域名（对齐 PHP toMedia 的 static/ 分支）
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
// 还原规则按输入形态与存储模式分支：
//   - 命中本地访问域名（LocalVisitDomain）：优先从文件类型目录（images/videos/offices 等）截取；
//     未命中类型目录与 static 时剥离 LocalDir 前缀（覆盖 /uploads/seed/** 等非类型目录路径）
//   - 命中远程访问域名（VisitDomain）：按对象 Key 语义从类型目录截取；
//     对象 Key 合法含 uploads 等目录段时不做强行剥离（远程 Key 与本地 LocalDir 无关）
//   - 未命中任何域名的 http(s) URL：local 存储模式下路径命中 /<LocalDir>/ 前缀即剥离
//     （覆盖历史按任意请求域名拼出的存量附件 URL）；其余视为外链原样返回，且不再剥除
//     查询参数，保证带参数外链（如占位图服务）不被破坏
//   - 纯路径形态（/uploads/... 或 uploads/...）：local 存储模式下剥离 LocalDir 前缀，
//     其余仅去掉开头的 /
//
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
//	att.ToSource("https://localhost:8080/uploads/images/logo.png")        => "images/logo.png"
//	att.ToSource("https://old-host.example/uploads/images/logo.png")      => "images/logo.png" // 历史请求域名存量，按路径前缀剥离
//	att.ToSource("https://localhost:8080/uploads/seed/x.png")             => "seed/x.png"      // 非类型目录路径剥离 LocalDir
//	att.ToSource("https://cdn.example.com/uploads/myzipfile/x.png")       => "uploads/myzipfile/x.png" // 远程 Key 不做 LocalDir 剥离
//	att.ToSource("https://placehold.co/300x300?text=Tomato")              => 原样返回（外链保留查询参数）
//	att.ToSource("https://host/attachment/avatar.png?v=123", false)       => "attachment/avatar.png?v=123"
func (a *Attachment) ToSource(src string, args ...bool) string {
	if src == "" {
		return ""
	}
	removeQuery := true
	if len(args) > 0 {
		removeQuery = args[0]
	}

	remoteDomain := ""
	localDomain := ""
	if a != nil && a.BaseConfig != nil {
		remoteDomain = strings.TrimSpace(a.BaseConfig.VisitDomain)
		localDomain = strings.TrimSpace(a.BaseConfig.LocalVisitDomain)
	}

	// 纯路径形态（无协议）：local 模式剥离 LocalDir 前缀，其余归一为相对路径
	if !strings.HasPrefix(src, "http://") && !strings.HasPrefix(src, "https://") {
		return a.sourceFromPathForm(src, removeQuery)
	}

	startsWithLocal := isUsableDomain(localDomain) && strings.HasPrefix(src, localDomain)
	startsWithRemote := isUsableDomain(remoteDomain) && strings.HasPrefix(src, remoteDomain)

	// 未命中任何附件域名：local 模式按 LocalDir 路径前缀剥离历史存量 URL，外链原样保留
	if !startsWithLocal && !startsWithRemote {
		if base := detPath(src); a.isLocalStorage() {
			if rel, ok := cutLocalDir(base, a.localDirName()); ok {
				return joinQuery(rel, tailQuery(src), removeQuery)
			}
		}
		return src
	}

	base := detPath(src)

	// 命中本地域名且路径含独立 static 目录段时属于内部资源，保留 static 起始路径（对应 PHP else 分支）
	staticIdx := -1
	if startsWithLocal && !startsWithRemote && base != "" {
		staticIdx = indexOfStaticSegment(base)
	}

	if base != "" {
		if staticIdx >= 0 {
			return joinQuery(strings.TrimLeft(base[staticIdx:], "/"), tailQuery(src), removeQuery)
		}
		// 默认定位到文件类型目录（对应 PHP 的 strpos($urls['path'], $type)），
		// 本地域名下的 LocalDir 前缀与远程域名后的 bucket/path 前缀均会被一并剥离
		if cut := indexOfFileType(base); cut >= 0 {
			return joinQuery(base[cut:], tailQuery(src), removeQuery)
		}
		// 本地域名命中但未命中类型目录：剥离 LocalDir 前缀得到服务目录之后的相对路径
		// （如 /uploads/seed/**；远程域名分支不处理，对象 Key 与 LocalDir 无关）
		if startsWithLocal && !startsWithRemote {
			if rel, ok := cutLocalDir(base, a.localDirName()); ok {
				return joinQuery(rel, tailQuery(src), removeQuery)
			}
		}
		// 兜底：detPath 已取到纯路径（域名部分自然去除），仅去掉开头的斜杠
		if startsWithRemote {
			return joinQuery(strings.TrimLeft(strings.TrimPrefix(base, strings.TrimRight(remoteDomain, "/")), "/"), tailQuery(src), removeQuery)
		}
		return joinQuery(strings.TrimLeft(strings.TrimPrefix(base, strings.TrimRight(localDomain, "/")), "/"), tailQuery(src), removeQuery)
	}

	if strings.HasPrefix(src, "/") {
		return strings.TrimLeft(src, "/")
	}
	return src
}

// sourceFromPathForm 处理无协议前缀的路径形态输入：
// local 存储模式剥离 LocalDir 前缀（/uploads/images/x → images/x，uploads/x → x），
// 协议相对链接（//host/x）非附件路径原样返回，其余仅去掉开头的 /
func (a *Attachment) sourceFromPathForm(src string, removeQuery bool) string {
	path := src
	tail := ""
	if idx := strings.IndexAny(src, "?#"); idx >= 0 {
		path = src[:idx]
		tail = src[idx:]
	}
	if strings.HasPrefix(path, "//") {
		return src
	}
	rel := strings.TrimLeft(path, "/")
	if a.isLocalStorage() {
		rel = strings.TrimPrefix(rel, a.localDirName()+"/")
	}
	if removeQuery {
		return rel
	}
	return rel + tail
}

// isLocalStorage 判断是否本地存储模式；StorageType 缺省（字面量构造未赋值）按 local 处理
// 与 AttachmentStruct.StorageType 的 default:local 一致
func (a *Attachment) isLocalStorage() bool {
	if a == nil || a.BaseConfig == nil {
		return true
	}
	st := a.BaseConfig.StorageType
	return st == "" || st == "local"
}

// localDirName 取本地存储目录名；未配置时回 AttachmentStruct 的默认值 attachment
func (a *Attachment) localDirName() string {
	if a == nil || a.BaseConfig == nil || a.BaseConfig.LocalDir == "" {
		return defaultLocalDir
	}
	if dir := strings.Trim(a.BaseConfig.LocalDir, "/"); dir != "" {
		return dir
	}
	return defaultLocalDir
}

// cutLocalDir 从 URL 路径中剥离 /<LocalDir>/ 服务目录前缀
// 参数：
//   - path: URL 的 path 部分
//   - dir: 本地存储目录名
//
// 返回值：
//   - string: 剥离后的相对路径
//   - bool: 是否命中前缀
func cutLocalDir(path, dir string) (string, bool) {
	if dir == "" {
		return "", false
	}
	return strings.CutPrefix(path, "/"+dir+"/")
}

// isUsableDomain 域名是否可用于 URL 前缀匹配（空与占位 "/" 视为未配置）
func isUsableDomain(domain string) bool {
	return domain != "" && domain != "/"
}

// tailQuery 取 URL 尾部的查询串与锚点片段，无则返回空串
func tailQuery(src string) string {
	if idx := strings.IndexAny(src, "?#"); idx >= 0 {
		return src[idx:]
	}
	return ""
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
//   - rawQuery: 原始输入中含 ?/# 的尾部片段，为空则说明无查询串可恢复
//   - removeQuery: 是否移除查询参数
//
// 返回值：
//   - 保留查询参数时返回路径加原始尾部（含锚点），否则仅返回路径部分
func joinQuery(path, rawQuery string, removeQuery bool) string {
	if removeQuery || rawQuery == "" {
		return path
	}
	return path + rawQuery
}

// indexOfFileType 定位路径中首个附件文件类型目录的起始下标。
// 仅匹配 /item/ 或路径开头 item/ 两种独立目录段形式，避免 myimages 等普通子串被误截。
// 参数：
//   - path: 待匹配的 URL 路径
//
// 返回值：
//   - 首个命中的类型目录在路径中的下标，未命中返回 -1
func indexOfFileType(path string) int {
	cut := -1
	for _, item := range attachmentFileTypes {
		// 匹配 /item/ 或路径开头 item/ 两种情况，确保是独立目录段
		if idx := strings.Index(path, "/"+item+"/"); idx >= 0 && (cut < 0 || idx+1 < cut) {
			cut = idx + 1 // 跳过前导斜杠，指向 item 起始位置
		}
		if strings.HasPrefix(path, item+"/") && cut != 0 {
			cut = 0
		}
	}
	return cut
}

// indexOfStaticSegment 定位路径中独立 static 目录段的起始下标。
// 仅匹配 /static/ 或路径开头 static/ 两种独立目录段形式，避免 oldstatic/ 等子串误命中。
// 参数：
//   - path: 待匹配的 URL 路径
//
// 返回值：
//   - static 目录段在路径中的起始下标，未命中返回 -1
func indexOfStaticSegment(path string) int {
	if idx := strings.Index(path, "/static/"); idx >= 0 {
		return idx + 1 // 跳过前导斜杠，指向 static 起始位置
	}
	if strings.HasPrefix(path, "static/") {
		return 0
	}
	return -1
}
