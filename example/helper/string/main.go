package main

import (
	"fmt"
	"strings"

	"github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
	fmt.Println("=== Helper 字符串处理工具使用示例 ===\n")

	// 1. 字符串截取示例
	fmt.Println("1. 字符串截取示例:")
	stringSubExample()

	// 2. 字符串替换示例
	fmt.Println("\n2. 字符串替换示例:")
	stringReplaceExample()

	// 3. 字符串分割示例
	fmt.Println("\n3. 字符串分割示例:")
	stringSplitExample()

	// 4. 字符串格式化示例
	fmt.Println("\n4. 字符串格式化示例:")
	stringFormatExample()
}

// stringSubExample 字符串截取示例
func stringSubExample() {
	text := "Hello, 世界！这是一个测试字符串"
	str := helper.NewStr(text)

	// 截取前10个字符
	sub1 := str.ByteSubstr(0, 10)
	fmt.Printf("原字符串: %s\n", text)
	fmt.Printf("截取前10个字符: %s\n", sub1)

	// 截取中间部分
	sub2 := str.ByteSubstr(7, 5)
	fmt.Printf("从第7个字符开始截取5个字符: %s\n", sub2)

	// 截取到末尾
	sub3 := str.ByteSubstr(10, -1)
	fmt.Printf("从第10个字符截取到末尾: %s\n", sub3)

	// 处理中文字符
	chineseText := "你好世界"
	chineseStr := helper.NewStr(chineseText)
	sub4 := chineseStr.ByteSubstr(0, 2)
	fmt.Printf("中文字符串: %s\n", chineseText)
	fmt.Printf("截取前2个字符: %s\n", sub4)

	// 字符串截断
	truncated := str.Truncate(15, "...")
	fmt.Printf("截断到15个字符: %s\n", truncated)
}

// stringReplaceExample 字符串替换示例
func stringReplaceExample() {
	text := "Hello World, Hello Go, Hello jcbaseGo"
	str := helper.NewStr(text)

	// 使用 strings.ReplaceAll 进行替换
	replaced := strings.ReplaceAll(text, "Hello", "Hi")
	fmt.Printf("原字符串: %s\n", text)
	fmt.Printf("替换后: %s\n", replaced)

	// 替换多个字符串
	text2 := "apple,banana,orange"
	replaced2 := strings.ReplaceAll(text2, ",", " | ")
	fmt.Printf("原字符串: %s\n", text2)
	fmt.Printf("替换后: %s\n", replaced2)

	// 大小写转换
	upper := str.ToUpper()
	lower := str.ToLower()
	fmt.Printf("转大写: %s\n", upper)
	fmt.Printf("转小写: %s\n", lower)
}

// stringSplitExample 字符串分割示例
func stringSplitExample() {
	text := "apple,banana,orange,grape"
	str := helper.NewStr(text)

	// 按逗号分割
	parts := str.Explode(",", true, false)
	fmt.Printf("原字符串: %s\n", text)
	fmt.Printf("分割结果: %v\n", parts)
	fmt.Printf("分割后元素数量: %d\n", len(parts))

	// 按空格分割
	text2 := "Hello World Go Language"
	str2 := helper.NewStr(text2)
	parts2 := str2.Explode(" ", true, false)
	fmt.Printf("原字符串: %s\n", text2)
	fmt.Printf("分割结果: %v\n", parts2)

	// 处理空字符串，跳过空值
	text3 := "a,,b,,c"
	str3 := helper.NewStr(text3)
	parts3 := str3.Explode(",", true, true)
	fmt.Printf("原字符串: %s\n", text3)
	fmt.Printf("分割结果（跳过空值）: %v\n", parts3)
}

// stringFormatExample 字符串格式化示例
func stringFormatExample() {
	// 格式化字符串
	format := "Hello %s, you are %d years old"
	result := fmt.Sprintf(format, "World", 25)
	fmt.Printf("格式化字符串: %s\n", result)

	// 数字格式化
	number := 123.456
	formatted := fmt.Sprintf("数字: %.2f", number)
	fmt.Printf("数字格式化: %s\n", formatted)

	// 多参数格式化
	formatted2 := fmt.Sprintf("姓名: %s, 年龄: %d, 分数: %.1f", "张三", 20, 95.5)
	fmt.Printf("多参数格式化: %s\n", formatted2)

	// 处理特殊字符
	special := fmt.Sprintf("路径: %s, 引号: \"%s\"", "/path/to/file", "quoted text")
	fmt.Printf("特殊字符: %s\n", special)

	// 字符串工具的其他功能
	text := "hello world"
	str := helper.NewStr(text)

	// 首字母大写
	ucFirst := str.MbUcFirst()
	fmt.Printf("首字母大写: %s\n", ucFirst)

	// 每个单词首字母大写
	ucWords := str.MbUcWords()
	fmt.Printf("每个单词首字母大写: %s\n", ucWords)

	// 驼峰转下划线
	camelCase := "userName"
	camelStr := helper.NewStr(camelCase)
	snakeCase := camelStr.ConvertCamelToSnake()
	fmt.Printf("驼峰转下划线: %s -> %s\n", camelCase, snakeCase)
}
