package php

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/helper"
)

// phpAvailable 标记 PHP 命令是否可用，在 init 中检测一次
var phpAvailable bool

func init() {
	_, err := exec.LookPath("php")
	phpAvailable = err == nil
}

// buildExpectedPath 按照 main.go 中的拼接规则构造期望路径
// 函数名：buildExpectedPath
// 参数：base string — RuntimePath 或 ConfigSource 所在目录
// 返回值：string — 期望的 PHP 运行时文件路径
// 异常：无
func buildExpectedPath(base string) string {
	return filepath.Join(helper.NewFile(&helper.File{Path: base}).DirName(), "tmp", "php", "main.php")
}

// TestNew_WithRuntimePath 验证指定 RuntimePath 时生成的 PHP 文件路径正确且文件被创建
// 函数名：TestNew_WithRuntimePath
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：路径不符合预期或文件未创建时测试失败
// 使用示例：go test ./component/php/...
func TestNew_WithRuntimePath(t *testing.T) {
	tmpDir := t.TempDir()
	opt := jcbaseGo.Option{RuntimePath: tmpDir}

	conf := New(opt)
	expected := buildExpectedPath(tmpDir)
	if conf.funcFilePath != expected {
		t.Fatalf("funcFilePath mismatch: got %q, want %q", conf.funcFilePath, expected)
	}

	if !helper.NewFile(&helper.File{Path: conf.funcFilePath}).Exists() {
		t.Fatalf("expected PHP file %q to be created", conf.funcFilePath)
	}

	content, err := os.ReadFile(conf.funcFilePath)
	if err != nil {
		t.Fatalf("read created PHP file failed: %v", err)
	}
	if !strings.Contains(string(content), "<?php") {
		t.Fatalf("created file does not look like a PHP script")
	}
}

// TestNew_WithConfigSource 验证未指定 RuntimePath 时根据 ConfigSource 生成 PHP 文件路径
// 函数名：TestNew_WithConfigSource
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：路径不符合预期或文件未创建时测试失败
// 使用示例：go test ./component/php/...
func TestNew_WithConfigSource(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "conf.ini")
	if err := os.WriteFile(configFile, []byte("[test]\n"), 0644); err != nil {
		t.Fatalf("create temp config file failed: %v", err)
	}

	opt := jcbaseGo.Option{ConfigSource: configFile}
	conf := New(opt)
	expected := buildExpectedPath(configFile)
	if conf.funcFilePath != expected {
		t.Fatalf("funcFilePath mismatch: got %q, want %q", conf.funcFilePath, expected)
	}

	if !helper.NewFile(&helper.File{Path: conf.funcFilePath}).Exists() {
		t.Fatalf("expected PHP file %q to be created", conf.funcFilePath)
	}
}

// TestNew_DoesNotOverwriteExistingFile 验证已存在的 PHP 文件不会被覆盖
// 函数名：TestNew_DoesNotOverwriteExistingFile
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：已有文件内容被覆盖时测试失败
// 使用示例：go test ./component/php/...
func TestNew_DoesNotOverwriteExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	opt := jcbaseGo.Option{RuntimePath: tmpDir}

	// 预先创建目标文件并写入自定义内容
	expectedPath := buildExpectedPath(tmpDir)
	if err := os.MkdirAll(filepath.Dir(expectedPath), 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	customContent := "<?php // custom marker\n"
	if err := os.WriteFile(expectedPath, []byte(customContent), 0644); err != nil {
		t.Fatalf("write custom PHP file failed: %v", err)
	}

	conf := New(opt)
	if conf.funcFilePath != expectedPath {
		t.Fatalf("funcFilePath mismatch: got %q, want %q", conf.funcFilePath, expectedPath)
	}

	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("read existing PHP file failed: %v", err)
	}
	if string(content) != customContent {
		t.Fatalf("existing PHP file should not be overwritten")
	}
}

// TestRunFunc_BuiltIn 验证可以成功调用 PHP 内置函数
// 函数名：TestRunFunc_BuiltIn
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：PHP 不可用或调用结果不符合预期时测试失败/跳过
// 使用示例：go test ./component/php/...
func TestRunFunc_BuiltIn(t *testing.T) {
	if !phpAvailable {
		t.Skip("php command not found in PATH")
	}

	tmpDir := t.TempDir()
	conf := New(jcbaseGo.Option{RuntimePath: tmpDir})

	result, err := conf.RunFunc("strtoupper", "hello")
	if err != nil {
		t.Fatalf("RunFunc failed: %v", err)
	}
	if strings.TrimSpace(result) != "HELLO" {
		t.Fatalf("unexpected result: got %q, want %q", strings.TrimSpace(result), "HELLO")
	}
}

// TestRunFunc_NotExists 验证调用不存在的 PHP 函数会返回错误并输出错误提示
// 函数名：TestRunFunc_NotExists
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：PHP 不可用、未返回错误，或输出不包含预期错误信息时测试失败/跳过
// 使用示例：go test ./component/php/...
func TestRunFunc_NotExists(t *testing.T) {
	if !phpAvailable {
		t.Skip("php command not found in PATH")
	}

	tmpDir := t.TempDir()
	conf := New(jcbaseGo.Option{RuntimePath: tmpDir})

	result, err := conf.RunFunc("jcbaseGo_non_existent_function_12345")
	if err == nil {
		t.Fatalf("expected error for non-existent function")
	}
	if !strings.Contains(result, "not exists") {
		t.Fatalf("expected output to contain 'not exists', got %q", result)
	}
}

// TestRunFunc_PhpExtension 验证生成的运行时文件扩展名为 .php 而非 .go
// 函数名：TestRunFunc_PhpExtension
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：扩展名不是 .php 时测试失败
// 使用示例：go test ./component/php/...
func TestRunFunc_PhpExtension(t *testing.T) {
	tmpDir := t.TempDir()
	conf := New(jcbaseGo.Option{RuntimePath: tmpDir})

	if !strings.HasSuffix(conf.funcFilePath, ".php") {
		t.Fatalf("expected funcFilePath to end with .php, got %q", conf.funcFilePath)
	}
	if strings.HasSuffix(conf.funcFilePath, ".go") {
		t.Fatalf("funcFilePath should not end with .go, got %q", conf.funcFilePath)
	}
}
