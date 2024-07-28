package crud

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/orm/mysql"
	"log"
	"net/http"
	"reflect"
	"time"
)

type Trait struct {
	PkId           string   `default:"id"` // 数据表主键
	ModelTableName string   // 模型表名
	ModelFields    []string // 模型所有字段
	OperateTime    string   // 操作时间

	Model      any             // 模型指针
	MysqlMain  *mysql.Instance // 数据库实例
	Controller interface{}     // 控制器
	GinContext *gin.Context    // 请求上下文
}

// 初始化crud，仅当初始化完成才可以使用
func (t *Trait) checkInit(c *gin.Context) {
	_ = helper.CheckAndSetDefault(t)
	t.GinContext = c

	// 判断模型是否为空
	if t.Model == nil {
		log.Panic("模型不能为空")
	}

	modelValue := reflect.ValueOf(t.Model)
	modelType := reflect.TypeOf(t.Model)

	// 检查是否传入的是指针
	if modelType.Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
		modelType = modelType.Elem()
	}

	// 确保获取到具体模型的名称
	model := reflect.New(modelType).Interface()

	// 解析模型（获取模型表名以及表字段）
	if modelParseProvider, ok := model.(interface {
		ModelParse(modelType reflect.Type) (tableName string, fields []string)
	}); ok {
		t.ModelTableName, t.ModelFields = modelParseProvider.ModelParse(modelType)
	} else {
		log.Panic("模型未实现 ModelParse 方法")
	}

	// 设置操作时间
	t.OperateTime = time.Now().Format("2006-01-02 15:04:05")
}

// 调用自定义方法，如果方法不存在则调用默认方法
func (t *Trait) callCustomMethod(methodName string, args ...interface{}) (results []interface{}) {
	// 调用自定义方法
	method := reflect.ValueOf(t.Controller).MethodByName(methodName)
	if method.IsValid() {
		in := make([]reflect.Value, len(args))
		for i, arg := range args {
			if arg == nil {
				// 如果参数为nil，则创建一个零值
				in[i] = reflect.Zero(method.Type().In(i))
			} else {
				in[i] = reflect.ValueOf(arg)
			}
		}
		resultValues := method.Call(in)
		for ind := 0; ind < len(resultValues); ind++ {
			results = append(results, resultValues[ind].Interface())
		}

		return results
	}

	// 调用默认方法
	defaultMethod := reflect.ValueOf(t).MethodByName(methodName)
	if !defaultMethod.IsValid() {
		log.Panic("默认方法不存在：" + methodName)
	}
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		if arg == nil {
			// 如果参数为nil，则创建一个零值
			in[i] = reflect.Zero(defaultMethod.Type().In(i))
		} else {
			in[i] = reflect.ValueOf(arg)
		}

	}
	resultsValue := defaultMethod.Call(in)
	for ind := 0; ind < len(resultsValue); ind++ {
		results = append(results, resultsValue[ind].Interface())
	}
	return
}

// ----- 公共方法 ----- /

// ExtractPkId 方法从不同类型的请求中提取 PkId
func (t *Trait) ExtractPkId() (pkValue uint, err error) {
	gpcInterface, GPCExists := t.GinContext.Get("GPC")
	if !GPCExists {
		return 0, err
	}
	gpc := gpcInterface.(map[string]map[string]any)["all"]

	idStr, ok := gpc[t.PkId]
	if !ok {
		return 0, err
	}
	pkValue = helper.Convert{Value: idStr}.ToUint()

	return
}

// Result 整理结果输出
func (t *Trait) Result(code int, msg string, args ...any) {
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
			//errs := helper.JsonStruct(data).ToMap(&resultMapData).Errors()
			//jcbaseGo.PanicIfError(errs)
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
			resultData = data.([]any)
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
		additionalParams := args[1]
		for k, v := range additionalParams.(map[string]any) {
			result[k] = v
		}
	}

	// 设置数据统计字段
	if reflect.TypeOf(result["data"]).Kind() == reflect.Map {
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
	}

	t.GinContext.JSON(http.StatusOK, result)
}
