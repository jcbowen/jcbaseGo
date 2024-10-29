package attachment

import (
	"strings"
	"time"
)

func (a *Attachment) toMedia(src string, args ...interface{}) string {
	isLocal := false // 是否为本地附件
	isCache := true  // 是否在尾部添加时间戳

	if len(args) > 0 {
		isLocal = args[0].(bool)
	}
	if len(args) > 1 {
		isCache, _ = args[1].(bool)
	}

	if len(src) == 0 {
		return ""
	}
	if !isCache {
		src += "?v=" + time.Now().Format("20060102150405")
	}

	if strings.Index(src, "http://") == 0 || strings.Index(src, "https://") == 0 {
		return src
	} else if strings.Index(src, "//") == 0 {
		return "http:" + src
	}

	src = strings.TrimPrefix(src, "/")

	if isLocal || a.Config.StorageType == "local" {
		src = a.Config.LocalVisitDomain + a.Config.LocalDir + "/" + src
	} else {
		src = a.Config.VisitDomain + src
	}

	return src
}
