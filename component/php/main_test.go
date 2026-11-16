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

// TestNew_TemplateRegenerated 验证旧版模板生成的运行时脚本会被自动升级为新模板
// 函数名：TestNew_TemplateRegenerated
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：升级后的内容不符合预期时测试失败
// 使用示例：go test ./component/php/...
func TestNew_TemplateRegenerated(t *testing.T) {
	tmpDir := t.TempDir()
	expectedPath := buildExpectedPath(tmpDir)
	if err := os.MkdirAll(filepath.Dir(expectedPath), 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	// 模拟旧版模板生成的脚本：含入口签名，但不含新模板的版本标记与 prelude 加载
	oldGenerated := "#!/usr/bin/env php\n<?php\n\n(new Console)->run();\n\nclass Console\n{\n}\n"
	if err := os.WriteFile(expectedPath, []byte(oldGenerated), 0644); err != nil {
		t.Fatalf("write old generated PHP file failed: %v", err)
	}

	New(jcbaseGo.Option{RuntimePath: tmpDir})

	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("read regenerated PHP file failed: %v", err)
	}
	if !strings.Contains(string(content), phpTemplateMarker) {
		t.Fatalf("old generated template should be upgraded to current template")
	}
	if !strings.Contains(string(content), "prelude.php") {
		t.Fatalf("current template should include prelude.php")
	}
}

// TestRunFunc_WithPrelude 验证 WithPrelude 注入的前置代码在主脚本执行前生效（可定义函数）
// 函数名：TestRunFunc_WithPrelude
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：PHP 不可用或前置函数调用失败时测试失败/跳过
// 使用示例：go test ./component/php/...
func TestRunFunc_WithPrelude(t *testing.T) {
	if !phpAvailable {
		t.Skip("php command not found in PATH")
	}

	tmpDir := t.TempDir()
	conf := New(jcbaseGo.Option{RuntimePath: tmpDir}).
		WithPrelude(`function jc_test_prelude_hello() { return 'hello from prelude'; }`)

	result, err := conf.RunFunc("jc_test_prelude_hello")
	if err != nil {
		t.Fatalf("RunFunc with prelude failed: %v", err)
	}
	if strings.TrimSpace(result) != "hello from prelude" {
		t.Fatalf("unexpected result: got %q", strings.TrimSpace(result))
	}
}

// TestRunFunc_WithClassAliases 验证 WithClassAliases 注入的类别名在主脚本执行前生效
// 函数名：TestRunFunc_WithClassAliases
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：PHP 不可用或别名未生效时测试失败/跳过
// 使用示例：go test ./component/php/...
func TestRunFunc_WithClassAliases(t *testing.T) {
	if !phpAvailable {
		t.Skip("php command not found in PATH")
	}

	tmpDir := t.TempDir()
	conf := New(jcbaseGo.Option{RuntimePath: tmpDir}).
		WithPrelude(`class jc_test_base_user { }`).
		WithClassAliases(map[string]string{"jc_test_alias_user": "jc_test_base_user"})

	result, err := conf.RunFunc("class_exists", "jc_test_alias_user")
	if err != nil {
		t.Fatalf("RunFunc with class aliases failed: %v", err)
	}
	if strings.TrimSpace(result) != "1" {
		t.Fatalf("alias class should exist, got %q", strings.TrimSpace(result))
	}
}

// TestRunFunc_TolerantUnserialize 验证「垫类 + 容错反序列化」可解析内嵌未知类对象的序列化数据。
// 复刻跨项目场景：数据由 Web 端 PHP 生成，内嵌 ArrayObject 容器对象；CLI 环境通过 eval
// 定义 extends \ArrayObject 的 stub 类后，对象存储可无损还原。
// 函数名：TestRunFunc_TolerantUnserialize
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：PHP 不可用或解析结果不符合预期时测试失败/跳过
// 使用示例：go test ./component/php/...
func TestRunFunc_TolerantUnserialize(t *testing.T) {
	if !phpAvailable {
		t.Skip("php command not found in PATH")
	}

	tmpDir := t.TempDir()
	conf := New(jcbaseGo.Option{RuntimePath: tmpDir}).WithPrelude(`
if (!class_exists('common\components\Request', false)) {
    eval('namespace common\components; class Request extends \ArrayObject {}');
}
function jc_test_unserialize_tolerant($str) {
    $v = @unserialize((string)$str);
    if (is_object($v)) {
        if (get_class($v) === '__PHP_Incomplete_Class') {
            $v = get_object_vars($v);
            unset($v['__PHP_Incomplete_Class_Name']);
        } elseif ($v instanceof ArrayObject) {
            $v = $v->getArrayCopy();
        } elseif (method_exists($v, 'getArrayCopy')) {
            $v = $v->getArrayCopy();
        } else {
            $v = get_object_vars($v);
        }
    }
    return $v;
}`)

	// Yii2 缓存包装形态：[值, 依赖]，值内嵌 common\components\Request 对象（O: 格式，
	// 槽位：i:0=flags、i:1=存储数组、i:2=对象属性数组）。类名 "common\components\Request" 长度 25。
	fixture := `a:2:{i:0;O:25:"common\components\Request":3:{i:0;i:0;i:1;a:2:{s:7:"user_id";i:439426;s:6:"openid";s:6:"oXyz12";}i:2;a:0:{}}i:1;N;}`

	result, err := conf.RunFunc("jc_test_unserialize_tolerant", fixture)
	if err != nil {
		t.Fatalf("tolerant unserialize failed: %v", err)
	}
	if !strings.Contains(result, "439426") {
		t.Fatalf("expected user_id in result, got %q", result)
	}
	if !strings.Contains(result, "oXyz12") {
		t.Fatalf("expected openid in result, got %q", result)
	}

	// 垫类未覆盖的未知类：不应 Fatal，降级为属性提取
	unknownFixture := `a:2:{i:0;O:9:"SomeThing":1:{s:3:"foo";i:1;}i:1;N;}`
	result, err = conf.RunFunc("jc_test_unserialize_tolerant", unknownFixture)
	if err != nil {
		t.Fatalf("unknown class should not fatal: %v", err)
	}
	if !strings.Contains(result, `"foo"`) {
		t.Fatalf("expected property extraction for unknown class, got %q", result)
	}
}
