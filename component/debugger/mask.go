package debugger

import (
	"net/url"
	"regexp"
	"strings"
)

// 凭据掩码：调试日志默认会把请求/响应原文入库并渲染到页面，
// 其中可能包含密码、验证码、令牌等敏感值。本文件在条目生成的出口统一做键名匹配替换
// （早于截断说明与落库），保证页面、导出、存储三条下游路径都拿不到原文。

// maskPlaceholder 掩码后的占位值
const maskPlaceholder = "***"

// defaultMaskHeaderKeys 内置默认掩码的请求/响应头名（大小写不敏感）
// 只保留头名、头值替换为 ***，便于确认「传了这个头」而不泄露内容
var defaultMaskHeaderKeys = []string{
	"authorization",
	"proxy-authorization",
	"cookie",
	"set-cookie",
	"x-api-key",
	"x-access-token",
	"x-auth-token",
}

// defaultMaskBodyKeys 内置默认掩码的字段名（大小写不敏感），作用于查询参数与请求/响应体
// 刻意不含裸 code：网关响应信封本身用 code 表示业务状态码，掩掉会让日志失去可读性；
// 验证码类字段用 sms_code / verify_code / verification_code / captcha 等具名键覆盖
var defaultMaskBodyKeys = []string{
	"password",
	"pwd",
	"passwd",
	"old_password",
	"new_password",
	"confirm_password",
	"re_password",
	"token",
	"access_token",
	"refresh_token",
	"id_token",
	"secret",
	"app_secret",
	"api_secret",
	"client_secret",
	"api_key",
	"apikey",
	"session_key",
	"session_id",
	"captcha",
	"sms_code",
	"verify_code",
	"verification_code",
	"id_card",
	"phone",
}

// 以下正则只编译一次，供逐请求调用复用
// 三处值捕获组（最后一个组）都包含需要整体替换的 span，替换文本由 maskValueSpan 依据
// 原值是否带引号决定，保证掩码后 JSON 结构仍然合法
var (
	// reMaskJSONString 匹配完整形式的字符串键值对："key": "value"（值组含首尾引号，值内允许转义引号）
	reMaskJSONString = regexp.MustCompile(`(?i)"([a-z0-9_$.\-]+)"(\s*:\s*)("(?:[^"\\]|\\.)*")`)
	// reMaskJSONStringTail 匹配因截断而未闭合的字符串值："key": "未闭合内容…（文本结尾）
	// 捕获窗口按上限截断时，凭据值可能被截半残留，故必须单独覆盖
	reMaskJSONStringTail = regexp.MustCompile(`(?i)"([a-z0-9_$.\-]+)"(\s*:\s*)("(?:[^"\\]|\\.)*)\z`)
	// reMaskJSONScalar 匹配数字/布尔/null 形式的键值对："key": 123456
	reMaskJSONScalar = regexp.MustCompile(`(?i)"([a-z0-9_$.\-]+)"(\s*:\s*)(-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][-+]?\d+)?|true|false|null)`)
	// reMaskMultipartField 匹配 multipart 摘要中的字段行：
	// [Field] Name: x, Size: n bytes, Value/Content: y
	// 已知局限：多行字段值只掩掉首行；本组件默认不记录部件内容（MultipartSkipFiles），
	// 该分支仅在调用方显式关闭时生效
	reMaskMultipartField = regexp.MustCompile(`(?m)^(\s*\[Field\] Name: )([^,]+?)(,[^\n]*?(?:Value|Content): ).*$`)
)

// buildMaskKeySet 归一化掩码键集合：内置默认 + 配置追加 - 配置排除，统一小写并去空白
// 参数：
//   - defaults: 内置默认键名
//   - extra: 调用方通过配置追加的键名（不覆盖默认值，只补充）
//   - exclusions: 调用方配置的排除键名，从并集中减去；优先级最高，用于放行被误伤的字段
//
// 返回值：
//   - map[string]struct{}: 供 O(1) 查表的集合，键已小写；空输入返回空集合（非 nil）
//
// 异常：不触发 panic
// 使用示例：
//
//	set := buildMaskKeySet(defaultMaskHeaderKeys, cfg.HeaderMaskKeys, cfg.HeaderMaskKeyExclusions)
func buildMaskKeySet(defaults, extra, exclusions []string) map[string]struct{} {
	set := make(map[string]struct{}, len(defaults)+len(extra))
	for _, list := range [][]string{defaults, extra} {
		for _, key := range list {
			normalized := strings.ToLower(strings.TrimSpace(key))
			if normalized != "" {
				set[normalized] = struct{}{}
			}
		}
	}
	for _, key := range exclusions {
		normalized := strings.ToLower(strings.TrimSpace(key))
		if normalized != "" {
			delete(set, normalized)
		}
	}
	return set
}

// ensureMaskSets 保证掩码键集合已构建（New 调用一次，sync.Once 保证并发安全且幂等）
func (d *Debugger) ensureMaskSets() {
	d.maskSetsOnce.Do(func() {
		var headerExtra, bodyExtra, headerExclusions, bodyExclusions []string
		if d.config != nil {
			headerExtra = d.config.HeaderMaskKeys
			bodyExtra = d.config.BodyMaskKeys
			headerExclusions = d.config.HeaderMaskKeyExclusions
			bodyExclusions = d.config.BodyMaskKeyExclusions
		}
		d.maskHeaderSet = buildMaskKeySet(defaultMaskHeaderKeys, headerExtra, headerExclusions)
		d.maskBodySet = buildMaskKeySet(defaultMaskBodyKeys, bodyExtra, bodyExclusions)
	})
}

// maskEnabled 返回是否启用凭据掩码
// BodyMaskEnabled 为 *bool 三态：nil（未配置）默认开启，显式 false 才关闭——
// 关闭会把凭据原文写入日志库与页面，仅建议本地排查一次性使用
func (d *Debugger) maskEnabled() bool {
	if d.config == nil {
		return true
	}
	d.ensureMaskSets()
	return d.config.BodyMaskEnabled == nil || *d.config.BodyMaskEnabled
}

// maskHeaders 就地掩码头值，保留头名
// 参数：
//   - headers: helper.ExtractHeaders 产出的头映射（每次调用新建，可安全就地修改）
//
// 返回值：
//   - map[string]string: 原映射（命中键的值替换为 ***）
//
// 异常：不触发 panic；入参为 nil 时原样返回
// 使用示例：
//
//	entry.RequestHeaders = d.maskHeaders(helper.ExtractHeaders(c.Request.Header))
func (d *Debugger) maskHeaders(headers map[string]string) map[string]string {
	if !d.maskEnabled() || len(d.maskHeaderSet) == 0 {
		return headers
	}
	for name := range headers {
		if headers[name] == "" {
			continue
		}
		if _, hit := d.maskHeaderSet[strings.ToLower(strings.TrimSpace(name))]; hit {
			headers[name] = maskPlaceholder
		}
	}
	return headers
}

// maskQueryParams 就地掩码命中的查询参数值
// 参数：
//   - params: 查询参数映射（键为原始参数名）
//
// 返回值：
//   - map[string]string: 原映射（命中键的值替换为 ***）
//
// 异常：不触发 panic
// 使用示例：
//
//	entry.QueryParams = d.maskQueryParams(extractQueryParams(c.Request.URL.Query()))
func (d *Debugger) maskQueryParams(params map[string]string) map[string]string {
	if !d.maskEnabled() || len(d.maskBodySet) == 0 {
		return params
	}
	for name := range params {
		if params[name] == "" {
			continue
		}
		if _, hit := d.maskBodySet[strings.ToLower(strings.TrimSpace(name))]; hit {
			params[name] = maskPlaceholder
		}
	}
	return params
}

// safeURL 生成可入库的请求 URL：命中掩码键的查询参数值替换为占位
// URL 列与请求级 logger 的 url 字段必须走本函数：GET 请求的凭据只体现在查询串上，
// 仅掩码 QueryParams 映射会让同一份原文留在 url 列和 logger_logs 里，检索明文即可命中
// 参数：
//   - u: 原始请求 URL（不修改入参）
//
// 返回值：
//   - string: 掩码后的 URL；掩码关闭、无查询串或无命中时与 u.String() 完全一致
//
// 异常：不触发 panic；u 为 nil 时返回空串
// 使用示例：
//
//	entry.URL = d.safeURL(c.Request.URL)
func (d *Debugger) safeURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	if u.RawQuery == "" || !d.maskEnabled() || len(d.maskBodySet) == 0 {
		return u.String()
	}
	masked := maskFormPairs(d.maskBodySet, u.RawQuery)
	if masked == u.RawQuery {
		return u.String()
	}
	clone := *u
	clone.RawQuery = masked
	return clone.String()
}

// maskBody 按内容类型掩码请求/响应体文本
// 调用时机必须在追加截断说明之前：捕获窗口按上限截断时凭据值可能被截半残留，
// 未闭合值只有在位于文本末尾时才会命中掩码规则，说明行先附上去就再也匹配不到
// 参数：
//   - text: 待掩码的正文
//   - contentType: 对应的 Content-Type 头值（可为空，空时按文本形态猜测）
//
// 返回值：
//   - string: 掩码后的正文；未命中或关闭掩码时原样返回
//
// 异常：不触发 panic（正则替换失败即返回原文，绝不改写结构）
// 使用示例：
//
//	entry.RequestBody = d.maskBody(body, c.Request.Header.Get("Content-Type"))
func (d *Debugger) maskBody(text, contentType string) string {
	if text == "" || !d.maskEnabled() {
		return text
	}

	lowerType := strings.ToLower(contentType)
	switch {
	case strings.Contains(lowerType, "multipart/"):
		text = maskMultipartFields(d.maskBodySet, text)
	case strings.Contains(lowerType, "application/x-www-form-urlencoded"), looksLikeForm(text):
		text = maskFormPairs(d.maskBodySet, text)
	}

	// 引号形态的键值对对三种内容类型都无害（form/multipart 文本不含 "key": 结构），
	// 统一在最后兜一层，覆盖未知 Content-Type 的正文
	return maskJSONPairs(d.maskBodySet, text)
}

// looksLikeForm 判断无 Content-Type 时文本是否为 key=value&key2=value2 形态
func looksLikeForm(text string) bool {
	if strings.ContainsRune(text, '{') || strings.Contains(text, "boundary=") {
		return false
	}
	first := text
	if idx := strings.IndexAny(text, "&"); idx >= 0 {
		first = text[:idx]
	}
	eq := strings.IndexByte(first, '=')
	if eq <= 0 {
		return false
	}
	_, err := url.QueryUnescape(first[:eq])
	return err == nil
}

// maskJSONPairs 掩码 JSON（及任意引号键值形态）正文中命中键的值
// 采用「只替换值片段」而非解析重排：保留原始缩进与格式，且天然兼容被截断的非法 JSON
// 参数：
//   - keys: 掩码键集合
//   - text: 正文
//
// 返回值：
//   - string: 掩码后的正文；JSON 结构保持合法
//
// 异常：不触发 panic
func maskJSONPairs(keys map[string]struct{}, text string) string {
	if len(keys) == 0 {
		return text
	}
	masked := replaceIfKeyHit(keys, reMaskJSONString, text, 1)
	masked = replaceIfKeyHit(keys, reMaskJSONScalar, masked, 1)
	return replaceIfKeyHit(keys, reMaskJSONStringTail, masked, 1)
}

// maskFormPairs 掩码 form-urlencoded 正文中命中键的值，保持原有顺序与编码形式
func maskFormPairs(keys map[string]struct{}, text string) string {
	if len(keys) == 0 {
		return text
	}
	pairs := strings.Split(text, "&")
	for i, pair := range pairs {
		eq := strings.IndexByte(pair, '=')
		if eq <= 0 {
			continue
		}
		name, err := url.QueryUnescape(pair[:eq])
		if err != nil {
			name = pair[:eq]
		}
		if _, hit := keys[strings.ToLower(strings.TrimSpace(name))]; hit {
			pairs[i] = pair[:eq+1] + maskPlaceholder
		}
	}
	return strings.Join(pairs, "&")
}

// maskMultipartFields 掩码 multipart 摘要文本中的字段行值（[Field] Name: x, Value: y）
// 文件部件在本组件默认配置下只记录元数据，不含内容，故无需处理
func maskMultipartFields(keys map[string]struct{}, text string) string {
	if len(keys) == 0 {
		return text
	}
	return reMaskMultipartField.ReplaceAllStringFunc(text, func(line string) string {
		groups := reMaskMultipartField.FindStringSubmatch(line)
		if groups == nil {
			return line
		}
		name := strings.ToLower(strings.TrimSpace(groups[2]))
		if _, hit := keys[name]; !hit {
			return line
		}
		return groups[1] + groups[2] + groups[3] + maskPlaceholder
	})
}

// replaceIfKeyHit 逐个命中做键名判定，命中则把值捕获组（固定为最后一个捕获组）替换为占位
// 参数：
//   - keys: 掩码键集合
//   - re: 需按「整体 + 键 + 分隔 + 值」四段捕获，即最后一个捕获组为待掩码值
//   - text: 正文
//   - keyGroup: 键名所在捕获组序号
//
// 返回值：
//   - string: 替换后的文本
//
// 异常：不触发 panic；捕获组不足时原样返回命中片段
func replaceIfKeyHit(keys map[string]struct{}, re *regexp.Regexp, text string, keyGroup int) string {
	return re.ReplaceAllStringFunc(text, func(match string) string {
		// 组数 = 整体 + 键 + 分隔 + 值 → 下标对至少 4 组 8 个元素
		indexes := re.FindStringSubmatchIndex(match)
		if len(indexes) < 8 {
			return match
		}
		key := strings.ToLower(match[indexes[2*keyGroup]:indexes[2*keyGroup+1]])
		if _, hit := keys[key]; !hit {
			return match
		}
		valueGroup := len(indexes)/2 - 1
		valueStart := indexes[2*valueGroup]
		return match[:valueStart] + maskValueSpan(match[valueStart:indexes[2*valueGroup+1]])
	})
}

// maskValueSpan 生成值 span 的替换文本
// 截断未闭合的字符串只补回开引号，维持原有的截断形态；
// 完整字符串与数字/布尔/null 一律替换为带引号占位，避免掩码后正文变成非法 JSON
func maskValueSpan(value string) string {
	if strings.HasPrefix(value, `"`) && !strings.HasSuffix(value, `"`) {
		return `"` + maskPlaceholder
	}
	return `"` + maskPlaceholder + `"`
}
