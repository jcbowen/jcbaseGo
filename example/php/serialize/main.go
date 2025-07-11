package main

import (
	"fmt"
	"log"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/php"
)

func main() {
	fmt.Println("=== PHP 序列化和反序列化示例 ===")

	// 创建 PHP 组件实例
	phpComponent := php.New(jcbaseGo.Option{})

	// 示例1: 基本序列化和反序列化
	fmt.Println("\n1. 基本序列化和反序列化:")

	// 序列化简单数组
	arrayData := `["apple","banana","orange"]`
	result, err := phpComponent.RunFunc("serialize", arrayData)
	if err != nil {
		log.Printf("调用 serialize 失败: %v", err)
	} else {
		fmt.Printf("serialize(['apple','banana','orange']) = %s\n", result)

		// 反序列化
		unserialized, err := phpComponent.RunFunc("unserialize", result)
		if err != nil {
			log.Printf("调用 unserialize 失败: %v", err)
		} else {
			fmt.Printf("unserialize('%s') = %s\n", result, unserialized)
		}
	}

	// 示例2: 关联数组序列化
	fmt.Println("\n2. 关联数组序列化:")

	assocArray := `{"name":"张三","age":25,"city":"北京"}`
	result, err = phpComponent.RunFunc("serialize", assocArray)
	if err != nil {
		log.Printf("调用 serialize 失败: %v", err)
	} else {
		fmt.Printf("serialize({'name':'张三','age':25,'city':'北京'}) = %s\n", result)

		// 反序列化
		unserialized, err := phpComponent.RunFunc("unserialize", result)
		if err != nil {
			log.Printf("调用 unserialize 失败: %v", err)
		} else {
			fmt.Printf("unserialize('%s') = %s\n", result, unserialized)
		}
	}

	// 示例3: 嵌套数组序列化
	fmt.Println("\n3. 嵌套数组序列化:")

	nestedArray := `{"user":{"name":"李四","profile":{"age":30,"city":"上海","hobbies":["读书","游泳","编程"]}}}`
	result, err = phpComponent.RunFunc("serialize", nestedArray)
	if err != nil {
		log.Printf("调用 serialize 失败: %v", err)
	} else {
		fmt.Printf("serialize(嵌套数组) = %s\n", result)

		// 反序列化
		unserialized, err := phpComponent.RunFunc("unserialize", result)
		if err != nil {
			log.Printf("调用 unserialize 失败: %v", err)
		} else {
			fmt.Printf("unserialize('%s') = %s\n", result, unserialized)
		}
	}

	// 示例4: 数字和布尔值序列化
	fmt.Println("\n4. 数字和布尔值序列化:")

	// 数字
	result, err = phpComponent.RunFunc("serialize", "123.45")
	if err != nil {
		log.Printf("调用 serialize 失败: %v", err)
	} else {
		fmt.Printf("serialize(123.45) = %s\n", result)

		unserialized, err := phpComponent.RunFunc("unserialize", result)
		if err != nil {
			log.Printf("调用 unserialize 失败: %v", err)
		} else {
			fmt.Printf("unserialize('%s') = %s\n", result, unserialized)
		}
	}

	// 布尔值
	result, err = phpComponent.RunFunc("serialize", "true")
	if err != nil {
		log.Printf("调用 serialize 失败: %v", err)
	} else {
		fmt.Printf("serialize(true) = %s\n", result)

		unserialized, err := phpComponent.RunFunc("unserialize", result)
		if err != nil {
			log.Printf("调用 unserialize 失败: %v", err)
		} else {
			fmt.Printf("unserialize('%s') = %s\n", result, unserialized)
		}
	}

	// 示例5: 字符串序列化
	fmt.Println("\n5. 字符串序列化:")

	testString := "Hello World 你好世界"
	result, err = phpComponent.RunFunc("serialize", testString)
	if err != nil {
		log.Printf("调用 serialize 失败: %v", err)
	} else {
		fmt.Printf("serialize('Hello World 你好世界') = %s\n", result)

		unserialized, err := phpComponent.RunFunc("unserialize", result)
		if err != nil {
			log.Printf("调用 unserialize 失败: %v", err)
		} else {
			fmt.Printf("unserialize('%s') = %s\n", result, unserialized)
		}
	}

	// 示例6: 空值序列化
	fmt.Println("\n6. 空值序列化:")

	result, err = phpComponent.RunFunc("serialize", "null")
	if err != nil {
		log.Printf("调用 serialize 失败: %v", err)
	} else {
		fmt.Printf("serialize(null) = %s\n", result)

		unserialized, err := phpComponent.RunFunc("unserialize", result)
		if err != nil {
			log.Printf("调用 unserialize 失败: %v", err)
		} else {
			fmt.Printf("unserialize('%s') = %s\n", result, unserialized)
		}
	}

	// 示例7: 复杂数据结构序列化
	fmt.Println("\n7. 复杂数据结构序列化:")

	complexData := `{"users":[{"id":1,"name":"用户1","active":true},{"id":2,"name":"用户2","active":false}],"meta":{"total":2,"page":1,"limit":10}}`
	result, err = phpComponent.RunFunc("serialize", complexData)
	if err != nil {
		log.Printf("调用 serialize 失败: %v", err)
	} else {
		fmt.Printf("serialize(复杂数据) = %s\n", result)

		unserialized, err := phpComponent.RunFunc("unserialize", result)
		if err != nil {
			log.Printf("调用 unserialize 失败: %v", err)
		} else {
			fmt.Printf("unserialize('%s') = %s\n", result, unserialized)
		}
	}

	// 示例8: 错误处理
	fmt.Println("\n8. 错误处理:")

	// 尝试反序列化无效数据
	invalidData := "a:1:{i:0;s:5:\"hello\""
	result, err = phpComponent.RunFunc("unserialize", invalidData)
	if err != nil {
		fmt.Printf("✅ 错误处理正常: %v\n", err)
	} else {
		fmt.Printf("❌ 错误处理异常: 应该返回错误但没有\n")
	}

	// 示例9: 序列化与JSON对比
	fmt.Println("\n9. 序列化与JSON对比:")

	testData := `{"name":"王五","age":28,"city":"广州","hobbies":["音乐","电影","旅行"]}`

	// PHP序列化
	serialized, err := phpComponent.RunFunc("serialize", testData)
	if err != nil {
		log.Printf("调用 serialize 失败: %v", err)
	} else {
		fmt.Printf("PHP序列化结果: %s\n", serialized)
	}

	// JSON序列化
	jsonResult, err := phpComponent.RunFunc("json_encode", testData)
	if err != nil {
		log.Printf("调用 json_encode 失败: %v", err)
	} else {
		fmt.Printf("JSON序列化结果: %s\n", jsonResult)
	}

	// 示例10: 性能测试
	fmt.Println("\n10. 性能测试:")

	largeArray := `[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20]`

	// 序列化大数组
	result, err = phpComponent.RunFunc("serialize", largeArray)
	if err != nil {
		log.Printf("调用 serialize 失败: %v", err)
	} else {
		fmt.Printf("大数组序列化: %s\n", result)

		// 反序列化大数组
		unserialized, err := phpComponent.RunFunc("unserialize", result)
		if err != nil {
			log.Printf("调用 unserialize 失败: %v", err)
		} else {
			fmt.Printf("大数组反序列化: %s\n", unserialized)
		}
	}

	fmt.Println("\n=== 序列化和反序列化示例完成 ===")
}
