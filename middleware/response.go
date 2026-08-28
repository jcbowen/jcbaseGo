package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/serializer"
	"github.com/jcbowen/jcbaseGo/errcode"
)

// Response 统一响应结构。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Options 为序列化处理选项的别名，保持与原有 middleware 调用方式一致。
type Options = serializer.Options

// Option 为序列化选项函数类型的别名。
type Option = serializer.Option

// WithEncryptID 注入 ID 加密函数。
var WithEncryptID = serializer.WithEncryptID

// WithIDField 注入 ID 字段判断函数。
var WithIDField = serializer.WithIDField

// WithTimeField 注入时间字段判断函数。
var WithTimeField = serializer.WithTimeField

// WithFormatTime 注入时间值格式化函数。
var WithFormatTime = serializer.WithFormatTime

// WithSkipField 注入字段跳过判断函数。
var WithSkipField = serializer.WithSkipField

// WithMaxDepth 注入递归处理最大深度。
var WithMaxDepth = serializer.WithMaxDepth

// JSON 返回统一 JSON 响应。
//
// 参数:
//   - c: gin 上下文
//   - code: 响应码
//   - message: 响应消息
//   - data: 响应数据
func JSON(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// Success 返回成功响应，并统一处理时间字段、按需加密 ID。
// 通过函数式选项注入时间格式化与 ID 加密策略；不注入则不处理。
//
// 参数:
//   - c: gin 上下文
//   - data: 响应数据
//   - opts: 可选的处理策略
func Success(c *gin.Context, data interface{}, opts ...Option) {
	JSON(c, errcode.Success, "ok", serializer.Process(data, opts...))
}

// Error 返回错误响应。
//
// 参数:
//   - c: gin 上下文
//   - code: 错误码
//   - message: 错误消息
func Error(c *gin.Context, code int, message string) {
	JSON(c, code, message, nil)
}
