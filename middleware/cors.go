package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// corsDefaultAllowHeaders 默认放行的请求头；客户端通过 Access-Control-Request-Headers
// 申请的头会反射追加在前（重复项浏览器会忽略）
const corsDefaultAllowHeaders = "Content-Type, Authorization, X-Requested-With, Accept, Origin, Cache-Control, Pragma, X-Environment, X-Version, X-Request-Time, X-Api-Key, X-Resource-Version, JcClient, Referer, User-Agent, sec-ch-ua, sec-ch-ua-mobile, sec-ch-ua-platform"

// corsExposeHeaders 暴露给前端 JS 可读的响应头
const corsExposeHeaders = "Content-Length, Content-Type, Authorization, X-Environment, X-Version, X-Request-Time, X-Api-Key, X-Resource-Version, JcClient"

// corsMaxAge 预检结果缓存时间（秒），48 小时
const corsMaxAge = "172800"

// CorsConfig 跨域中间件的配置项
type CorsConfig struct {
	// AllowAllOrigins 放行任意来源并回显 Origin（Cors() 的历史行为）。
	// 开启后 AllowOrigins 白名单被忽略
	AllowAllOrigins bool

	// AllowOrigins 来源白名单，仅在 AllowAllOrigins 为 false 时生效。
	// 支持两种写法：
	//   - 精确匹配：完整 Origin 字符串，含 scheme 与端口，如 "https://a.example.com:8443"
	//   - 子域通配符："*.example.com"，匹配 example.com 及其任意子域，scheme 不限
	// 白名单为空且未开启 AllowAllOrigins 时，任何来源都不会被授予 CORS 头
	AllowOrigins []string

	// AllowCredentials 是否允许携带 Cookie 等凭证。
	// 仅在来源被授予且 Allow-Origin 不是 "*" 时下发，避免与通配符冲突
	AllowCredentials bool
}

// Cors 放行任意来源的跨域中间件（历史行为，向后兼容保留）
// 等价于 CorsWithConfig(CorsConfig{AllowAllOrigins: true, AllowCredentials: true})
// 注意：该模式等价于"任意网站可携带 Cookie 跨域读写"，生产环境建议改用
// CorsWithConfig 配置白名单收紧
// 返回值：
//   - gin.HandlerFunc: 注册到引擎的中间件
//
// 使用示例：
//
//	base := middleware.Base{}
//	r.Use(base.Cors())
func (b Base) Cors() gin.HandlerFunc {
	return b.CorsWithConfig(CorsConfig{AllowAllOrigins: true, AllowCredentials: true})
}

// CorsWithConfig 按配置处理跨域的中间件
// 白名单未命中的来源不下发任何 CORS 响应头，由浏览器自行拦截（预检同样返回 204 但不带授权头），
// 服务端路由仍会正常执行，不在预检阶段回 403 以缩小探测面
// 参数：
//   - conf: CORS 配置，见 CorsConfig 各字段说明
//
// 返回值：
//   - gin.HandlerFunc: 注册到引擎的中间件
//
// 异常：不触发 panic
// 使用示例：
//
//	base := middleware.Base{}
//	r.Use(base.CorsWithConfig(middleware.CorsConfig{
//	    AllowOrigins:     []string{"https://www.example.com", "*.api.example.com"},
//	    AllowCredentials: true,
//	}))
func (b Base) CorsWithConfig(conf CorsConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 让缓存正确区分不同的跨域响应；用 Add 而非 set，
		// 避免覆盖压缩、静态文件等下游中间件已写入的 Vary
		c.Writer.Header().Add("Vary", "Origin")
		c.Writer.Header().Add("Vary", "Access-Control-Request-Method")
		c.Writer.Header().Add("Vary", "Access-Control-Request-Headers")

		if allowOrigin, granted := resolveCorsOrigin(origin, conf); granted {
			c.Header("Access-Control-Allow-Origin", allowOrigin)

			// 凭证与 "*" 不兼容，仅在回显具体来源时下发
			if conf.AllowCredentials && allowOrigin != "*" {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			// 允许的方法
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

			// 允许的请求头：反射客户端申请的头并附加默认列表（允许重复，浏览器会忽略）
			allowHeaders := corsDefaultAllowHeaders
			if reqHeaders := strings.TrimSpace(c.Request.Header.Get("Access-Control-Request-Headers")); reqHeaders != "" {
				allowHeaders = reqHeaders + ", " + corsDefaultAllowHeaders
			}
			c.Header("Access-Control-Allow-Headers", allowHeaders)

			// 暴露的响应头（便于前端读取自定义头）
			c.Header("Access-Control-Expose-Headers", corsExposeHeaders)

			// 预检缓存时间
			c.Header("Access-Control-Max-Age", corsMaxAge)

			// 允许私有网络的跨域（Chrome PNA，按需支持）
			if c.Request.Header.Get("Access-Control-Request-Private-Network") == "true" {
				c.Header("Access-Control-Allow-Private-Network", "true")
			}
		}

		// 放行 OPTIONS 预检请求：命中白名单时带授权头返回 204，
		// 未命中时不带授权头返回 204，由浏览器拦截
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 处理后续请求
		c.Next()
	}
}

// resolveCorsOrigin 判断请求是否被授予 CORS 头，并给出应回显的 Allow-Origin 值
// 参数：
//   - origin: 请求的 Origin 头，同源请求可能为空
//   - conf: CORS 配置
//
// 返回值：
//   - allowOrigin: 应写入 Access-Control-Allow-Origin 的值；未授予时为空串
//   - granted: 是否授予 CORS 头
func resolveCorsOrigin(origin string, conf CorsConfig) (allowOrigin string, granted bool) {
	// 全放行模式：回显具体来源以支持凭证；无 Origin（同源请求）则为 *
	if conf.AllowAllOrigins {
		if origin != "" {
			return origin, true
		}
		return "*", true
	}

	// 白名单模式：仅命中白名单的具体来源被授予
	if origin != "" && isOriginAllowed(origin, conf.AllowOrigins) {
		return origin, true
	}
	return "", false
}

// isOriginAllowed 判断 origin 是否命中白名单
// 精确匹配比较完整 Origin 字符串；"*.example.com" 形式匹配 example.com 及其任意子域（scheme 不限）
// 参数：
//   - origin: 请求的 Origin 头
//   - allowlist: 白名单模式串列表
//
// 返回值：
//   - bool: 是否命中
func isOriginAllowed(origin string, allowlist []string) bool {
	if len(allowlist) == 0 {
		return false
	}

	// 取 Origin 的主机名（去端口、转小写）用于通配符比较
	hostname := strings.ToLower(origin)
	if u, err := url.Parse(origin); err == nil && u.Hostname() != "" {
		hostname = u.Hostname()
	}

	for _, pattern := range allowlist {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}

		// 精确匹配（含 scheme 与端口）
		if pattern == origin {
			return true
		}

		// 子域通配符：*.example.com 匹配 example.com 及其任意子域
		if after, ok := strings.CutPrefix(pattern, "*."); ok {
			patternHost := after
			if hostname == patternHost || strings.HasSuffix(hostname, "."+patternHost) {
				return true
			}
		}
	}
	return false
}
