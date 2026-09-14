package helper

import (
	"testing"
	"time"
)

// InterfaceHolder 用于验证 interface 字段承载 map / 结构体 / 结构体指针时的默认值补充
type InterfaceHolder struct {
	Any interface{} `json:"any"`
}

// NestedMapCfg 指针字段与 interface 承载场景使用的子配置结构体
// 内嵌 map[string]SubDefaultConfig，用于验证 *struct 字段内的 map 元素可被递归补充
type NestedMapCfg struct {
	Items map[string]SubDefaultConfig `json:"items"`
	Name  string                      `json:"name" default:"inner-cfg"`
}

// TestCheckAndSetDefault_InterfaceHoldingMap 验证 interface 字段承载 map 时元素默认值可被补充
// 函数名：TestCheckAndSetDefault_InterfaceHoldingMap
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_InterfaceHoldingMap
func TestCheckAndSetDefault_InterfaceHoldingMap(t *testing.T) {
	cfg := &InterfaceHolder{Any: map[string]SubDefaultConfig{"main": {}}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items, ok := cfg.Any.(map[string]SubDefaultConfig)
	if !ok {
		t.Fatalf("interface 承载类型被改变: %T", cfg.Any)
	}
	if items["main"].Host != "127.0.0.1" || items["main"].Port != 3306 {
		t.Fatalf("interface 承载的 map 元素未补充默认值: %+v", items["main"])
	}
}

// TestCheckAndSetDefault_InterfaceHoldingStruct 验证 interface 字段承载值结构体时的默认值补充
// 函数名：TestCheckAndSetDefault_InterfaceHoldingStruct
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_InterfaceHoldingStruct
func TestCheckAndSetDefault_InterfaceHoldingStruct(t *testing.T) {
	cfg := &InterfaceHolder{Any: SubDefaultConfig{}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sub, ok := cfg.Any.(SubDefaultConfig)
	if !ok {
		t.Fatalf("interface 承载类型被改变: %T", cfg.Any)
	}
	if sub.Host != "127.0.0.1" || sub.Port != 3306 || sub.Timeout != 500*time.Millisecond {
		t.Fatalf("interface 承载的结构体未补充默认值: %+v", sub)
	}
}

// TestCheckAndSetDefault_InterfaceHoldingPtrStruct 验证 interface 字段承载结构体指针时的默认值补充
// 函数名：TestCheckAndSetDefault_InterfaceHoldingPtrStruct
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_InterfaceHoldingPtrStruct
func TestCheckAndSetDefault_InterfaceHoldingPtrStruct(t *testing.T) {
	cfg := &InterfaceHolder{Any: &SubDefaultConfig{}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := cfg.Any.(*SubDefaultConfig)
	if !ok {
		t.Fatalf("interface 承载类型被改变: %T", cfg.Any)
	}
	if got.Host != "127.0.0.1" || got.Port != 3306 {
		t.Fatalf("interface 承载的结构体指针未补充默认值: %+v", got)
	}
}

// TestCheckAndSetDefault_InterfaceNonStruct 验证 interface 承载基础类型时不被改动
// 函数名：TestCheckAndSetDefault_InterfaceNonStruct
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_InterfaceNonStruct
func TestCheckAndSetDefault_InterfaceNonStruct(t *testing.T) {
	cfg := &InterfaceHolder{Any: "plain"}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Any != "plain" {
		t.Fatalf("interface 承载的基础类型不应被改动: %+v", cfg.Any)
	}

	empty := &InterfaceHolder{}
	if err := CheckAndSetDefault(empty); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if empty.Any != nil {
		t.Fatalf("nil interface 字段不应被发现或实例化: %+v", empty.Any)
	}
}

// TestCheckAndSetDefault_PtrStructFieldMap 验证 *struct 字段内的 map 元素默认值可被补充
// 函数名：TestCheckAndSetDefault_PtrStructFieldMap
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_PtrStructFieldMap
func TestCheckAndSetDefault_PtrStructFieldMap(t *testing.T) {
	type Cfg struct {
		Inner *NestedMapCfg `json:"inner"`
	}

	cfg := &Cfg{Inner: &NestedMapCfg{Items: map[string]SubDefaultConfig{"main": {}}}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Inner.Name != "inner-cfg" {
		t.Fatalf("指针字段自身默认值未补充: %+v", cfg.Inner)
	}
	if cfg.Inner.Items["main"].Host != "127.0.0.1" || cfg.Inner.Items["main"].Port != 3306 {
		t.Fatalf("*struct 字段内的 map 元素未补充默认值: %+v", cfg.Inner.Items["main"])
	}
}

// TestCheckAndSetDefault_NilPtrStructField 验证 nil 结构体指针字段保持 nil，不被自动实例化
// 函数名：TestCheckAndSetDefault_NilPtrStructField
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：nil 表达「字段不存在 / 未启用」，自动 New 会改变调用方语义，故约定保持原样
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_NilPtrStructField
func TestCheckAndSetDefault_NilPtrStructField(t *testing.T) {
	type Cfg struct {
		Inner *NestedMapCfg `json:"inner"`
	}

	cfg := &Cfg{}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Inner != nil {
		t.Fatalf("nil 指针字段应保持 nil 而不被实例化: %+v", cfg.Inner)
	}
}

// TestCheckAndSetDefault_PtrToScalar 验证指向基础类型的指针字段不被递归处理
// 函数名：TestCheckAndSetDefault_PtrToScalar
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_PtrToScalar
func TestCheckAndSetDefault_PtrToScalar(t *testing.T) {
	type Cfg struct {
		NumPtr *int `json:"num_ptr" default:"99"`
	}

	num := 0
	cfg := &Cfg{NumPtr: &num}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 基础类型指针不在递归范围内，保持原值
	if cfg.NumPtr == nil || *cfg.NumPtr != 0 {
		t.Fatalf("基础类型指针不应被处理: %+v", cfg.NumPtr)
	}
}

// buildNestedMapChain 构造 map -> interface 交替嵌套的链路，用于深度上限边界测试
// 函数名：buildNestedMapChain
// 参数：depth int — 嵌套层数
// 返回值：map[string]interface{} — 外层映射，最内层为 SubDefaultConfig 值
// 异常：不触发panic
// 使用示例：
//
//	chain := buildNestedMapChain(30)
func buildNestedMapChain(depth int) map[string]interface{} {
	var node interface{} = SubDefaultConfig{}
	for i := 0; i < depth; i++ {
		node = map[string]interface{}{"child": node}
	}
	return map[string]interface{}{"root": node}
}

// digNestedMapChain 沿 buildNestedMapChain 的结构下钻，取回最内层的 SubDefaultConfig
// 函数名：digNestedMapChain
// 参数：
// - data map[string]interface{} — buildNestedMapChain 产出的映射
// - depth int — 嵌套层数，需与构造时一致
// 返回值：
// - SubDefaultConfig — 最内层结构体
// - bool — 下钻是否成功
// 异常：不触发panic
// 使用示例：
//
//	sub, ok := digNestedMapChain(chain, 30)
func digNestedMapChain(data map[string]interface{}, depth int) (SubDefaultConfig, bool) {
	var cur interface{} = data["root"]
	for i := 0; i < depth; i++ {
		m, ok := cur.(map[string]interface{})
		if !ok {
			return SubDefaultConfig{}, false
		}
		cur = m["child"]
	}
	sub, ok := cur.(SubDefaultConfig)
	return sub, ok
}

// buildNestedMapChainPtr 构造 map -> interface 交替嵌套的链路，用于指针元素深度边界测试
// 函数名：buildNestedMapChainPtr
// 参数：depth int — 嵌套层数
// 返回值：map[string]interface{} — 外层映射，最内层为 *SubDefaultConfig
// 异常：不触发panic
// 使用示例：
//
//	chain := buildNestedMapChainPtr(30)
func buildNestedMapChainPtr(depth int) map[string]interface{} {
	var node interface{} = &SubDefaultConfig{}
	for i := 0; i < depth; i++ {
		node = map[string]interface{}{"child": node}
	}
	return map[string]interface{}{"root": node}
}

// digNestedMapChainAny 沿嵌套 map 链下钻并取回最内层的原始值
// 函数名：digNestedMapChainAny
// 参数：
// - data map[string]interface{} — 嵌套链路的外层映射
// - depth int — 嵌套层数，需与构造时一致
// 返回值：
// - interface{} — 最内层原始值（可能是结构体、结构体指针等）
// - bool — 下钻是否成功
// 异常：不触发panic
// 使用示例：
//
//	raw, ok := digNestedMapChainAny(chain, 30)
func digNestedMapChainAny(data map[string]interface{}, depth int) (interface{}, bool) {
	var cur interface{} = data["root"]
	for i := 0; i < depth; i++ {
		m, ok := cur.(map[string]interface{})
		if !ok {
			return nil, false
		}
		cur = m["child"]
	}
	return cur, true
}

// TestCheckAndSetDefault_DepthBoundary 验证深度上限的精确边界行为
// 函数名：TestCheckAndSetDefault_DepthBoundary
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：
//   - 深度语义统一为「容器嵌套层数」：每深入一层 struct / map 元素 / 指针指向的 struct 消耗一层，
//     interface 仅作拆包不消耗深度，字段自身的 map 容器也不消耗（其元素才计入）
//   - 上限内必须正常填充；超过上限必须静默截断（既不 panic、不栈溢出，也确实不再补充默认值），
//     两侧都要断言，避免出现「把上限改成极大值也照样通过」的弱断言
//   - 值元素与指针元素必须落在同一条边界上（深度口径统一的回归保护）
//
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_DepthBoundary
func TestCheckAndSetDefault_DepthBoundary(t *testing.T) {
	type Cfg struct {
		Data map[string]interface{} `json:"data"`
	}

	// 远离上限的层数：上限内必须生效、明显超限必须被截断（同时验证不 panic、不栈溢出）
	for _, depth := range []int{1, 8, 16} {
		cfg := &Cfg{Data: buildNestedMapChain(depth)}
		if err := CheckAndSetDefault(cfg); err != nil {
			t.Fatalf("depth=%d unexpected error: %v", depth, err)
		}
		sub, ok := digNestedMapChain(cfg.Data, depth)
		if !ok {
			t.Fatalf("depth=%d 下钻失败，结构被改变", depth)
		}
		if sub.Host != "127.0.0.1" {
			t.Fatalf("depth=%d 应在上限内被补充默认值: %+v", depth, sub)
		}
	}
	for _, depth := range []int{40, 80, 200} {
		cfg := &Cfg{Data: buildNestedMapChain(depth)}
		if err := CheckAndSetDefault(cfg); err != nil {
			t.Fatalf("depth=%d unexpected error: %v", depth, err)
		}
		sub, ok := digNestedMapChain(cfg.Data, depth)
		if !ok {
			t.Fatalf("depth=%d 下钻失败，结构被改变", depth)
		}
		if sub.Host != "" {
			t.Fatalf("depth=%d 超过深度上限后不应再补充默认值: %+v", depth, sub)
		}
	}

	// 紧邻上限的相邻两层必须一升一降。层数与深度的对应关系：
	// buildNestedMapChain(d) 共 d+1 层 map，最内层 map 的元素按深度 d 处理，
	// 若最内层是值结构体则再消耗一层（按 d+1 递归），故最后一层可用深度为 max-2
	lastOK := maxStructDefaultDepth - 2
	firstOver := maxStructDefaultDepth - 1
	for _, depth := range []int{lastOK, firstOver} {
		want := ""
		if depth == lastOK {
			want = "127.0.0.1"
		}

		// 值结构体元素
		cfgValue := &Cfg{Data: buildNestedMapChain(depth)}
		if err := CheckAndSetDefault(cfgValue); err != nil {
			t.Fatalf("depth=%d unexpected error: %v", depth, err)
		}
		sub, ok := digNestedMapChain(cfgValue.Data, depth)
		if !ok {
			t.Fatalf("depth=%d 下钻失败，结构被改变", depth)
		}
		if sub.Host != want {
			t.Fatalf("depth=%d 值元素边界不符: got %q want %q", depth, sub.Host, want)
		}

		// 结构体指针元素必须与值元素落在同一层边界（深度口径统一的回归保护）
		cfgPtr := &Cfg{Data: buildNestedMapChainPtr(depth)}
		if err := CheckAndSetDefault(cfgPtr); err != nil {
			t.Fatalf("ptr depth=%d unexpected error: %v", depth, err)
		}
		raw, ok := digNestedMapChainAny(cfgPtr.Data, depth)
		if !ok {
			t.Fatalf("ptr depth=%d 下钻失败，结构被改变", depth)
		}
		got, ok := raw.(*SubDefaultConfig)
		if !ok {
			t.Fatalf("ptr depth=%d 元素类型被改变: %T", depth, raw)
		}
		if got.Host != want {
			t.Fatalf("ptr depth=%d 指针元素边界与值元素不一致: got %q want %q", depth, got.Host, want)
		}
	}
}

// TestCheckAndSetDefault_SharedPointerNotSkipped 验证同一指针先经深链被访问后，浅层引用仍会被补充默认值
// 函数名：TestCheckAndSetDefault_SharedPointerNotSkipped
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：
//   - visited 语义为「当前递归路径」集合（进入登记、返回注销），深链处理完即注销，
//     不会把同一指针在浅层字段/map 键上的引用误判为「已处理」而静默漏填
//   - 深度耗尽时 setDefaultForPtrField 直接返回且不登记地址，避免在 visited 中残留未真正处理过的记录
//
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_SharedPointerNotSkipped
func TestCheckAndSetDefault_SharedPointerNotSkipped(t *testing.T) {
	// max-1 是最容易暴露该缺陷的层数：深链上的指针刚好会在登记地址后被深度上限截断，
	// 若 visited 不复位就会残留下一条「未真正处理过」的记录，把浅层引用吞掉
	depths := []int{10, 30, 40, 80, maxStructDefaultDepth - 1, maxStructDefaultDepth, maxStructDefaultDepth + 1}
	for _, depth := range depths {
		ptr := &SubDefaultConfig{}
		var node interface{} = ptr
		for i := 0; i < depth; i++ {
			node = map[string]interface{}{"child": node}
		}

		type Cfg struct {
			Deep    map[string]interface{} `json:"deep"`
			Shallow *SubDefaultConfig      `json:"shallow"`
		}
		cfg := &Cfg{
			Deep:    map[string]interface{}{"root": node},
			Shallow: ptr,
		}
		if err := CheckAndSetDefault(cfg); err != nil {
			t.Fatalf("depth=%d unexpected error: %v", depth, err)
		}
		if cfg.Shallow != ptr {
			t.Fatalf("depth=%d 浅层指针被替换: %+v", depth, cfg.Shallow)
		}
		if cfg.Shallow.Host != "127.0.0.1" || cfg.Shallow.Port != 3306 {
			t.Fatalf("depth=%d 浅层引用的默认值被深链路径吞掉: %+v", depth, cfg.Shallow)
		}
	}
}
