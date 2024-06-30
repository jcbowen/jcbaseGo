package jcbaseGo

import "os"

// Config 为Config添加默认数据
var Config interface{}

// isDevelopment 是否为开发环境(既运行环境)
var IsDevelopment = os.Getenv("GOBIN") == ""
