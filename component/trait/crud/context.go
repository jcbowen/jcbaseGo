package crud

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/security"
	"github.com/jcbowen/jcbaseGo/errcode"
)

type Context struct {
	ActionName string       // Action名称(如: "create", "update", "delete", "detail", "list", "all", "set-value")
	Debug      bool         // 调试模式
	GinContext *gin.Context // 请求上下文
}

// NewContextOpt 目前由于配置项没有分离，所以直接用相等，后续有可能会分离
type NewContextOpt = Context

// ----- 公共方法 ----- /

// NewContext 创建一个新的控制器上下文对象
// 这个方法用于初始化一个新的上下文对象，包含Gin上下文和调试模式。
// 参数说明：
//   - opt *NewContextOpt: 上下文选项，包含Gin上下文和调试模式
//
// 返回值：
//   - *Context: 新创建的上下文对象
func NewContext(opt *NewContextOpt) *Context {
	ctx := &Context{
		Debug:      opt.Debug,
		GinContext: opt.GinContext,
	}
	return ctx
}

// Success 返回成功的响应
// 这个方法用于简化成功响应的构建，接收可变参数，
// 并根据参数数量确定响应的data、additionalParams和message字段的值。
// 参数：
//   - ginContext *gin.Context: Gin框架的上下文对象
//   - message string: 返回的消息内容
//   - data any: 返回的数据
//   - additionalParams any: 附加数据
//
// 仅1个参数时，如果是字符串则作为message输出，否则作为data输出；
// 更多参数时，第一个参数为data，第二个参数为message，第三个参数为additionalParams；
func (ctx *Context) Success(args ...any) {
	var (
		message          = "success"
		data             any
		additionalParams map[string]any
	)
	switch len(args) {
	case 1:
		// 如果是字符串，则作为message输出
		if msg, ok := args[0].(string); ok {
			message = msg
		} else {
			data = args[0]
		}
	case 2:
		data = args[0]
		message = args[1].(string)
	case 3:
		data = args[0]
		additionalParams = args[2].(map[string]any)
		var ok bool
		message, ok = args[1].(string)
		if !ok {
			message = "ok"
		}
	}
	ctx.Result(errcode.Success, message, data, additionalParams)
}

// Failure 返回失败的响应
// 这个方法用于简化失败响应的构建，接收可变参数
// 并根据参数数量确定响应的data、message、code字段的值。
// 参数：
//   - ginContext *gin.Context: Gin框架的上下文对象
//   - message string: 返回的消息内容
//   - data any: 返回的数据
//   - code int: 错误码
//
// 仅1个参数时，如果是字符串则作为message输出，否则作为data输出；
// 更多参数时，第一个参数为message，第二个参数为data，第三个参数为code；
func (ctx *Context) Failure(args ...any) {
	var (
		code             = errcode.BadRequest
		message          = "failure"
		data             any
		additionalParams map[string]any
	)
	switch len(args) {
	case 1:
		// 如果是字符串，则作为message输出
		if msg, ok := args[0].(string); ok {
			message = msg
		} else {
			data = args[0]
		}
	case 2:
		message = args[0].(string)
		data = args[1]
	case 3:
		message = args[0].(string)
		data = args[1]
		code = args[2].(int)
	}
	ctx.Result(code, message, data, additionalParams)
}

// Result 整理结果输出
// 这个方法用于统一返回API响应结果。接收状态码、消息以及可选的额外参数，
// 并根据传入的数据类型对结果进行格式化和处理，最终返回JSON格式的响应。
// 参数：
//   - ginContext *gin.Context: Gin框架的上下文对象
//   - code int: 状态码，通常为HTTP状态码。
//   - msg string: 返回的消息内容。
//   - data any 选填，主要数据内容，可以是结构体、map、string或slice。
//   - additionalParams map[string]any 选填，附加参数
func (ctx *Context) Result(code int, msg string, args ...any) {
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
			resultData = ctx.convertToInterfaceSlice(data)
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
				if k == "data" {
					if resultData2, exist := result["data"].(map[string]any); exist {
						for k2, v2 := range v.(map[string]any) {
							resultData2[k2] = v2
						}
						result["data"] = resultData2
					}
				} else {
					result[k] = v
				}
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

	ctx.GinContext.JSON(http.StatusOK, result)
}

// convertToInterfaceSlice 将特定类型的切片转换为通用的 interface{} 切片
// 这个方法用于将任意类型的切片转换为 []interface{} 类型，
// 这在需要将不同类型的切片合并到一个统一处理的场景中非常有用。
// 参数：
//   - slice interface{}: 任意类型的切片，必须是切片类型。
//
// 返回值：
//   - []interface{}: 转换后的通用接口切片。
func (ctx *Context) convertToInterfaceSlice(slice interface{}) []interface{} {
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

// GetSafeMapGPC 安全获取map类型GPC数据
// 这个方法用于从Gin上下文的GPC（全局请求上下文）中安全获取map[string]any类型的数据。
// 它首先检查GPC是否存在，然后根据提供的键（默认为"all"）获取对应的数据。
// 如果键不存在或类型断言失败，会返回空map。
// 参数：
//   - key ...string: 可选参数，指定要获取的GPC键，枚举值为"query", "header", "cookie", "data"，"all"。如果为空，默认使用"all"。
//
// 返回值：
//   - mapData map[string]any: 安全获取到的map[string]any类型数据。
func (ctx *Context) GetSafeMapGPC(key ...string) (mapData map[string]any) {
	mapKey := "all"
	if len(key) > 0 {
		mapKey = key[0]
	}

	gpcInterface, GPCExists := ctx.GinContext.Get("GPC")
	if !GPCExists {
		log.Println("GPC data not found")
		return
	}

	formDataMap, ok := gpcInterface.(map[string]map[string]any)[mapKey]
	if !ok {
		log.Printf("Type assertion failed for formDataMap, expected map[string]map[string]any, got %T\n", gpcInterface)
		return
	}

	// 安全过滤
	sanitizedMapData := security.Input{Value: formDataMap}.Sanitize().(map[interface{}]interface{})

	// 转换为 map[string]any
	mapData = make(map[string]any)
	for k, value := range sanitizedMapData {
		strKey, ok := k.(string)
		if !ok {
			log.Printf("Type assertion failed for key in sanitizedMapData, got %T\n", k)
			continue
		}
		mapData[strKey] = value
	}

	if ctx.Debug {
		log.Printf("mapData: %v\n", mapData)
	}

	return
}

// GetSafeGPCValStr 安全获取GPC字符串值
// 这个方法用于从Gin上下文的GPC（全局请求上下文）中安全获取指定路径的字段值。
// 它首先调用GetSafeMapGPC方法获取所有GPC数据，然后使用helper.NewMap方法将其转换为可操作的Map类型，
// 最后使用ExtractString方法根据指定的字段路径提取对应的值。
// 参数：
//   - fieldPath string: 要获取的字段路径，支持嵌套路径（如"user.name"）。
//
// 返回值：
//   - value any: 安全获取到的字段值，如果路径不存在或类型断言失败，返回空字符串。
func (ctx *Context) GetSafeGPCValStr(fieldPath string) (value any) {
	mapData := ctx.GetSafeMapGPC("all")

	return helper.NewMap(mapData).ExtractString(fieldPath)
}

// GetSafeGPCValInt 安全获取GPC整数值
// 这个方法用于从Gin上下文的GPC（全局请求上下文）中安全获取指定路径的字段值。
// 它首先调用GetSafeMapGPC方法获取所有GPC数据，然后使用helper.NewMap方法将其转换为可操作的Map类型，
// 最后使用ExtractInt方法根据指定的字段路径提取对应的值。
// 参数：
//   - fieldPath string: 要获取的字段路径，支持嵌套路径（如"user.age"）。
//
// 返回值：
//   - value any: 安全获取到的字段值，如果路径不存在或类型断言失败，返回0。
func (ctx *Context) GetSafeGPCValInt(fieldPath string) (value any) {
	mapData := ctx.GetSafeMapGPC("all")

	return helper.NewMap(mapData).ExtractInt(fieldPath)
}

// GetSafeGPCValBool 安全获取GPC布尔值
// 这个方法用于从Gin上下文的GPC（全局请求上下文）中安全获取指定路径的字段值。
// 它首先调用GetSafeMapGPC方法获取所有GPC数据，然后使用helper.NewMap方法将其转换为可操作的Map类型，
// 最后使用ExtractBool方法根据指定的字段路径提取对应的值。
// 参数：
//   - fieldPath string: 要获取的字段路径，支持嵌套路径（如"user.isActive"）。
//
// 返回值：
//   - value any: 安全获取到的字段值，如果路径不存在或类型断言失败，返回false。
func (ctx *Context) GetSafeGPCValBool(fieldPath string) (value any) {
	mapData := ctx.GetSafeMapGPC("all")

	return helper.NewMap(mapData).ExtractBool(fieldPath)
}

// GetSafeGPCValTime 安全获取GPC时间值
// 这个方法用于从Gin上下文的GPC（全局请求上下文）中安全获取指定路径的字段值。
// 它首先调用GetSafeMapGPC方法获取所有GPC数据，然后使用helper.NewMap方法将其转换为可操作的Map类型，
// 最后使用ExtractTime方法根据指定的字段路径提取对应的值。
// 参数：
//   - fieldPath string: 要获取的字段路径，支持嵌套路径（如"user.birthday"）。
//
// 返回值：
//   - value any: 安全获取到的字段值，如果路径不存在或类型断言失败，返回time.Time零值。
func (ctx *Context) GetSafeGPCValTime(fieldPath string) (value any) {
	mapData := ctx.GetSafeMapGPC("all")

	return helper.NewMap(mapData).ExtractTime(fieldPath)
}

// GetSafeGPCVal 安全获取GPC值
// 这个方法用于从Gin上下文的GPC（全局请求上下文）中安全获取指定路径的字段值。
// 它首先调用GetSafeMapGPC方法获取所有GPC数据，然后使用helper.NewMap方法将其转换为可操作的Map类型，
// 最后使用Extract方法根据指定的字段路径提取对应的值。
// 参数：
//   - fieldPath string: 要获取的字段路径，支持嵌套路径（如"user.name"）。
//
// 返回值：
//   - value any: 安全获取到的字段值，如果路径不存在或类型断言失败，返回nil。
func (ctx *Context) GetSafeGPCVal(fieldPath string) (value any) {
	mapData := ctx.GetSafeMapGPC("all")

	return helper.NewMap(mapData).Extract(fieldPath)
}
