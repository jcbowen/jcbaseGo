package helper

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
)

type JsonHelper struct {
	Struct   interface{}            // 结构体
	String   string                 // JSON字符串
	Map      map[string]interface{} // Map
	initType jsonDataType           // 初始化时的数据类型
	needFile bool                   // 是否需要输出JSON文件
	filePath string                 // 输出JSON文件路径
	errors   []error                // 错误信息列表
}

type jsonDataType int

const (
	unknownType jsonDataType = iota
	structType
	stringType
	mapType
)

// ----- 实例化，Begin ----- /

// JsonStruct 根据传入的结构体实例化JsonHelper
func JsonStruct(jsonStruct interface{}) *JsonHelper {
	return &JsonHelper{Struct: jsonStruct, initType: structType}
}

// JsonFile 根据传入的文件路径读取JSON文件并实例化JsonHelper
func JsonFile(path string) *JsonHelper {
	jh := &JsonHelper{initType: stringType}
	absPath, _ := NewFile(&File{Path: path}).GetAbsPath()

	if !NewFile(&File{Path: path}).Exists() {
		jh.errors = append(jh.errors, errors.New("JSON文件不存在\n文件路径: "+absPath))
		return jh
	}

	file, err := os.ReadFile(absPath)
	if err != nil {
		jh.errors = append(jh.errors, err)
		return jh
	}

	jh.String = string(file)
	return jh
}

// JsonString 根据传入的JSON字符串实例化JsonHelper
func JsonString(jsonString string) *JsonHelper {
	return &JsonHelper{String: jsonString, initType: stringType}
}

// JsonMap 根据传入的Map实例化JsonHelper
func JsonMap(jsonMap map[string]interface{}) *JsonHelper {
	return &JsonHelper{Map: jsonMap, initType: mapType}
}

// JsonAny 根据传入的未知类型实例化JsonHelper
func JsonAny(input interface{}) *JsonHelper {
	jh := &JsonHelper{}
	switch v := input.(type) {
	case string:
		jh.String = v
		jh.initType = stringType
	case map[string]interface{}:
		jh.Map = v
		jh.initType = mapType
	case []byte:
		jh.String = string(v)
		jh.initType = stringType
	default:
		if reflect.ValueOf(input).Kind() == reflect.Ptr {
			jh.Struct = input
			jh.initType = structType
		} else {
			jh.errors = append(jh.errors, errors.New("不支持的输入类型"))
			jh.initType = unknownType
		}
	}
	return jh
}

// ----- 实例化，End ----- /

// ----- 参数配置，Begin ----- /

// MakeFile 设置是否生成JSON文件及其路径
func (jh *JsonHelper) MakeFile(filepath string) *JsonHelper {
	if filepath == "" {
		filepath = "./file.json"
	}
	jh.needFile = true
	absFilePath, err := NewFile(&File{Path: filepath}).GetAbsPath()
	if err != nil {
		jh.errors = append(jh.errors, err)
	}
	jh.filePath = absFilePath
	return jh
}

// ----- 参数配置，End ----- /

// ----- 转换，Begin ----- /

// 将任意数据转换为JSON字符串
func (jh *JsonHelper) toJSONString() {
	if jh.String == "" {
		var data interface{}
		switch jh.initType {
		case structType:
			data = jh.Struct
		case mapType:
			data = jh.Map
		default:
			jh.errors = append(jh.errors, errors.New("没有提供可供转换的数据"))
			return
		}

		jsonBytes, err := json.Marshal(data)
		if err != nil {
			jh.errors = append(jh.errors, err)
			return
		}
		jh.String = string(jsonBytes)
	}
}

// ToStruct 将JSON数据转换为结构体
func (jh *JsonHelper) ToStruct(newStruct interface{}) *JsonHelper {
	if jh.initType == mapType {
		jh.toJSONString()
	}

	if jh.String == "" {
		jh.errors = append(jh.errors, errors.New("没有提供可供转换的数据"))
		return jh
	}

	if err := json.Unmarshal([]byte(jh.String), newStruct); err != nil {
		jh.errors = append(jh.errors, err)
		return jh
	}
	jh.Struct = newStruct
	jh.initType = structType

	if jh.needFile {
		jh.ToFile()
	}

	return jh
}

// ToString 将结构体或Map转换为JSON字符串
func (jh *JsonHelper) ToString(newStr *string) *JsonHelper {
	jh.toJSONString()

	if jh.String == "" {
		return jh
	}

	if jh.needFile {
		jh.ToFile()
	}

	*newStr = jh.String
	return jh
}

// ToMap 将JSON字符串或结构体转换为Map
func (jh *JsonHelper) ToMap(newMap *map[string]interface{}) *JsonHelper {
	if jh.initType == structType {
		jh.toJSONString()
	}

	if jh.String == "" {
		jh.errors = append(jh.errors, errors.New("没有提供可供转换的数据"))
		return jh
	}

	if err := json.Unmarshal([]byte(jh.String), newMap); err != nil {
		jh.errors = append(jh.errors, err)
		return jh
	}
	jh.Map = *newMap
	jh.initType = mapType

	if jh.needFile {
		jh.ToFile()
	}

	return jh
}

// ToFile 将JSON字符串输出为文件
func (jh *JsonHelper) ToFile() *JsonHelper {
	jh.toJSONString()

	if jh.String == "" {
		jh.errors = append(jh.errors, errors.New("转换出字符串类型JSON时出错"))
		return jh
	}

	if err := NewFile(&File{Path: jh.filePath, Perm: os.ModePerm}).CreateFile([]byte(jh.String), true); err != nil {
		jh.errors = append(jh.errors, err)
	}

	return jh
}

// ----- 转换，End ----- /

// HasError 判断是否有错误
func (jh *JsonHelper) HasError() bool {
	return len(jh.errors) > 0
}

// Errors 获取错误信息列表
func (jh *JsonHelper) Errors() []error {
	return jh.errors
}
