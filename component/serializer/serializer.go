// Package serializer 提供响应数据的递归序列化能力。
//
// 支持将结构体、切片、map、指针等嵌套数据转换为统一的 map/slice 结构，
// 并在此过程中按需格式化时间字段、加密 uint64 类型的 ID 字段。
// 通过最大深度限制与循环引用检测，避免自引用结构体或极深嵌套导致栈溢出。
package serializer

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// defaultMaxDepth 为 Process 递归处理的默认最大深度。
const defaultMaxDepth = 32

// Options 序列化处理可配置项。所有回调均为可选注入，业务策略由调用方提供。
type Options struct {
	// EncryptID 可选：对 uint64 ID 字段加密（如 iMe 注入 utils.EncryptID）
	EncryptID func(uint64) string
	// IsIDField 可选：判断某字段是否需加密（默认空，即不加密）
	IsIDField func(string) bool
	// IsTimeField 可选：判断某字段是否需格式化为北京时间（默认空，即不处理时间字段）
	IsTimeField func(string) bool
	// FormatTime 可选：将时间值格式化为北京时间字符串
	FormatTime func(interface{}) (string, bool)
	// SkipField 可选：判断某结构体字段是否跳过（如跳过匿名嵌入的 orm base model）
	SkipField func(reflect.StructField) bool
	// MaxDepth 可选：递归处理的最大深度，小于等于 0 时使用默认值 32
	MaxDepth int
	// LegacyMode 可选：使用旧版 json.Marshal/Unmarshal 路径处理结构体与切片，
	// 保持与重构前一致的输出行为。
	LegacyMode bool
}

// maxDepth 返回实际使用的最大递归深度。
func (o *Options) maxDepth() int {
	if o == nil || o.MaxDepth <= 0 {
		return defaultMaxDepth
	}
	return o.MaxDepth
}

// Option 函数式选项类型。
type Option func(*Options)

// WithEncryptID 注入 ID 加密函数。
func WithEncryptID(fn func(uint64) string) Option {
	return func(o *Options) { o.EncryptID = fn }
}

// WithIDField 注入 ID 字段判断函数。
func WithIDField(fn func(string) bool) Option {
	return func(o *Options) { o.IsIDField = fn }
}

// WithTimeField 注入时间字段判断函数。
func WithTimeField(fn func(string) bool) Option {
	return func(o *Options) { o.IsTimeField = fn }
}

// WithFormatTime 注入时间值格式化函数。
func WithFormatTime(fn func(interface{}) (string, bool)) Option {
	return func(o *Options) { o.FormatTime = fn }
}

// WithSkipField 注入字段跳过判断函数。
func WithSkipField(fn func(reflect.StructField) bool) Option {
	return func(o *Options) { o.SkipField = fn }
}

// WithMaxDepth 注入递归处理最大深度。
//
// 参数:
//   - depth: 最大递归深度，小于等于 0 时使用默认值 32。
func WithMaxDepth(depth int) Option {
	return func(o *Options) { o.MaxDepth = depth }
}

// WithLegacyMode 注入旧版序列化模式。
//
// 开启后，结构体通过 json.Marshal/Unmarshal 转为 map（与原 CRUD 行为一致），
// 切片仅转换为 []interface{} 而不递归处理元素，map 保持原样返回。
// 该模式主要用于向后兼容，避免已上线项目因序列化逻辑变更而出现响应差异。
func WithLegacyMode(enabled bool) Option {
	return func(o *Options) { o.LegacyMode = enabled }
}

// visitState 用于循环引用检测的 DFS 访问状态。
type visitState int

const (
	unvisited visitState = iota
	visiting
	visited
)

// Process 递归处理数据：统一格式化时间字段，并按需加密 uint64 类型的 ID 字段。
// 支持 map、slice、struct、指针、嵌套结构。
//
// 通过最大深度限制与循环引用检测，避免自引用结构体或极深嵌套导致栈溢出。
// 超出最大深度时返回原始数据，避免子树数据缺失。
//
// 参数:
//   - data: 待处理的数据
//   - opts: 可选的处理策略
//
// 返回值:
//   - 处理后的数据；若检测到循环引用，对应位置返回 nil 以打破循环。
func Process(data interface{}, opts ...Option) interface{} {
	o := &Options{}
	for _, opt := range opts {
		opt(o)
	}
	if o.LegacyMode {
		return legacyProcess(data)
	}
	return processDataRecursive(data, o, 0, make(map[uintptr]visitState))
}

// legacyProcess 使用重构前的 json.Marshal/Unmarshal 路径处理数据，
// 用于在兼容模式下保持旧行为。
func legacyProcess(data interface{}) interface{} {
	val := reflect.ValueOf(data)

	// 如果是指针类型，循环解引用直到非指针
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
		data = val.Interface()
	}

	switch val.Kind() {
	case reflect.Struct:
		resultMapData := make(map[string]interface{})
		jsonData, err := json.Marshal(data)
		if err != nil {
			return data
		}
		if err := json.Unmarshal(jsonData, &resultMapData); err != nil {
			return data
		}
		return resultMapData
	case reflect.Map:
		// 保持与旧逻辑一致：gin.H 与 map[string]any 直接返回
		if h, ok := data.(gin.H); ok {
			return map[string]interface{}(h)
		}
		if m, ok := data.(map[string]interface{}); ok {
			return m
		}
		// 其他 map 类型尝试 json 往返转换，失败时返回原值
		resultMapData := make(map[string]interface{})
		jsonData, err := json.Marshal(data)
		if err != nil {
			return data
		}
		if err := json.Unmarshal(jsonData, &resultMapData); err != nil {
			return data
		}
		return resultMapData
	case reflect.String:
		return data.(string)
	case reflect.Array, reflect.Slice:
		return convertToInterfaceSlice(data)
	}
	return data
}

// convertToInterfaceSlice 将特定类型的切片转换为通用的 interface{} 切片。
func convertToInterfaceSlice(slice interface{}) []interface{} {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil
	}

	interfaceSlice := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		interfaceSlice[i] = v.Index(i).Interface()
	}

	return interfaceSlice
}

// processDataRecursive 为 Process 的内部递归实现。
//
// 参数:
//   - data: 待处理的数据
//   - o: 处理选项
//   - depth: 当前递归深度
//   - states: 结构体指针的 DFS 访问状态，用于检测循环引用
//
// 返回值:
//   - 处理后的数据；若超出最大深度返回原始数据，若检测到循环引用返回 nil。
func processDataRecursive(data interface{}, o *Options, depth int, states map[uintptr]visitState) interface{} {
	if data == nil {
		return nil
	}
	if depth >= o.maxDepth() {
		return data
	}

	// 优先处理已知的 interface 类型（gin.H 等）
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(v))
		for key, val := range v {
			result[key] = processValueRecursive(key, val, o, depth+1, states)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = processDataRecursive(item, o, depth+1, states)
		}
		return result
	}

	// 处理结构体、结构体切片/数组、以及带具体类型的 map
	rv := reflect.ValueOf(data)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	// 循环引用检测：结构体处理前检查指针地址
	if rv.Kind() == reflect.Struct && rv.CanAddr() {
		ptr := rv.Addr().Pointer()
		switch states[ptr] {
		case visiting:
			// 检测到循环引用，返回 nil 打破循环
			return nil
		case visited:
			// 已处理过的共享节点（DAG），直接返回原始值
			return data
		}
		states[ptr] = visiting
		result := structToMapRecursive(rv, o, depth, states)
		states[ptr] = visited
		return result
	}

	switch rv.Kind() {
	case reflect.Struct:
		return structToMapRecursive(rv, o, depth, states)
	case reflect.Slice, reflect.Array:
		result := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = processDataRecursive(rv.Index(i).Interface(), o, depth+1, states)
		}
		return result
	case reflect.Map:
		result := make(map[string]interface{}, rv.Len())
		for _, key := range rv.MapKeys() {
			keyStr := fmt.Sprint(key.Interface())
			result[keyStr] = processValueRecursive(keyStr, rv.MapIndex(key).Interface(), o, depth+1, states)
		}
		return result
	}

	return data
}

// structToMapRecursive 将结构体按 json tag 转换为 map，并递归处理字段。
//
// 参数:
//   - v: 结构体 reflect.Value
//   - o: 处理选项
//   - depth: 当前递归深度
//   - states: 结构体指针的 DFS 访问状态
//
// 返回值:
//   - 转换后的 map。
func structToMapRecursive(v reflect.Value, o *Options, depth int, states map[uintptr]visitState) map[string]interface{} {
	rt := v.Type()
	result := make(map[string]interface{}, rt.NumField())

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)

		// 跳过未导出字段
		if !field.IsExported() {
			continue
		}

		// 跳过匿名嵌入的 base model 等字段（由调用方注入判断）
		if field.Anonymous && o.SkipField != nil && o.SkipField(field) {
			continue
		}

		key := jsonFieldName(field)
		if key == "-" {
			continue
		}

		fv := v.Field(i)
		// nil 指针直接置为 nil 以便 JSON 输出 null；否则保留指针原始值，
		// 使递归过程中循环引用检测能命中同一内存地址。
		if fv.Kind() == reflect.Ptr && fv.IsNil() {
			result[key] = nil
		} else {
			result[key] = processValueRecursive(key, fv.Interface(), o, depth+1, states)
		}
	}

	return result
}

// jsonFieldName 从结构体字段的 json tag 中提取字段名。
func jsonFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}

	parts := strings.Split(tag, ",")
	name := strings.TrimSpace(parts[0])
	if name == "-" {
		return "-"
	}
	if name == "" {
		return field.Name
	}
	return name
}

// ProcessValue 根据字段名处理单个字段：优先格式化时间，再按需加密 ID。
//
// 参数:
//   - key: 字段名
//   - val: 字段值
//   - opts: 可选的处理策略
//
// 返回值:
//   - 处理后的字段值。
func ProcessValue(key string, val interface{}, opts ...Option) interface{} {
	o := &Options{}
	for _, opt := range opts {
		opt(o)
	}
	return processValueRecursive(key, val, o, 0, make(map[uintptr]visitState))
}

// processValueRecursive 为 ProcessValue 的内部递归实现。
//
// 参数:
//   - key: 字段名
//   - val: 字段值
//   - o: 处理选项
//   - depth: 当前递归深度
//   - states: 结构体指针的 DFS 访问状态
//
// 返回值:
//   - 处理后的字段值。
func processValueRecursive(key string, val interface{}, o *Options, depth int, states map[uintptr]visitState) interface{} {
	// 进入递归前检查深度，超过 MaxDepth 时直接返回原始值，避免无限递归导致栈溢出，
	// 同时也防止对过深的嵌套结构执行无意义的解指针和时间格式化。
	if depth >= o.maxDepth() {
		return val
	}

	// 保存原始值，以便递归嵌套结构体时保留指针地址信息，用于循环引用检测。
	originalVal := val

	// 解指针以获取实际值，用于类型判断、时间格式化和 ID 加密。
	rv := reflect.ValueOf(val)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
		val = rv.Interface()
	}

	// 时间字段优先格式化
	if o.IsTimeField != nil && o.IsTimeField(key) && o.FormatTime != nil {
		if formatted, ok := o.FormatTime(val); ok {
			return formatted
		}
	}

	switch v := val.(type) {
	case uint64:
		if o.EncryptID != nil && o.IsIDField != nil && o.IsIDField(key) {
			return o.EncryptID(v)
		}
		return v
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = processDataRecursive(item, o, depth+1, states)
		}
		return result
	case map[string]interface{}:
		return processDataRecursive(v, o, depth, states)
	default:
		// 对 reflect.Kind 为 uint64 的命名类型（如 utils.EncryptedID）做兜底
		if reflect.TypeOf(val) != nil && reflect.TypeOf(val).Kind() == reflect.Uint64 {
			if o.EncryptID != nil && o.IsIDField != nil && o.IsIDField(key) {
				return o.EncryptID(reflect.ValueOf(val).Uint())
			}
		}

		// 嵌套结构体或具体类型切片递归处理
		// 优先使用原始值（保留指针），使循环引用检测能命中同一内存地址。
		rv := reflect.ValueOf(originalVal)
		for rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return originalVal
			}
			rv = rv.Elem()
		}
		switch rv.Kind() {
		case reflect.Struct, reflect.Slice, reflect.Array:
			return processDataRecursive(originalVal, o, depth, states)
		}

		return originalVal
	}
}
