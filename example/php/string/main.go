package main

import (
	"fmt"
	"log"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/php"
)

func main() {
	fmt.Println("=== PHP 字符串处理示例 ===")

	// 创建 PHP 组件实例
	phpComponent := php.New(jcbaseGo.Option{})

	// 示例1: 字符串大小写转换
	fmt.Println("\n1. 字符串大小写转换:")

	testString := "Hello World 你好世界"

	// 转换为大写
	result, err := phpComponent.RunFunc("strtoupper", testString)
	if err != nil {
		log.Printf("调用 strtoupper 失败: %v", err)
	} else {
		fmt.Printf("strtoupper('%s') = %s\n", testString, result)
	}

	// 转换为小写
	result, err = phpComponent.RunFunc("strtolower", testString)
	if err != nil {
		log.Printf("调用 strtolower 失败: %v", err)
	} else {
		fmt.Printf("strtolower('%s') = %s\n", testString, result)
	}

	// 首字母大写
	result, err = phpComponent.RunFunc("ucfirst", "hello world")
	if err != nil {
		log.Printf("调用 ucfirst 失败: %v", err)
	} else {
		fmt.Printf("ucfirst('hello world') = %s\n", result)
	}

	// 每个单词首字母大写
	result, err = phpComponent.RunFunc("ucwords", "hello world")
	if err != nil {
		log.Printf("调用 ucwords 失败: %v", err)
	} else {
		fmt.Printf("ucwords('hello world') = %s\n", result)
	}

	// 示例2: 字符串长度和截取
	fmt.Println("\n2. 字符串长度和截取:")

	// 获取字符串长度
	result, err = phpComponent.RunFunc("strlen", testString)
	if err != nil {
		log.Printf("调用 strlen 失败: %v", err)
	} else {
		fmt.Printf("strlen('%s') = %s\n", testString, result)
	}

	// 截取字符串
	result, err = phpComponent.RunFunc("substr", testString, "0", "5")
	if err != nil {
		log.Printf("调用 substr 失败: %v", err)
	} else {
		fmt.Printf("substr('%s', 0, 5) = %s\n", testString, result)
	}

	// 从末尾截取
	result, err = phpComponent.RunFunc("substr", testString, "-5")
	if err != nil {
		log.Printf("调用 substr 失败: %v", err)
	} else {
		fmt.Printf("substr('%s', -5) = %s\n", testString, result)
	}

	// 示例3: 字符串查找和替换
	fmt.Println("\n3. 字符串查找和替换:")

	searchString := "Hello World"

	// 查找子字符串位置
	result, err = phpComponent.RunFunc("strpos", searchString, "World")
	if err != nil {
		log.Printf("调用 strpos 失败: %v", err)
	} else {
		fmt.Printf("strpos('%s', 'World') = %s\n", searchString, result)
	}

	// 查找子字符串位置（不区分大小写）
	result, err = phpComponent.RunFunc("stripos", searchString, "world")
	if err != nil {
		log.Printf("调用 stripos 失败: %v", err)
	} else {
		fmt.Printf("stripos('%s', 'world') = %s\n", searchString, result)
	}

	// 字符串替换
	result, err = phpComponent.RunFunc("str_replace", "World", "PHP", searchString)
	if err != nil {
		log.Printf("调用 str_replace 失败: %v", err)
	} else {
		fmt.Printf("str_replace('World', 'PHP', '%s') = %s\n", searchString, result)
	}

	// 示例4: 字符串分割和合并
	fmt.Println("\n4. 字符串分割和合并:")

	// 按空格分割字符串
	result, err = phpComponent.RunFunc("explode", " ", "apple banana orange")
	if err != nil {
		log.Printf("调用 explode 失败: %v", err)
	} else {
		fmt.Printf("explode(' ', 'apple banana orange') = %s\n", result)
	}

	// 按多个分隔符分割
	result, err = phpComponent.RunFunc("preg_split", "/[,\\s]+/", "apple,banana orange,grape")
	if err != nil {
		log.Printf("调用 preg_split 失败: %v", err)
	} else {
		fmt.Printf("preg_split('/[,\\s]+/', 'apple,banana orange,grape') = %s\n", result)
	}

	// 数组合并为字符串
	result, err = phpComponent.RunFunc("implode", "-", `["apple","banana","orange"]`)
	if err != nil {
		log.Printf("调用 implode 失败: %v", err)
	} else {
		fmt.Printf("implode('-', ['apple','banana','orange']) = %s\n", result)
	}

	// 示例5: 字符串格式化
	fmt.Println("\n5. 字符串格式化:")

	// 格式化字符串
	result, err = phpComponent.RunFunc("sprintf", "Hello %s, you are %d years old", "张三", "25")
	if err != nil {
		log.Printf("调用 sprintf 失败: %v", err)
	} else {
		fmt.Printf("sprintf('Hello %%s, you are %%d years old', '张三', 25) = %s\n", result)
	}

	// 数字格式化
	result, err = phpComponent.RunFunc("number_format", "1234.5678", "2", ".", ",")
	if err != nil {
		log.Printf("调用 number_format 失败: %v", err)
	} else {
		fmt.Printf("number_format(1234.5678, 2, '.', ',') = %s\n", result)
	}

	// 示例6: 字符串清理和验证
	fmt.Println("\n6. 字符串清理和验证:")

	// 去除首尾空白字符
	result, err = phpComponent.RunFunc("trim", "  Hello World  ")
	if err != nil {
		log.Printf("调用 trim 失败: %v", err)
	} else {
		fmt.Printf("trim('  Hello World  ') = '%s'\n", result)
	}

	// 去除左侧空白字符
	result, err = phpComponent.RunFunc("ltrim", "  Hello World  ")
	if err != nil {
		log.Printf("调用 ltrim 失败: %v", err)
	} else {
		fmt.Printf("ltrim('  Hello World  ') = '%s'\n", result)
	}

	// 去除右侧空白字符
	result, err = phpComponent.RunFunc("rtrim", "  Hello World  ")
	if err != nil {
		log.Printf("调用 rtrim 失败: %v", err)
	} else {
		fmt.Printf("rtrim('  Hello World  ') = '%s'\n", result)
	}

	// 检查是否为数字
	result, err = phpComponent.RunFunc("is_numeric", "123.45")
	if err != nil {
		log.Printf("调用 is_numeric 失败: %v", err)
	} else {
		fmt.Printf("is_numeric('123.45') = %s\n", result)
	}

	// 检查是否为字母
	result, err = phpComponent.RunFunc("ctype_alpha", "HelloWorld")
	if err != nil {
		log.Printf("调用 ctype_alpha 失败: %v", err)
	} else {
		fmt.Printf("ctype_alpha('HelloWorld') = %s\n", result)
	}

	// 示例7: 字符串编码和转义
	fmt.Println("\n7. 字符串编码和转义:")

	// HTML 实体编码
	result, err = phpComponent.RunFunc("htmlspecialchars", "<script>alert('test')</script>")
	if err != nil {
		log.Printf("调用 htmlspecialchars 失败: %v", err)
	} else {
		fmt.Printf("htmlspecialchars('<script>alert(\\'test\\')</script>') = %s\n", result)
	}

	// URL 编码
	result, err = phpComponent.RunFunc("urlencode", "Hello World 你好")
	if err != nil {
		log.Printf("调用 urlencode 失败: %v", err)
	} else {
		fmt.Printf("urlencode('Hello World 你好') = %s\n", result)
	}

	// Base64 编码
	result, err = phpComponent.RunFunc("base64_encode", "Hello World")
	if err != nil {
		log.Printf("调用 base64_encode 失败: %v", err)
	} else {
		fmt.Printf("base64_encode('Hello World') = %s\n", result)
	}

	// Base64 解码
	result, err = phpComponent.RunFunc("base64_decode", "SGVsbG8gV29ybGQ=")
	if err != nil {
		log.Printf("调用 base64_decode 失败: %v", err)
	} else {
		fmt.Printf("base64_decode('SGVsbG8gV29ybGQ=') = %s\n", result)
	}

	fmt.Println("\n=== 字符串处理示例完成 ===")
}
