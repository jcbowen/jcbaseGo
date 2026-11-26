package content

import (
	"html"
	"regexp"
)

// regexpImgTag 匹配单个 img 标签（大小写不敏感、跨行容错）。
var regexpImgTag = regexp.MustCompile(`(?is)<img[^>]*>`)

// regexpImgSrc 在 img 标签内定位 src 属性。
// Go RE2 不支持反向引用，用双引号/单引号双分支表达成对引号：
// 子匹配 1=双引号值，2=单引号值（同一时刻仅其一命中）。
var regexpImgSrc = regexp.MustCompile(`(?is)\ssrc\s*=\s*(?:"([^"]*)"|'([^']*)')`)

// htmlUnescape HTML 实体解码（标准库实现，语义覆盖 ENT_QUOTES）。
func htmlUnescape(s string) string {
	return html.UnescapeString(s)
}
