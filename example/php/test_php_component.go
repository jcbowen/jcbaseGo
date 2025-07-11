package main

import (
	"fmt"
	"log"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/php"
)

func main() {
	fmt.Println("=== PHP 组件功能测试 ===")

	// 创建 PHP 组件实例
	phpComponent := php.New(jcbaseGo.Option{})

	// 测试1: 基本字符串函数
	fmt.Println("\n1. 测试基本字符串函数:")

	result, err := phpComponent.RunFunc("strtoupper", "hello world")
	if err != nil {
		log.Printf("❌ strtoupper 测试失败: %v", err)
	} else {
		fmt.Printf("✅ strtoupper('hello world') = %s\n", result)
	}

	// 测试2: 数学函数
	fmt.Println("\n2. 测试数学函数:")

	result, err = phpComponent.RunFunc("pow", "2", "3")
	if err != nil {
		log.Printf("❌ pow 测试失败: %v", err)
	} else {
		fmt.Printf("✅ pow(2, 3) = %s\n", result)
	}

	// 测试3: 数组函数
	fmt.Println("\n3. 测试数组函数:")

	result, err = phpComponent.RunFunc("count", `["a","b","c"]`)
	if err != nil {
		log.Printf("❌ count 测试失败: %v", err)
	} else {
		fmt.Printf("✅ count(['a','b','c']) = %s\n", result)
	}

	// 测试4: JSON 函数
	fmt.Println("\n4. 测试 JSON 函数:")

	testData := `{"name":"测试","age":25}`
	result, err = phpComponent.RunFunc("json_encode", testData)
	if err != nil {
		log.Printf("❌ json_encode 测试失败: %v", err)
	} else {
		fmt.Printf("✅ json_encode 测试成功: %s\n", result)
	}

	// 测试5: 日期函数
	fmt.Println("\n5. 测试日期函数:")

	result, err = phpComponent.RunFunc("date", "Y-m-d")
	if err != nil {
		log.Printf("❌ date 测试失败: %v", err)
	} else {
		fmt.Printf("✅ 当前日期: %s\n", result)
	}

	// 测试6: 错误处理
	fmt.Println("\n6. 测试错误处理:")

	result, err = phpComponent.RunFunc("non_existent_function")
	if err != nil {
		fmt.Printf("✅ 错误处理正常: %v\n", err)
	} else {
		fmt.Printf("❌ 错误处理异常: 应该返回错误但没有\n")
	}

	fmt.Println("\n=== 测试完成 ===")

	// 总结
	fmt.Println("\n测试总结:")
	fmt.Println("✅ 如果所有测试都显示 ✅，说明 PHP 组件工作正常")
	fmt.Println("❌ 如果有测试显示 ❌，请检查:")
	fmt.Println("   - PHP 是否正确安装")
	fmt.Println("   - PHP 命令是否在 PATH 中")
	fmt.Println("   - 生成的 PHP 文件是否有执行权限")
}
