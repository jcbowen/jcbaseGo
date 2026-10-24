package middleware

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jcbowen/jcbaseGo/component/helper"
)

type Base struct {
}

// Cors 开放所有接口的OPTIONS方法
func (b Base) Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")

		// 让缓存正确区分不同的跨域预检
		c.Header("Vary", "Origin, Access-Control-Request-Method, Access-Control-Request-Headers")

		// Origin 处理：回显具体来源以支持凭证；无来源则为 *
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		// 允许的方法
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

		// === 自定义允许的请求头 ===
		defaultAllowHeaders := "Content-Type, Authorization, X-Requested-With, Accept, Origin, Cache-Control, Pragma, X-Environment, X-Version, X-Request-Time, X-Api-Key, X-Resource-Version, JcClient, Referer, User-Agent, sec-ch-ua, sec-ch-ua-mobile, sec-ch-ua-platform"
		reqHeaders := strings.TrimSpace(c.Request.Header.Get("Access-Control-Request-Headers"))
		if reqHeaders != "" {
			// 返回客户端请求的头部，同时附加我们的默认列表（允许重复，浏览器会忽略）
			c.Header("Access-Control-Allow-Headers", reqHeaders+", "+defaultAllowHeaders)
		} else {
			c.Header("Access-Control-Allow-Headers", defaultAllowHeaders)
		}

		// 暴露的响应头（加入常用与自定义，便于前端读取）
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization, X-Environment, X-Version, X-Request-Time, X-Api-Key, X-Resource-Version, JcClient")

		// 预检缓存时间
		c.Header("Access-Control-Max-Age", "172800")

		// 仅当存在具体的 Origin 时允许携带凭证（与 * 不兼容）
		if origin != "" {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 允许私有网络的跨域（Chrome 新增，按需支持）
		if c.Request.Header.Get("Access-Control-Request-Private-Network") == "true" {
			c.Header("Access-Control-Allow-Private-Network", "true")
		}

		// 放行所有 OPTIONS 预检请求
		if method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 处理后续请求
		c.Next()
	}
}

// SetRealIP 设置真实IP到gin上下文
func (b Base) SetRealIP(useCDN bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		realIP := GetRealIP(c, useCDN)
		c.Set("RealIP", realIP)
		c.Next()
	}
}

// GetRealIP 获取真实客户端IP地址
// 当useCDN为true时，优先从代理头信息中获取IP；为false时，直接使用gin的ClientIP方法
// 参数:
//   - c: gin上下文
//   - useCDN: 是否使用CDN或代理，true表示使用，false表示不使用
//
// 返回值:
//   - realIP: 真实客户端IP地址
func GetRealIP(c *gin.Context, useCDN bool) (realIP string) {
	return GetRealIPWithFilter(c, useCDN, nil, nil)
}

// GetRealIPWithFilter 获取真实客户端IP地址（带过滤功能）
// 参数:
//   - c: gin上下文
//   - useCDN: 是否使用CDN或代理，true表示使用，false表示不使用
//   - whitelist: IP白名单列表，如果提供则只允许白名单内的IP
//   - blacklist: IP黑名单列表，如果提供则拒绝黑名单内的IP
//
// 返回值:
//   - realIP: 真实客户端IP地址
func GetRealIPWithFilter(c *gin.Context, useCDN bool, whitelist, blacklist []string) (realIP string) {
	// 如果没有使用CDN，直接返回gin的客户端IP
	if !useCDN {
		return c.ClientIP()
	}

	// 使用CDN时，优先从代理头信息中获取真实IP

	// 构建HTTP头信息映射
	headers := map[string]string{
		"Ali-Cdn-Real-Ip":     c.GetHeader("Ali-Cdn-Real-Ip"),
		"CF-Connecting-IP":    c.GetHeader("CF-Connecting-IP"),
		"True-Client-IP":      c.GetHeader("True-Client-IP"),
		"X-Client-IP":         c.GetHeader("X-Client-IP"),
		"X-Forwarded-For":     c.GetHeader("X-Forwarded-For"),
		"X-Real-IP":           c.GetHeader("X-Real-IP"),
		"X-Originating-IP":    c.GetHeader("X-Originating-IP"),
		"X-Cluster-Client-IP": c.GetHeader("X-Cluster-Client-IP"),
	}

	// 使用helper包中的函数获取真实IP
	realIP = helper.GetRealIPFromHeaders(headers)

	// 如果从头部获取的IP为空或无效，回退到gin的客户端IP
	if realIP == "" || helper.NewIP(realIP).IsLoopback() {
		realIP = c.ClientIP()
	}

	// 应用IP过滤规则
	if !helper.NewIP(realIP).IsAllowed(whitelist, blacklist) {
		// 如果IP被过滤，回退到gin的客户端IP
		realIP = c.ClientIP()
	}

	return
}

// SetGPC 把 query、cookie、header 与请求体参数汇总到上下文的 "GPC" 键
// 请求体按 Content-Type 分派解析，读到的字节一律回灌，业务 handler 仍可重复读取原始请求体
// 参数：
//   - e: gin 引擎，取用其 MaxMultipartMemory 作为 multipart 的内存上限
//
// 返回值：
//   - gin.HandlerFunc: 注册到引擎的中间件
//
// 异常：不触发 panic；请求体解析失败只记录日志并给出空的 data，GPC 结构始终完整写入
// 使用示例：
//
//	base := middleware.Base{}
//	e := gin.Default()
//	e.Use(base.SetGPC(e))
func (b Base) SetGPC(e *gin.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error

		// 初始化GPC map
		GPC := map[string]map[string]any{
			"query":  make(map[string]any),
			"header": make(map[string]any),
			"cookie": make(map[string]any),
			"data":   make(map[string]any),
			"all":    make(map[string]any), // 所有的参数都将被合并到这里，优先级为data\cookie\header\query
		}

		// 获取query
		mergeValues(GPC["query"], c.Request.URL.Query())

		// 获取Cookies
		for _, cookie := range c.Request.Cookies() {
			GPC["cookie"][cookie.Name] = cookie.Value
		}

		// 获取Headers
		for key, values := range c.Request.Header {
			if len(values) > 0 {
				GPC["header"][key] = values[0]
			}
		}

		formDataMap := make(map[string]any)

		// 获取请求数据并解析；每个分支都把请求体原文回灌给后续处理链，业务侧仍可读到完整字节
		switch c.ContentType() {
		case "application/json":
			formDataMap, err = parseJSONData(c)
		case "application/x-www-form-urlencoded":
			formDataMap, err = parseFormData(c)
		case "multipart/form-data":
			formDataMap, err = parseMultipartData(c, e.MaxMultipartMemory)
		case "text/xml", "application/xml":
			// 支持微信开放平台等XML格式请求
			formDataMap, err = parseXMLData(c)
		default:
			if c.ContentType() != "" {
				log.Println("Unsupported Content-Type：", c.ContentType())
			}
			/*err = gin.Error{
				Err:  errors.New("unsupported request method"),
				Type: gin.ErrorTypeBind,
			}*/
		}

		// 解析失败只降级为空参数并记录日志：GPC 必须完整写入，否则下游 c.MustGet("GPC") 会 panic；
		// 提前 return 拦不住业务 handler（gin 引擎层会继续走链），只会留下缺 key 的上下文
		if err != nil {
			log.Println("GPC error:", err)
		}

		GPC["data"] = formDataMap

		// 辅助函数，用于将数据合并到all字段中
		mergeToall := func(source map[string]any) {
			for key, value := range source {
				GPC["all"][key] = value
			}
		}
		mergeToall(GPC["query"])
		mergeToall(GPC["header"])
		mergeToall(GPC["cookie"])
		mergeToall(GPC["data"])

		// log.Println("GPC：", GPC)

		c.Set("GPC", GPC)

		// 处理请求
		c.Next()
	}
}

// gpcBodyParseLimit 是 SetGPC 解析请求体的字节上限，对齐 net/http parsePostForm 的 10MB 事实标准
// 超限的请求体只回灌不解析，避免大 POST 把整份 body 常驻内存
const gpcBodyParseLimit = int64(10 << 20)

// readBody 有界读取请求体原文，并把读到的字节原样回灌给后续处理链
// 参数：
//   - c: gin 上下文，c.Request.Body 会被替换成「已读片段 + 未读残余」的组合读取器
//
// 返回值：
//   - []byte: 请求体字节；空请求体或超过 gpcBodyParseLimit 时返回 nil（超限会记录日志）
//   - error: 读取错误；出错时已读片段同样完成回灌，不会让下游挂起
//
// 异常：不触发 panic
// 使用示例：
//
//	body, err := readBody(c)
func readBody(c *gin.Context) ([]byte, error) {
	if c.Request.Body == nil {
		return nil, nil
	}

	// 多读 1 字节用于区分「正好等于上限」与「超过上限」
	buf := &bytes.Buffer{}
	_, readErr := buf.ReadFrom(io.LimitReader(c.Request.Body, gpcBodyParseLimit+1))

	// 先回灌再判断：已读片段与未读残余都要还给后续处理链，验签、原文哈希这类
	// 场景依赖完整字节，出错路径同样不能吞掉
	c.Request.Body = io.NopCloser(io.MultiReader(bytes.NewReader(buf.Bytes()), c.Request.Body))

	if readErr != nil {
		return nil, readErr
	}

	size := buf.Len()
	if size == 0 {
		return nil, nil
	}
	if int64(size) > gpcBodyParseLimit {
		log.Printf("GPC: 请求体超过解析上限 %d 字节，只回灌不解析", gpcBodyParseLimit)
		return nil, nil
	}

	return buf.Bytes(), nil
}

// mergeValues 按 jcbaseGo 的参数约定把多值参数写入目标 map
// 键名以 [] 结尾时保留完整数组并去掉后缀，其余只取首个值；空值键跳过
// 参数：
//   - dst: 目标 map，即 GPC 的 query 或 data 分区
//   - src: 多值参数集合，query 传 url.Values，multipart 传 MultipartForm.Value
//
// 返回值：无
// 异常：不触发 panic
// 使用示例：
//
//	mergeValues(GPC["query"], c.Request.URL.Query())
func mergeValues(dst map[string]any, src map[string][]string) {
	for key, values := range src {
		if len(values) == 0 {
			continue
		}
		if strings.HasSuffix(key, "[]") {
			dst[strings.TrimSuffix(key, "[]")] = values
			continue
		}
		dst[key] = values[0]
	}
}

// parseJSONData 解析 application/json 请求体，读取后回灌原文
// 参数：
//   - c: gin 上下文
//
// 返回值：
//   - map[string]any: 解析出的参数；失败时为空 map
//   - error: 读取或反序列化错误；非对象 JSON（数组、标量）会返回类型错误
//
// 异常：不触发 panic
// 使用示例：
//
//	formDataMap, err = parseJSONData(c)
func parseJSONData(c *gin.Context) (map[string]any, error) {
	data := make(map[string]any)

	body, err := readBody(c)
	if err != nil || len(body) == 0 {
		return data, err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return data, err
	}
	return data, nil
}

// parseFormData 解析 application/x-www-form-urlencoded 请求体，读取后回灌原文
// 自行解析原文而不用 Request.ParseForm：后者会把 body 读干且不留下可回灌的字节
// 参数：
//   - c: gin 上下文
//
// 返回值：
//   - map[string]any: 解析出的参数；失败时为空 map
//   - error: 读取或 URL 解码错误，例如坏转义 a=%%zz
//
// 异常：不触发 panic
// 使用示例：
//
//	formDataMap, err = parseFormData(c)
func parseFormData(c *gin.Context) (map[string]any, error) {
	data := make(map[string]any)

	body, err := readBody(c)
	if err != nil || len(body) == 0 {
		return data, err
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return data, err
	}
	mergeValues(data, values)
	return data, nil
}

// parseMultipartData 解析 multipart/form-data 请求体
// 文本字段按 [] 约定展开，文件字段取 *multipart.FileHeader；不做原文回灌——
// 回灌需要整块缓冲上传文件，而 c.FormFile/MultipartForm 走的是已解析状态，换不到收益
// 参数：
//   - c: gin 上下文
//   - maxMemory: multipart 的内存上限，通常传 gin.Engine.MaxMultipartMemory
//
// 返回值：
//   - map[string]any: 解析出的参数；失败时为空 map
//   - error: 解析错误，例如 boundary 不匹配
//
// 异常：不触发 panic
// 使用示例：
//
//	formDataMap, err = parseMultipartData(c, e.MaxMultipartMemory)
func parseMultipartData(c *gin.Context, maxMemory int64) (map[string]any, error) {
	data := make(map[string]any)

	if err := c.Request.ParseMultipartForm(maxMemory); err != nil {
		return data, err
	}
	if c.Request.MultipartForm == nil {
		return data, nil
	}

	mergeValues(data, c.Request.MultipartForm.Value)

	// 处理文件上传字段
	for key := range c.Request.MultipartForm.File {
		data[key], _ = c.FormFile(key)
	}
	return data, nil
}

// parseXMLData 解析 text/xml 与 application/xml 请求体，读取后回灌原文
// 支持微信开放平台等场景，按元素名扁平化为字符串 map；嵌套元素只能取到空文本，属既有约定
// 参数：
//   - c: gin 上下文
//
// 返回值：
//   - map[string]any: 元素名到文本内容的 map；解析中断时返回已取到的部分
//   - error: 读取错误；XML 自身的 token/decode 错误只记日志，不向上抛
//
// 异常：不触发 panic
// 使用示例：
//
//	formDataMap, err = parseXMLData(c)
func parseXMLData(c *gin.Context) (map[string]any, error) {
	result := make(map[string]any)

	body, err := readBody(c)
	if err != nil || len(body) == 0 {
		return result, err
	}

	decoder := xml.NewDecoder(bytes.NewReader(body))

	// 解析XML到通用的map结构
	for {
		token, tokenErr := decoder.Token()
		if tokenErr == io.EOF {
			break
		}
		if tokenErr != nil {
			log.Println("XML token error:", tokenErr)
			break
		}

		se, ok := token.(xml.StartElement)
		if !ok || se.Name.Local == "xml" {
			continue
		}

		// 读取元素内容作为字符串
		var content string
		if decodeErr := decoder.DecodeElement(&content, &se); decodeErr != nil {
			log.Println("XML decode error:", decodeErr)
			continue
		}

		// 将元素名和内容添加到结果中
		result[se.Name.Local] = content
	}

	return result, nil
}
