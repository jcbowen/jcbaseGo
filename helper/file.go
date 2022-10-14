package helper

import (
	"errors"
	"io"
	"os"
	"strings"
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

// IsDir 判断是否是目录
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// DirName 获取文件所在目录
func DirName(path string) string {
	if len(path) < 1 {
		return path
	}
	pathRune := []rune(path)
	if os.IsPathSeparator(uint8(pathRune[len(pathRune)-1])) {
		pathRune = pathRune[len(pathRune)-1:]
	}
	path = string(pathRune)
	tmp := strings.Split(path, string(os.PathSeparator))
	newPath := strings.Join(tmp[:len(tmp)-1], string(os.PathSeparator))
	return newPath
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
			}
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CreateFileIfNotExist 判断文件是否存在，不存在则根据传入的文件内容创建，可设置文件权限，可设置是否覆盖
func CreateFileIfNotExist(path string, content []byte, perm os.FileMode, overwrite bool) error {
	if perm == 0 {
		perm = defaultPerm
	}
	checkDir, err := DirExists(path, true, perm)
	if checkDir && err == nil {
		if !overwrite {
			return errors.New("file already exists")
		}
	}
	err = os.WriteFile(path, content, perm)

	return err
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
