package crud

import (
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/orm"
	"github.com/jcbowen/jcbaseGo/component/trait/controller"
)

type Trait struct {
	// ----- 基础配置 ----- /
	PkId               string       `default:"id"` // 数据表主键
	Model              any          // 模型指针
	ModelTableAlias    string       // 模型表别名
	DBI                orm.Instance // 数据库实例
	ListResultStruct   interface{}  // 列表返回结构体
	DetailResultStruct interface{}  // 详情返回结构体
	Controller         interface{}  // 控制器

	// ----- 初始化时生成 ----- /
	ModelTableName      string   // 模型表名
	ModelFields         []string // 模型所有字段
	SoftDeleteField     string   // 软删除字段名（如: "deleted_at"、"is_deleted" 等）
	SoftDeleteCondition string   // 软删除判断条件（如: "IS NULL" 或 "= '0000-00-00 00:00:00'"，不包含字段名）
	OperateTime         string   // 操作时间
	TableAlias          string   // 表别名（仅用于拼接查询语句，配置别名请用ModelTableAlias）

	// ----- 非基础配置 ----- /
	BaseControllerTrait controller.Base
}

// InitCrud 初始化CRUD，仅当初始化完成才可以使用
// 参数说明：
//   - c *gin.Context: Gin框架的上下文对象，包含请求和响应信息
func (t *Trait) InitCrud(c *gin.Context) {
	_ = helper.CheckAndSetDefault(t)
	t.BaseControllerTrait.GinContext = c

	// 设置json响应头
	c.Set("Content-type", "application/json;charset=utf-8")

	// 如果控制器中有CheckInit方法，就调用
	method := reflect.ValueOf(t.Controller).MethodByName("CheckInit")
	if method.IsValid() {
		in := make([]reflect.Value, 1)
		in[0] = reflect.ValueOf(c)
		method.Call(in)
	}

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

	// 解析模型（获取模型表名、表字段、软删除字段名以及软删除条件）
	if modelParseProvider, ok := model.(interface {
		ModelParse(modelType reflect.Type) (tableName string, fields []string, softDeleteField string, softDeleteCondition string)
	}); ok {
		var SoftDeleteField, SoftDeleteCondition string
		t.ModelTableName, t.ModelFields, SoftDeleteField, SoftDeleteCondition = modelParseProvider.ModelParse(modelType)
		// 如果在初始化crud时配置过了，优先以配置的来（旧版只支持配置）
		if t.SoftDeleteField == "" {
			t.SoftDeleteField = SoftDeleteField
		}
		if t.SoftDeleteCondition == "" {
			t.SoftDeleteCondition = SoftDeleteCondition
		}
	} else {
		log.Panic("模型未实现 ModelParse 方法")
	}

	// 设置操作时间
	t.OperateTime = time.Now().Format("2006-01-02 15:04:05")

	// 存放表别名到上下文，方便查询时调用
	t.TableAlias = t.ModelTableName
	if t.ModelTableAlias != "" {
		t.TableAlias = t.ModelTableAlias
	}
	t.TableAlias += "."
}

// callCustomMethod 调用自定义方法，如果方法不存在则调用默认方法
// 参数说明：
//   - methodName string: 要调用的方法名称
//   - args ...interface{}: 传递给方法的可变参数列表
//
// 返回值：
//   - results []interface{}: 方法调用的返回值列表
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

// ExtractPkId 从不同类型的请求中提取主键ID
// 返回值：
//   - pkValue uint: 提取到的主键ID值
//   - err error: 提取过程中的错误信息
func (t *Trait) ExtractPkId() (pkValue uint, err error) {
	gpcInterface, GPCExists := t.BaseControllerTrait.GinContext.Get("GPC")
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

// Failure 返回失败的响应
// 这个方法用于简化失败响应的构建，接收可变参数
// 并根据参数数量确定响应的data、message、code字段的值。
// 参数：
//   - message string: 返回的消息内容
//   - data any: 返回的数据
//   - code int: 错误码
//
// 仅1个参数时，如果是字符串则作为message输出，否则作为data输出；
// 更多参数时，第一个参数为message，第二个参数为data，第三个参数为code；
func (t *Trait) Failure(args ...any) {
	t.BaseControllerTrait.Failure(args...)
}

// Success 返回成功的响应
// 这个方法用于简化成功响应的构建，接收可变参数，
// 并根据参数数量确定响应的data、additionalParams和message字段的值。
// 参数：
//   - message string: 返回的消息内容
//   - data any: 返回的数据
//   - additionalParams any: 附加数据
//
// 仅1个参数时，如果是字符串则作为message输出，否则作为data输出；
// 更多参数时，第一个参数为data，第二个参数为message，第三个参数为additionalParams；
func (t *Trait) Success(args ...any) {
	t.BaseControllerTrait.Success(args...)
}

// Result 整理结果输出
// 这个方法用于统一返回API响应结果。接收状态码、消息以及可选的额外参数，
// 并根据传入的数据类型对结果进行格式化和处理，最终返回JSON格式的响应。
// 参数：
//   - code int: 状态码，通常为HTTP状态码。
//   - msg string: 返回的消息内容。
//   - data any 选填，主要数据内容，可以是结构体、map、string或slice。
//   - additionalParams map[string]any 选填，附加参数
func (t *Trait) Result(code int, msg string, args ...any) {
	t.BaseControllerTrait.Result(code, msg, args...)
}

// BindMapToStruct 将map数据绑定到struct，并处理类型转换
// 参数说明：
//   - mapData map[string]any: 要绑定的map数据
//   - modelValue interface{}: 目标struct的指针，必须是非nil的指针类型
//
// 返回值：
//   - error: 绑定过程中的错误信息
func (t *Trait) BindMapToStruct(mapData map[string]any, modelValue interface{}) error {
	val := reflect.ValueOf(modelValue)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New("modelValue must be a non-nil pointer")
	}

	modelVal := val.Elem()
	modelType := modelVal.Type()

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		fieldName := field.Tag.Get("json")
		if fieldName == "" {
			fieldName = field.Name
		}

		if val, ok := mapData[fieldName]; ok {
			fieldVal := modelVal.Field(i)
			if !fieldVal.CanSet() {
				log.Printf("Field %s cannot be set\n", fieldName)
				continue
			}

			err := t.setValue(fieldVal, val)
			if err != nil {
				log.Printf("Error setting field %s: %v\n", fieldName, err)
				return err
			}
		}
	}

	return nil
}

// setValue 根据字段类型设置值
// 参数说明：
//   - fieldVal reflect.Value: 目标字段的反射值对象
//   - val interface{}: 要设置的值
//
// 返回值：
//   - error: 设置过程中的错误信息
func (t *Trait) setValue(fieldVal reflect.Value, val interface{}) error {
	// 如果传入值为nil，直接设置零值并返回，避免反射panic
	if val == nil {
		fieldVal.Set(reflect.Zero(fieldVal.Type()))
		return nil
	}
	// 统一处理指针类型：如果目标是指针，则创建元素并对元素赋值
	if fieldVal.Kind() == reflect.Ptr {
		// 如果还未分配，先创建一个新的元素
		if fieldVal.IsNil() {
			newElem := reflect.New(fieldVal.Type().Elem())
			fieldVal.Set(newElem)
		}
		// 对指针指向的元素赋值
		return t.setValue(fieldVal.Elem(), val)
	}

	switch fieldVal.Kind() {
	case reflect.String:
		strVal := helper.Convert{Value: val}.ToString()
		fieldVal.SetString(strVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal := helper.Convert{Value: val}.ToInt64()
		fieldVal.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal := helper.Convert{Value: val}.ToUint64()
		fieldVal.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal := helper.Convert{Value: val}.ToFloat64()
		fieldVal.SetFloat(floatVal)
	case reflect.Bool:
		// 处理布尔类型，兼容多种输入：bool、字符串、数字等
		boolVal := false
		switch v := val.(type) {
		case bool:
			boolVal = v
		case string:
			s := strings.TrimSpace(strings.ToLower(v))
			switch s {
			case "1", "true", "yes", "y", "on":
				boolVal = true
			case "0", "false", "no", "n", "off", "":
				boolVal = false
			default:
				// 尝试按数字解析
				boolVal = helper.Convert{Value: v}.ToInt64() != 0
			}
		case int, int8, int16, int32, int64:
			boolVal = helper.Convert{Value: v}.ToInt64() != 0
		case uint, uint8, uint16, uint32, uint64:
			boolVal = helper.Convert{Value: v}.ToUint64() != 0
		case float32, float64:
			boolVal = helper.Convert{Value: v}.ToFloat64() != 0
		default:
			// 其他类型统一按字符串或数字再判断
			boolVal = helper.Convert{Value: v}.ToInt64() != 0
		}
		fieldVal.SetBool(boolVal)
	case reflect.Map:
		// 检查输入val是否为map类型
		valReflected := reflect.ValueOf(val)
		if !valReflected.IsValid() {
			fieldVal.Set(reflect.Zero(fieldVal.Type()))
			return nil
		}
		if valReflected.Kind() == reflect.Map {
			valType := valReflected.Type()
			fieldType := fieldVal.Type()

			// map[interface{}]interface{} 类型需要转换为 map[string]interface{}
			if valType.Key().Kind() == reflect.Interface && valType.Elem().Kind() == reflect.Interface {
				// 将 map[interface{}]interface{} 转换为 map[string]interface{}
				convertedMap := make(map[string]interface{})
				for _, key := range valReflected.MapKeys() {
					strKey := helper.Convert{Value: key.Interface()}.ToString() // 转换 key 为 string
					convertedMap[strKey] = valReflected.MapIndex(key).Interface()
				}
				val = convertedMap
				valReflected = reflect.ValueOf(val)
				valType = valReflected.Type()
			}

			// 检查类型是否匹配
			if valType.Key().AssignableTo(fieldType.Key()) && valType.Elem().AssignableTo(fieldType.Elem()) {
				// 类型匹配，进行赋值
				fieldVal.Set(valReflected)
			} else {
				log.Printf("Map field type mismatch: expected %s but got %s\n", fieldType, valType)
				// 类型不匹配，设置为nil
				fieldVal.Set(reflect.Zero(fieldType))
			}
		} else {
			if valReflected.IsValid() {
				log.Printf("Expected map type for field, got %s\n", valReflected.Kind())
			}
			// 不是map类型，设置为零值
			fieldVal.Set(reflect.Zero(fieldVal.Type()))
		}
	case reflect.Slice:
		// 处理 slice 类型，增强兼容：支持 []map -> []struct，JSON 字符串 -> slice
		fieldType := fieldVal.Type()
		valReflected := reflect.ValueOf(val)
		if !valReflected.IsValid() {
			fieldVal.Set(reflect.Zero(fieldVal.Type()))
			return nil
		}

		// 如果传入的是字符串，尝试按 JSON 反序列化
		if valReflected.Kind() == reflect.String {
			newSlice := reflect.MakeSlice(fieldType, 0, 0)
			// 使用一个同类型的临时切片来接收
			tmpPtr := reflect.New(fieldType)
			tmpPtr.Elem().Set(newSlice)
			if err := json.Unmarshal([]byte(valReflected.String()), tmpPtr.Interface()); err == nil {
				fieldVal.Set(tmpPtr.Elem())
				return nil
			}
		}

		if valReflected.Kind() == reflect.Slice {
			valType := valReflected.Type()
			// 元素类型完全可赋值，直接设置
			if valType.Elem().AssignableTo(fieldType.Elem()) {
				fieldVal.Set(valReflected)
				return nil
			}

			// 如果目标元素是 struct 或 *struct，尝试把每个元素的 map 映射为 struct
			destElemType := fieldType.Elem()
			isPtrElem := false
			if destElemType.Kind() == reflect.Ptr && destElemType.Elem().Kind() == reflect.Struct {
				isPtrElem = true
				destElemType = destElemType.Elem()
			}
			if destElemType.Kind() == reflect.Struct {
				converted := reflect.MakeSlice(fieldType, 0, valReflected.Len())
				for i := 0; i < valReflected.Len(); i++ {
					ele := valReflected.Index(i).Interface()
					// 接受 map[string]any 或 map[any]any
					m, ok := ele.(map[string]interface{})
					if !ok {
						if mv, ok2 := ele.(map[interface{}]interface{}); ok2 {
							m = make(map[string]interface{}, len(mv))
							for k, v := range mv {
								m[helper.Convert{Value: k}.ToString()] = v
							}
						} else {
							// 不支持的元素，跳过或置零
							continue
						}
					}

					// 创建目标元素
					var elemVal reflect.Value
					if isPtrElem {
						elemVal = reflect.New(destElemType)
						if err := t.assignMapToStruct(elemVal.Elem(), m); err != nil {
							return err
						}
						converted = reflect.Append(converted, elemVal)
					} else {
						elemVal = reflect.New(destElemType).Elem()
						if err := t.assignMapToStruct(elemVal, m); err != nil {
							return err
						}
						converted = reflect.Append(converted, elemVal)
					}
				}
				fieldVal.Set(converted)
				return nil
			}

			// 兼容 interface{} -> string 的转换场景
			if valType.Elem().Kind() == reflect.Interface && fieldType.Elem().Kind() == reflect.String {
				convertedSlice := reflect.MakeSlice(fieldType, valReflected.Len(), valReflected.Len())
				for i := 0; i < valReflected.Len(); i++ {
					strVal := helper.Convert{Value: valReflected.Index(i).Interface()}.ToString()
					convertedSlice.Index(i).SetString(strVal)
				}
				fieldVal.Set(convertedSlice)
				return nil
			}

			log.Printf("Slice element type mismatch: expected %s elements but got %s elements\n", fieldType.Elem(), valType.Elem())
			fieldVal.Set(reflect.Zero(fieldType))
			return nil
		}

		if valReflected.IsValid() {
			log.Printf("Expected slice type for field, got %s\n", valReflected.Kind())
		}
		fieldVal.Set(reflect.Zero(fieldVal.Type()))
		return nil
	case reflect.Struct:
		valReflected := reflect.ValueOf(val)
		// 1) 如果传入 map，则按字段名/JSON标签进行映射
		if valReflected.Kind() == reflect.Map {
			// 先把 map[any]any 统一转换为 map[string]any
			var m map[string]interface{}
			if mv, ok := val.(map[string]interface{}); ok {
				m = mv
			} else if mv2, ok2 := val.(map[interface{}]interface{}); ok2 {
				m = make(map[string]interface{}, len(mv2))
				for k, v := range mv2 {
					m[helper.Convert{Value: k}.ToString()] = v
				}
			} else {
				// 其他 map 类型，尝试通过反射读取
				m = make(map[string]interface{})
				for _, key := range valReflected.MapKeys() {
					m[helper.Convert{Value: key.Interface()}.ToString()] = valReflected.MapIndex(key).Interface()
				}
			}
			if err := t.assignMapToStruct(fieldVal, m); err != nil {
				return err
			}
			return nil
		}

		// 2) 如果传入字符串，尝试 JSON 反序列化
		if valReflected.Kind() == reflect.String {
			tmpPtr := reflect.New(fieldVal.Type())
			if err := json.Unmarshal([]byte(valReflected.String()), tmpPtr.Interface()); err == nil {
				fieldVal.Set(tmpPtr.Elem())
				return nil
			}
		}

		// 3) 仍是 struct，则按字段名逐个拷贝
		if valReflected.Kind() != reflect.Struct {
			log.Printf("Expected struct type for assignment but got %s\n", valReflected.Type())
			fieldVal.Set(reflect.Zero(fieldVal.Type()))
			return nil
		}

		for i := 0; i < fieldVal.NumField(); i++ {
			field := fieldVal.Type().Field(i)
			subFieldVal := fieldVal.Field(i)

			if !subFieldVal.CanSet() {
				log.Printf("Field %s cannot be set. \n", field.Name)
				continue
			}
			subValField := valReflected.FieldByName(field.Name)
			if !subValField.IsValid() {
				continue
			}

			// 递归地为结构体字段赋值
			if err := t.setValue(subFieldVal, subValField.Interface()); err != nil {
				return err
			}
		}
	default:
		return errors.New("unsupported field type：" + fieldVal.Kind().String())
	}

	return nil
}

// assignMapToStruct 将 map[string]any 映射到结构体字段
// 支持按照字段名和 `json` 标签匹配，并递归处理嵌套结构体、指针、切片
func (t *Trait) assignMapToStruct(structVal reflect.Value, data map[string]interface{}) error {
	structType := structVal.Type()
	for i := 0; i < structVal.NumField(); i++ {
		field := structType.Field(i)
		if !structVal.Field(i).CanSet() {
			continue
		}

		// 优先匹配 json 标签
		jsonTag := field.Tag.Get("json")
		candidateKeys := []string{}
		if jsonTag != "" {
			candidateKeys = append(candidateKeys, strings.Split(jsonTag, ",")[0])
		}
		candidateKeys = append(candidateKeys, field.Name)

		var value interface{}
		found := false
		for _, k := range candidateKeys {
			if k == "-" || k == "" {
				continue
			}
			if v, ok := data[k]; ok {
				value = v
				found = true
				break
			}
			// 兼容下划线命名
			lower := strings.ToLower(k)
			if v, ok := data[lower]; ok {
				value = v
				found = true
				break
			}
		}
		if !found {
			continue
		}

		if err := t.setValue(structVal.Field(i), value); err != nil {
			return err
		}
	}
	return nil
}

// GetSafeMapGPC 安全获取map类型的GPC数据
// 参数说明：
//   - key ...string: 可选的键名，用于指定获取特定的GPC数据类型（如"get", "post", "all"等）
//
// 返回值：
//   - mapData map[string]any: 获取到的GPC数据映射
func (t *Trait) GetSafeMapGPC(key ...string) (mapData map[string]any) {
	return t.BaseControllerTrait.GetSafeMapGPC(key...)
}

// getFieldType 获取模型字段的类型
// 参数说明：
//   - fieldName string: 字段名称
//
// 返回值：
//   - string: 字段类型字符串，如 "string"、"time.Time" 等
func (t *Trait) getFieldType(fieldName string) string {
	if t.Model == nil {
		return "string" // 默认返回字符串类型
	}

	modelType := reflect.TypeOf(t.Model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 遍历模型字段查找指定字段
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// 检查字段名是否匹配（忽略大小写）
		if strings.EqualFold(field.Name, fieldName) {
			return field.Type.String()
		}

		// 检查 gorm 标签中的列名
		gormTag := field.Tag.Get("gorm")
		if gormTag != "" {
			// 解析 gorm 标签获取列名
			tags := strings.Split(gormTag, ";")
			for _, tag := range tags {
				if strings.HasPrefix(tag, "column:") {
					columnName := strings.TrimPrefix(tag, "column:")
					if columnName == fieldName {
						return field.Type.String()
					}
				}
			}
		}
	}

	return "string" // 默认返回字符串类型
}
