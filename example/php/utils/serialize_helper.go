package utils

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
// 参数：
//   - data: 要序列化的数据（JSON 字符串格式）
//
// 返回：
//   - string: 序列化后的字符串
//   - error: 序列化失败时的错误信息
func (h *SerializeHelper) Serialize(data string) (string, error) {
	result, err := h.phpComponent.RunFunc("serialize", data)
	if err != nil {
		return "", fmt.Errorf("序列化失败: %w", err)
	}
	return result, nil
}

// Unserialize 反序列化 PHP 格式的数据
// 参数：
//   - serializedData: 序列化的数据字符串
//
// 返回：
//   - string: 反序列化后的 JSON 字符串
//   - error: 反序列化失败时的错误信息
func (h *SerializeHelper) Unserialize(serializedData string) (string, error) {
	result, err := h.phpComponent.RunFunc("unserialize", serializedData)
	if err != nil {
		return "", fmt.Errorf("反序列化失败: %w", err)
	}
	return result, nil
}

// SerializeArray 序列化数组
// 参数：
//   - items: 数组项
//
// 返回：
//   - string: 序列化后的字符串
//   - error: 序列化失败时的错误信息
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
// 参数：
//   - data: 映射数据
//
// 返回：
//   - string: 序列化后的字符串
//   - error: 序列化失败时的错误信息
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
// 参数：
//   - data: 要检查的数据
//
// 返回：
//   - bool: 是否为有效的序列化数据
func (h *SerializeHelper) IsValidSerializedData(data string) bool {
	_, err := h.Unserialize(data)
	return err == nil
}

// GetSerializedDataInfo 获取序列化数据的信息
// 参数：
//   - data: 序列化数据
//
// 返回：
//   - string: 数据信息
//   - error: 获取信息失败时的错误信息
func (h *SerializeHelper) GetSerializedDataInfo(data string) (string, error) {
	// 尝试反序列化获取数据类型
	unserialized, err := h.Unserialize(data)
	if err != nil {
		return "", fmt.Errorf("无法解析序列化数据: %w", err)
	}

	// 获取数据类型信息
	result, err := h.phpComponent.RunFunc("gettype", unserialized)
	if err != nil {
		return "", fmt.Errorf("无法获取数据类型: %w", err)
	}

	return fmt.Sprintf("数据类型: %s, 内容: %s", result, unserialized), nil
}

// ConvertJSONToSerialized 将 JSON 转换为 PHP 序列化格式
// 参数：
//   - jsonData: JSON 字符串
//
// 返回：
//   - string: PHP 序列化字符串
//   - error: 转换失败时的错误信息
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
// 参数：
//   - serializedData: PHP 序列化字符串
//
// 返回：
//   - string: JSON 字符串
//   - error: 转换失败时的错误信息
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

// Example 使用示例
func Example() {
	helper := NewSerializeHelper()

	// 示例1: 序列化数组
	fmt.Println("=== 序列化数组示例 ===")
	arrayData := []string{"apple", "banana", "orange"}
	serialized, err := helper.SerializeArray(arrayData)
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

	// 示例2: 序列化映射
	fmt.Println("\n=== 序列化映射示例 ===")
	mapData := map[string]interface{}{
		"name":   "张三",
		"age":    25,
		"city":   "北京",
		"active": true,
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

	// 示例3: JSON 转换
	fmt.Println("\n=== JSON 转换示例 ===")
	jsonData := `{"name":"李四","age":30,"hobbies":["读书","游泳"]}`
	serialized, err = helper.ConvertJSONToSerialized(jsonData)
	if err != nil {
		log.Printf("JSON 转序列化失败: %v", err)
	} else {
		fmt.Printf("JSON 转序列化结果: %s\n", serialized)

		// 转回 JSON
		jsonResult, err := helper.ConvertSerializedToJSON(serialized)
		if err != nil {
			log.Printf("序列化转 JSON 失败: %v", err)
		} else {
			fmt.Printf("序列化转 JSON 结果: %s\n", jsonResult)
		}
	}

	// 示例4: 数据验证
	fmt.Println("\n=== 数据验证示例 ===")
	validData := "a:3:{i:0;s:5:\"apple\";i:1;s:6:\"banana\";i:2;s:6:\"orange\";}"
	invalidData := "a:1:{i:0;s:5:\"hello\""

	fmt.Printf("有效数据验证: %t\n", helper.IsValidSerializedData(validData))
	fmt.Printf("无效数据验证: %t\n", helper.IsValidSerializedData(invalidData))

	// 获取数据信息
	info, err := helper.GetSerializedDataInfo(validData)
	if err != nil {
		log.Printf("获取数据信息失败: %v", err)
	} else {
		fmt.Printf("数据信息: %s\n", info)
	}
}
