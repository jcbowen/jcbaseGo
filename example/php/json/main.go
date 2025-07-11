package main

import (
	"fmt"
	"log"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/php"
)

func main() {
	fmt.Println("=== PHP JSON 处理示例 ===")

	// 创建 PHP 组件实例
	phpComponent := php.New(jcbaseGo.Option{})

	// 示例1: JSON 编码
	fmt.Println("\n1. JSON 编码:")

	// 编码简单数组
	result, err := phpComponent.RunFunc("json_encode", `["apple","banana","orange"]`)
	if err != nil {
		log.Printf("调用 json_encode 失败: %v", err)
	} else {
		fmt.Printf("json_encode(['apple','banana','orange']) = %s\n", result)
	}

	// 编码关联数组
	result, err = phpComponent.RunFunc("json_encode", `{"name":"张三","age":25,"city":"北京"}`)
	if err != nil {
		log.Printf("调用 json_encode 失败: %v", err)
	} else {
		fmt.Printf("json_encode({'name':'张三','age':25,'city':'北京'}) = %s\n", result)
	}

	// 编码嵌套数组
	nestedArray := `{"user":{"name":"李四","profile":{"age":30,"city":"上海","hobbies":["读书","游泳","编程"]}}}`
	result, err = phpComponent.RunFunc("json_encode", nestedArray)
	if err != nil {
		log.Printf("调用 json_encode 失败: %v", err)
	} else {
		fmt.Printf("json_encode(嵌套数组) = %s\n", result)
	}

	// 示例2: JSON 解码
	fmt.Println("\n2. JSON 解码:")

	// 解码为关联数组
	jsonString := `{"name":"王五","age":28,"city":"广州"}`
	result, err = phpComponent.RunFunc("json_decode", jsonString, "true")
	if err != nil {
		log.Printf("调用 json_decode 失败: %v", err)
	} else {
		fmt.Printf("json_decode('%s', true) = %s\n", jsonString, result)
	}

	// 解码为对象
	result, err = phpComponent.RunFunc("json_decode", jsonString, "false")
	if err != nil {
		log.Printf("调用 json_decode 失败: %v", err)
	} else {
		fmt.Printf("json_decode('%s', false) = %s\n", jsonString, result)
	}

	// 解码复杂 JSON
	complexJson := `{"data":{"users":[{"id":1,"name":"用户1"},{"id":2,"name":"用户2"}],"total":2}}`
	result, err = phpComponent.RunFunc("json_decode", complexJson, "true")
	if err != nil {
		log.Printf("调用 json_decode 失败: %v", err)
	} else {
		fmt.Printf("json_decode(复杂JSON, true) = %s\n", result)
	}

	// 示例3: JSON 错误处理
	fmt.Println("\n3. JSON 错误处理:")

	// 检查 JSON 语法错误
	result, err = phpComponent.RunFunc("json_last_error")
	if err != nil {
		log.Printf("调用 json_last_error 失败: %v", err)
	} else {
		fmt.Printf("json_last_error() = %s\n", result)
	}

	// 获取 JSON 错误信息
	result, err = phpComponent.RunFunc("json_last_error_msg")
	if err != nil {
		log.Printf("调用 json_last_error_msg 失败: %v", err)
	} else {
		fmt.Printf("json_last_error_msg() = %s\n", result)
	}

	// 示例4: JSON 格式化
	fmt.Println("\n4. JSON 格式化:")

	// 美化 JSON 输出
	compactJson := `{"name":"赵六","age":32,"city":"深圳","hobbies":["音乐","电影","旅行"]}`
	result, err = phpComponent.RunFunc("json_encode", compactJson, "JSON_PRETTY_PRINT")
	if err != nil {
		log.Printf("调用 json_encode 失败: %v", err)
	} else {
		fmt.Printf("json_encode(美化输出) = %s\n", result)
	}

	// 示例5: JSON 特殊字符处理
	fmt.Println("\n5. JSON 特殊字符处理:")

	// 处理中文和特殊字符
	chineseJson := `{"message":"你好世界","symbols":"!@#$%^&*()","unicode":"🎉🎊🎈"}`
	result, err = phpComponent.RunFunc("json_encode", chineseJson, "JSON_UNESCAPED_UNICODE")
	if err != nil {
		log.Printf("调用 json_encode 失败: %v", err)
	} else {
		fmt.Printf("json_encode(中文和特殊字符) = %s\n", result)
	}

	// 示例6: JSON 数组操作
	fmt.Println("\n6. JSON 数组操作:")

	// 创建 JSON 数组
	arrayJson := `[1,2,3,4,5]`
	result, err = phpComponent.RunFunc("json_encode", arrayJson)
	if err != nil {
		log.Printf("调用 json_encode 失败: %v", err)
	} else {
		fmt.Printf("json_encode([1,2,3,4,5]) = %s\n", result)
	}

	// 解码数组并获取长度
	result, err = phpComponent.RunFunc("count", `[1,2,3,4,5]`)
	if err != nil {
		log.Printf("调用 count 失败: %v", err)
	} else {
		fmt.Printf("count([1,2,3,4,5]) = %s\n", result)
	}

	// 示例7: JSON 对象操作
	fmt.Println("\n7. JSON 对象操作:")

	// 检查是否为对象
	objectJson := `{"key":"value"}`
	result, err = phpComponent.RunFunc("is_object", objectJson)
	if err != nil {
		log.Printf("调用 is_object 失败: %v", err)
	} else {
		fmt.Printf("is_object('{\"key\":\"value\"}') = %s\n", result)
	}

	// 获取对象属性
	result, err = phpComponent.RunFunc("property_exists", objectJson, "key")
	if err != nil {
		log.Printf("调用 property_exists 失败: %v", err)
	} else {
		fmt.Printf("property_exists('{\"key\":\"value\"}', 'key') = %s\n", result)
	}

	// 示例8: JSON 数据验证
	fmt.Println("\n8. JSON 数据验证:")

	// 验证 JSON 格式
	validJson := `{"valid":true}`
	result, err = phpComponent.RunFunc("json_validate", validJson)
	if err != nil {
		log.Printf("调用 json_validate 失败: %v", err)
	} else {
		fmt.Printf("json_validate('{\"valid\":true}') = %s\n", result)
	}

	// 示例9: JSON 文件操作
	fmt.Println("\n9. JSON 文件操作:")

	// 从字符串创建 JSON 文件内容
	fileContent := `{"config":{"debug":true,"timeout":30,"database":{"host":"localhost","port":3306}}}`
	result, err = phpComponent.RunFunc("json_encode", fileContent, "JSON_PRETTY_PRINT")
	if err != nil {
		log.Printf("调用 json_encode 失败: %v", err)
	} else {
		fmt.Printf("配置文件 JSON:\n%s\n", result)
	}

	// 示例10: JSON 数据转换
	fmt.Println("\n10. JSON 数据转换:")

	// 数组转 JSON
	arrayData := `["red","green","blue"]`
	result, err = phpComponent.RunFunc("json_encode", arrayData)
	if err != nil {
		log.Printf("调用 json_encode 失败: %v", err)
	} else {
		fmt.Printf("数组转 JSON: %s\n", result)
	}

	// JSON 转数组
	result, err = phpComponent.RunFunc("json_decode", `["red","green","blue"]`, "true")
	if err != nil {
		log.Printf("调用 json_decode 失败: %v", err)
	} else {
		fmt.Printf("JSON 转数组: %s\n", result)
	}

	// 示例11: JSON 深度操作
	fmt.Println("\n11. JSON 深度操作:")

	// 深度复制 JSON 数据
	deepJson := `{"level1":{"level2":{"level3":{"value":"deep"}}}}`
	result, err = phpComponent.RunFunc("json_encode", deepJson, "JSON_PRETTY_PRINT")
	if err != nil {
		log.Printf("调用 json_encode 失败: %v", err)
	} else {
		fmt.Printf("深度 JSON 结构:\n%s\n", result)
	}

	// 示例12: JSON 性能测试
	fmt.Println("\n12. JSON 性能测试:")

	// 大数组 JSON 处理
	largeArray := `[1,2,3,4,5,6,7,8,9,10]`
	result, err = phpComponent.RunFunc("json_encode", largeArray)
	if err != nil {
		log.Printf("调用 json_encode 失败: %v", err)
	} else {
		fmt.Printf("大数组 JSON 编码: %s\n", result)
	}

	// 大对象 JSON 处理
	largeObject := `{"a":1,"b":2,"c":3,"d":4,"e":5,"f":6,"g":7,"h":8,"i":9,"j":10}`
	result, err = phpComponent.RunFunc("json_encode", largeObject, "JSON_PRETTY_PRINT")
	if err != nil {
		log.Printf("调用 json_encode 失败: %v", err)
	} else {
		fmt.Printf("大对象 JSON 编码:\n%s\n", result)
	}

	fmt.Println("\n=== JSON 处理示例完成 ===")
}
