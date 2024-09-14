package attachment

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Attachment 附件结构体，包含文件头信息、扩展名、保存目录和错误列表
type Attachment struct {
	FileHeader *multipart.FileHeader // multipart 文件头部
	FileMD5    string                // 文件MD5
	Opt        *Options              // 附件实例化时的参数选项

	ext     string  // 文件扩展名
	saveDir string  // 文件保存目录
	errors  []error // 错误信息列表
}

// Options 附件实例化时的参数选项
type Options struct {
	FileData interface{} // 文件数据，支持base64字符串或*multipart.FileHeader
	FileType string      // 文件类型

	SaveDir string // 文件保存目录
}

// New 创建一个新的附件实例，处理初始化和文件类型解析
func New(opt *Options) *Attachment {
	a := &Attachment{}
	a.initOpt(opt) // 初始化选项，设置默认值

	return a
}

// initOpt 初始化选项，提供默认值设置
func (a *Attachment) initOpt(opt *Options) {
	if opt.FileType == "" {
		opt.FileType = "image"
	}

	if opt.SaveDir == "" {
		opt.SaveDir = fmt.Sprintf("./attachment/%ss/%s/", opt.FileType, time.Now().Format("2006/01"))
	}
	a.saveDir, _ = filepath.Abs(opt.SaveDir)

	a.Opt = opt
}

// Save 保存文件到指定的附件目录
func (a *Attachment) Save() *Attachment {
	if len(a.errors) > 0 {
		return a
	}

	switch v := a.Opt.FileData.(type) {
	case *multipart.FileHeader: // 处理multipart文件头
		if v.Size == 0 {
			a.errors = append(a.errors, fmt.Errorf("无效的参数，文件为空"))
			return a
		} else {
			a.FileHeader = v
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

	// 打开源文件（已经是临时文件的内容）
	srcFile, err := os.Open(a.FileHeader.Filename) // 直接打开临时文件
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

	// 生成文件MD5和复制文件内容
	hash := md5.New()
	reader := io.TeeReader(srcFile, hash)

	_, fullDstFilePath, _ := a.fileRandomName(a.saveDir)

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

	return a
}

// Error 返回第一个捕获的错误
func (a *Attachment) Error() error {
	if len(a.errors) == 0 {
		return nil
	} else {
		return a.errors[0]
	}
}

// getExt 获取文件扩展名，若未设置则从文件名解析
func (a *Attachment) getExt() string {
	if a.ext == "" {
		a.ext = strings.ToLower(filepath.Ext(a.FileHeader.Filename))
	}
	return a.ext
}

// fileRandomName 生成随机文件名，确保文件名在指定目录下是唯一的
func (a *Attachment) fileRandomName(dir string) (filename, fullDstFile string, err error) {
	for {
		dateStr := time.Now().Format("02150405")
		randomStr := helper.Random(22)
		filename = fmt.Sprintf("%s%s%s", dateStr, randomStr, a.ext)
		fullDstFile = filepath.Join(dir, filename)
		if _, err = os.Stat(fullDstFile); os.IsNotExist(err) {
			break
		} else if err != nil {
			return
		}
	}
	return
}

// parseBase64ToMultipart 解析base64字符串到multipart.FileHeader
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
	a.ext, ok = extMap[mediaType]
	if !ok {
		// 使用 MIME 类型解析库提供的扩展名作为后备选项
		exts, err := mime.ExtensionsByType(mediaType)
		if err != nil || len(exts) == 0 {
			return nil, fmt.Errorf("no suitable extension found for MIME type: %s", mediaType)
		}
		a.ext = exts[0] // 使用 mime 提供的第一个扩展名
	}

	tmpFile, err := os.CreateTemp("", "attachment-*"+a.ext)
	if err != nil {
		return tmpFile, err
	}

	if _, err = tmpFile.Write(decodedData); err != nil {
		tmpFile.Close()
		return nil, err
	}

	if _, err = tmpFile.Seek(0, 0); err != nil {
		tmpFile.Close()
		return nil, err
	}

	// 这里手动设置 Filename 为临时文件的路径，确保 FileHeader 可以正确打开文件
	a.FileHeader = &multipart.FileHeader{
		Filename: tmpFile.Name(), // 设置 Filename 为 tmpFile.Name()
		Size:     int64(len(decodedData)),
	}

	return tmpFile, nil
}

// addError 添加错误到错误列表
func (a *Attachment) addError(err error) {
	if err != nil {
		a.errors = append(a.errors, err)
	}
}
