package helper

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// ----- 结构体转json，Begin -----/

// JsonHelper
type JsonHelper struct {
	Struct   interface{}            // 结构体
	String   string                 // json字符串
	Map      map[string]interface{} // map
	Sort     bool                   // 是否需要排序
	NeedFile bool                   // 是否需要输出json文件
	FilePath string                 // 输出json文件路径
	errors   []error
}

// JsonStruct
// 初始化json结构体
func JsonStruct(jsonStruct interface{}) *JsonHelper {
	return &JsonHelper{Struct: jsonStruct}
}

// JsonFile
// 初始化文件中的json字符串
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

// JsonString 初始化json字符串
func JsonString(jsonString string) *JsonHelper {
	return &JsonHelper{String: jsonString}
}

// DoSort
// 对json结构根据key重新进行排序
func (opt *JsonHelper) DoSort() *JsonHelper {
	opt.Sort = true
	return opt
}

// File
// 将json输出到
func (opt *JsonHelper) File(filepath string) *JsonHelper {
	opt.NeedFile = true
	absFilePath, err := GetAbsPath(filepath)
	if err != nil {
		opt.errors = append(opt.errors, err)
	}
	opt.FilePath = absFilePath
	return opt
}

func (opt *JsonHelper) ToStruct(data interface{}) *JsonHelper {
	if opt.Struct == nil { // 如果没有传入结构体，则将字符串转为结构体
		if opt.String == "" { // 如果结构体和字符串都为空，则返回错误
			err := errors.New("没有传入结构体或json字符串")
			opt.errors = append(opt.errors, err)
			return opt
		}
		// 将字符串转为结构体
		err := json.Unmarshal([]byte(opt.String), &data)
		if err != nil {
			opt.errors = append(opt.errors, err)
			return opt
		}
		opt.Struct = data
	}

	if opt.NeedFile {
		_, err := DirExists(opt.FilePath, true, os.ModePerm)
		if err != nil {
			opt.errors = append(opt.errors, err)
			return opt
		}
		err = CreateFile(opt.FilePath, []byte(opt.String), os.ModePerm, true)
		if err != nil {
			opt.errors = append(opt.errors, err)
			return opt
		}
	}

	return opt
}

// Get 支持以.分割的key获取json中的值
// key: json的key, 例如: "a.b.c"
//func (opt *JsonHelper) Get(key string) (interface{}, []error) {
//
//}

// HasError 判断是否有错误
func (opt *JsonHelper) HasError() bool {
	return len(opt.errors) > 0
}

// Errors
// 获取错误信息
func (opt *JsonHelper) Errors() []error {
	return opt.errors
}

// ToString
// 将json字符串以字符串的形式返回
func (opt *JsonHelper) ToString(str *string) *JsonHelper {
	jsonStr := opt.String

	if jsonStr == "" { // 如果json字符串为空，则将结构体转为json字符串
		if opt.Struct == nil { // 如果结构体和字符串都为空，则返回错误
			err := errors.New("json字符串及json结构体都为空")
			opt.errors = append(opt.errors, err)
			return opt
		}
		// 将结构体转换为json字符串
		jsonByte, err := json.Marshal(opt.Struct)
		jsonStr = string(jsonByte)
		if err != nil {
			opt.errors = append(opt.errors, err)
			return opt
		}
	}

	if opt.Sort { // 如果需要排序
		jsonStr = JsonStrSort(jsonStr)
	}

	// 判断是否需要输出json文件
	if opt.NeedFile {
		_ = os.MkdirAll(filepath.Dir(opt.FilePath), os.ModePerm)
		cfgFile, err := os.Create(opt.FilePath)
		if err != nil {
			opt.errors = append(opt.errors, err)
			return opt
		}
		defer func(cfgFile *os.File) {
			err = cfgFile.Close()
			if err != nil {
				opt.errors = append(opt.errors, err)
			}
		}(cfgFile)

		// 编码写入配置文件;
		cfgEncoder := json.NewEncoder(cfgFile)
		cfgEncoder.SetIndent("", "\t")
		if err = cfgEncoder.Encode(opt.Struct); err != nil {
			opt.errors = append(opt.errors, err)
			return opt
		}
	}

	*str = jsonStr
	return opt
}

// ----- 结构体转json，End -----/

// ----- Json -----/

func JsonStr2Map(str string) map[string]interface{} {
	var tempMap map[string]interface{}
	err := json.Unmarshal([]byte(str), &tempMap)
	if err != nil {
		panic(err)
	}
	return tempMap
}

// JsonStrSort 根据map的key进行排序
func JsonStrSort(jsonStr string) string {
	jsonMap := JsonStr2Map(jsonStr)
	nData := SetMapStrInterface(jsonMap).DoSort().GetData()
	jsonByte, _ := json.Marshal(nData)
	return string(jsonByte)
}

// ------------------------ 以下是弃用了的函数，将在后续版本中被移除 ------------------------ /

// SetStruct
// Deprecated: 请使用 JsonStruct
func SetStruct(jsonStruct interface{}) *JsonHelper {
	return JsonStruct(jsonStruct)
}

// ToJson
// Deprecated: 请使用 ToString
func (opt *JsonHelper) ToJson() (string, error) {
	jsonStr := ""
	result := opt.ToString(&jsonStr)
	if result.HasError() {
		return "", result.Errors()[0]
	}
	return jsonStr, nil
}
