package content

import (
	"strings"
	"testing"
)

// mockConverter 测试用媒体转换器（模拟附件组件行为：域名前缀剥/补）。
type mockConverter struct {
	domain string
}

func (m *mockConverter) ToMedia(src string, args ...interface{}) string {
	if src == "" {
		return ""
	}
	if len(src) > 7 && (src[:7] == "http://" || (len(src) > 8 && src[:8] == "https://")) {
		return src // 已是完整 URL 透传
	}
	return m.domain + src
}

func (m *mockConverter) ToSource(src string, args ...bool) string {
	if src == "" {
		return ""
	}
	p := m.domain
	if len(src) > len(p) && src[:len(p)] == p {
		return src[len(p):]
	}
	return src // 相对路径/外链透传
}

const testDomain = "https://oss.example.com/"

func TestToSave(t *testing.T) {
	conv := &mockConverter{domain: testDomain}

	// raw HTML + 绝对 URL src：去域名 + 实体编码
	got := ToSave(`<p><img src="`+testDomain+`images/a.jpg" alt="图"></p>`, conv, true)
	want := `&lt;p&gt;&lt;img src=&quot;images/a.jpg&quot; alt=&quot;图&quot;&gt;&lt;/p&gt;`
	if got != want {
		t.Fatalf("ToSave raw:\n got: %s\nwant: %s", got, want)
	}

	// 外链 src 透传不剥
	got = ToSave(`<img src="https://third.com/b.jpg">`, conv, true)
	if !contains(got, "https://third.com/b.jpg") {
		t.Fatalf("ToSave 外链应透传: %s", got)
	}

	// 相对引用（. 开头）原样保留
	got = ToSave(`<img src="./attachment/c.jpg">`, conv, true)
	if !contains(got, `src=&quot;./attachment/c.jpg&quot;`) {
		t.Fatalf("ToSave 相对引用应原样保留: %s", got)
	}

	// 单引号 src（引号风格保留，编码为 &#039;）
	got = ToSave(`<img src='`+testDomain+`images/d.jpg'>`, conv, true)
	if !contains(got, `src=&#039;images/d.jpg&#039;`) {
		t.Fatalf("ToSave 单引号 src 应处理: %s", got)
	}

	// 已编码形态输入（存量库形态）：幂等
	encoded := `&lt;p&gt;&lt;img src=&quot;images/a.jpg&quot;&gt;&lt;/p&gt;`
	got = ToSave(encoded, conv, true)
	if got != encoded {
		t.Fatalf("ToSave 已编码输入应幂等:\n got: %s\nwant: %s", got, encoded)
	}

	// entities=false：仅去域名不编码
	got = ToSave(`<img src="`+testDomain+`images/e.jpg">`, conv, false)
	if got != `<img src="images/e.jpg">` {
		t.Fatalf("ToSave entities=false: %s", got)
	}

	// 空输入
	if ToSave("", conv, true) != "" {
		t.Fatal("ToSave 空输入应返回空串")
	}
}

func TestToShow(t *testing.T) {
	conv := &mockConverter{domain: testDomain}

	// 存量库形态（实体编码 + 相对路径）：解码 + 补域名
	got := ToShow(`&lt;p&gt;&lt;img src=&quot;images/a.jpg&quot;&gt;&lt;/p&gt;`, conv, false)
	want := `<p><img src="` + testDomain + `images/a.jpg"></p>`
	if got != want {
		t.Fatalf("ToShow 存量形态:\n got: %s\nwant: %s", got, want)
	}

	// 已是完整 URL 透传
	got = ToShow(`<img src="https://cdn.host/x.jpg">`, conv, false)
	if got != `<img src="https://cdn.host/x.jpg">` {
		t.Fatalf("ToShow 完整 URL 应透传: %s", got)
	}

	// entities=true：补域名后重新编码
	got = ToShow(`&lt;img src=&quot;images/a.jpg&quot;&gt;`, conv, true)
	if got != `&lt;img src=&quot;`+testDomain+`images/a.jpg&quot;&gt;` {
		t.Fatalf("ToShow entities=true: %s", got)
	}

	// 空/无 img 输入原样
	if ToShow("", conv, false) != "" {
		t.Fatal("ToShow 空输入应返回空串")
	}
	if got := ToShow(`<p>纯文本</p>`, conv, false); got != `<p>纯文本</p>` {
		t.Fatalf("ToShow 无 img 应原样: %s", got)
	}
}

func TestRoundTrip(t *testing.T) {
	conv := &mockConverter{domain: testDomain}

	raw := `<p><img style="max-width: 100%;" src="` + testDomain + `images/2026/11/a.jpg" alt="t&amp;t" width="495"></p>`
	saved := ToSave(raw, conv, true)
	shown := ToShow(saved, conv, false)
	if shown != raw {
		t.Fatalf("往返不一致:\n in:   %s\nshown: %s", raw, shown)
	}
}

func TestNilConverter(t *testing.T) {
	// conv 为 nil：仅实体转换，img src 原样
	got := ToSave(`<img src="`+testDomain+`images/a.jpg">`, nil, true)
	if got != `&lt;img src=&quot;`+testDomain+`images/a.jpg&quot;&gt;` {
		t.Fatalf("nil converter: %s", got)
	}
}

func TestHTMLEntities(t *testing.T) {
	cases := map[string]string{
		`<p>"a"&'b'</p>`: `&lt;p&gt;&quot;a&quot;&amp;&#039;b&#039;&lt;/p&gt;`,
		"":               "",
		"a<b>c":          "a&lt;b&gt;c",
	}
	for in, want := range cases {
		if got := HTMLEntities(in); got != want {
			t.Fatalf("HTMLEntities(%q) = %q, want %q", in, got, want)
		}
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
