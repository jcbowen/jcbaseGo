package crud

import (
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/debugger"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/orm"
)

// ControllerInterface 定义控制器接口
// 该接口用于规范CRUD控制器必须实现的方法
// 必选方法：CheckInit - 用于控制器初始化检查
type ControllerInterface interface {
	// CheckInit 控制器初始化确认方法
	// 参数：
	//   - ctx any: crud上下文对象(*crud.Context或者*gin.Context,不同的传入，处理不同的逻辑)
	// 功能：在CRUD初始化时调用，用于控制器级别的初始化确认
	CheckInit(ctx any) *Context

	// GetLogger 获取日志记录器
	// 返回值：
	//   - debugger.LoggerInterface: 日志记录器接口实例
	//   - bool: 是否成功获取日志记录器
	GetLogger() debugger.LoggerInterface
}

type Trait struct {
	// ----- 基础配置 ----- /

	PkId               string              `default:"id"` // 数据表主键
	Model              any                 // 模型指针
	ModelTableAlias    string              // 模型表别名
	DBI                orm.Instance        // 数据库实例
	ListResultStruct   interface{}         // 列表返回结构体
	DetailResultStruct interface{}         // 详情返回结构体
	Controller         ControllerInterface // 控制器

	// ----- 初始化时生成 ----- /

	ModelTableName      string   // 模型表名
	ModelFields         []string // 模型所有字段
	SoftDeleteField     string   // 软删除字段名（如: "deleted_at"、"is_deleted" 等）
	SoftDeleteCondition string   // 软删除判断条件（如: "IS NULL" 或 "= '0000-00-00 00:00:00'"，不包含字段名）
	OperateTime         string   // 操作时间
	TableAlias          string   // 表别名（仅用于拼接查询语句，配置别名请用ModelTableAlias）

	// ----- 非基础配置 ----- /
}

// InitCrud 初始化CRUD，仅当初始化完成才可以使用
// 参数说明：
//   - c *gin.Context: Gin框架的上下文对象，包含请求和响应信息
//   - args ...any: 可选的参数列表
//     -- actionName string: 操作名称（如: "create", "update", "delete", "detail", "list", "all", "set-value"）
//
// 返回值：
//   - *Context: 新创建的上下文对象
func (t *Trait) InitCrud(ginContext *gin.Context, args ...any) (ctx *Context) {
	_ = helper.CheckAndSetDefault(t)

	nCtxOpt := &NewContextOpt{
		GinContext: ginContext,
	}

	// 设置ActionName
	if len(args) > 0 {
		nCtxOpt.ActionName, _ = args[0].(string)
	}

	// 创建上下文
	ctx = NewContext(nCtxOpt)

	// 设置json响应头
	ginContext.Set("Content-type", "application/json;charset=utf-8")

	// 调用控制器初始化确认方法
	if t.Controller != nil {
		t.Controller.CheckInit(ctx)
		// 给DBI设置日志记录器
		if t.DBI != nil {
			// 获取日志记录器
			logger := t.Controller.GetLogger()
			t.DBI.SetDebuggerLogger(logger)
		}
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

	// 存放表别名（以点号结尾）到上下文，方便补充到字段前
	t.TableAlias = t.ModelTableName
	if t.ModelTableAlias != "" {
		t.TableAlias = t.ModelTableAlias
	}
	t.TableAlias += "."
	return
}

// callCustomMethod 调用自定义方法（优先控制器方法），否则调用默认方法
// 功能：
//   - 优先在控制器上查找并调用名为 methodName 的方法；若不存在则调用 Trait 自身同名默认方法
//   - 自动适配入参（含类型转换与零值补齐），完整支持可变参方法（使用 CallSlice）
//   - 安全提取返回值，避免因不可导出类型导致崩溃
//
// 参数：
//   - methodName string: 方法名称
//   - args ...interface{}: 入参数组（按目标方法签名顺序传入）
//
// 返回值：
//   - results []interface{}: 方法调用返回的所有结果值（按声明顺序）
//
// 异常：
//   - 当默认方法不存在时触发 panic（保留原行为以提示缺失实现）
//
// 使用示例：
//   - t.callCustomMethod("List", ctx)
//   - t.callCustomMethod("Format", 1, "a")
//   - t.callCustomMethod("Sum", 1, 2, 3) // 对变参方法自动打包
func (t *Trait) callCustomMethod(methodName string, args ...interface{}) (results []interface{}) {
	// 优先尝试控制器方法
	var ctrlMethod reflect.Value
	if t.Controller != nil {
		rv := reflect.ValueOf(t.Controller)
		ctrlMethod = rv.MethodByName(methodName)
		if !ctrlMethod.IsValid() {
			// 兼容指针接收者方法
			if rv.Kind() != reflect.Ptr {
				if rv.CanAddr() {
					ctrlMethod = rv.Addr().MethodByName(methodName)
				} else {
					// 创建可寻址副本以获取指针方法集（副本调用）
					tmp := reflect.New(rv.Type())
					if tmp.Elem().CanSet() {
						tmp.Elem().Set(rv)
					}
					ctrlMethod = tmp.MethodByName(methodName)
				}
			}
		}
	}

	if ctrlMethod.IsValid() {
		in, ok, useSlice := buildCallInputs(ctrlMethod, args)
		if ok {
			var resultValues []reflect.Value
			if useSlice {
				resultValues = ctrlMethod.CallSlice(in)
			} else {
				resultValues = ctrlMethod.Call(in)
			}
			for i := 0; i < len(resultValues); i++ {
				if resultValues[i].CanInterface() {
					results = append(results, resultValues[i].Interface())
				} else {
					results = append(results, nil)
				}
			}
			return results
		}
		// 参数不匹配时回退到默认方法
	}

	// 默认方法路径（保持原有 panic 行为）
	defaultMethod := reflect.ValueOf(t).MethodByName(methodName)
	if !defaultMethod.IsValid() {
		log.Panic("默认方法不存在：" + methodName)
	}
	in, ok, useSlice := buildCallInputs(defaultMethod, args)
	if !ok {
		// 参数不匹配时，用零值补齐或截断以避免崩溃
		in, _, useSlice = buildCallInputs(defaultMethod, nil)
	}
	var resultsValue []reflect.Value
	if useSlice {
		resultsValue = defaultMethod.CallSlice(in)
	} else {
		resultsValue = defaultMethod.Call(in)
	}
	for i := 0; i < len(resultsValue); i++ {
		if resultsValue[i].CanInterface() {
			results = append(results, resultsValue[i].Interface())
		} else {
			results = append(results, nil)
		}
	}
	return
}

// buildCallInputs 构造反射调用入参，支持可变参与类型转换
// 功能：
//   - 非变参：按方法参数列表顺序构造入参；当传入不足或类型不匹配时使用参数类型零值
//   - 变参：固定参数位按上诉规则构造；剩余参数自动打包为目标切片类型并使用 CallSlice 调用
//   - 若调用者已传入一个与目标切片类型一致的值，则直接使用该切片
//
// 参数：
//   - method reflect.Value: 目标方法
//   - args []interface{}: 原始参数
//
// 返回值：
//   - in []reflect.Value: 适配后的入参
//   - ok bool: 是否成功适配（始终返回 true；保守策略避免崩溃）
//   - useSlice bool: 是否需使用 CallSlice（仅在变参方法时为 true）
//
// 异常：
//   - 无
//
// 使用示例：
//   - in, ok, useSlice := buildCallInputs(m, []interface{}{1, "a"})
func buildCallInputs(method reflect.Value, args []interface{}) (in []reflect.Value, ok bool, useSlice bool) {
	mt := method.Type()
	nin := mt.NumIn()
	variadic := mt.IsVariadic()

	if !variadic {
		in = make([]reflect.Value, nin)
		for i := 0; i < nin; i++ {
			var a interface{}
			if i < len(args) {
				a = args[i]
			}
			pt := mt.In(i)
			if a == nil {
				in[i] = reflect.Zero(pt)
				continue
			}
			av := reflect.ValueOf(a)
			if av.Type().AssignableTo(pt) {
				in[i] = av
				continue
			}
			if av.Type().ConvertibleTo(pt) {
				in[i] = av.Convert(pt)
				continue
			}
			in[i] = reflect.Zero(pt)
		}
		return in, true, false
	}

	// 处理变参：前置固定参数 + 最后一个切片参数
	fixed := nin - 1
	in = make([]reflect.Value, nin)
	for i := 0; i < fixed; i++ {
		var a interface{}
		if i < len(args) {
			a = args[i]
		}
		pt := mt.In(i)
		if a == nil {
			in[i] = reflect.Zero(pt)
			continue
		}
		av := reflect.ValueOf(a)
		if av.Type().AssignableTo(pt) {
			in[i] = av
			continue
		}
		if av.Type().ConvertibleTo(pt) {
			in[i] = av.Convert(pt)
			continue
		}
		in[i] = reflect.Zero(pt)
	}

	sliceT := mt.In(nin - 1)
	elemT := sliceT.Elem()
	remain := 0
	if len(args) > fixed {
		remain = len(args) - fixed
	}
	// 单个切片参数直接使用，否则构造切片
	if remain == 1 {
		last := args[fixed]
		if last != nil {
			lv := reflect.ValueOf(last)
			if lv.Type().AssignableTo(sliceT) {
				in[nin-1] = lv
				return in, true, true
			}
		}
	}

	sl := reflect.MakeSlice(sliceT, remain, remain)
	for i := 0; i < remain; i++ {
		var a interface{}
		if fixed+i < len(args) {
			a = args[fixed+i]
		}
		if a == nil {
			sl.Index(i).Set(reflect.Zero(elemT))
			continue
		}
		av := reflect.ValueOf(a)
		if av.Type().AssignableTo(elemT) {
			sl.Index(i).Set(av)
			continue
		}
		if av.Type().ConvertibleTo(elemT) {
			sl.Index(i).Set(av.Convert(elemT))
			continue
		}
		sl.Index(i).Set(reflect.Zero(elemT))
	}
	in[nin-1] = sl
	return in, true, true
}

// ----- 公共方法 ----- /

// ExtractPkId 从不同类型的请求中提取主键ID
// 参数：
//   - c *gin.Context: Gin框架的上下文对象
//
// 返回值：
//   - pkValue uint: 提取到的主键ID值
//   - err error: 提取过程中的错误信息
func (t *Trait) ExtractPkId(ctx *Context) (pkValue uint, err error) {
	gpcInterface, GPCExists := ctx.GinContext.Get("GPC")
	if !GPCExists {
		return 0, errors.New("GPC data not found")
	}
	gpc := gpcInterface.(map[string]map[string]any)["all"]

	idStr, ok := gpc[t.PkId]
	if !ok {
		return 0, errors.New("PkId not found in GPC data")
	}
	pkValue = helper.Convert{Value: idStr}.ToUint()

	return
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
