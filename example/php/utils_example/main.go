package main

import (
	"fmt"
	"log"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/php"
)

// SerializeHelper PHP 序列化辅助工具
type SerializeHelper struct {
	phpComponent *php.ConfigStruct
}

// NewSerializeHelper 创建序列化辅助工具实例
func NewSerializeHelper() *SerializeHelper {
	return &SerializeHelper{
		phpComponent: php.New(jcbaseGo.Option{}),
	}
}

// Serialize 序列化数据为 PHP 格式
func (h *SerializeHelper) Serialize(data string) (string, error) {
	result, err := h.phpComponent.RunFunc("serialize", data)
	if err != nil {
		return "", fmt.Errorf("序列化失败: %w", err)
	}
	return result, nil
}

// Unserialize 反序列化 PHP 格式的数据
func (h *SerializeHelper) Unserialize(serializedData string) (string, error) {
	result, err := h.phpComponent.RunFunc("unserialize", serializedData)
	if err != nil {
		return "", fmt.Errorf("反序列化失败: %w", err)
	}
	return result, nil
}

// SerializeArray 序列化数组
func (h *SerializeHelper) SerializeArray(items []string) (string, error) {
	// 构建 JSON 数组字符串
	arrayStr := "["
	for i, item := range items {
		if i > 0 {
			arrayStr += ","
		}
		arrayStr += fmt.Sprintf(`"%s"`, item)
	}
	arrayStr += "]"

	return h.Serialize(arrayStr)
}

// SerializeMap 序列化映射
func (h *SerializeHelper) SerializeMap(data map[string]interface{}) (string, error) {
	// 构建 JSON 对象字符串
	mapStr := "{"
	first := true
	for key, value := range data {
		if !first {
			mapStr += ","
		}
		first = false

		switch v := value.(type) {
		case string:
			mapStr += fmt.Sprintf(`"%s":"%s"`, key, v)
		case int:
			mapStr += fmt.Sprintf(`"%s":%d`, key, v)
		case float64:
			mapStr += fmt.Sprintf(`"%s":%f`, key, v)
		case bool:
			mapStr += fmt.Sprintf(`"%s":%t`, key, v)
		default:
			mapStr += fmt.Sprintf(`"%s":"%v"`, key, v)
		}
	}
	mapStr += "}"

	return h.Serialize(mapStr)
}

// IsValidSerializedData 检查是否为有效的序列化数据
func (h *SerializeHelper) IsValidSerializedData(data string) bool {
	_, err := h.Unserialize(data)
	return err == nil
}

// ConvertJSONToSerialized 将 JSON 转换为 PHP 序列化格式
func (h *SerializeHelper) ConvertJSONToSerialized(jsonData string) (string, error) {
	// 先解码 JSON 确保格式正确
	_, err := h.phpComponent.RunFunc("json_decode", jsonData, "true")
	if err != nil {
		return "", fmt.Errorf("JSON 格式无效: %w", err)
	}

	// 然后序列化
	return h.Serialize(jsonData)
}

// ConvertSerializedToJSON 将 PHP 序列化格式转换为 JSON
func (h *SerializeHelper) ConvertSerializedToJSON(serializedData string) (string, error) {
	// 先反序列化
	unserialized, err := h.Unserialize(serializedData)
	if err != nil {
		return "", fmt.Errorf("反序列化失败: %w", err)
	}

	// 然后编码为 JSON
	result, err := h.phpComponent.RunFunc("json_encode", unserialized)
	if err != nil {
		return "", fmt.Errorf("JSON 编码失败: %w", err)
	}

	return result, nil
}

func main() {
	fmt.Println("=== PHP 序列化辅助工具示例 ===")

	// 创建序列化辅助工具实例
	helper := NewSerializeHelper()

	// 示例1: 基本序列化和反序列化
	fmt.Println("\n1. 基本序列化和反序列化:")

	// 序列化简单数据
	data := `{"name":"张三","age":25}`
	serialized, err := helper.Serialize(data)
	if err != nil {
		log.Printf("序列化失败: %v", err)
	} else {
		fmt.Printf("序列化结果: %s\n", serialized)

		// 反序列化
		unserialized, err := helper.Unserialize(serialized)
		if err != nil {
			log.Printf("反序列化失败: %v", err)
		} else {
			fmt.Printf("反序列化结果: %s\n", unserialized)
		}
	}

	// 示例2: 序列化数组
	fmt.Println("\n2. 序列化数组:")

	arrayData := []string{"apple", "banana", "orange", "grape"}
	serialized, err = helper.SerializeArray(arrayData)
	if err != nil {
		log.Printf("序列化数组失败: %v", err)
	} else {
		fmt.Printf("数组序列化结果: %s\n", serialized)

		// 反序列化
		unserialized, err := helper.Unserialize(serialized)
		if err != nil {
			log.Printf("反序列化失败: %v", err)
		} else {
			fmt.Printf("数组反序列化结果: %s\n", unserialized)
		}
	}

	// 示例3: 序列化映射
	fmt.Println("\n3. 序列化映射:")

	mapData := map[string]interface{}{
		"name":   "李四",
		"age":    30,
		"city":   "上海",
		"active": true,
		"score":  95.5,
	}
	serialized, err = helper.SerializeMap(mapData)
	if err != nil {
		log.Printf("序列化映射失败: %v", err)
	} else {
		fmt.Printf("映射序列化结果: %s\n", serialized)

		// 反序列化
		unserialized, err := helper.Unserialize(serialized)
		if err != nil {
			log.Printf("反序列化失败: %v", err)
		} else {
			fmt.Printf("映射反序列化结果: %s\n", unserialized)
		}
	}

	// 示例4: JSON 转换
	fmt.Println("\n4. JSON 转换:")

	jsonData := `{"user":{"name":"王五","profile":{"age":28,"hobbies":["音乐","电影","旅行"]}}}`

	// JSON 转序列化
	serialized, err = helper.ConvertJSONToSerialized(jsonData)
	if err != nil {
		log.Printf("JSON 转序列化失败: %v", err)
	} else {
		fmt.Printf("JSON 转序列化结果: %s\n", serialized)

		// 序列化转 JSON
		jsonResult, err := helper.ConvertSerializedToJSON(serialized)
		if err != nil {
			log.Printf("序列化转 JSON 失败: %v", err)
		} else {
			fmt.Printf("序列化转 JSON 结果: %s\n", jsonResult)
		}
	}

	// 示例5: 数据验证
	fmt.Println("\n5. 数据验证:")

	// 有效数据
	validData := "a:3:{i:0;s:5:\"apple\";i:1;s:6:\"banana\";i:2;s:6:\"orange\";}"
	fmt.Printf("有效数据验证: %t\n", helper.IsValidSerializedData(validData))

	// 无效数据
	invalidData := "a:1:{i:0;s:5:\"hello\""
	fmt.Printf("无效数据验证: %t\n", helper.IsValidSerializedData(invalidData))

	// 示例6: 复杂数据结构处理
	fmt.Println("\n6. 复杂数据结构处理:")

	complexData := `{
		"users": [
			{"id": 1, "name": "用户1", "active": true, "roles": ["admin", "user"]},
			{"id": 2, "name": "用户2", "active": false, "roles": ["user"]}
		],
		"meta": {
			"total": 2,
			"page": 1,
			"limit": 10
		}
	}`

	serialized, err = helper.ConvertJSONToSerialized(complexData)
	if err != nil {
		log.Printf("复杂数据序列化失败: %v", err)
	} else {
		fmt.Printf("复杂数据序列化结果: %s\n", serialized)

		// 反序列化
		unserialized, err := helper.Unserialize(serialized)
		if err != nil {
			log.Printf("复杂数据反序列化失败: %v", err)
		} else {
			fmt.Printf("复杂数据反序列化结果: %s\n", unserialized)
		}
	}

	// 示例7: 错误处理
	fmt.Println("\n7. 错误处理:")

	// 尝试反序列化无效数据
	invalidSerialized := "invalid_serialized_data"
	_, err = helper.Unserialize(invalidSerialized)
	if err != nil {
		fmt.Printf("✅ 错误处理正常: %v\n", err)
	} else {
		fmt.Printf("❌ 错误处理异常: 应该返回错误但没有\n")
	}

	// 尝试序列化无效 JSON
	invalidJSON := `{"name": "test", "age": 25,}`
	_, err = helper.ConvertJSONToSerialized(invalidJSON)
	if err != nil {
		fmt.Printf("✅ JSON 验证正常: %v\n", err)
	} else {
		fmt.Printf("❌ JSON 验证异常: 应该返回错误但没有\n")
	}

	// 示例8: 性能测试
	fmt.Println("\n8. 性能测试:")

	// 大数组测试
	largeArray := make([]string, 100)
	for i := 0; i < 100; i++ {
		largeArray[i] = fmt.Sprintf("item_%d", i)
	}

	serialized, err = helper.SerializeArray(largeArray)
	if err != nil {
		log.Printf("大数组序列化失败: %v", err)
	} else {
		fmt.Printf("大数组序列化成功，长度: %d\n", len(serialized))

		// 反序列化
		unserialized, err := helper.Unserialize(serialized)
		if err != nil {
			log.Printf("大数组反序列化失败: %v", err)
		} else {
			fmt.Printf("大数组反序列化成功，长度: %d\n", len(unserialized))
		}
	}

	fmt.Println("\n=== 序列化辅助工具示例完成 ===")
}
