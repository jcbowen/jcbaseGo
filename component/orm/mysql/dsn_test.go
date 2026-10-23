package mysql

import (
	"net/url"
	"strings"
	"testing"

	"github.com/jcbowen/jcbaseGo"
)

// testDSNBase 构造一个最小可用的 DbStruct，供 DSN 相关用例复用。
func testDSNBase() jcbaseGo.DbStruct {
	return jcbaseGo.DbStruct{
		Username: "u",
		Password: "p",
		Protocol: "tcp",
		Host:     "127.0.0.1",
		Port:     "3306",
		Dbname:   "db",
		Charset:  "utf8mb4",
	}
}

// TestGetDSN_Loc 校验 DSN 中 loc 参数的生成规则。
//
// loc 决定驱动如何解释 DATETIME/TIMESTAMP：
//   - 未配置时必须回退为 Local，保证与历史行为逐字节一致（下游项目的时间语义依赖此约定）；
//   - 支持 Local / UTC（大小写不敏感）与 IANA 时区名；
//   - loc 是 DSN query 的参数值，必须无条件 url.QueryEscape。仅转义 "/" 是不够的：
//     "&" 会被当作参数分隔符导致取值截断，"%" 会让驱动的 url.QueryUnescape 直接失败。
func TestGetDSN_Loc(t *testing.T) {
	base := testDSNBase()

	cases := []struct {
		name      string
		parseTime string
		loc       string
		want      string
	}{
		{
			name:      "未配置Loc时回退Local（保持历史行为）",
			parseTime: "True",
			loc:       "",
			want:      "u:p@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true&loc=Local",
		},
		{
			name:      "显式Local",
			parseTime: "True",
			loc:       "Local",
			want:      "u:p@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true&loc=Local",
		},
		{
			name:      "小写local规范化为Local",
			parseTime: "True",
			loc:       "local",
			want:      "u:p@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true&loc=Local",
		},
		{
			name:      "显式UTC",
			parseTime: "True",
			loc:       "UTC",
			want:      "u:p@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true&loc=UTC",
		},
		{
			name:      "小写utc规范化为UTC",
			parseTime: "True",
			loc:       "utc",
			want:      "u:p@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true&loc=UTC",
		},
		{
			name:      "IANA时区名转义",
			parseTime: "True",
			loc:       "Asia/Shanghai",
			want:      "u:p@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai",
		},
		{
			name:      "IANA时区名含下划线同样只转义斜杠",
			parseTime: "True",
			loc:       "America/New_York",
			want:      "u:p@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true&loc=America%2FNew_York",
		},
		{
			name:      "字母数字型时区名不转义",
			parseTime: "True",
			loc:       "PRC",
			want:      "u:p@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true&loc=PRC",
		},
		{
			name:      "Loc含百分号需转义（否则驱动报invalid URL escape）",
			parseTime: "True",
			loc:       "100%",
			want:      "u:p@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true&loc=100%25",
		},
		{
			name:      "Loc含&需转义（否则取值被截断）",
			parseTime: "True",
			loc:       "a&b",
			want:      "u:p@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true&loc=a%26b",
		},
		{
			name:      "Loc含等号需转义",
			parseTime: "True",
			loc:       "a=b",
			want:      "u:p@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true&loc=a%3Db",
		},
		{
			name:      "Loc中间空格转义为加号",
			parseTime: "True",
			loc:       "a b",
			want:      "u:p@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true&loc=a+b",
		},
		{
			name:      "Loc含空格时裁剪",
			parseTime: "True",
			loc:       "  UTC  ",
			want:      "u:p@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true&loc=UTC",
		},
		{
			name:      "parseTime关闭时不受影响",
			parseTime: "False",
			loc:       "UTC",
			want:      "u:p@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=false&loc=UTC",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := base
			cfg.ParseTime = c.parseTime
			cfg.Loc = c.loc
			if got := getDSN(cfg); got != c.want {
				t.Errorf("getDSN() = %q, want %q", got, c.want)
			}
		})
	}
}

// TestGetDSN_LocRoundTrip 校验生成的 loc 参数能被 query 解码无损还原为原时区名。
//
// 驱动解析 DSN 时固定对 loc 执行 url.QueryUnescape，因此“编码后能否无损还原”
// 是 loc 处理是否正确的核心不变量。只在遇到 "/" 才转义的旧实现会在此暴露问题：
//   - 含 "%" 时 ParseQuery 直接失败；
//   - 含 "&" 时取值被静默截断（a&b 会被解析成 a）。
func TestGetDSN_LocRoundTrip(t *testing.T) {
	base := testDSNBase()

	cases := []struct {
		name string
		loc  string
		want string
	}{
		{name: "空值回退Local", loc: "", want: "Local"},
		{name: "Local", loc: "Local", want: "Local"},
		{name: "小写local", loc: "local", want: "Local"},
		{name: "UTC", loc: "UTC", want: "UTC"},
		{name: "小写utc", loc: "utc", want: "UTC"},
		{name: "两端空格裁剪", loc: "  UTC  ", want: "UTC"},
		{name: "PRC", loc: "PRC", want: "PRC"},
		{name: "Asia/Shanghai", loc: "Asia/Shanghai", want: "Asia/Shanghai"},
		{name: "America/New_York", loc: "America/New_York", want: "America/New_York"},
		{name: "百分号", loc: "100%", want: "100%"},
		{name: "与符号", loc: "a&b", want: "a&b"},
		{name: "等号", loc: "a=b", want: "a=b"},
		{name: "中间空格", loc: "a b", want: "a b"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := base
			cfg.Loc = c.loc
			dsn := getDSN(cfg)

			i := strings.Index(dsn, "?")
			if i < 0 {
				t.Fatalf("getDSN() 生成的 DSN 缺少 query 部分: %q", dsn)
			}

			q, err := url.ParseQuery(dsn[i+1:])
			if err != nil {
				t.Fatalf("loc 未正确转义，DSN query 无法解析: dsn=%q err=%v", dsn, err)
			}
			if got := q.Get("loc"); got != c.want {
				t.Errorf("loc 解码后 = %q, want %q (dsn=%q)", got, c.want, dsn)
			}
		})
	}
}

// TestGetDSN_ParseTime 校验 parsetime 的显式解析与「无法识别即保持历史口径」的兜底。
//
// parseTime 决定驱动把 DATETIME/TIMESTAMP 读成 time.Time 还是裸串，进而决定下游项目
// 模型 string 字段拿到的是 `2026-10-21 14:19:47` 还是 RFC3339；
// 「不等于 false 就开启」的老写法会让 Flase 这类拼错静默变成开启，
// 因此真假两种写法都必须认全，拼错时保持原行为并交由日志告警暴露。
func TestGetDSN_ParseTime(t *testing.T) {
	base := testDSNBase()

	cases := []struct {
		name       string
		parseTime  string
		wantParam  string
		recognized bool
	}{
		{name: "True", parseTime: "True", wantParam: "true", recognized: true},
		{name: "全小写true", parseTime: "true", wantParam: "true", recognized: true},
		{name: "全大写TRUE", parseTime: "TRUE", wantParam: "true", recognized: true},
		{name: "数字1", parseTime: "1", wantParam: "true", recognized: true},
		{name: "yes", parseTime: "yes", wantParam: "true", recognized: true},
		{name: "on", parseTime: "on", wantParam: "true", recognized: true},
		{name: "单字符t", parseTime: "t", wantParam: "true", recognized: true},
		{name: "单字符y", parseTime: "y", wantParam: "true", recognized: true},
		{name: "False", parseTime: "False", wantParam: "false", recognized: true},
		{name: "全大写FALSE", parseTime: "FALSE", wantParam: "false", recognized: true},
		{name: "数字0关闭（老实现会误判为开启）", parseTime: "0", wantParam: "false", recognized: true},
		{name: "no关闭（老实现会误判为开启）", parseTime: "no", wantParam: "false", recognized: true},
		{name: "off关闭（老实现会误判为开启）", parseTime: "off", wantParam: "false", recognized: true},
		{name: "单字符f关闭（老实现会误判为开启）", parseTime: "f", wantParam: "false", recognized: true},
		{name: "单字符n关闭（老实现会误判为开启）", parseTime: "n", wantParam: "false", recognized: true},
		{name: "两端空格裁剪", parseTime: "  False  ", wantParam: "false", recognized: true},
		{name: "空值保持开启且不判为配置错误", parseTime: "", wantParam: "true", recognized: true},
		{name: "拼错的Flase保持历史口径并标记待告警", parseTime: "Flase", wantParam: "true", recognized: false},
		{name: "越界取值2保持历史口径并标记待告警", parseTime: "2", wantParam: "true", recognized: false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := base
			cfg.ParseTime = c.parseTime
			got := getDSN(cfg)
			if !strings.Contains(got, "parseTime="+c.wantParam) {
				t.Errorf("getDSN() = %q, want parseTime=%s", got, c.wantParam)
			}
			if _, recognized := normalizeParseTime(c.parseTime); recognized != c.recognized {
				t.Errorf("normalizeParseTime(%q) 可识别性 = %v, want %v", c.parseTime, recognized, c.recognized)
			}
		})
	}
}
