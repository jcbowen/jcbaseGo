# PHP 组件

在 Go 程序中调用 PHP 函数的组件：自动生成 PHP 引导脚本，通过 `php` 命令行执行并回传结果。适用于在 Go 项目中复用现有 PHP 函数库或与 PHP 体系（如 Yii2 项目）共享计算的场景。

## 概述

PHP 组件以「生成引导脚本 + CLI 调用」的方式打通 Go 与 PHP：

- **零侵入调用**：Go 侧只需 `New` + `RunFunc` 两行，即可调用任意 PHP 函数（内置函数或自定义函数）
- **自动生成引导脚本**：首次 `New` 时按内置模板生成 `tmp/php/main.php`，已存在则不覆盖，可自行定制
- **参数自动转换**：字符串参数中的 `true/false` 转 bool、数字转 int、`{...}` 转 JSON 数组，无需手动构造
- **结果回传**：PHP 函数返回数组时自动 `json_encode`（不转义 Unicode）后经 stdout 回传
- **混合输出**：stdout 与 stderr 合并返回，错误信息不丢失

## 功能特性

### 实例化

- `New(opt jcbaseGo.Option)`：推导引导脚本路径并确保其存在
  - 默认取 `filepath.Dir(opt.ConfigSource)/tmp/php/main.php`
  - `opt.RuntimePath` 非空时优先取 `filepath.Dir(opt.RuntimePath)/tmp/php/main.php`
  - 文件不存在时按内置模板 `TmpJcbasePHP` 创建（自动创建多级目录）

### 函数调用

- `RunFunc(funcName string, args ...string) (string, error)`：执行 PHP 函数，返回合并输出

### 引导脚本能力（内置模板）

- 命令行选项：`-h/--help`、`-v/--version`、`-f/--func <函数名>`
- 参数过滤：自动剔除选项参数，仅保留位置参数
- 参数类型转换：`true/false` → bool、纯数字 → int、`{...}` 形态 → `json_decode` 数组
- 结果输出：数组结果 `json_encode` + `stripslashes`；标量直接输出
- 错误路径：函数不存在时输出 `fatal:Function xxx not exists` 到 stderr 并以退出码 1 结束

## 安装指南

### 依赖要求

- Go 1.26.0+
- PHP 7.0+（需支持命令行执行，`php` 命令在系统 PATH 中可用）

### 环境检查

```bash
php --version
```

### 安装依赖

组件已包含在 jcbaseGo 模块中，无需额外安装：

```bash
go get github.com/jcbowen/jcbaseGo
```

## 使用示例

### 基本使用

```go
package main

import (
    "fmt"
    "log"

    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/php"
)

func main() {
    // 使用默认配置（引导脚本生成在 ./tmp/php/main.php）
    phpComponent := php.New(jcbaseGo.Option{})

    // 调用无参 PHP 函数
    result, err := phpComponent.RunFunc("phpinfo")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result)
}
```

### 指定运行目录

```go
phpComponent := php.New(jcbaseGo.Option{
    // RuntimePath 非空时优先，引导脚本生成在其目录的 tmp/php/ 下
    RuntimePath: "./data/runtime/",
    ConfigSource: "./data/conf.ini",
})
```

### 带参数调用

```go
// 字符串参数
result, err := phpComponent.RunFunc("strtoupper", "hello world")
// 输出: HELLO WORLD

// 数字参数（自动转 int）
result, err = phpComponent.RunFunc("pow", "2", "3")
// 输出: 8

// JSON 对象参数（自动转数组）
result, err = phpComponent.RunFunc("json_encode", `{"name":"张三","age":25}`)
// 输出: {"name":"张三","age":25}
```

### 调用自定义 PHP 函数

引导脚本首次生成后不会覆盖，可直接编辑 `tmp/php/main.php`，在文件中定义自己的函数：

```php
// tmp/php/main.php（在 Console 类定义之外追加）
function calculateSum($a, $b) {
    return $a + $b;
}
```

```go
result, err := phpComponent.RunFunc("calculateSum", "10", "20")
// 输出: 30
```

也可以引用外部 PHP 文件，复用既有代码库：

```php
// tmp/php/main.php 顶部追加
require_once '/path/to/your/functions.php';
```

### 自定义引导脚本模板

内置模板是导出变量 `php.TmpJcbasePHP`，在调用 `New` 之前整体替换即可（例如注入项目共用的 Console 基类）：

```go
php.TmpJcbasePHP = myCustomTemplate
phpComponent := php.New(jcbaseGo.Option{})
```

### 错误处理

```go
result, err := phpComponent.RunFunc("not_exists_func", "a")
if err != nil {
    // err 非 nil：php 进程退出码非 0（如函数不存在、PHP 语法错误）
    log.Printf("执行失败: %v, 输出: %s", err, result)
    return
}
// err 为 nil 但 result 中含 stderr 内容：PHP warning/notice 等非致命输出
```

## 详细功能说明

### 引导脚本路径推导规则

| 配置 | 生成的引导脚本路径 |
|---|---|
| `Option{}`（全空） | `./tmp/php/main.php` |
| `ConfigSource: "./data/conf.ini"` | `./data/tmp/php/main.php` |
| `RuntimePath: "./data/runtime/"` | `./data/runtime/tmp/php/main.php`（优先于 ConfigSource） |

注意推导用的是**目录部分**（`filepath.Dir`）：`RuntimePath` 传目录时带不带结尾斜杠结果一致。

### 参数转换明细（PHP 侧）

| Go 侧传入 | PHP 侧收到 |
|---|---|
| `"true"` / `"false"` | bool |
| `"123"` / `"3.14"` | int（数字串统一转 int） |
| `{"name":"张三"}` | 关联数组（`json_decode(..., true)`） |
| 其他字符串 | string |

浮点参数注意：数字串统一按 int 转换，需要浮点语义时请在 PHP 函数内自行 `floatval`。

### 执行链路

```
RunFunc("strtoupper", "hello")
  → command.Run("php", ["<funcFilePath>", "--func=strtoupper", "hello"])
    → php 进程执行引导脚本，getopt 解析 --func
      → 位置参数过滤与类型转换
        → call_user_func_array("strtoupper", ["hello"])
          → stdout 回传 "HELLO"
```

`command.Run` 使用 `CombinedOutput`，stdout 与 stderr 合并返回；`command.CmdPath` 可指定 php 进程的工作目录。

### 跨语言配合

组件与 PHP 版基础库 `jcbowen/jcbase-yii2` 配合时，可在 PHP 侧引入该库的函数文件，复用其加解密、校验等实现，与 Go 侧保持同一套算法参数（如 SM4 CBC 的 key/iv 约定），实现跨语言互解。

## API 参考

### 构造函数

- `New(opt jcbaseGo.Option) *ConfigStruct` - 创建 PHP 组件实例，确保引导脚本存在

### 实例方法

- `RunFunc(funcName string, args ...string) (string, error)` - 调用 PHP 函数，返回合并输出（stdout+stderr）

### 导出变量

- `TmpJcbasePHP string` - 内置引导脚本模板，可在 `New` 之前替换以定制引导逻辑

## 常见问题

### 提示 command not found: php？

系统未安装 PHP 或不在 PATH 中。安装 php-cli 后重试（macOS: `brew install php`；Ubuntu: `sudo apt-get install php-cli`）。

### 引导脚本生成在非预期目录？

按「引导脚本路径推导规则」检查 `ConfigSource` / `RuntimePath`；生成使用的是目录部分，传文件路径与目录路径效果不同。

### 修改了 main.php 但不生效？

引导脚本只在**不存在时**生成，已存在则跳过。直接编辑已生成的文件即可；如需重新生成，先删除旧文件或在 `New` 前替换 `php.TmpJcbasePHP`。

### PHP warning 混在输出里怎么办？

`RunFunc` 返回的是合并输出。生产环境建议在引导脚本或 php.ini 中关闭 `display_errors`，将错误写入日志，保持 stdout 纯净。

### 函数不存在时返回什么？

PHP 侧以退出码 1 结束，stderr 输出 `fatal:Function xxx not exists`，Go 侧 `err` 非 nil，`result` 中含该错误文本。

## 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个组件。

## 许可证

MIT License
