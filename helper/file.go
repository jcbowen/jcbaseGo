package helper

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
)

var defaultPerm os.FileMode = 0755

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// ReadJsonFile 读取json文件，并解析到结构体
func ReadJsonFile(filePath string, data any) error {
	// 读取json配置文件
	file, fErr := os.ReadFile(filePath)
	if fErr != nil {
		return fErr
	}
	fileDataString := string(file)

	err := json.Unmarshal([]byte(fileDataString), &data)
	return err
}

// IsDir 判断是否是目录
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// GetAbsPath 获取绝对路径
func GetAbsPath(path string) (string, error) {
	return filepath.Abs(path)
}

// DirName 获取目录部分
func DirName(path string) string {
	return filepath.Dir(path)
}

// IsFile 判断是否是文件
func IsFile(path string) bool {
	return !IsDir(path)
}

// IsEmptyDir 判断目录是否为空
func IsEmptyDir(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)
	_, err = f.Readdirnames(1)
	if err == os.ErrNotExist {
		return true
	}
	return false
}

// IsEmptyFile 判断文件是否为空
func IsEmptyFile(path string) bool {
	f, err := os.Open(path)
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
func IsEmpty(path string) bool {
	if IsDir(path) {
		return IsEmptyDir(path)
	}
	return IsEmptyFile(path)
}

// IsReadable 判断文件是否可读
func IsReadable(path string) bool {
	_, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return false
	}
	return true
}

// IsWritable 判断文件是否可写
func IsWritable(path string) bool {
	_, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		return false
	}
	return true
}

// IsExecutable 判断文件是否可执行
func IsExecutable(path string) bool {
	_, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return false
	}
	return true
}

// IsSymlink 判断是否是软链接
func IsSymlink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// IsHidden 判断文件是否隐藏
func IsHidden(path string) bool {
	return len(path) > 1 && path[0] == '.'
}

// DirExists 判断目录是否存在
func DirExists(path string, create bool, perm os.FileMode) (bool, error) {
	// 判断path是否为一个目录，如果不是目录则取出目录部分
	if !IsDir(path) {
		path = DirName(path)
		if path == "." || path == "/" {
			return false, errors.New("请输入正确的目录路径(不能为当前目录或根目录;目录必须以/结尾，否则目录名会被当做文件处理)")
		}
	}
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if create {
				if perm == 0 {
					perm = defaultPerm
				}
				err = os.MkdirAll(path, perm)
				if err != nil {
					return false, err
				}
				return true, nil
			}
			return false, nil
		}
		return true, err
	}
	return true, nil
}

// CreateFileIfNotExist 判断文件是否存在，不存在则根据传入的文件内容创建，可设置文件权限
//
// Deprecated: As of jcbaseGo 0.2.1, this function simply calls CreateFile.
func CreateFileIfNotExist(path string, content []byte, perm os.FileMode, overwrite bool) error {
	return CreateFile(path, content, perm, false)
}

// CreateFile 创建文件，可设置文件权限，可设置是否覆盖
func CreateFile(path string, content []byte, perm os.FileMode, overwrite bool) error {
	// 如果已经存在且不需要覆盖则返回错误
	if exists := FileExists(path); exists {
		if !overwrite {
			return errors.New("文件已存在，路径：" + path)
		}
	}

	// 如果没有设置权限则使用默认权限
	if perm == 0 {
		perm = defaultPerm
	}

	// 检查目录是否存在，不存在则创建
	_, err := DirExists(path, true, perm)
	if err != nil {
		return err
	}

	// 创建文件
	return os.WriteFile(path, content, perm)
}

// Remove 删除文件或目录
func Remove(path string) error {
	return os.RemoveAll(path)
}

// CopyFileAttr 复制文件属性
func CopyFileAttr(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

// CopyFile 复制文件，可设置是否覆盖，可设置文件权限，可设置是否复制文件属性
func CopyFile(src, dst string, overwrite bool, perm os.FileMode, copyAttr bool) error {
	if !FileExists(src) {
		return errors.New("file not exists")
	}
	if FileExists(dst) {
		if overwrite {
			err := os.Remove(dst)
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(srcFile *os.File) {
		err := srcFile.Close()
		if err != nil {

		}
	}(srcFile)
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, perm)
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
		err = CopyFileAttr(src, dst)
		if err != nil {
			return err
		}
	}
	return nil
}

// ParseIP 解析IP地址，输出是ipv4或ipv6
// 0: invalid ip
// 4: ipv4
// 6: ipv6
func ParseIP(s string) (net.IP, int) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, 0
	}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return ip, 4
		case ':':
			return ip, 6
		}
	}
	return nil, 0
}
