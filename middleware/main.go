package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
)

type Base struct {
}

// Cors 开放所有接口的OPTIONS方法
func (b Base) Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method               //请求方法
		origin := c.Request.Header.Get("Origin") //请求头部
		var headerKeys []string                  // 声明请求头keys
		for k := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("Access-Control-Allow-Origin, Access-Control-Allow-Headers, %s", headerStr)
		} else {
			headerStr = "Access-Control-Allow-Origin, Access-Control-Allow-Headers"
		}
		if origin != "" {
			origin := c.Request.Header.Get("Origin")
			c.Header("Access-Control-Allow-Origin", origin)                                    // 这是允许访问所有域
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE") //服务器支持的所有跨域请求的方法,为了避免浏览次请求的多次'预检'请求
			// header的类型
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma, Code")
			// 允许跨域设置 可以返回其他子段
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar") // 跨域关键设置 让浏览器可以解析
			c.Header("Access-Control-Max-Age", "172800")                                                                                                                                                           // 缓存请求信息 单位为秒
			c.Header("Access-Control-Allow-Credentials", "false")                                                                                                                                                  // 跨域请求是否需要带cookie信息 默认设置为true
			c.Set("Content-type", "application/json;charset=utf-8")                                                                                                                                                // 设置返回格式是json
		}

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}

		// 处理请求
		c.Next() // 处理请求
	}
}

// RealIP 获取真实IP
func (b Base) RealIP(useCDN bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		realIP := getRealIP(c, useCDN)
		c.Set("ClientIP", realIP)
		c.Next()
	}
}

// 如果开启了CDN之类的，获取真实IP需要从头部读取
func getRealIP(c *gin.Context, useCDN bool) (realIP string) {
	// 从上下文中获取客户端IP
	realIP = c.ClientIP()

	// 如果没有使用CDN或穿透什么的
	if !useCDN {
		return
	}

	// 尝试从 X-Forwarded-For 中获取
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For 可能包含多个IP地址，用逗号分隔
		ips := splitIps(xForwardedFor)
		if len(ips) > 0 {
			realIP = ips[0] // 获取第一个IP
			return
		}
	}

	// 尝试从 X-Real-IP 中获取
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" {
		realIP = xRealIP
	}

	return
}

// splitIps 分割多个IP
func splitIps(s string) []string {
	var result []string
	for _, item := range strings.Split(s, ",") {
		result = append(result, strings.TrimSpace(item))
	}
	return result
}

// SetGPC 设置响应头
func (b Base) SetGPC() gin.HandlerFunc {
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
			err = c.ShouldBindJSON(&formDataMap)
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
			err = c.Request.ParseMultipartForm(32 << 20) // 32MB
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
			}
		default:
			log.Println("Unsupported Content-Type：", c.ContentType())
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
