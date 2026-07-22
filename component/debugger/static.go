package debugger

import (
	"embed"
	"io/fs"
	"net/http"
)

// staticFiles 嵌入调试器前端所需的静态资源文件
// 通过 embed 将静态文件编译到二进制中，避免运行时依赖外部 CDN 或文件系统
//
//go:embed static/*
var staticFiles embed.FS

// getStaticFileSystem 获取调试器静态文件的 http.FileSystem
// 返回以 static 目录为根的只读文件系统，供 Gin 的 StaticFS 使用
func getStaticFileSystem() http.FileSystem {
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		// 嵌入失败时回退到原始文件系统，避免服务完全不可用
		return http.FS(staticFiles)
	}
	return http.FS(sub)
}
