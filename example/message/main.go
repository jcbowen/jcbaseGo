package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/message"
)

func main() {
	r := gin.Default()

	// 起始页：提供三种进入消息页的方式
	r.GET("/from", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>消息组件示例入口</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; padding: 40px; }
    .card { max-width: 720px; margin: 0 auto; background: #fff; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); padding: 24px; }
    h1 { margin: 0 0 16px; font-size: 22px; }
    p { color: #666; }
    ul { list-style: none; padding: 0; }
    li { margin: 8px 0; }
    a { text-decoration: none; color: #10aeff; }
  </style>
</head>
<body>
  <div class="card">
    <h1>消息组件示例入口</h1>
    <p>请选择一个示例进行体验：</p>
    <ul>
      <li><a href="/message/next">成功消息，5秒后自动跳转到下一页</a></li>
      <li><a href="/message/back">成功消息，5秒后自动返回上一页（依赖 Referer）</a></li>
      <li><a href="/message/error">错误消息，不自动跳转，手动返回上一页</a></li>
    </ul>
  </div>
</body>
</html>`)
	})

	// 示例：显式设置 RedirectURL，5秒后自动跳到 /next
	r.GET("/message/next", func(c *gin.Context) {
		message.Success(
			c,
			"操作成功",
			"这是一条测试消息，用于预览自动跳转到下一页。",
			message.WithRedirect("/next"),
			message.WithAutoRedirect(true),
		)
	})

	// 示例：未设置 RedirectURL，5秒后自动返回上一页（Referer）
	r.GET("/message/back", func(c *gin.Context) {
		message.Success(
			c,
			"操作成功",
			"这是一条测试消息，用于预览自动返回上一页。",
			message.WithAutoRedirect(true),
		)
	})

	// 示例：错误消息，不自动跳转，展示返回按钮（Referer 存在时）
	r.GET("/message/error", func(c *gin.Context) {
		message.Error(
			c,
			"操作失败",
			"发生错误，不会自动跳转。请点击按钮返回上一页。",
			message.WithAutoRedirect(false),
		)
	})

	// 跳转目标页
	r.GET("/next", func(c *gin.Context) {
		c.String(200, "到达下一页")
	})

	addr := ":8080"
	log.Println("预览入口: http://localhost:8080/from")
	log.Println("直接预览自动跳转到下一页: http://localhost:8080/message/next")
	log.Println("直接预览自动返回上一页: 需从 /from 点击进入 /message/back")
	_ = r.Run(addr)
}
