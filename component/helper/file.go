package helper

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
)

type FileHelper struct {
	Path string      `json:"path"`
	Perm os.FileMode `json:"perm" default:"0755"`
}

func NewFileHelper(args ...any) *FileHelper {
	var fileHelper *FileHelper
	if len(args) > 0 {
		fileHelper = args[0].(*FileHelper)
	} else {
		fileHelper = &FileHelper{}
	}
	fileHelper.init()
	return fileHelper
}

func (fh *FileHelper) init() {
	_ = CheckAndSetDefault(fh)
	if fh.Perm <= 0 {
		fh.Perm = 0755
	}
}

// Exists 检查文件是否存在
func (fh *FileHelper) Exists() bool {
	if fh.Path == "" {
		return false
	}
	_, err := os.Stat(fh.Path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// JsonToData 将文件中的json数据解析到data中
func (fh *FileHelper) JsonToData(data *interface{}) error {
	// 读取json配置文件
	file, fErr := os.ReadFile(fh.Path)
	if fErr != nil {
		return fErr
	}
	fileDataString := string(file)

	err := json.Unmarshal([]byte(fileDataString), data)
	return err
}

// IsDir 判断是否是目录
func (fh *FileHelper) IsDir() bool {
	s, err := os.Stat(fh.Path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// GetAbsPath 获取绝对路径
func (fh *FileHelper) GetAbsPath() (string, error) {
	return filepath.Abs(fh.Path)
}

// DirName 获取目录部分
func (fh *FileHelper) DirName() string {
	return filepath.Dir(fh.Path)
}

// IsFile 判断是否是文件
func (fh *FileHelper) IsFile() bool {
	return !fh.IsDir()
}

// IsEmptyDir 判断目录是否为空
func (fh *FileHelper) IsEmptyDir() bool {
	f, err := os.Open(fh.Path)
	if err != nil {
		return false
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	_, err = f.Readdirnames(1)
	if errors.Is(err, os.ErrNotExist) {
		return true
	}
	return false
}

// IsEmptyFile 判断文件是否为空
func (fh *FileHelper) IsEmptyFile() bool {
	f, err := os.Open(fh.Path)
	if err != nil {
		return false
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Size() == 0
}

// IsEmpty 判断文件或目录是否为空
func (fh *FileHelper) IsEmpty() bool {
	if fh.IsDir() {
		return fh.IsEmptyDir()
	}
	return fh.IsEmptyFile()
}

// IsReadable 判断文件是否可读
func (fh *FileHelper) IsReadable() bool {
	_, err := os.OpenFile(fh.Path, os.O_RDONLY, 0666)
	if err != nil {
		return false
	}
	return true
}

// IsWritable 判断文件是否可写
func (fh *FileHelper) IsWritable() bool {
	_, err := os.OpenFile(fh.Path, os.O_WRONLY, 0666)
	if err != nil {
		return false
	}
	return true
}

// IsExecutable 判断文件是否可执行
func (fh *FileHelper) IsExecutable() bool {
	_, err := os.OpenFile(fh.Path, os.O_RDONLY, 0666)
	if err != nil {
		return false
	}
	return true
}

// IsSymlink 判断是否是软链接
func (fh *FileHelper) IsSymlink() bool {
	fi, err := os.Lstat(fh.Path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// IsHidden 判断文件是否隐藏
func (fh *FileHelper) IsHidden() bool {
	return len(fh.Path) > 1 && fh.Path[0] == '.'
}

// DirExists 判断目录是否存在，可选在不存在时是否创建目录
func (fh *FileHelper) DirExists(createIfNotExists bool) (exists bool, err error) {
	// 判断path是否为一个目录，如果不是目录则取出目录部分
	if !fh.IsDir() {
		fh.Path = fh.DirName()
		if fh.Path == "." || fh.Path == "/" {
			return false, errors.New("请输入正确的目录路径(不能为当前目录或根目录;目录必须以/结尾，否则目录名会被当做文件处理)")
		}
	}
	_, err = os.Stat(fh.Path)
	if err != nil {
		if os.IsNotExist(err) {
			if createIfNotExists {
				err = os.MkdirAll(fh.Path, fh.Perm)
				if err != nil {
					return false, err
				}
				return true, nil
			}
			return false, err
		}
		return true, err
	}
	return true, nil
}

// CreateFile 创建文件，可设置文件权限，可设置是否覆盖
func (fh *FileHelper) CreateFile(content []byte, overwrite bool) error {
	// 如果已经存在且不需要覆盖则返回错误
	if exists := fh.Exists(); exists {
		if !overwrite {
			return errors.New("文件已存在，路径：" + fh.Path)
		}
	}

	// 检查目录是否存在，不存在则创建
	_, err := fh.DirExists(true)
	if err != nil {
		return err
	}

	// 创建文件
	return os.WriteFile(fh.Path, content, fh.Perm)
}

// Remove 删除文件或目录
func (fh *FileHelper) Remove() error {
	return os.RemoveAll(fh.Path)
}

// CopyFileAttr 复制文件属性到目标文件
func (fh *FileHelper) CopyFileAttr(targetFile string) error {
	srcInfo, err := os.Stat(fh.Path)
	if err != nil {
		return err
	}
	return os.Chmod(targetFile, srcInfo.Mode())
}

// CopyFile 复制文件到指定位置，可设置是否覆盖，可设置是否复制文件属性
func (fh *FileHelper) CopyFile(targetPath string, overwrite bool, copyAttr bool) error {
	if !fh.Exists() {
		return errors.New("file not exists")
	}
	if fh.Exists() {
		if overwrite {
			err := os.Remove(targetPath)
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
	srcFile, err := os.Open(fh.Path)
	if err != nil {
		return err
	}
	defer func(srcFile *os.File) {
		err := srcFile.Close()
		if err != nil {

		}
	}(srcFile)
	dstFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, fh.Perm)
	if err != nil {
		return err
	}
	defer func(dstFile *os.File) {
		err := dstFile.Close()
		if err != nil {

		}
	}(dstFile)
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	if copyAttr {
		err = fh.CopyFileAttr(targetPath)
		if err != nil {
			return err
		}
	}
	return nil
}
