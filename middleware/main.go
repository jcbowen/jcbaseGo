package middleware

import (
	"bytes"
	"encoding/xml"
	"io"
	"log"
	"net/http"
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
		"X-Real-IP":        c.GetHeader("X-Real-IP"),
		"X-Forwarded-For":  c.GetHeader("X-Forwarded-For"),
		"X-Forwarded-Host": c.GetHeader("X-Forwarded-Host"),
		"X-Originating-IP": c.GetHeader("X-Originating-IP"),
		"True-Client-IP":   c.GetHeader("True-Client-IP"),
	}

	// 使用helper包中的函数获取真实IP
	realIP = helper.GetRealIPFromHeaders(headers)

	// 如果从头部获取的IP为空或无效，回退到gin的客户端IP
	if realIP == "" || !helper.NewIP(realIP).IsValid() {
		realIP = c.ClientIP()
	}

	// 应用IP过滤规则
	if !helper.NewIP(realIP).IsAllowed(whitelist, blacklist) {
		// 如果IP被过滤，回退到gin的客户端IP
		realIP = c.ClientIP()
	}

	return
}

// SetGPC 设置响应头
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
		queryParams := c.Request.URL.Query()
		for key, values := range queryParams {
			if len(values) > 0 {
				if strings.HasSuffix(key, "[]") {
					cleanKey := strings.TrimSuffix(key, "[]")
					GPC["query"][cleanKey] = values
				} else {
					//if len(values) > 1 {
					//	GPC["query"][key] = values
					//} else {
					GPC["query"][key] = values[0]
					//}
				}
			}
		}

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

		// 不是根目录也不是web目录的情况才输出为json格式
		formDataMap := make(map[string]any)

		// 获取请求数据并解析
		switch c.ContentType() {
		case "application/json":
			if c.Request.ContentLength > 0 {
				err = c.ShouldBindJSON(&formDataMap)
			}
		case "application/x-www-form-urlencoded":
			err = c.Request.ParseForm()
			if err == nil {
				for key, values := range c.Request.PostForm {
					if len(values) > 0 {
						if strings.HasSuffix(key, "[]") {
							cleanKey := strings.TrimSuffix(key, "[]")
							formDataMap[cleanKey] = values
						} else {
							//if len(values) > 1 {
							//	formDataMap[key] = values
							//} else {
							formDataMap[key] = values[0]
							//}
						}
					}
				}
			}
		case "multipart/form-data":
			err = c.Request.ParseMultipartForm(e.MaxMultipartMemory)
			if err == nil {
				for key, values := range c.Request.MultipartForm.Value {
					if len(values) > 0 {
						if strings.HasSuffix(key, "[]") {
							cleanKey := strings.TrimSuffix(key, "[]")
							formDataMap[cleanKey] = values
						} else {
							//if len(values) > 1 {
							//	formDataMap[key] = values
							//} else {
							formDataMap[key] = values[0]
							//}
						}
					}
				}

				// 处理文件上传字段
				for key := range c.Request.MultipartForm.File {
					formDataMap[key], _ = c.FormFile(key)
				}
			}
		case "text/xml", "application/xml":
			// 支持微信开放平台等XML格式请求
			if c.Request.ContentLength > 0 {
				// 读取原始数据，避免消耗body
				bodyData, err := c.GetRawData()
				if err != nil {
					log.Println("GetRawData error:", err)
					break
				}

				// 重置body，确保后续处理可以读取
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyData))

				result := make(map[string]any)
				decoder := xml.NewDecoder(bytes.NewReader(bodyData))

				// 解析XML到通用的map结构
				for {
					token, err := decoder.Token()
					if err == io.EOF {
						break
					}
					if err != nil {
						log.Println("XML token error:", err)
						break
					}

					switch se := token.(type) {
					case xml.StartElement:
						// 处理开始元素，跳过根元素
						if se.Name.Local == "xml" {
							continue
						}

						// 读取元素内容作为字符串
						var content string
						err = decoder.DecodeElement(&content, &se)
						if err != nil {
							log.Println("XML decode error:", err)
							continue
						}

						// 将元素名和内容添加到结果中
						result[se.Name.Local] = content
					}
				}
				formDataMap = result
			}
		default:
			if c.ContentType() != "" {
				log.Println("Unsupported Content-Type：", c.ContentType())
			}
			/*err = gin.Error{
				Err:  errors.New("unsupported request method"),
				Type: gin.ErrorTypeBind,
			}*/
		}

		if err != nil {
			log.Println("GPC error:", err)
			//c.AbortWithStatusJSON(http.StatusOK, jcbaseGo.Result{
			//	Code: errcode.BadRequest,
			//	Msg:  err.Error(),
			//})
			return
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
