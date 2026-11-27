// Package content 富文本内容处理器。
//
// 对齐 PHP 端 Jcbowen\JcbaseYii2\components\Content 的 toSave / toShow 语义：
//   - ToSave：富文本转入库口径——img src 去附件域名（仅剥本端附件域名，外链透传），
//     并做 HTML 实体编码；
//   - ToShow：富文本转展示口径——HTML 实体解码，img src 补全附件域名。
//
// 媒体路径转换通过 MediaConverter 接口注入（由 component/attachment 提供
// ToMedia/ToSource），本包不直接依赖附件组件，保持职责单一。
//
// 与 PHP 版的行为差异（均为有意省略）：
//   - PHP toSave 对以 "." 开头的相对引用（如 ../attachment/）有专门的
//     "_attachment/" 前缀清理分支，Go 版直接原样保留该类 src；
//   - PHP 版依赖 Yii params 的 attachment/attachment_local 双域名判定，
//     Go 版的域名判定由注入的 MediaConverter（附件组件配置）完成。
package content

import (
	"strings"
)

// MediaConverter 媒体路径转换接口（由 component/attachment 的 Attachment 实现）。
type MediaConverter interface {
	// ToMedia 附件相对路径补全为完整 URL（已是完整 URL 的输入透传）。
	ToMedia(src string, args ...interface{}) string
	// ToSource 附件完整 URL 还原为相对路径（相对路径与外链透传）。
	ToSource(src string, args ...bool) string
}

// imgTagPattern 匹配单个 img 标签（大小写不敏感，跨行容错）。
var imgTagPattern = regexpImgTag

// imgSrcPattern 在 img 标签内定位 src 属性值（双引号或单引号）。
var imgSrcPattern = regexpImgSrc

// ToSave 将富文本 HTML 转换为入库口径：img src 去附件域名，entities 为 true 时
// 做 HTML 实体编码（对齐 PHP Content::toSave 默认行为；存量库实证 goods content
// 以实体编码形态存储）。
//
// 输入若已是实体编码形态（含 &lt;/&gt;），先解码再处理，保证幂等。
//
// 参数：
//   - html: 富文本 HTML（raw 或实体编码形态均可）
//   - conv: 媒体路径转换器（nil 时仅做实体转换，不改写 img src）
//   - entities: 是否 HTML 实体编码输出
//
// 返回值：
//   - string: 入库口径的富文本；空输入返回空串
//
// 使用示例：
//
//	content.ToSave(`<p><img src="https://oss.host/images/a.jpg"></p>`, att, true)
//	// => &lt;p&gt;&lt;img src=&quot;images/a.jpg&quot;&gt;&lt;/p&gt;
func ToSave(html string, conv MediaConverter, entities bool) string {
	return process(html, conv, entities, true)
}

// ToShow 将入库口径的富文本转换为展示口径：实体解码，img src 补全附件域名。
//
// 参数：
//   - html: 富文本 HTML（入库口径或 raw 均可）
//   - conv: 媒体路径转换器（nil 时仅做实体解码）
//   - entities: 为 true 时输出前重新实体编码（对齐 PHP toShow 的 entities 参数；
//     API 返回场景一般传 false，输出可直接渲染的 raw HTML）
//
// 返回值：
//   - string: 展示口径的富文本；空输入返回空串
//
// 使用示例：
//
//	content.ToShow(`&lt;p&gt;&lt;img src=&quot;images/a.jpg&quot;&gt;&lt;/p&gt;`, att, false)
//	// => <p><img src="https://oss.host/images/a.jpg"></p>
func ToShow(html string, conv MediaConverter, entities bool) string {
	return process(html, conv, entities, false)
}

// process 统一处理入口（toSave: toMedia=false；toShow: toMedia=true）。
func process(html string, conv MediaConverter, entities, toSave bool) string {
	if html == "" {
		return ""
	}

	// 已是实体编码形态（含 &lt;/&gt;）先解码（对齐 PHP 防重复解码保护）
	html = HTMLEntitiesDecode(html)

	// 逐 img 标签改写 src（对齐 PHP 的 tag 级 str_replace 语义）
	html = imgTagPattern.ReplaceAllStringFunc(html, func(tag string) string {
		m := imgSrcPattern.FindStringSubmatchIndex(tag)
		if m == nil {
			return tag
		}
		// 双分支模式：仅其一命中；group1=双引号值(m[2:m[3]])，group2=单引号值(m[4:m[5]])
		// 两个分支都会完整覆盖 start/end，无需初值（避免 ineffassign 告警）
		var start, end int
		if m[2] >= 0 {
			start, end = m[2], m[3]
		} else if m[4] >= 0 {
			start, end = m[4], m[5]
		} else {
			return tag
		}
		src := tag[start:end]
		var converted string
		if toSave {
			// 以 "." 开头的相对引用原样保留（PHP 版为 _attachment/ 前缀清理分支）
			if strings.HasPrefix(src, ".") {
				return tag
			}
			if conv == nil {
				return tag
			}
			converted = conv.ToSource(src)
		} else {
			if conv == nil {
				return tag
			}
			converted = conv.ToMedia(src)
		}
		if converted == src {
			return tag
		}
		return tag[:start] + converted + tag[end:]
	})

	if entities {
		return HTMLEntities(html)
	}
	return html
}

// HTMLEntitiesDecode HTML 实体解码（仅在含 &lt; 或 &gt; 时执行，避免重复解码）。
//
// 参数：
//   - html: 待解码字符串
//
// 返回值：
//   - string: 解码结果
func HTMLEntitiesDecode(html string) string {
	if strings.Contains(html, "&lt;") || strings.Contains(html, "&gt;") {
		return htmlUnescape(html)
	}
	return html
}

// htmlEntityReplacer HTML 实体编码替换表（ENT_QUOTES 语义：&amp; &lt; &gt; &quot; &#039;）。
// strings.Replacer 为单趟替换，不会产生二次编码。
var htmlEntityReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
	"'", "&#039;",
)

// HTMLEntities HTML 实体编码（对齐 PHP htmlentities + ENT_QUOTES：编码 &amp; &lt; &gt; " '）。
//
// 参数：
//   - html: 待编码字符串
//
// 返回值：
//   - string: 编码结果
func HTMLEntities(html string) string {
	if html == "" {
		return ""
	}
	return htmlEntityReplacer.Replace(html)
}
