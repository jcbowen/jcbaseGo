package crud

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/orm/mysql"
	"github.com/jcbowen/jcbaseGo/component/trait/controller"
	"log"
	"reflect"
	"time"
)

type Trait struct {
	// ----- 基础配置 ----- /
	PkId               string          `default:"id"` // 数据表主键
	Model              any             // 模型指针
	ModelTableAlias    string          // 模型表别名
	MysqlMain          *mysql.Instance // 数据库实例
	ListResultStruct   interface{}     // 列表返回结构体
	DetailResultStruct interface{}     // 详情返回结构体
	Controller         interface{}     // 控制器

	// ----- 初始化时生成 ----- /
	ModelTableName string   // 模型表名
	ModelFields    []string // 模型所有字段
	OperateTime    string   // 操作时间
	TableAlias     string   // 表别名（仅用于拼接查询语句，配置别名请用ModelTableAlias）

	// ----- 非基础配置 ----- /
	BaseControllerTrait controller.Base
}

// InitCrud 初始化crud，仅当初始化完成才可以使用
func (t *Trait) InitCrud(c *gin.Context) {
	_ = helper.CheckAndSetDefault(t)
	t.BaseControllerTrait.GinContext = c

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

	// 存放表别名到上下文，方便查询时调用
	t.TableAlias = t.ModelTableName
	if t.ModelTableAlias != "" {
		t.TableAlias = t.ModelTableAlias
	}
	t.TableAlias += "."
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
// 传递1个参数时，如果是字符串则作为message输出，否则作为data输出；
// 传递2个参数时，第一个参数为message，第二个参数为data；
// 传递3个参数时，第一个参数为message，第二个参数为data，第三个参数为code；
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
// 传递1个参数时，如果是字符串则作为message输出，否则作为data输出；
// 传递2个参数时，第一个参数为data，第二个参数为message；
// 传递3个参数时，第一个参数为data，第二个参数为additionalParams，第三个参数为message；
func (t *Trait) Success(args ...any) {
	t.BaseControllerTrait.Success(args...)
}

// Result 整理结果输出
func (t *Trait) Result(code int, msg string, args ...any) {
	t.BaseControllerTrait.Result(code, msg, args...)
}

// BindMapToStruct 将 map 数据绑定到 struct，并处理类型转换
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
func (t *Trait) setValue(fieldVal reflect.Value, val interface{}) error {
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
	case reflect.Map:
		// 检查输入val是否为map类型
		valReflected := reflect.ValueOf(val)
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
			log.Printf("Expected map type for field, got %s\n", valReflected.Type())
			// 不是map类型，设置为nil
			fieldVal.Set(reflect.Zero(fieldVal.Type()))
		}
	case reflect.Slice:
		// 处理slice类型
		valReflected := reflect.ValueOf(val)
		if valReflected.Kind() == reflect.Slice {
			valType := valReflected.Type()
			fieldType := fieldVal.Type()

			if valType.Elem().AssignableTo(fieldType.Elem()) {
				fieldVal.Set(valReflected)
			} else if valType.Elem().Kind() == reflect.Interface && fieldType.Elem().Kind() == reflect.String {
				// 如果slice中的元素是interface{}，而目标类型是string，则进行转换
				convertedSlice := reflect.MakeSlice(fieldType, valReflected.Len(), valReflected.Cap())
				for i := 0; i < valReflected.Len(); i++ {
					strVal := helper.Convert{Value: valReflected.Index(i).Interface()}.ToString()
					convertedSlice.Index(i).SetString(strVal)
				}
				fieldVal.Set(convertedSlice)
			} else {
				log.Printf("Slice element type mismatch: expected %s elements but got %s elements\n", fieldType.Elem(), valType.Elem())
				fieldVal.Set(reflect.Zero(fieldType))
			}
		} else {
			log.Printf("Expected slice type for field, got %s\n", valReflected.Type())
			fieldVal.Set(reflect.Zero(fieldVal.Type()))
		}
	case reflect.Struct:
		valReflected := reflect.ValueOf(val)
		if valReflected.Kind() != reflect.Struct {
			log.Printf("Expected struct type for assignment but got %s\n", valReflected.Type())
			fieldVal.Set(reflect.Zero(fieldVal.Type())) // 设置为零值
			return nil                                  // 返回 nil，不中断程序
		}

		// 遍历结构体中的每个字段
		for i := 0; i < fieldVal.NumField(); i++ {
			field := fieldVal.Type().Field(i)
			subFieldVal := fieldVal.Field(i)

			if !subFieldVal.CanSet() {
				log.Printf("Field %s cannot be set. \n", field.Name)
				continue
			}

			// 检查val中是否存在此字段
			subValField := valReflected.FieldByName(field.Name)
			if !subValField.IsValid() {
				continue // 或者根据需要处理不存在的字段
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

// GetSafeMapGPC 安全获取map类型GPC
func (t *Trait) GetSafeMapGPC(key ...string) (mapData map[string]any) {
	return t.BaseControllerTrait.GetSafeMapGPC(key...)
}
