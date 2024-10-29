package attachment

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo"
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

	Config     *jcbaseGo.AttachmentStruct // 附件配置信息
	GinContext *gin.Context               // gin上下文

	FileType       string // 附件类型
	FileName       string // 附件名
	FileSize       int64  // 附件大小
	FileAttachment string // 附件相对路径
	FileMD5        string // 附件MD5
	FileExt        string // 文件扩展名
	Width          int    // 图片宽
	Height         int    // 图片高

	saveDir    string                   // 文件保存目录
	errors     []error                  // 错误信息列表
	beforeSave func(a *Attachment) bool // 保存前的回调函数，可选
}

// Options 附件实例化时的参数选项
type Options struct {
	Group string // 附件组，默认不分组（分组会文件类型目录前多一级分组目录）

	FileData interface{} // 文件数据，支持 base64 字符串、*multipart.FileHeader、[]byte
	FileType string      // 文件类型，默认为 image
	MaxSize  int64       // 最大文件大小
	AllowExt []string    // 允许的文件扩展名
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
		AllowExt: []string{".gif", ".jpg", ".jpeg", ".bmp", ".png", ".ico"},
		MaxSize:  5 * 1024 * 1024, // 5MB
	},
	"voice": {
		TypeName: "音频",
		AllowExt: []string{".mp3", ".wma", ".wav", ".amr"},
		MaxSize:  50 * 1024 * 1024, // 50MB
	},
	"video": {
		TypeName: "视频",
		AllowExt: []string{".rm", ".rmvb", ".wmv", ".avi", ".mpg", ".mpeg", ".mp4"},
		MaxSize:  300 * 1024 * 1024, // 300MB
	},
	"office": {
		TypeName: "办公文件",
		AllowExt: []string{
			".wps", ".wpt", ".doc", ".dot", ".docx", ".docm", ".dotm",
			".et", ".ett", ".xls", ".xlt", ".xlsx", ".xlsm", ".xltx", ".xltm", ".xlsb",
			".dps", ".dpt", ".ppt", ".pps", ".pot", ".pptx", ".ppsx", ".potx",
			".txt", ".csv", ".prn", ".pdf", ".xml",
		},
		MaxSize: 50 * 1024 * 1024, // 50MB
	},
	"zip": {
		TypeName: "压缩文件",
		AllowExt: []string{".zip", ".rar"},
		MaxSize:  500 * 1024 * 1024, // 500MB
	},
}

// New 创建一个新的附件实例
// args:
//   - *gin.Context
//   - *jcbaseGo.AttachmentStruct
func New(args ...interface{}) *Attachment {

	a := &Attachment{}

	// 第一个参数为 gin 上下文
	if len(args) > 0 {
		a.GinContext, _ = args[0].(*gin.Context)
	}

	// 第二个参数为附件配置
	if len(args) > 1 {
		a.Config, _ = args[1].(*jcbaseGo.AttachmentStruct)
	} else {
		a.Config = &jcbaseGo.AttachmentStruct{}
	}
	_ = helper.CheckAndSetDefault(a.Config)

	return a
}

func (a *Attachment) initConfig() {
	// 自定义访问域名情况下不进行处理
	if a.Config.VisitDomain != "/" && a.Config.VisitDomain != "" {
		return
	}

	if a.GinContext != nil {
		a.Config.LocalVisitDomain = helper.GetHostInfo(a.GinContext.Request)
	} else {
		log.Println("gin上下文为空，无法获取域名信息")
	}

	switch a.Config.StorageType {
	case "local":
		if a.GinContext == nil {
			return
		}
		// 从gin上下文中获取域名信息
		a.Config.VisitDomain = a.Config.LocalVisitDomain
	case "cos":
		remoteConf, ok := a.Config.Remote.(jcbaseGo.COSStruct)
		if !ok || remoteConf.SecretId == "" {
			log.Println("cos配置错误")
			return
		}
		a.Config.VisitDomain = remoteConf.CustomizeVisitDomain
	case "oss":
		remoteConf, ok := a.Config.Remote.(jcbaseGo.OSSStruct)
		if !ok || remoteConf.AccessKeyId == "" {
			log.Println("oss配置错误")
			return
		}
		if remoteConf.CustomizeVisitDomain != "" {
			a.Config.VisitDomain = remoteConf.CustomizeVisitDomain
		} else {
			a.Config.VisitDomain = fmt.Sprintf("https://%s.%s/", remoteConf.BucketName, remoteConf.Endpoint)
		}
	case "sftp":
		remoteConf, ok := a.Config.Remote.(jcbaseGo.SFTPStruct)
		if !ok || remoteConf.Address == "" {
			log.Println("sftp配置错误")
			return
		}
		a.Config.VisitDomain = remoteConf.CustomizeVisitDomain
	case "ftp":
		remoteConf, ok := a.Config.Remote.(jcbaseGo.FTPStruct)
		if !ok || remoteConf.Address == "" {
			log.Println("ftp配置错误")
			return
		}
		a.Config.VisitDomain = remoteConf.CustomizeVisitDomain
	}
}

// Upload 初始化文件上传参数
func (a *Attachment) Upload(opt *Options) *Attachment {
	a.Opt = opt

	if a.Opt.FileType == "" {
		a.Opt.FileType = "image"
	}
	a.FileType = a.Opt.FileType

	// 整理存储目录
	a.saveDir = fmt.Sprintf("./%s/%ss/%s/", a.Config.LocalDir, a.Opt.FileType, time.Now().Format("2006/01"))
	if a.Opt.Group != "" {
		a.saveDir = fmt.Sprintf("./%s/%s/%ss/%s/", a.Config.LocalDir, a.Opt.Group, a.Opt.FileType, time.Now().Format("2006/01"))
	}
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

	return a
}

// SetBeforeSave 设置保存文件前的回调函数
// - 回调函数返回 true 则继续保存，回调函数返回 false 则跳过
func (a *Attachment) SetBeforeSave(fn func(a *Attachment) bool) *Attachment {
	a.beforeSave = fn
	return a
}

// Save 保存文件到指定的附件目录
func (a *Attachment) Save() *Attachment {
	if a.HasError() {
		return a
	}

	// 如果没有初始化opt参数，应当返回错误
	if a.Opt == nil {
		a.addError(fmt.Errorf("无效的参数，opt 不能为空"))
		return a
	}

	var srcFile io.ReadSeeker
	var err error

	switch v := a.Opt.FileData.(type) {
	case *multipart.FileHeader:
		// 处理 *multipart.FileHeader 类型
		if v.Size == 0 {
			a.addError(fmt.Errorf("无效的参数，文件为空"))
			return a
		}
		a.FileSize = v.Size
		a.FileExt = strings.ToLower(filepath.Ext(v.Filename))
		srcFile, err = v.Open()
		if err != nil {
			a.addError(fmt.Errorf("打开源文件失败: %v", err))
			return a
		}
		defer func() {
			if c, ok := srcFile.(io.Closer); ok {
				_ = c.Close()
			}
		}()
	case multipart.FileHeader:
		// 处理 multipart.FileHeader 值类型
		if v.Size == 0 {
			a.addError(fmt.Errorf("无效的参数，文件为空"))
			return a
		}
		a.FileSize = v.Size
		a.FileExt = strings.ToLower(filepath.Ext(v.Filename))
		srcFile, err = v.Open()
		if err != nil {
			a.addError(fmt.Errorf("打开源文件失败: %v", err))
			return a
		}
		defer func() {
			if c, ok := srcFile.(io.Closer); ok {
				_ = c.Close()
			}
		}()
	case string:
		// 处理 Base64 编码的文件数据
		decodedData, err := a.parseBase64Data(v)
		if err != nil {
			a.addError(fmt.Errorf("解析 Base64 失败：%v", err))
			return a
		}
		srcFile = bytes.NewReader(decodedData)
		a.FileSize = int64(len(decodedData))
		// FileExt 已在 parseBase64Data 中获取
	case []byte:
		// 处理 []byte 数据
		srcFile = bytes.NewReader(v)
		a.FileSize = int64(len(v))
		// 需要确定文件扩展名
		if len(a.Opt.AllowExt) == 1 {
			a.FileExt = a.Opt.AllowExt[0]
		} else {
			a.addError(fmt.Errorf("无法确定文件扩展名，请在 Options 中指定 AllowExt"))
			return a
		}
	default:
		a.addError(fmt.Errorf("不支持的文件数据类型: %T", v))
		return a
	}

	// 计算文件 MD5
	if _, err := srcFile.Seek(0, io.SeekStart); err != nil {
		a.addError(fmt.Errorf("无法重置文件指针: %v", err))
		return a
	}
	hash := md5.New()
	if _, err := io.Copy(hash, srcFile); err != nil {
		a.addError(fmt.Errorf("计算文件 MD5 失败：%v", err))
		return a
	}
	a.FileMD5 = hex.EncodeToString(hash.Sum(nil))

	// 重置文件指针到文件开头
	if _, err = srcFile.Seek(0, io.SeekStart); err != nil {
		a.addError(fmt.Errorf("无法重置文件指针: %v", err))
		return a
	}

	// 如果设置了 beforeSave 回调函数，调用它
	if a.beforeSave != nil {
		if !a.beforeSave(a) {
			// 回调函数返回 false，不继续保存
			return a
		}
	}

	// 校验文件扩展名
	if len(a.Opt.AllowExt) > 0 && !helper.InArray(a.FileExt, a.Opt.AllowExt) {
		a.addError(fmt.Errorf("不支持的文件【%s】", a.FileExt))
		return a
	}

	// 判断文件大小是否超出限制
	if a.Opt.MaxSize > 0 && a.FileSize > a.Opt.MaxSize {
		a.addError(fmt.Errorf("文件大小不能超出[%s]", a.formatFileSize(a.Opt.MaxSize)))
		return a
	}

	// 如果是图片，应当获取宽高
	if a.FileType == "image" {
		img, _, err := image.Decode(srcFile)
		if err != nil {
			log.Println("解码图片失败: ", err)
			a.addError(err)
			return a
		}
		// 获取图片尺寸
		bounds := img.Bounds()
		a.Width = bounds.Dx()
		a.Height = bounds.Dy()
		// 重置文件指针到文件开头
		if _, err = srcFile.Seek(0, io.SeekStart); err != nil {
			a.addError(fmt.Errorf("无法重置文件指针: %v", err))
			return a
		}
	}

	// 提前创建文件目录，避免后续操作报错
	err = os.MkdirAll(a.saveDir, os.ModePerm)
	if err != nil {
		a.addError(fmt.Errorf("创建目录失败：%v", err))
		return a
	}

	// 生成随机文件名，如果目录不存在将会自动创建目录
	var fullDstFilePath string
	a.FileName, fullDstFilePath, err = a.fileRandomName(a.saveDir)
	if err != nil {
		a.addError(fmt.Errorf("生成随机文件名失败：%v", err))
		return a
	}

	// 创建目标文件
	dstFile, err := os.Create(fullDstFilePath)
	if err != nil {
		a.addError(fmt.Errorf("创建目标文件失败：%v", err))
		return a
	}
	defer func() {
		if err = dstFile.Close(); err != nil {
			log.Println("关闭文件失败: ", err)
		}
	}()

	// 复制文件内容
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		a.addError(fmt.Errorf("拷贝文件内容失败：%v", err))
		return a
	}

	// 获取附件相对路径
	index := strings.Index(fullDstFilePath, a.Config.LocalDir+"/")
	if index == -1 {
		log.Println("未在路径中找到" + a.Config.LocalDir)
		a.addError(fmt.Errorf("未在路径中找到%s", a.Config.LocalDir))
	} else {
		a.FileAttachment = fullDstFilePath[index+len(a.Config.LocalDir+"/"):]
	}

	return a
}

// parseBase64Data 解析 Base64 字符串，返回解码后的数据，并设置文件扩展名
func (a *Attachment) parseBase64Data(base64Data string) ([]byte, error) {
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

	// 自定义 MIME 类型到文件扩展名的映射表
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
			return nil, fmt.Errorf("无法获取 MIME 类型的扩展名: %s", mediaType)
		}
		a.FileExt = exts[0] // 使用 mime 提供的第一个扩展名
	}

	return decodedData, nil
}

// fileRandomName 生成随机文件名，确保文件名在指定目录下是唯一的
func (a *Attachment) fileRandomName(dir string) (string, string, error) {
	var filename, fullDstFile string
	for {
		dateStr := time.Now().Format("02150405")
		randomStr := helper.Random(22)
		filename = fmt.Sprintf("%s%s%s", dateStr, randomStr, a.FileExt)
		fullDstFile = filepath.Join(dir, filename)
		_, err := os.Stat(fullDstFile)
		if err == nil {
			// 文件已存在，继续生成新的文件名
			continue
		}
		if os.IsNotExist(err) {
			// 文件不存在，可以使用该文件名
			break
		} else {
			// 其他错误，返回错误信息
			return "", "", fmt.Errorf("检查文件状态时出错: %v", err)
		}
	}
	return filename, fullDstFile, nil
}

// formatFileSize 格式化文件大小，将字节单位转换为适当的单位（KB, MB, GB 等）并保留两位小数
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
	return len(a.errors) > 0
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
