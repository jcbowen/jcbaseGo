package attachment

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"image"
	_ "image/gif"  // 导入 GIF 支持
	_ "image/jpeg" // 导入 JPEG 支持
	_ "image/png"  // 导入 PNG 支持
	"io"
	"log"
	"math"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Attachment 附件结构体，包含文件头信息、扩展名、保存目录和错误列表
type Attachment struct {
	Opt *Options // 附件实例化时的参数选项

	// 临时
	TmpFileHeader *multipart.FileHeader

	FileType       string // 附件类型
	FileName       string // 附件名
	FileSize       int64  // 附件大小
	FileAttachment string // 附件相对路径
	FileMD5        string // 附件MD5
	FileExt        string // 文件扩展名
	Width          int    // 图片宽
	Height         int    // 图片高

	saveDir string  // 文件保存目录
	errors  []error // 错误信息列表
}

// Options 附件实例化时的参数选项
type Options struct {
	FileData      interface{} // 文件数据，支持base64字符串或*multipart.FileHeader
	FileType      string      // 文件类型，默认为image
	AttachmentDir string      // 附件目录，默认为attachment
	MaxSize       int64       // 最大文件大小
	AllowExt      []string    // 允许的文件扩展名
}

// typeInfo 附件类型信息
type typeInfo struct {
	TypeName string   // 类型名
	MaxSize  int64    // 最大文件大小
	AllowExt []string // 允许的文件扩展名
}

// Types 附件类型
var Types = map[string]*typeInfo{
	"image": {
		TypeName: "图片",
		AllowExt: []string{"gif", "jpg", "jpeg", "bmp", "png", "ico"},
		MaxSize:  5 * 1024 * 1024, // 5MB
	},
	"voice": {
		TypeName: "音频",
		AllowExt: []string{"mp3", "wma", "wav", "amr"},
		MaxSize:  50 * 1024 * 1024, // 50MB
	},
	"video": {
		TypeName: "视频",
		AllowExt: []string{"rm", "rmvb", "wmv", "avi", "mpg", "mpeg", "mp4"},
		MaxSize:  300 * 1024 * 1024, // 300MB
	},
	"office": {
		TypeName: "办公文件",
		AllowExt: []string{
			"wps", "wpt", "doc", "dot", "docx", "docm", "dotm",
			"et", "ett", "xls", "xlt", "xlsx", "xlsm", "xltx", "xltm", "xlsb",
			"dps", "dpt", "ppt", "pps", "pot", "pptx", "ppsx", "potx",
			"txt", "csv", "prn", "pdf", "xml",
		},
		MaxSize: 50 * 1024 * 1024, // 50MB
	},
	"zip": {
		TypeName: "压缩文件",
		AllowExt: []string{"zip", "rar"},
		MaxSize:  500 * 1024 * 1024, // 500MB
	},
}

// New 创建一个新的附件实例，处理初始化和文件类型解析
func New(opt *Options) *Attachment {
	a := &Attachment{}
	a.initOpt(opt) // 初始化选项，设置默认值

	return a
}

// initOpt 初始化选项，提供默认值设置
func (a *Attachment) initOpt(opt *Options) {
	a.Opt = opt

	if a.Opt.FileType == "" {
		a.Opt.FileType = "image"
	}
	a.FileType = a.Opt.FileType

	if a.Opt.AttachmentDir == "" {
		a.Opt.AttachmentDir = "attachment"
	}

	a.saveDir = fmt.Sprintf("./%s/%ss/%s/", a.Opt.AttachmentDir, a.Opt.FileType, time.Now().Format("2006/01"))
	a.saveDir, _ = filepath.Abs(a.saveDir)

	// 判断是否为文件限制补充默认配置
	typeDefaultInfo, ok := Types[a.Opt.FileType]
	if ok {
		if a.Opt.MaxSize == 0 && typeDefaultInfo.MaxSize > 0 {
			a.Opt.MaxSize = typeDefaultInfo.MaxSize
		}
		if len(a.Opt.AllowExt) == 0 && len(typeDefaultInfo.AllowExt) > 0 {
			a.Opt.AllowExt = typeDefaultInfo.AllowExt
		}
	}
}

// Save 保存文件到指定的附件目录
func (a *Attachment) Save() *Attachment {
	if a.HasError() {
		return a
	}

	switch v := a.Opt.FileData.(type) {
	case *multipart.FileHeader: // 处理multipart文件头
		if v.Size == 0 {
			a.errors = append(a.errors, fmt.Errorf("无效的参数，文件为空"))
			return a
		} else {
			a.TmpFileHeader = v
		}
	case string:
		// 处理base64编码的文件数据
		tmpFile, err := a.parseBase64ToMultipart(v)
		if tmpFile != nil {
			// 处理 Seek 的错误
			if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
				a.errors = append(a.errors, fmt.Errorf("无法重置临时文件的指针: %v", err))
				return a
			}

			// 确保在操作完成后删除临时文件
			defer func(name string) {
				err := os.Remove(name)
				if err != nil {
					log.Println("删除临时文件失败:", err)
				}
			}(tmpFile.Name())
		}
		if err != nil {
			a.errors = append(a.errors, err)
			return a
		}
	default:
		a.errors = append(a.errors, fmt.Errorf("不支持的文件数据类型: %T", v))
		return a
	}

	a.getExt()
	if len(a.Opt.AllowExt) > 0 && !helper.InArray(a.FileExt, a.Opt.AllowExt) {
		a.errors = append(a.errors, fmt.Errorf("不支持的文件【%s】", a.FileExt))
		return a
	}

	// 赋值文件大小
	a.FileSize = a.TmpFileHeader.Size

	// 判断文件大小是否超出限制
	if a.Opt.MaxSize > 0 && a.FileSize > a.Opt.MaxSize {
		a.errors = append(a.errors, fmt.Errorf("文件大小不能超出[%s]", a.formatFileSize(a.Opt.MaxSize)))
		return a
	}

	// 打开源文件（已经是临时文件的内容）
	srcFile, err := os.Open(a.TmpFileHeader.Filename) // 直接打开临时文件
	if err != nil {
		log.Println("打开文件失败: ", err)
		a.addError(err)
		return a
	}

	// 确保源文件被正确关闭
	defer func(srcFile *os.File) {
		err = srcFile.Close()
		if err != nil {
			log.Println(err)
		}
	}(srcFile)

	// 如果是图片，应当获取宽高
	if a.Opt.FileType == "image" {
		// 解码图片
		img, _, err := image.Decode(srcFile)
		if err != nil {
			log.Println("解码图片失败: ", err)
			a.addError(err)
		} else {
			// 获取图片尺寸
			bounds := img.Bounds()
			a.Width = bounds.Dx()
			a.Height = bounds.Dy()
		}
	}

	// 生成文件MD5和复制文件内容
	hash := md5.New()
	reader := io.TeeReader(srcFile, hash)

	var fullDstFilePath string
	a.FileName, fullDstFilePath, _ = a.fileRandomName(a.saveDir)

	// 创建目标文件之前，确保目录存在
	err = os.MkdirAll(filepath.Dir(fullDstFilePath), os.ModePerm)
	if err != nil {
		a.addError(fmt.Errorf("创建目录失败: %v", err))
		return a
	}

	dstFile, err := os.Create(fullDstFilePath) // 创建目标文件
	if err != nil {
		a.addError(err)
		return a
	}
	defer func(dstFile *os.File) {
		err = dstFile.Close()
		if err != nil {
			log.Println(err)
		}
	}(dstFile)

	// 复制文件内容并生成 MD5
	if _, err = io.Copy(dstFile, reader); err != nil {
		a.addError(err)
		return a
	}
	a.FileMD5 = hex.EncodeToString(hash.Sum(nil))

	// 获取附件相对路径
	index := strings.Index(fullDstFilePath, a.Opt.AttachmentDir+"/")
	if index == -1 {
		log.Println("未在路径中找到" + a.Opt.AttachmentDir)
		a.addError(fmt.Errorf("未在路径中找到%s", a.Opt.AttachmentDir))
	} else {
		a.FileAttachment = fullDstFilePath[index+len(a.Opt.AttachmentDir+"/"):]
	}

	return a
}

// getExt 获取文件扩展名，若未设置则从文件名解析
func (a *Attachment) getExt() string {
	if a.FileExt == "" {
		a.FileExt = strings.ToLower(filepath.Ext(a.TmpFileHeader.Filename))
	}
	return a.FileExt
}

// fileRandomName 生成随机文件名，确保文件名在指定目录下是唯一的
func (a *Attachment) fileRandomName(dir string) (filename, fullDstFile string, err error) {
	for {
		dateStr := time.Now().Format("02150405")
		randomStr := helper.Random(22)
		filename = fmt.Sprintf("%s%s%s", dateStr, randomStr, a.getExt())
		fullDstFile = filepath.Join(dir, filename)
		if _, err = os.Stat(fullDstFile); os.IsNotExist(err) {
			break
		} else if err != nil {
			return
		}
	}
	return
}

// parseBase64ToMultipart 解析base64字符串到multipart.TmpFileHeader
func (a *Attachment) parseBase64ToMultipart(base64Data string) (*os.File, error) {
	parts := strings.SplitN(base64Data, ",", 2)
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "data:") {
		return nil, fmt.Errorf("invalid base64 data")
	}

	mediaTypeSpec := strings.SplitN(parts[0][5:], ";", 2)
	mediaType := mediaTypeSpec[0]

	decodedData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	// 自定义MIME类型到文件扩展的映射表
	var extMap = map[string]string{
		"application/x-jpg": ".jpg",
		"image/jpg":         ".jpg",
		"image/jpeg":        ".jpg",
		"image/png":         ".png",
		"image/gif":         ".gif",
		"video/mp4":         ".mp4",
		"video/mpeg4":       ".mp4",
		"video/x-ms-wmv":    ".wmv",
		"audio/mpeg":        ".mp3",
		"audio/mp4":         ".mp4",
		"audio/x-ms-wma":    ".wma",
	}

	var ok bool
	a.FileExt, ok = extMap[mediaType]
	if !ok {
		// 使用 MIME 类型解析库提供的扩展名作为后备选项
		exts, err := mime.ExtensionsByType(mediaType)
		if err != nil || len(exts) == 0 {
			return nil, fmt.Errorf("no suitable extension found for MIME type: %s", mediaType)
		}
		a.FileExt = exts[0] // 使用 mime 提供的第一个扩展名
	}

	tmpFile, err := os.CreateTemp("", "attachment-*"+a.getExt())
	if err != nil {
		return tmpFile, err
	}

	if _, err = tmpFile.Write(decodedData); err != nil {
		_ = tmpFile.Close()
		return nil, err
	}

	if _, err = tmpFile.Seek(0, 0); err != nil {
		_ = tmpFile.Close()
		return nil, err
	}

	// 这里手动设置 Filename 为临时文件的路径，确保 TmpFileHeader 可以正确打开文件
	a.TmpFileHeader = &multipart.FileHeader{
		Filename: tmpFile.Name(), // 设置 Filename 为 tmpFile.Name()
		Size:     int64(len(decodedData)),
	}

	return tmpFile, nil
}

// formatFileSize 格式化文件大小，将字节单位转换为适当的单位（KB, MB, GB等）并保留两位小数。
func (a *Attachment) formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	ceilToTwoDecimals := func(num float64) float64 {
		return math.Ceil(num*100) / 100
	}

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", ceilToTwoDecimals(float64(size)/float64(TB)))
	case size >= GB:
		return fmt.Sprintf("%.2f GB", ceilToTwoDecimals(float64(size)/float64(GB)))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", ceilToTwoDecimals(float64(size)/float64(MB)))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", ceilToTwoDecimals(float64(size)/float64(KB)))
	default:
		return fmt.Sprintf("%d Bytes", size)
	}
}

// addError 添加错误到错误列表
func (a *Attachment) addError(err error) {
	if err != nil {
		a.errors = append(a.errors, err)
	}
}

// HasError 是否有错误
func (a *Attachment) HasError() bool {
	return a.HasError()
}

// Error 返回第一个捕获的错误
func (a *Attachment) Error() error {
	if len(a.errors) == 0 {
		return nil
	} else {
		return a.errors[0]
	}
}

// Errors 返回所有捕获的错误
func (a *Attachment) Errors() []error {
	return a.errors
}
