package php

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/command"
	"github.com/jcbowen/jcbaseGo/component/helper"
)

// templateEntrySignature 生成模板的入口签名，用于区分「库生成的运行时脚本」与「用户自定义脚本」
const templateEntrySignature = "(new Console)->run();"

// phpTemplateMarker 模板版本标记，随模板演进更新；
// 运行时已生成脚本若含该标记且内容与当前模板不一致，将在 New 时自动重写（免手动清缓存升级）
const phpTemplateMarker = "@jcbasePHP-template v2"

type ConfigStruct struct {
	funcFilePath string
	// prelude 前置 PHP 代码（不含 <?php 标签），在主脚本执行前加载。
	// 典型用途：为 CLI 环境垫上目标系统中不存在的类（跨项目解析序列化数据）、注册容错解析函数。
	// 以 prelude.php 形式落盘到主脚本同目录，由主脚本 @include_once 加载。
	prelude string
}

func New(opt jcbaseGo.Option) *ConfigStruct {
	baseDir := helper.NewFile(&helper.File{Path: opt.ConfigSource}).DirName()
	if opt.RuntimePath != "" {
		baseDir = helper.NewFile(&helper.File{Path: opt.RuntimePath}).DirName()
	}
	funcFilePath := filepath.Join(baseDir, "tmp", "php", "main.php")

	conf := &ConfigStruct{
		funcFilePath: funcFilePath,
	}

	conf.syncTemplate()

	return conf
}

// syncTemplate 同步运行时 PHP 脚本：
//   - 文件不存在 → 写入当前模板；
//   - 文件由旧版模板生成（含入口签名但内容与当前模板不一致）→ 重写为新模板，保证库升级后自动生效；
//   - 用户自定义内容（不含入口签名）→ 保持不动。
//
// 注意：若用户在生成的脚本基础上做过定制（仍含入口签名），升级时会被新模板覆盖，请自行备份。
func (c *ConfigStruct) syncTemplate() {
	file := helper.NewFile(&helper.File{Path: c.funcFilePath})
	if !file.Exists() {
		if err := file.CreateFile([]byte(TmpJcbasePHP), true); err != nil {
			log.Panic(err)
		}
		return
	}

	content, err := os.ReadFile(c.funcFilePath)
	if err != nil {
		log.Panic(err)
	}
	if strings.Contains(string(content), templateEntrySignature) && string(content) != TmpJcbasePHP {
		if err := file.CreateFile([]byte(TmpJcbasePHP), true); err != nil {
			log.Panic(err)
		}
	}
}

// WithClassAliases 注册类别名：在主脚本执行前执行 class_alias(real, alias)（目标类已存在时跳过）。
// 注意：class_alias 的源类必须是用户自定义类，内部类（如 ArrayObject）不支持，此类场景请用
// WithPrelude 配合 eval 定义 extends 该内部类的 stub 类。
//
// 参数:
//   - aliases: map[别名]真实类名
//
// 返回值:
//   - *ConfigStruct: 支持链式调用
//
// 示例:
//
//	conf.WithClassAliases(map[string]string{"app\\models\\User": "app\\models\\BaseUser"})
func (c *ConfigStruct) WithClassAliases(aliases map[string]string) *ConfigStruct {
	var sb strings.Builder
	for alias, real := range aliases {
		fmt.Fprintf(&sb, "if (!class_exists(%q, false)) { class_alias(%q, %q); }\n", alias, real, alias)
	}
	return c.appendPrelude(sb.String())
}

// WithPrelude 附加原始前置 PHP 代码（不含 <?php 标签），在主脚本执行前加载。
// 可用于定义类别名（eval 方式）、注册容错解析函数等；多个 WithPrelude 调用按调用顺序拼接。
//
// 参数:
//   - code: PHP 代码，不应定义名为 Console 的类（与主脚本冲突）
//
// 返回值:
//   - *ConfigStruct: 支持链式调用
//
// 示例:
//
//	conf.WithPrelude(`if (!class_exists('app\\models\\User', false)) {
//	    eval('namespace app\\models; class User extends \\ArrayObject {}');
//	}`)
func (c *ConfigStruct) WithPrelude(code string) *ConfigStruct {
	return c.appendPrelude(code)
}

// appendPrelude 追加前置代码，自动剥离调用方误带的 <?php 标签
func (c *ConfigStruct) appendPrelude(code string) *ConfigStruct {
	code = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(code), "<?php"))
	if code == "" {
		return c
	}
	if c.prelude != "" {
		c.prelude += "\n"
	}
	c.prelude += code
	return c
}

// preludeFilePath 前置代码落盘路径（与主脚本同目录）
func (c *ConfigStruct) preludeFilePath() string {
	return filepath.Join(filepath.Dir(c.funcFilePath), "prelude.php")
}

// syncPrelude 将前置代码落盘（内容变化才写，避免无谓的 mtime 变更）；
// 未配置前置代码时清理遗留文件，避免污染同机其他实例。
//
// 返回值:
//   - string: 落盘路径（未配置时为空串）
//   - error: 写入/清理失败时返回错误
func (c *ConfigStruct) syncPrelude() (string, error) {
	preludePath := c.preludeFilePath()
	if strings.TrimSpace(c.prelude) == "" {
		if helper.NewFile(&helper.File{Path: preludePath}).Exists() {
			if err := os.Remove(preludePath); err != nil {
				return "", err
			}
		}
		return "", nil
	}

	content := "<?php\n" + c.prelude + "\n"
	existing, err := os.ReadFile(preludePath)
	if err == nil && string(existing) == content {
		return preludePath, nil
	}
	if err := helper.NewFile(&helper.File{Path: preludePath}).CreateFile([]byte(content), true); err != nil {
		return "", err
	}
	return preludePath, nil
}

// RunFunc 调用 PHP 函数并返回结果
//
// 参数:
//   - funcName: 全局 PHP 函数名（前置代码中定义的函数同样可用）
//   - args: 传给函数的参数（字符串形式）
//
// 返回值:
//   - string: 函数返回值（数组会先 json_encode）
//   - error: 执行失败时返回错误
func (c *ConfigStruct) RunFunc(funcName string, args ...string) (string, error) {

	funcName = "--func=" + funcName

	if _, err := c.syncPrelude(); err != nil {
		return "", fmt.Errorf("写入前置 PHP 代码失败: %w", err)
	}

	// 往args前面追加"jcbasePHP", funcName
	args = append([]string{c.funcFilePath, funcName}, args...)

	return command.Run("php", args...)
}
