package helper

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// ExtractHeaders 提取HTTP头信息并转换为map[string]string格式
// 该方法将http.Header中的每个键值对转换为字符串，多个值用逗号分隔
//
// 参数:
//
//	header - HTTP请求头
//
// 返回值:
//
//	map[string]string - 转换后的头信息
func ExtractHeaders(header http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range header {
		result[key] = strings.Join(values, ", ")
	}
	return result
}

// GetHostInfo 从HTTP请求中获取主机信息
// 该方法会检查TLS状态和X-Scheme头来确定协议，并优先使用代理头获取主机名
//
// 参数:
//
//	req - HTTP请求
//
// 返回值:
//
//	string - 完整的主机信息（协议://主机:端口）
func GetHostInfo(req *http.Request) string {
	// 确定协议
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}

	// 检查X-Scheme头
	if xScheme := req.Header.Get("X-Scheme"); xScheme != "" {
		scheme = xScheme
	}

	// 获取主机名，优先使用代理头
	host := req.Host
	if xForwardedHost := req.Header.Get("X-Forwarded-Host"); xForwardedHost != "" {
		host = xForwardedHost
	} else if xOriginalHost := req.Header.Get("X-Original-Host"); xOriginalHost != "" {
		host = xOriginalHost
	} else if xHost := req.Header.Get("X-Host"); xHost != "" {
		host = xHost
	}

	// 如果host为空，使用URL中的主机
	if host == "" && req.URL != nil {
		host = req.URL.Host
	}

	// 移除默认端口
	if (scheme == "http" && strings.HasSuffix(host, ":80")) ||
		(scheme == "https" && strings.HasSuffix(host, ":443")) {
		host = strings.Split(host, ":")[0]
	}

	return scheme + "://" + host
}

// FormatHeaders 格式化HTTP头信息为可读的字符串格式
// 该方法将map[string]string格式的头信息转换为键值对字符串，每行一个头信息
//
// 参数:
//
//	headers - HTTP头信息
//
// 返回值:
//
//	string - 格式化后的头信息字符串
func FormatHeaders(headers map[string]string) string {
	if len(headers) == 0 {
		return "无"
	}

	var buf bytes.Buffer
	for key, value := range headers {
		buf.WriteString(fmt.Sprintf("%s: %s\n", key, html.EscapeString(value)))
	}

	return buf.String()
}

// GetFullURL 获取完整的URL（包括协议、主机、路径和查询参数）
// 参数:
//   - req: HTTP请求对象
//
// 返回值:
//   - string: 完整的URL字符串
func GetFullURL(req *http.Request) string {
	if req == nil {
		return ""
	}

	host := GetHostInfo(req)

	// 构建完整URL
	return fmt.Sprintf("%s%s", host, req.URL.RequestURI())
}

// URLEncode 对字符串进行URL编码
// 参数:
//   - s: 需要编码的字符串
//
// 返回值:
//   - string: URL编码后的字符串
func URLEncode(s string) string {
	return url.QueryEscape(s)
}

// URLDecode 对URL编码的字符串进行解码
// 参数:
//   - s: URL编码的字符串
//
// 返回值:
//   - string: 解码后的字符串
//   - error: 解码错误
func URLDecode(s string) (string, error) {
	return url.QueryUnescape(s)
}

// IsAjaxRequest 检查是否为Ajax请求
// 参数:
//   - req: HTTP请求对象
//
// 返回值:
//   - bool: 如果是Ajax请求返回true，否则返回false
func IsAjaxRequest(req *http.Request) bool {
	if req == nil {
		return false
	}
	return req.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

// IsJSONRequest 检查是否为JSON请求
// 参数:
//   - req: HTTP请求对象
//
// 返回值:
//   - bool: 如果是JSON请求返回true，否则返回false
func IsJSONRequest(req *http.Request) bool {
	if req == nil {
		return false
	}
	contentType := req.Header.Get("Content-Type")
	return strings.Contains(contentType, "application/json")
}

// IsFormRequest 检查是否为表单请求
// 参数:
//   - req: HTTP请求对象
//
// 返回值:
//   - bool: 如果是表单请求返回true，否则返回false
func IsFormRequest(req *http.Request) bool {
	if req == nil {
		return false
	}
	contentType := req.Header.Get("Content-Type")
	return strings.Contains(contentType, "application/x-www-form-urlencoded") ||
		strings.Contains(contentType, "multipart/form-data")
}

// GetQueryParam 获取URL查询参数
// 参数:
//   - req: HTTP请求对象
//   - key: 参数键名
//
// 返回值:
//   - string: 参数值，如果不存在则返回空字符串
func GetQueryParam(req *http.Request, key string) string {
	if req == nil || req.URL == nil {
		return ""
	}
	return req.URL.Query().Get(key)
}

// GetQueryParams 获取所有URL查询参数
// 参数:
//   - req: HTTP请求对象
//
// 返回值:
//   - map[string][]string: 所有查询参数
func GetQueryParams(req *http.Request) map[string][]string {
	if req == nil || req.URL == nil {
		return make(map[string][]string)
	}
	return req.URL.Query()
}

// GetHeader 获取HTTP请求头信息
// 参数:
//   - req: HTTP请求对象
//   - key: 头信息键名
//
// 返回值:
//   - string: 头信息值，如果不存在则返回空字符串
func GetHeader(req *http.Request, key string) string {
	if req == nil {
		return ""
	}
	return req.Header.Get(key)
}

// GetCookie 获取HTTP Cookie值
// 参数:
//   - req: HTTP请求对象
//   - name: Cookie名称
//
// 返回值:
//   - string: Cookie值，如果不存在则返回空字符串
func GetCookie(req *http.Request, name string) string {
	if req == nil {
		return ""
	}
	cookie, err := req.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// BuildQueryString 构建查询字符串
// 参数:
//   - params: 参数映射
//
// 返回值:
//   - string: 查询字符串
func BuildQueryString(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}

	values := url.Values{}
	for key, value := range params {
		values.Add(key, value)
	}

	return values.Encode()
}

// ParseQueryString 解析查询字符串
// 参数:
//   - query: 查询字符串
//
// 返回值:
//   - map[string]string: 解析后的参数映射
func ParseQueryString(query string) map[string]string {
	result := make(map[string]string)

	values, err := url.ParseQuery(query)
	if err != nil {
		return result
	}

	for key, value := range values {
		if len(value) > 0 {
			result[key] = value[0]
		}
	}

	return result
}

// IsHTTPS 检查是否为HTTPS请求
// 参数:
//   - req: HTTP请求对象
//
// 返回值:
//   - bool: 如果是HTTPS请求返回true，否则返回false
func IsHTTPS(req *http.Request) bool {
	if req == nil {
		return false
	}

	if req.TLS != nil {
		return true
	}

	// 检查代理头
	if req.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}

	if req.Header.Get("X-Forwarded-Ssl") == "on" {
		return true
	}

	return false
}

// GetHostname 获取主机名（不含端口）
// 参数:
//   - req: HTTP请求对象
//
// 返回值:
//   - string: 主机名
func GetHostname(req *http.Request) string {
	if req == nil {
		return ""
	}

	host := req.Host
	if host == "" && req.URL != nil {
		host = req.URL.Host
	}

	// 移除端口
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	return host
}

// GetPort 获取端口号
// 参数:
//   - req: HTTP请求对象
//
// 返回值:
//   - string: 端口号
func GetPort(req *http.Request) string {
	if req == nil {
		return ""
	}

	host := req.Host
	if host == "" && req.URL != nil {
		host = req.URL.Host
	}

	// 提取端口
	if idx := strings.Index(host, ":"); idx != -1 {
		return host[idx+1:]
	}

	// 默认端口
	if IsHTTPS(req) {
		return "443"
	}
	return "80"
}

// SetHeader 设置HTTP响应头
// 参数:
//   - w: HTTP响应写入器
//   - key: 头信息键名
//   - value: 头信息值
func SetHeader(w http.ResponseWriter, key, value string) {
	if w != nil {
		w.Header().Set(key, value)
	}
}

// AddHeader 添加HTTP响应头
// 参数:
//   - w: HTTP响应写入器
//   - key: 头信息键名
//   - value: 头信息值
func AddHeader(w http.ResponseWriter, key, value string) {
	if w != nil {
		w.Header().Add(key, value)
	}
}

// SetContentType 设置内容类型
// 参数:
//   - w: HTTP响应写入器
//   - contentType: 内容类型
func SetContentType(w http.ResponseWriter, contentType string) {
	SetHeader(w, "Content-Type", contentType)
}

// SetJSONContentType 设置JSON内容类型
// 参数:
//   - w: HTTP响应写入器
func SetJSONContentType(w http.ResponseWriter) {
	SetContentType(w, "application/json; charset=utf-8")
}

// SetHTMLContentType 设置HTML内容类型
// 参数:
//   - w: HTTP响应写入器
func SetHTMLContentType(w http.ResponseWriter) {
	SetContentType(w, "text/html; charset=utf-8")
}

// SetTextContentType 设置文本内容类型
// 参数:
//   - w: HTTP响应写入器
func SetTextContentType(w http.ResponseWriter) {
	SetContentType(w, "text/plain; charset=utf-8")
}

// SetStatusCode 设置HTTP状态码
// 参数:
//   - w: HTTP响应写入器
//   - code: 状态码
func SetStatusCode(w http.ResponseWriter, code int) {
	if w != nil {
		w.WriteHeader(code)
	}
}

// Redirect 重定向到指定URL
// 参数:
//   - w: HTTP响应写入器
//   - req: HTTP请求对象
//   - url: 重定向URL
//   - code: 重定向状态码（默认302）
func Redirect(w http.ResponseWriter, req *http.Request, url string, code int) {
	if w == nil {
		return
	}

	if code == 0 {
		code = http.StatusFound
	}

	http.Redirect(w, req, url, code)
}

// WriteJSON 写入JSON响应
// 参数:
//   - w: HTTP响应写入器
//   - data: 要写入的数据
//
// 返回值:
//   - error: 写入错误
func WriteJSON(w http.ResponseWriter, data interface{}) error {
	SetJSONContentType(w)
	return json.NewEncoder(w).Encode(data)
}

// WriteJSONWithStatus 写入JSON响应并设置状态码
// 参数:
//   - w: HTTP响应写入器
//   - code: 状态码
//   - data: 要写入的数据
//
// 返回值:
//   - error: 写入错误
func WriteJSONWithStatus(w http.ResponseWriter, code int, data interface{}) error {
	SetStatusCode(w, code)
	return WriteJSON(w, data)
}

// WriteString 写入字符串响应
// 参数:
//   - w: HTTP响应写入器
//   - s: 要写入的字符串
//
// 返回值:
//   - int: 写入的字节数
//   - error: 写入错误
func WriteString(w http.ResponseWriter, s string) (int, error) {
	SetTextContentType(w)
	return w.Write([]byte(s))
}

// WriteStringWithStatus 写入字符串响应并设置状态码
// 参数:
//   - w: HTTP响应写入器
//   - code: 状态码
//   - s: 要写入的字符串
//
// 返回值:
//   - int: 写入的字节数
//   - error: 写入错误
func WriteStringWithStatus(w http.ResponseWriter, code int, s string) (int, error) {
	SetStatusCode(w, code)
	return WriteString(w, s)
}

// WriteHTML 写入HTML响应
// 参数:
//   - w: HTTP响应写入器
//   - html: 要写入的HTML内容
//
// 返回值:
//   - int: 写入的字节数
//   - error: 写入错误
func WriteHTML(w http.ResponseWriter, html string) (int, error) {
	SetHTMLContentType(w)
	return w.Write([]byte(html))
}

// WriteHTMLWithStatus 写入HTML响应并设置状态码
// 参数:
//   - w: HTTP响应写入器
//   - code: 状态码
//   - html: 要写入的HTML内容
//
// 返回值:
//   - int: 写入的字节数
//   - error: 写入错误
func WriteHTMLWithStatus(w http.ResponseWriter, code int, html string) (int, error) {
	SetStatusCode(w, code)
	return WriteHTML(w, html)
}

// GetFileExtension 获取文件扩展名
// 参数:
//   - filename: 文件名
//
// 返回值:
//   - string: 文件扩展名
func GetFileExtension(filename string) string {
	if filename == "" {
		return ""
	}

	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

// GetMimeType 根据文件扩展名获取MIME类型
// 参数:
//   - extension: 文件扩展名
//
// 返回值:
//   - string: MIME类型
func GetMimeType(extension string) string {
	mimeTypes := map[string]string{
		"html": "text/html",
		"htm":  "text/html",
		"css":  "text/css",
		"js":   "application/javascript",
		"json": "application/json",
		"xml":  "application/xml",
		"txt":  "text/plain",
		"pdf":  "application/pdf",
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"svg":  "image/svg+xml",
		"ico":  "image/x-icon",
		"zip":  "application/zip",
		"rar":  "application/x-rar-compressed",
		"doc":  "application/msword",
		"docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"xls":  "application/vnd.ms-excel",
		"xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"ppt":  "application/vnd.ms-powerpoint",
		"pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	}

	if mimeType, exists := mimeTypes[strings.ToLower(extension)]; exists {
		return mimeType
	}

	return "application/octet-stream"
}

// SetContentDisposition 设置内容处置头（用于文件下载）
// 参数:
//   - w: HTTP响应写入器
//   - filename: 文件名
//   - dispositionType: 处置类型（"attachment" 或 "inline"）
func SetContentDisposition(w http.ResponseWriter, filename, dispositionType string) {
	if dispositionType == "" {
		dispositionType = "attachment"
	}

	headerValue := fmt.Sprintf("%s; filename=\"%s\"", dispositionType, filename)
	SetHeader(w, "Content-Disposition", headerValue)
}

// ParseMultipartForm 解析多部分表单数据
// 参数:
//   - req: HTTP请求对象
//   - maxMemory: 最大内存大小（字节）
//
// 返回值:
//   - error: 解析错误
func ParseMultipartForm(req *http.Request, maxMemory int64) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}

	if maxMemory <= 0 {
		maxMemory = 32 << 20 // 32MB
	}

	err := req.ParseMultipartForm(maxMemory)
	if err != nil {
		return err
	}

	return nil
}

// GetFormValue 获取表单值
// 参数:
//   - req: HTTP请求对象
//   - key: 表单键名
//
// 返回值:
//   - string: 表单值
func GetFormValue(req *http.Request, key string) string {
	if req == nil {
		return ""
	}
	return req.FormValue(key)
}

// GetFormFile 获取上传的文件
// 参数:
//   - req: HTTP请求对象
//   - key: 文件字段名
//
// 返回值:
//   - multipart.File: 文件对象
//   - *multipart.FileHeader: 文件头信息
//   - error: 获取错误
func GetFormFile(req *http.Request, key string) (multipart.File, *multipart.FileHeader, error) {
	if req == nil {
		return nil, nil, fmt.Errorf("request is nil")
	}
	return req.FormFile(key)
}

// SaveUploadedFile 保存上传的文件
// 参数:
//   - fileHeader: 文件头信息
//   - dst: 目标文件路径
//
// 返回值:
//   - error: 保存错误
func SaveUploadedFile(fileHeader *multipart.FileHeader, dst string) error {
	if fileHeader == nil {
		return fmt.Errorf("file header is nil")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, src)
	return err
}

// IsMobileUserAgent 检查是否为移动设备User-Agent
// 参数:
//   - userAgent: User-Agent字符串
//
// 返回值:
//   - bool: 如果是移动设备返回true，否则返回false
func IsMobileUserAgent(userAgent string) bool {
	if userAgent == "" {
		return false
	}

	mobileKeywords := []string{
		"Mobile", "Android", "iPhone", "iPad", "iPod", "BlackBerry",
		"Windows Phone", "Opera Mini", "IEMobile", "Mobile Safari",
	}

	userAgent = strings.ToLower(userAgent)
	for _, keyword := range mobileKeywords {
		if strings.Contains(userAgent, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

// IsBotUserAgent 检查是否为爬虫/机器人User-Agent
// 参数:
//   - userAgent: User-Agent字符串
//
// 返回值:
//   - bool: 如果是爬虫返回true，否则返回false
func IsBotUserAgent(userAgent string) bool {
	if userAgent == "" {
		return false
	}

	botKeywords := []string{
		"bot", "crawler", "spider", "scraper", "slurp", "search",
		"googlebot", "bingbot", "baiduspider", "yandexbot", "duckduckbot",
		"facebookexternalhit", "twitterbot", "linkedinbot", "telegrambot",
	}

	userAgent = strings.ToLower(userAgent)
	for _, keyword := range botKeywords {
		if strings.Contains(userAgent, keyword) {
			return true
		}
	}

	return false
}

// GenerateETag 生成ETag
// 参数:
//   - content: 内容字节数组
//
// 返回值:
//   - string: ETag字符串
func GenerateETag(content []byte) string {
	if len(content) == 0 {
		return ""
	}

	hash := sha1.Sum(content)
	return fmt.Sprintf("\"%x\"", hash)
}

// CheckETag 检查ETag是否匹配
// 参数:
//   - req: HTTP请求对象
//   - etag: 当前ETag
//
// 返回值:
//   - bool: 如果ETag匹配返回true，否则返回false
func CheckETag(req *http.Request, etag string) bool {
	if req == nil || etag == "" {
		return false
	}

	ifNoneMatch := req.Header.Get("If-None-Match")
	return ifNoneMatch == etag
}

// SetCacheHeaders 设置缓存头
// 参数:
//   - w: HTTP响应写入器
//   - maxAge: 最大缓存时间（秒）
//   - public: 是否公开缓存
func SetCacheHeaders(w http.ResponseWriter, maxAge int, public bool) {
	if w == nil {
		return
	}

	cacheControl := fmt.Sprintf("max-age=%d", maxAge)
	if public {
		cacheControl = "public, " + cacheControl
	} else {
		cacheControl = "private, " + cacheControl
	}

	SetHeader(w, "Cache-Control", cacheControl)
	SetHeader(w, "Expires", time.Now().Add(time.Duration(maxAge)*time.Second).UTC().Format(time.RFC1123))
}

// SetNoCacheHeaders 设置无缓存头
// 参数:
//   - w: HTTP响应写入器
func SetNoCacheHeaders(w http.ResponseWriter) {
	if w == nil {
		return
	}

	SetHeader(w, "Cache-Control", "no-cache, no-store, must-revalidate")
	SetHeader(w, "Pragma", "no-cache")
	SetHeader(w, "Expires", "0")
}

// HTTPClientConfig HTTP客户端配置
type HTTPClientConfig struct {
	Timeout    time.Duration
	ProxyURL   string
	UserAgent  string
	SkipVerify bool
}

// NewHTTPClient 创建新的HTTP客户端
// 参数:
//   - config: 客户端配置
//
// 返回值:
//   - *http.Client: HTTP客户端
func NewHTTPClient(config *HTTPClientConfig) *http.Client {
	if config == nil {
		config = &HTTPClientConfig{}
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	// 设置代理
	if config.ProxyURL != "" {
		proxyURL, err := url.Parse(config.ProxyURL)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	// 跳过TLS验证
	if config.SkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{
		Transport: transport,
	}

	// 设置超时
	if config.Timeout > 0 {
		client.Timeout = config.Timeout
	}

	return client
}

// HTTPRequestConfig HTTP请求配置
type HTTPRequestConfig struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
	Timeout time.Duration
}

// SendHTTPRequest 发送HTTP请求
// 参数:
//   - config: 请求配置
//
// 返回值:
//   - *http.Response: HTTP响应
//   - error: 请求错误
func SendHTTPRequest(config *HTTPRequestConfig) (*http.Response, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	if config.Method == "" {
		config.Method = "GET"
	}

	req, err := http.NewRequest(config.Method, config.URL, bytes.NewReader(config.Body))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// 创建客户端
	client := &http.Client{}
	if config.Timeout > 0 {
		client.Timeout = config.Timeout
	}

	return client.Do(req)
}

// GetHTTP 发送GET请求
// 参数:
//   - url: 请求URL
//   - headers: 请求头
//
// 返回值:
//   - *http.Response: HTTP响应
//   - error: 请求错误
func GetHTTP(url string, headers map[string]string) (*http.Response, error) {
	config := &HTTPRequestConfig{
		Method:  "GET",
		URL:     url,
		Headers: headers,
	}
	return SendHTTPRequest(config)
}

// PostHTTP 发送POST请求
// 参数:
//   - url: 请求URL
//   - body: 请求体
//   - headers: 请求头
//
// 返回值:
//   - *http.Response: HTTP响应
//   - error: 请求错误
func PostHTTP(url string, body []byte, headers map[string]string) (*http.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}

	if _, exists := headers["Content-Type"]; !exists {
		headers["Content-Type"] = "application/json"
	}

	config := &HTTPRequestConfig{
		Method:  "POST",
		URL:     url,
		Body:    body,
		Headers: headers,
	}
	return SendHTTPRequest(config)
}

// PostFormHTTP 发送表单POST请求
// 参数:
//   - urlStr: 请求URL
//   - data: 表单数据
//   - headers: 请求头
//
// 返回值:
//   - *http.Response: HTTP响应
//   - error: 请求错误
func PostFormHTTP(urlStr string, data map[string]string, headers map[string]string) (*http.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}

	headers["Content-Type"] = "application/x-www-form-urlencoded"

	// 手动构建表单数据
	var formData strings.Builder
	first := true
	for key, value := range data {
		if !first {
			formData.WriteByte('&')
		}
		first = false
		formData.WriteString(url.QueryEscape(key))
		formData.WriteByte('=')
		formData.WriteString(url.QueryEscape(value))
	}

	config := &HTTPRequestConfig{
		Method:  "POST",
		URL:     urlStr,
		Body:    []byte(formData.String()),
		Headers: headers,
	}
	return SendHTTPRequest(config)
}

// ReadResponseBody 读取HTTP响应体
// 参数:
//   - resp: HTTP响应
//
// 返回值:
//   - []byte: 响应体内容
//   - error: 读取错误
func ReadResponseBody(resp *http.Response) ([]byte, error) {
	if resp == nil {
		return nil, fmt.Errorf("response is nil")
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// ReadResponseBodyString 读取HTTP响应体为字符串
// 参数:
//   - resp: HTTP响应
//
// 返回值:
//   - string: 响应体字符串
//   - error: 读取错误
func ReadResponseBodyString(resp *http.Response) (string, error) {
	body, err := ReadResponseBody(resp)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// DownloadFile 下载文件
// 参数:
//   - url: 文件URL
//   - filepath: 保存路径
//
// 返回值:
//   - error: 下载错误
func DownloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// ValidateURL 验证URL格式
// 参数:
//   - urlStr: URL字符串
//
// 返回值:
//   - bool: 如果是有效URL返回true，否则返回false
func ValidateURL(urlStr string) bool {
	_, err := url.ParseRequestURI(urlStr)
	return err == nil
}

// NormalizeURL 规范化URL
// 参数:
//   - urlStr: URL字符串
//
// 返回值:
//   - string: 规范化后的URL
//   - error: 规范化错误
func NormalizeURL(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	if u.Scheme == "" {
		u.Scheme = "http"
	}

	return u.String(), nil
}

// ExtractDomain 提取域名
// 参数:
//   - urlStr: URL字符串
//
// 返回值:
//   - string: 域名
//   - error: 提取错误
func ExtractDomain(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	return u.Hostname(), nil
}

// ExtractPath 提取路径
// 参数:
//   - urlStr: URL字符串
//
// 返回值:
//   - string: 路径
//   - error: 提取错误
func ExtractPath(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	return u.Path, nil
}

// ExtractQueryParams 提取查询参数
// 参数:
//   - urlStr: URL字符串
//
// 返回值:
//   - map[string]string: 查询参数映射
//   - error: 提取错误
func ExtractQueryParams(urlStr string) (map[string]string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for key, values := range u.Query() {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}

	return result, nil
}

// BuildURL 构建URL
// 参数:
//   - scheme: 协议
//   - host: 主机
//   - path: 路径
//   - query: 查询参数
//
// 返回值:
//   - string: 构建的URL
func BuildURL(scheme, host, path string, query map[string]string) string {
	u := &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}

	if query != nil {
		q := u.Query()
		for key, value := range query {
			q.Add(key, value)
		}
		u.RawQuery = q.Encode()
	}

	return u.String()
}

// IsURLSecure 检查URL是否安全（HTTPS）
// 参数:
//   - urlStr: URL字符串
//
// 返回值:
//   - bool: 如果是HTTPS返回true，否则返回false
func IsURLSecure(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	return u.Scheme == "https"
}

// GetHTTPStatusText 获取HTTP状态码文本描述
// 参数:
//   - code: HTTP状态码
//
// 返回值:
//   - string: 状态码文本描述
func GetHTTPStatusText(code int) string {
	statusTexts := map[int]string{
		100: "Continue",
		101: "Switching Protocols",
		200: "OK",
		201: "Created",
		202: "Accepted",
		203: "Non-Authoritative Information",
		204: "No Content",
		205: "Reset Content",
		206: "Partial Content",
		300: "Multiple Choices",
		301: "Moved Permanently",
		302: "Found",
		303: "See Other",
		304: "Not Modified",
		305: "Use Proxy",
		307: "Temporary Redirect",
		400: "Bad Request",
		401: "Unauthorized",
		402: "Payment Required",
		403: "Forbidden",
		404: "Not Found",
		405: "Method Not Allowed",
		406: "Not Acceptable",
		407: "Proxy Authentication Required",
		408: "Request Timeout",
		409: "Conflict",
		410: "Gone",
		411: "Length Required",
		412: "Precondition Failed",
		413: "Payload Too Large",
		414: "URI Too Long",
		415: "Unsupported Media Type",
		416: "Range Not Satisfiable",
		417: "Expectation Failed",
		418: "I'm a teapot",
		421: "Misdirected Request",
		426: "Upgrade Required",
		428: "Precondition Required",
		429: "Too Many Requests",
		431: "Request Header Fields Too Large",
		451: "Unavailable For Legal Reasons",
		500: "Internal Server Error",
		501: "Not Implemented",
		502: "Bad Gateway",
		503: "Service Unavailable",
		504: "Gateway Timeout",
		505: "HTTP Version Not Supported",
		506: "Variant Also Negotiates",
		507: "Insufficient Storage",
		508: "Loop Detected",
		510: "Not Extended",
		511: "Network Authentication Required",
	}

	if text, exists := statusTexts[code]; exists {
		return text
	}

	return "Unknown Status"
}
