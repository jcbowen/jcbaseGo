package main

import (
	"fmt"
	"log"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/php"
)

func main() {
	fmt.Println("=== PHP 组件基本用法示例 ===")

	// 创建 PHP 组件实例
	phpComponent := php.New(jcbaseGo.Option{})

	// 示例1: 调用无参数的 PHP 函数
	fmt.Println("\n1. 调用 phpinfo() 函数:")
	result, err := phpComponent.RunFunc("phpinfo")
	if err != nil {
		log.Printf("调用 phpinfo 失败: %v", err)
	} else {
		fmt.Printf("PHP 信息长度: %d 字符\n", len(result))
	}

	// 示例2: 调用字符串处理函数
	fmt.Println("\n2. 调用字符串处理函数:")

	// strtoupper - 转换为大写
	result, err = phpComponent.RunFunc("strtoupper", "hello world")
	if err != nil {
		log.Printf("调用 strtoupper 失败: %v", err)
	} else {
		fmt.Printf("strtoupper('hello world') = %s\n", result)
	}

	// strlen - 获取字符串长度
	result, err = phpComponent.RunFunc("strlen", "你好世界")
	if err != nil {
		log.Printf("调用 strlen 失败: %v", err)
	} else {
		fmt.Printf("strlen('你好世界') = %s\n", result)
	}

	// substr - 字符串截取
	result, err = phpComponent.RunFunc("substr", "Hello World", "0", "5")
	if err != nil {
		log.Printf("调用 substr 失败: %v", err)
	} else {
		fmt.Printf("substr('Hello World', 0, 5) = %s\n", result)
	}

	// 示例3: 调用数学函数
	fmt.Println("\n3. 调用数学函数:")

	// pow - 幂运算
	result, err = phpComponent.RunFunc("pow", "2", "3")
	if err != nil {
		log.Printf("调用 pow 失败: %v", err)
	} else {
		fmt.Printf("pow(2, 3) = %s\n", result)
	}

	// sqrt - 平方根
	result, err = phpComponent.RunFunc("sqrt", "16")
	if err != nil {
		log.Printf("调用 sqrt 失败: %v", err)
	} else {
		fmt.Printf("sqrt(16) = %s\n", result)
	}

	// 示例4: 调用数组函数
	fmt.Println("\n4. 调用数组函数:")

	// count - 数组长度
	result, err = phpComponent.RunFunc("count", `["apple","banana","orange"]`)
	if err != nil {
		log.Printf("调用 count 失败: %v", err)
	} else {
		fmt.Printf("count(['apple','banana','orange']) = %s\n", result)
	}

	// 示例5: 调用日期时间函数
	fmt.Println("\n5. 调用日期时间函数:")

	// date - 格式化日期
	result, err = phpComponent.RunFunc("date", "Y-m-d H:i:s")
	if err != nil {
		log.Printf("调用 date 失败: %v", err)
	} else {
		fmt.Printf("当前时间: %s\n", result)
	}

	// time - 获取时间戳
	result, err = phpComponent.RunFunc("time")
	if err != nil {
		log.Printf("调用 time 失败: %v", err)
	} else {
		fmt.Printf("当前时间戳: %s\n", result)
	}

	// 示例6: 调用 JSON 函数
	fmt.Println("\n6. 调用 JSON 函数:")

	// json_encode - 编码为 JSON
	testData := `{"name":"张三","age":25,"city":"北京"}`
	result, err = phpComponent.RunFunc("json_encode", testData)
	if err != nil {
		log.Printf("调用 json_encode 失败: %v", err)
	} else {
		fmt.Printf("JSON 编码结果: %s\n", result)
	}

	// json_decode - 解码 JSON
	result, err = phpComponent.RunFunc("json_decode", testData, "true")
	if err != nil {
		log.Printf("调用 json_decode 失败: %v", err)
	} else {
		fmt.Printf("JSON 解码结果: %s\n", result)
	}

	fmt.Println("\n=== 基本用法示例完成 ===")
}
