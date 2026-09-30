package attachment

import "testing"

// TestTypesCommonExtensions 校验各类型白名单包含全部常用扩展名，
// 作为 Types 调整（新增/重构/误删）的回归防线。
//
// 使用示例：
//
//	go test ./component/attachment -run TestTypesCommonExtensions -v
func TestTypesCommonExtensions(t *testing.T) {
	// 各类型必须包含的扩展名集合（含历史对齐成果 apk/7z/wgt）
	cases := map[string][]string{
		"image": {".gif", ".jpg", ".jpeg", ".bmp", ".png", ".ico", ".webp", ".heic"},
		"voice": {".mp3", ".wma", ".wav", ".amr", ".m4a", ".aac", ".ogg", ".flac"},
		"video": {".rm", ".rmvb", ".wmv", ".avi", ".mpg", ".mpeg", ".mp4",
			".mov", ".mkv", ".flv", ".webm", ".3gp", ".m4v"},
		"office": {".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".pdf", ".txt", ".csv"},
		"zip":    {".zip", ".rar", ".apk", ".7z", ".wgt", ".tar", ".gz", ".bz2"},
	}

	for fileType, requiredExts := range cases {
		info, ok := Types[fileType]
		if !ok {
			t.Fatalf("Types 中不存在 %s 类型", fileType)
		}
		existing := make(map[string]bool, len(info.AllowExt))
		for _, ext := range info.AllowExt {
			existing[ext] = true
		}
		for _, ext := range requiredExts {
			if !existing[ext] {
				t.Errorf("%s 类型白名单缺少 %s，当前: %v", fileType, ext, info.AllowExt)
			}
		}
	}
}
