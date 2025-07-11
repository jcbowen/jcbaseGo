#!/bin/bash

# PHP 组件示例运行脚本
# 确保系统已安装 PHP 和 Go

echo "=== PHP 组件示例运行脚本 ==="
echo ""

# 检查 PHP 是否可用
if ! command -v php &> /dev/null; then
    echo "错误: PHP 未安装或不在 PATH 中"
    echo "请先安装 PHP:"
    echo "  macOS: brew install php"
    echo "  Ubuntu/Debian: sudo apt-get install php-cli"
    echo "  CentOS/RHEL: sudo yum install php-cli"
    exit 1
fi

# 检查 Go 是否可用
if ! command -v go &> /dev/null; then
    echo "错误: Go 未安装或不在 PATH 中"
    exit 1
fi

echo "PHP 版本: $(php --version | head -n 1)"
echo "Go 版本: $(go version)"
echo ""

# 设置工作目录
cd "$(dirname "$0")"

# 运行基本用法示例
echo "运行基本用法示例..."
cd basic
go run main.go
echo ""

# 运行字符串处理示例
echo "运行字符串处理示例..."
cd ../string
go run main.go
echo ""

# 运行数学计算示例
echo "运行数学计算示例..."
cd ../math
go run main.go
echo ""

# 运行 JSON 处理示例
echo "运行 JSON 处理示例..."
cd ../json
go run main.go
echo ""

# 运行序列化和反序列化示例
echo "运行序列化和反序列化示例..."
cd ../serialize
go run main.go
echo ""

# 运行自定义函数示例
echo "运行自定义函数示例..."
cd ../custom
go run main.go
echo ""

# 运行序列化辅助工具示例
echo "运行序列化辅助工具示例..."
cd ../utils_example
go run main.go
echo ""

echo "=== 所有示例运行完成 ==="
echo ""
echo "注意事项:"
echo "1. 确保 PHP 命令行环境正确配置"
echo "2. 某些 PHP 扩展可能需要额外安装"
echo "3. 自定义函数示例仅展示函数定义，需要在实际环境中使用"
echo "4. 如果遇到权限问题，请检查文件权限设置" 