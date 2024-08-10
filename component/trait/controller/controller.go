package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/security"
	"github.com/jcbowen/jcbaseGo/errcode"
	"log"
	"net/http"
	"reflect"
)

type Base struct {
	GinContext *gin.Context // 请求上下文
	Debug      bool         // 调试模式
}

// ----- 公共方法 ----- /

// Success 返回成功的响应
// 这个方法用于简化成功响应的构建，接收可变参数，
// 并根据参数数量确定响应的data、additionalParams和message字段的值。
// 传递2个以内参数时，第一个参数为data，第二个参数为message。
// 传递3个参数时，第一个参数为data，第二个参数为additionalParams，第三个参数为message
func (c Base) Success(args ...any) {
	var (
		message          = "success"
		data             any
		additionalParams map[string]any
	)
	switch len(args) {
	case 1:
		data = args[0]
	case 2:
		data = args[0]
		message = args[1].(string)
	case 3:
		data = args[0]
		additionalParams = args[1].(map[string]any)
		message = args[1].(string)
	}
	c.Result(errcode.Success, message, data, additionalParams)
}

// Failure 返回失败的响应
// 这个方法用于简化失败响应的构建，接收可变参数
// 参数：
//   - msg string: 返回的消息内容
//   - data any: 返回的数据
//   - code int: 错误码
func (c Base) Failure(args ...any) {
	var (
		code             = errcode.BadRequest
		message          = "failure"
		data             any
		additionalParams map[string]any
	)
	switch len(args) {
	case 1:
		message = args[0].(string)
	case 2:
		message = args[0].(string)
		data = args[1]
	case 3:
		message = args[0].(string)
		data = args[1]
		code = args[2].(int)
	}
	c.Result(code, message, data, additionalParams)
}

// Result 整理结果输出
// 这个方法用于统一返回API响应结果。接收状态码、消息以及可选的额外参数，
// 并根据传入的数据类型对结果进行格式化和处理，最终返回JSON格式的响应。
// 参数：
//   - code int: 状态码，通常为HTTP状态码。
//   - msg string: 返回的消息内容。
//   - data any 选填，主要数据内容，可以是结构体、map、string或slice。
//   - additionalParams map[string]any 选填，附加参数
func (c Base) Result(code int, msg string, args ...any) {
	// 虽然定义的是any，但是约定只能为map/string/[]any
	var resultData any
	resultMapData := make(map[string]any)

	if len(args) > 0 && !helper.IsEmptyValue(args[0]) {
		data := args[0]
		val := reflect.ValueOf(data)

		// 如果是指针类型，获取指针指向的数据类型
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
			data = val.Interface()
		}

		if val.Kind() == reflect.Struct {
			// 将data转换为map
			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Panic(err)
			}
			// Convert JSON to map
			err = json.Unmarshal(jsonData, &resultMapData)
			if err != nil {
				log.Panic(err)
			}
			resultData = resultMapData
		} else if val.Kind() == reflect.Map {
			// 检查是否为gin.H类型
			if _, ok := data.(gin.H); ok {
				resultData = map[string]any(data.(gin.H))
			} else {
				resultData = data.(map[string]any)
			}
		} else if val.Kind() == reflect.String {
			resultData = data.(string)
		} else if val.Kind() == reflect.Array || val.Kind() == reflect.Slice {
			resultData = c.convertToInterfaceSlice(data)
		} else {
			log.Panic("不支持的数据类型：" + val.Kind().String())
		}
	} else {
		resultData = make(map[string]any)
	}

	// 获取resultData的类型
	resultDataType := reflect.TypeOf(resultData)
	if resultDataType == nil {
		resultData = nil
	}

	// 构建结果map
	result := map[string]any{
		"code":    code,
		"message": msg,
		"data":    resultData,
	}

	// 合并附加参数
	if len(args) > 1 && !helper.IsEmptyValue(args[1]) {
		additionalParams, ok := args[1].(map[string]any)
		if ok {
			for k, v := range additionalParams {
				result[k] = v
			}
		}
	}

	// log.Println(reflect.TypeOf(result["data"]).Kind())

	// 设置数据统计字段
	dataKind := reflect.TypeOf(result["data"]).Kind()
	if dataKind == reflect.Map {
		if list, exists := result["data"].(map[string]any)["list"].([]any); exists {
			total := len(list)
			if countParam, exists := result["data"].(map[string]any)["total"]; exists {
				result["data"].(map[string]any)["total"] = countParam
			} else {
				result["data"].(map[string]any)["total"] = total
			}
		} else {
			if countParam, exists := result["total"]; exists {
				result["total"] = countParam
			} else {
				if resultDataMap, ok := result["data"].(map[string]any); ok {
					result["total"] = len(resultDataMap)
				} else if resultDataSlice, ok := result["data"].([]any); ok {
					result["total"] = len(resultDataSlice)
				}
			}
		}
	} else if dataKind == reflect.Slice {
		if countParam, exists := result["total"]; exists {
			result["total"] = countParam
		} else {
			if resultDataSlice, ok := result["data"].([]any); ok {
				result["total"] = len(resultDataSlice)
			}
		}
	}

	c.GinContext.JSON(http.StatusOK, result)
}

// convertToInterfaceSlice 将特定类型的切片转换为通用的 interface{} 切片
func (c Base) convertToInterfaceSlice(slice interface{}) []interface{} {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		panic("convertToInterfaceSlice: not a slice")
	}

	interfaceSlice := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		interfaceSlice[i] = v.Index(i).Interface()
	}

	return interfaceSlice
}

// GetSafeMapGPC 安全获取map类型GPC
func (c Base) GetSafeMapGPC(key ...string) (mapData map[string]any) {
	mapKey := "all"
	if len(key) > 0 {
		mapKey = key[0]
	}

	gpcInterface, GPCExists := c.GinContext.Get("GPC")
	if !GPCExists {
		return
	}
	formDataMap := gpcInterface.(map[string]map[string]any)[mapKey]

	// 安全过滤
	sanitizedMapData := security.Input{Value: formDataMap}.Sanitize().(map[interface{}]interface{})
	// 格式转换
	mapData = make(map[string]any)
	for k, value := range sanitizedMapData {
		strKey := k.(string)
		mapData[strKey] = value
	}

	if c.Debug {
		log.Printf("mapData: %v\n", mapData)
	}

	return
}
