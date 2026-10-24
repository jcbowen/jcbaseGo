package helper

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestFileIsEmptyDir 测试目录空判定
//
// 覆盖: 空目录、含条目目录、路径不存在、路径是普通文件
func TestFileIsEmptyDir(t *testing.T) {
	root := t.TempDir()

	emptyDir := filepath.Join(root, "empty")
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("创建空目录失败: %v", err)
	}

	filledDir := filepath.Join(root, "filled")
	if err := os.Mkdir(filledDir, 0755); err != nil {
		t.Fatalf("创建非空目录失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(filledDir, "a.txt"), []byte("x"), 0644); err != nil {
		t.Fatalf("创建目录内文件失败: %v", err)
	}

	regularFile := filepath.Join(root, "file.txt")
	if err := os.WriteFile(regularFile, []byte("x"), 0644); err != nil {
		t.Fatalf("创建普通文件失败: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"空目录返回 true", emptyDir, true},
		{"含条目目录返回 false", filledDir, false},
		{"路径不存在返回 false", filepath.Join(root, "missing"), false},
		{"普通文件不是空目录", regularFile, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fh := NewFile(&File{Path: tt.path})
			if got := fh.IsEmptyDir(); got != tt.expected {
				t.Errorf("IsEmptyDir(%s) = %v, 期望 %v", tt.path, got, tt.expected)
			}
		})
	}
}

// executableCase 可执行判定的单个用例
type executableCase struct {
	name     string
	path     string
	expected bool
}

// TestFileIsExecutable 测试可执行判定
//
// 平台差异: Unix/macOS 看权限位，Windows 看扩展名，因此分平台断言
func TestFileIsExecutable(t *testing.T) {
	root := t.TempDir()

	dir := filepath.Join(root, "some-dir")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}

	tests := []executableCase{
		{"路径不存在返回 false", filepath.Join(root, "missing"), false},
		{"目录不是可执行文件", dir, false},
	}

	if runtime.GOOS == "windows" {
		tests = append(tests,
			executableCase{"exe 文件按扩展名判定为可执行", newTempFile(t, root, "app.exe"), true},
			executableCase{"大写扩展名同样判定为可执行", newTempFile(t, root, "APP.EXE"), true},
			executableCase{"批处理文件判定为可执行", newTempFile(t, root, "run.bat"), true},
			executableCase{"文本文件不可执行", newTempFile(t, root, "readme.txt"), false},
		)
	} else {
		tests = append(tests,
			executableCase{"权限位含可执行位", newTempFileWithPerm(t, root, "run.sh", 0755), true},
			executableCase{"仅属主可执行", newTempFileWithPerm(t, root, "owner-only", 0700), true},
			executableCase{"普通六四四文件不可执行", newTempFileWithPerm(t, root, "plain.txt", 0644), false},
		)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fh := NewFile(&File{Path: tt.path})
			if got := fh.IsExecutable(); got != tt.expected {
				t.Errorf("IsExecutable(%s) = %v, 期望 %v", tt.path, got, tt.expected)
			}
		})
	}
}

// newTempFile 在 dir 下创建空文件并返回其完整路径
//
// 参数: t 测试句柄; dir 目录; name 文件名
// 返回值: 新建文件的完整路径
// 异常: 创建失败时直接终止当前用例
func newTempFile(t *testing.T, dir, name string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, nil, 0644); err != nil {
		t.Fatalf("创建文件 %s 失败: %v", name, err)
	}
	return path
}

// newTempFileWithPerm 在 dir 下创建文件并设置指定权限，返回其完整路径
//
// 参数: t 测试句柄; dir 目录; name 文件名; perm 文件权限
// 返回值: 新建文件的完整路径
// 异常: 创建或设置权限失败时直接终止当前用例
func newTempFileWithPerm(t *testing.T, dir, name string, perm os.FileMode) string {
	t.Helper()

	path := newTempFile(t, dir, name)
	if err := os.Chmod(path, perm); err != nil {
		t.Fatalf("设置权限 %o 失败: %v", perm, err)
	}
	return path
}
