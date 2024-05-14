package helper

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

type JsonHelper struct {
	Struct   interface{}            // 结构体
	String   string                 // json字符串
	Map      map[string]interface{} // map
	needFile bool                   // 是否需要输出json文件
	filePath string                 // 输出json文件路径
	errors   []error
	// sort     bool                   // 是否需要排序（废弃：不再支持排序，有需求可通过map去进行排序）
}

// ----- 实例化，Begin ----- /

func JsonStruct(jsonStruct interface{}) *JsonHelper {
	return &JsonHelper{Struct: jsonStruct}
}

// JsonFile 根据传入的文件路径自动读取文件中的json内容
// path: json文件路径
func JsonFile(path string) *JsonHelper {
	jsonStruct := &JsonHelper{}
	// 获取绝对路径
	absPath, _ := GetAbsPath(path)
	// 判断文件是否存在
	if !FileExists(path) {
		jsonStruct.errors = append(jsonStruct.errors, errors.New("json文件不存在\n文件路径: "+absPath))
		return jsonStruct
	}

	file, fErr := os.ReadFile(absPath)
	if fErr != nil {
		jsonStruct.errors = append(jsonStruct.errors, fErr)
		return jsonStruct
	}
	fileDataString := string(file)

	jsonStruct.String = fileDataString

	return jsonStruct
}

func JsonString(jsonString string) *JsonHelper {
	return &JsonHelper{String: jsonString}
}

func JsonMap(jsonMap map[string]interface{}) *JsonHelper {
	return &JsonHelper{Map: jsonMap}
}

// ----- 实例化，End ----- /

// ----- 参数配置，Begin ----- /

// MakeFile
// 是否生成json文件
func (jh *JsonHelper) MakeFile(filepath string) *JsonHelper {
	if filepath == "" {
		filepath = "./file.json"
	}
	jh.needFile = true
	absFilePath, err := GetAbsPath(filepath)
	if err != nil {
		jh.errors = append(jh.errors, err)
	}
	jh.filePath = absFilePath
	return jh
}

// ----- 参数配置，End ----- /

// ----- 转换，Begin ----- /

func (jh *JsonHelper) ToStruct(newStruct interface{}) *JsonHelper {
	if len(jh.Map) == 0 && jh.String == "" {
		err := errors.New("没有提供可供转换的数据")
		jh.errors = append(jh.errors, err)
		return jh
	}

	// 如果字符串为空但是map不为空，需要将map转为字符串
	if jh.String == "" {
		jh = jh.ToString(&jh.String)
	}

	// 如果字符串还是为空，多半转换就出现问题了
	if jh.String == "" {
		err := errors.New("转换出错，请检查数据格式")
		jh.errors = append(jh.errors, err)
		return jh
	}

	// 将字符串转为结构体
	err := json.Unmarshal([]byte(jh.String), &newStruct)
	if err != nil {
		jh.errors = append(jh.errors, err)
		return jh
	}
	jh.Struct = newStruct

	// 判断是否需要输出json文件
	if jh.needFile {
		jh = jh.ToFile()
	}

	return jh
}

func (jh *JsonHelper) ToString(newStr *string) *JsonHelper {
	jsonStr := jh.String

	if len(jh.Map) == 0 && jh.Struct == nil { // 如果结构体和字符串都为空，则返回错误
		err := errors.New("没有提供可供转换的数据")
		jh.errors = append(jh.errors, err)
		return jh
	}

	// 判断是用结构体还是map去转
	var oData interface{}
	if jh.Struct != nil {
		oData = jh.Struct
	} else {
		oData = jh.Map
	}
	jsonByte, err := json.Marshal(oData)
	if err != nil {
		jh.errors = append(jh.errors, err)
		return jh
	}
	jsonStr = string(jsonByte)

	// 判断是否需要输出json文件
	if jh.needFile {
		jh = jh.ToFile()
	}

	// 不相同时才进行赋值
	if *newStr != jsonStr {
		*newStr = jsonStr
	}

	return jh
}

func (jh *JsonHelper) ToMap(newMap *map[string]interface{}) *JsonHelper {
	if jh.String == "" {
		jh = jh.ToString(&jh.String)
	}

	if jh.String == "" {
		err := errors.New("转换出错，请检查数据格式")
		jh.errors = append(jh.errors, err)
		return jh
	}

	err := json.Unmarshal([]byte(jh.String), newMap)
	if err != nil {
		jh.errors = append(jh.errors, err)
		return jh
	}
	jh.Map = *newMap

	return jh
}

func (jh *JsonHelper) ToFile() *JsonHelper {
	if jh.String == "" {
		// 避免在获取json字符串时重复执行本方法
		if jh.needFile {
			jh.needFile = false
		}

		jh = jh.ToString(&jh.String)
	}

	// 还是为空肯定意味着哪里出问题了
	if jh.String == "" {
		// 有错误的话，在上面的转换步骤就已经写入错误了，这里无需重新写入
		if len(jh.Errors()) == 0 {
			jh.errors = append(jh.errors, errors.New("转换出字符串类型json时出错"))
		}
		return jh
	}

	// 都走到这步了，肯定是需要输出为文件的，改变状态值以便他用
	if !jh.needFile {
		jh.needFile = true
	}

	// 检查目录是否存在，不存在则创建
	_, err := DirExists(jh.filePath, true, os.ModePerm)
	if err != nil {
		jh.errors = append(jh.errors, err)
		return jh
	}

	// 生成json文件
	err = CreateFile(jh.filePath, []byte(jh.String), os.ModePerm, true)
	if err != nil {
		jh.errors = append(jh.errors, err)
	}

	return jh
}

// ----- 转换，End ----- /

// Get 支持以.分割的key获取json中的值
// key: json的key, 例如: "a.b.c"
//func (jh *JsonHelper) Get(key string) (interface{}, []error) {
//
//}

// HasError 判断是否有错误
func (jh *JsonHelper) HasError() bool {
	return len(jh.errors) > 0
}

// Errors
// 获取错误信息
func (jh *JsonHelper) Errors() []error {
	return jh.errors
}

// ------------------------ 以下是弃用了的函数，将在后续版本中被移除 ------------------------ /

// SetStruct
// Deprecated: 请使用 JsonStruct
func SetStruct(jsonStruct interface{}) *JsonHelper {
	return JsonStruct(jsonStruct)
}

// DoSort 输出json字符串时是否根据key排序
// Deprecated: 不再支持自动排序，有需求时通过MapHelper去排序
func (jh *JsonHelper) DoSort() *JsonHelper {
	return jh
}

// ToJson
// Deprecated: 请使用 JsonHelper.ToString
func (jh *JsonHelper) ToJson() (string, error) {
	jsonStr := ""
	result := jh.ToString(&jsonStr)
	if result.HasError() {
		return "", result.Errors()[0]
	}
	return jsonStr, nil
}

// JsonStrSort 根据map的key进行排序
// Deprecated: 请使用
func JsonStrSort(jsonStr string) string {
	jsonMap := JsonStr2Map(jsonStr)
	nData := NewMap(jsonMap).DoSort().GetData()
	jsonByte, _ := json.Marshal(nData)
	return string(jsonByte)
}

// JsonStr2Map
// Deprecated: 请使用 JsonHelper.ToMap
func JsonStr2Map(str string) map[string]interface{} {
	var tempMap map[string]interface{}

	jsonHelper := JsonString(str).ToMap(&tempMap)
	if jsonHelper.HasError() {
		for _, err := range jsonHelper.Errors() {
			log.Panic(err)
		}
	}

	return tempMap
}
