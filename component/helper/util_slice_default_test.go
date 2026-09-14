package helper

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

// SliceHolder 切片默认值测试使用的宿主结构体
// 集中声明各类切片字段，便于逐场景断言
type SliceHolder struct {
	Name    string                        `json:"name" default:"holder"`
	Items   []SubDefaultConfig            `json:"items"`
	Ptrs    []*SubDefaultConfig           `json:"ptrs"`
	Anys    []interface{}                 `json:"anys"`
	Nested  [][]SubDefaultConfig          `json:"nested"`
	Groups  map[string][]SubDefaultConfig `json:"groups"`
	Labels  []string                      `json:"labels" default:"a,b"`
	Weights []int                         `json:"weights"`
}

// ---- 结构体字段：[]SubConfig ----

// TestCheckAndSetDefault_SliceStructValue 验证 []SubConfig 字段的元素默认值被递归补充
// 函数名：TestCheckAndSetDefault_SliceStructValue
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_SliceStructValue
func TestCheckAndSetDefault_SliceStructValue(t *testing.T) {
	type Cfg struct {
		Items []SubDefaultConfig `json:"items"`
	}

	cfg := &Cfg{Items: []SubDefaultConfig{
		{},
		{Host: "10.0.0.1", Port: 3307},
	}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Items) != 2 {
		t.Fatalf("切片长度不应改变: %d", len(cfg.Items))
	}
	// 空元素被补齐全部默认值
	empty := cfg.Items[0]
	if empty.Host != "127.0.0.1" || empty.Port != 3306 || empty.Timeout != 500*time.Millisecond {
		t.Fatalf("值元素未补充默认值: %+v", empty)
	}
	// 已有值保留，仅补齐缺失字段
	custom := cfg.Items[1]
	if custom.Host != "10.0.0.1" || custom.Port != 3307 || custom.Timeout != 500*time.Millisecond {
		t.Fatalf("值元素既有字段不应被覆盖: %+v", custom)
	}
}

// ---- 结构体字段：[]*SubConfig ----

// TestCheckAndSetDefault_SlicePtrStructValue 验证 []*SubConfig 字段的非 nil 指针就地补齐、nil 指针保持原样
// 函数名：TestCheckAndSetDefault_SlicePtrStructValue
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：nil 指针表达「元素不存在」，自动 New 会改变调用方语义，故约定保持原样（与 map 值指针一致）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_SlicePtrStructValue
func TestCheckAndSetDefault_SlicePtrStructValue(t *testing.T) {
	type Cfg struct {
		Ptrs []*SubDefaultConfig `json:"ptrs"`
	}

	first := &SubDefaultConfig{}
	custom := &SubDefaultConfig{Host: "10.0.0.9"}
	nilItem := (*SubDefaultConfig)(nil)

	cfg := &Cfg{Ptrs: []*SubDefaultConfig{first, custom, nilItem}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 非 nil 指针被就地补齐（指针地址不变）
	if cfg.Ptrs[0] != first {
		t.Fatalf("元素指针被替换: %p != %p", cfg.Ptrs[0], first)
	}
	if first.Host != "127.0.0.1" || first.Port != 3306 || first.Timeout != 500*time.Millisecond {
		t.Fatalf("指针元素未补充默认值: %+v", first)
	}
	// 既有值保留
	if custom.Host != "10.0.0.9" || custom.Port != 3306 {
		t.Fatalf("指针元素既有字段不应被覆盖: %+v", custom)
	}
	// nil 指针保持 nil，不被实例化
	if cfg.Ptrs[2] != nil {
		t.Fatalf("nil 指针元素应保持 nil 而不被实例化: %+v", cfg.Ptrs[2])
	}
}

// ---- 结构体字段：[]interface{} ----

// TestCheckAndSetDefault_SliceInterfaceValue 验证 []interface{} 字段按元素实际类型分派
// 函数名：TestCheckAndSetDefault_SliceInterfaceValue
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：interface 元素承载结构体时走「拷贝 → 补充 → Set 写回」；承载结构体指针时就地补充；
// 承载基础类型时不改动。三种情况都必须校验类型未被改变
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_SliceInterfaceValue
func TestCheckAndSetDefault_SliceInterfaceValue(t *testing.T) {
	type Cfg struct {
		Anys []interface{} `json:"anys"`
	}

	ptrItem := &SubDefaultConfig{}
	cfg := &Cfg{Anys: []interface{}{SubDefaultConfig{}, ptrItem, "plain", 10, nil}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 值结构体元素：类型不变且默认值补齐
	sub, ok := cfg.Anys[0].(SubDefaultConfig)
	if !ok {
		t.Fatalf("interface 值元素类型被改变: %T", cfg.Anys[0])
	}
	if sub.Host != "127.0.0.1" || sub.Port != 3306 {
		t.Fatalf("interface 值元素未补充默认值: %+v", sub)
	}

	// 结构体指针元素：就地补齐
	gotPtr, ok := cfg.Anys[1].(*SubDefaultConfig)
	if !ok {
		t.Fatalf("interface 指针元素类型被改变: %T", cfg.Anys[1])
	}
	if gotPtr != ptrItem {
		t.Fatalf("interface 指针元素指针被替换: %p != %p", gotPtr, ptrItem)
	}
	if gotPtr.Host != "127.0.0.1" || gotPtr.Port != 3306 {
		t.Fatalf("interface 指针元素未补充默认值: %+v", gotPtr)
	}

	// 基础类型与 nil 元素不被改动
	if cfg.Anys[2] != "plain" || cfg.Anys[3] != 10 || cfg.Anys[4] != nil {
		t.Fatalf("非结构体元素不应被改动: %+v", cfg.Anys)
	}
}

// ---- 结构体字段：[][]SubConfig ----

// TestCheckAndSetDefault_SliceNestedSlice 验证嵌套切片的元素默认值被逐层向内补充
// 函数名：TestCheckAndSetDefault_SliceNestedSlice
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_SliceNestedSlice
func TestCheckAndSetDefault_SliceNestedSlice(t *testing.T) {
	type Cfg struct {
		Nested [][]SubDefaultConfig `json:"nested"`
	}

	cfg := &Cfg{Nested: [][]SubDefaultConfig{
		{{}, {Host: "1.1.1.1"}},
		{},
	}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Nested) != 2 || len(cfg.Nested[0]) != 2 {
		t.Fatalf("嵌套切片结构不应改变: %+v", cfg.Nested)
	}
	if cfg.Nested[0][0].Host != "127.0.0.1" || cfg.Nested[0][0].Port != 3306 {
		t.Fatalf("内层切片元素未补充默认值: %+v", cfg.Nested[0][0])
	}
	if cfg.Nested[0][1].Host != "1.1.1.1" || cfg.Nested[0][1].Port != 3306 {
		t.Fatalf("内层切片元素既有字段不应被覆盖: %+v", cfg.Nested[0][1])
	}
}

// ---- map[string][]SubConfig ----

// TestCheckAndSetDefault_MapOfSlice 验证映射值类型为切片时逐层向内递归补充
// 函数名：TestCheckAndSetDefault_MapOfSlice
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：链路为 map → slice → struct，要求 map 元素（切片）按 depth+1 进入切片分派
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_MapOfSlice
func TestCheckAndSetDefault_MapOfSlice(t *testing.T) {
	type Cfg struct {
		Groups map[string][]SubDefaultConfig `json:"groups"`
	}

	cfg := &Cfg{Groups: map[string][]SubDefaultConfig{
		"a": {{}, {Host: "2.2.2.2"}},
		"b": {},
	}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Groups["a"][0].Host != "127.0.0.1" || cfg.Groups["a"][0].Port != 3306 {
		t.Fatalf("map 值切片元素未补充默认值: %+v", cfg.Groups["a"][0])
	}
	if cfg.Groups["a"][1].Host != "2.2.2.2" || cfg.Groups["a"][1].Port != 3306 {
		t.Fatalf("map 值切片元素既有字段不应被覆盖: %+v", cfg.Groups["a"][1])
	}
	// 空切片不应被凭空加元素
	if len(cfg.Groups["b"]) != 0 {
		t.Fatalf("空切片不应被填充元素: %+v", cfg.Groups["b"])
	}
}

// ---- 基础元素切片：行为不变 ----

// TestCheckAndSetDefault_SliceBasicElements 验证基础元素切片的既有行为不变
// 函数名：TestCheckAndSetDefault_SliceBasicElements
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：[]string / []int 等不会被逐元素改写：
// - 带 default 标签的空切片仍按标签整体填充（新增能力未破坏原路径）
// - 无标签的 nil 切片仍按原逻辑初始化为空切片
// - 非空切片保持原值
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_SliceBasicElements
func TestCheckAndSetDefault_SliceBasicElements(t *testing.T) {
	type Cfg struct {
		Labels  []string `json:"labels" default:"a,b"`
		Weights []int    `json:"weights"`
		Set     []string `json:"set"`
	}

	cfg := &Cfg{Set: []string{"x"}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Labels) != 2 || cfg.Labels[0] != "a" || cfg.Labels[1] != "b" {
		t.Fatalf("带标签的空切片应按标签整体填充: %+v", cfg.Labels)
	}
	if cfg.Weights == nil || len(cfg.Weights) != 0 {
		t.Fatalf("无标签的 nil 切片应初始化为空切片: %+v", cfg.Weights)
	}
	if len(cfg.Set) != 1 || cfg.Set[0] != "x" {
		t.Fatalf("非空切片不应被改动: %+v", cfg.Set)
	}
}

// ---- 无收益类型跳过 ----

// TestCheckAndSetDefault_SliceThirdPartyNotMutated 验证切片中的第三方类型不会被静默改写
// 函数名：TestCheckAndSetDefault_SliceThirdPartyNotMutated
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：sliceElemMayBeConfigStruct 会把 []sync.Mutex 这类无标签类型整段跳过，
// 避免无意义的「拷贝 → 递归 → Set 写回」（反射路径绕过了 go vet 的 copylocks 检查）；
// []time.Time 与 map[string]time.Time 口径一致，均不被逐元素处理
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_SliceThirdPartyNotMutated
func TestCheckAndSetDefault_SliceThirdPartyNotMutated(t *testing.T) {
	type Cfg struct {
		Name    string       `json:"name" default:"holder"`
		Mutexes []sync.Mutex `json:"mutexes"`
		Times   []time.Time  `json:"times"`
	}

	cfg := &Cfg{}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Name != "holder" {
		t.Fatalf("宿主结构体自身默认值应被补充: %+v", cfg.Name)
	}
	// 宿主带 default 标签，故 nil 切片仍按原逻辑初始化为空切片；关键是不做逐元素拷贝
	if len(cfg.Mutexes) != 0 {
		t.Fatalf("[]sync.Mutex 不应被填充元素: %+v", cfg.Mutexes)
	}
	if len(cfg.Times) != 0 {
		t.Fatalf("[]time.Time 不应被填充元素: %+v", cfg.Times)
	}

	// 元素级判定矩阵
	cases := []struct {
		typ  string
		want bool
	}{
		{"[]SubDefaultConfig", true},
		{"[]*SubDefaultConfig", true},
		{"[]interface{}", true},
		{"[][]SubDefaultConfig", true},
		{"[]map[string]SubDefaultConfig", true},
		{"[]sync.Mutex", false},
		{"[]time.Time", false},
		{"[]string", false},
		{"[]int", false},
	}
	for _, c := range cases {
		var got bool
		switch c.typ {
		case "[]SubDefaultConfig":
			got = sliceElemMayBeConfigStruct(typeOfSliceElem[SubDefaultConfig]())
		case "[]*SubDefaultConfig":
			got = sliceElemMayBeConfigStruct(typeOfSliceElem[*SubDefaultConfig]())
		case "[]interface{}":
			got = sliceElemMayBeConfigStruct(typeOfSliceElem[interface{}]())
		case "[][]SubDefaultConfig":
			got = sliceElemMayBeConfigStruct(typeOfSliceElem[[]SubDefaultConfig]())
		case "[]map[string]SubDefaultConfig":
			got = sliceElemMayBeConfigStruct(typeOfSliceElem[map[string]SubDefaultConfig]())
		case "[]sync.Mutex":
			got = sliceElemMayBeConfigStruct(typeOfSliceElem[sync.Mutex]())
		case "[]time.Time":
			got = sliceElemMayBeConfigStruct(typeOfSliceElem[time.Time]())
		case "[]string":
			got = sliceElemMayBeConfigStruct(typeOfSliceElem[string]())
		case "[]int":
			got = sliceElemMayBeConfigStruct(typeOfSliceElem[int]())
		}
		if got != c.want {
			t.Fatalf("%s 的元素判定不符: got %v want %v", c.typ, got, c.want)
		}
	}
}

// typeOfSliceElem 返回 []T 的元素类型，用于元素判定矩阵的断言
// 函数名：typeOfSliceElem
// 类型参数：T — 切片元素类型
// 参数：无
// 返回值：reflect.Type — []T 的元素类型（即 T）
// 异常：不触发panic
// 使用示例：
//
//	typ := typeOfSliceElem[SubDefaultConfig]()
func typeOfSliceElem[T any]() reflect.Type {
	return reflect.TypeOf([]T{}).Elem()
}

// ---- 环保护 ----

// SliceSelfRefNode 切片与指针交织的自引用节点，用于验证切片路径的环保护
type SliceSelfRefNode struct {
	Name     string              `json:"name" default:"node"`
	Children []*SliceSelfRefNode `json:"children"`
}

// TestCheckAndSetDefault_SliceSelfReference 验证切片中的自引用环不会无限递归
// 函数名：TestCheckAndSetDefault_SliceSelfReference
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：链路为切片元素指针 → 该指针的 Children 切片 → 指回自身，
// 由 visited 指针地址路径集合收敛；新增切片分派不得破坏该保护
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_SliceSelfReference
func TestCheckAndSetDefault_SliceSelfReference(t *testing.T) {
	type Cfg struct {
		Nodes []*SliceSelfRefNode `json:"nodes"`
	}

	root := &SliceSelfRefNode{}
	child := &SliceSelfRefNode{Name: "child"}
	// 构造 root -> child -> root 的环
	root.Children = []*SliceSelfRefNode{child}
	child.Children = []*SliceSelfRefNode{root}

	cfg := &Cfg{Nodes: []*SliceSelfRefNode{root}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Nodes[0].Name != "node" {
		t.Fatalf("自引用节点默认值未补充: %+v", cfg.Nodes[0])
	}
	if cfg.Nodes[0].Children[0].Name != "child" {
		t.Fatalf("自引用子节点既有值不应被覆盖: %+v", cfg.Nodes[0].Children[0])
	}
}

// buildNestedSliceChain 构造「切片元素为 interface 承载的内层切片」的交替链路，用于切片深度边界测试
// 函数名：buildNestedSliceChain
// 参数：depth int — 嵌套层数
// 返回值：[]interface{} — 外层切片，最内层为 SubDefaultConfig
// 异常：不触发panic
// 使用示例：
//
//	chain := buildNestedSliceChain(30)
func buildNestedSliceChain(depth int) []interface{} {
	var node interface{} = SubDefaultConfig{}
	for i := 0; i < depth; i++ {
		node = []interface{}{node}
	}
	return []interface{}{node}
}

// digNestedSliceChain 沿 buildNestedSliceChain 的结构下钻，取回最内层的 SubDefaultConfig
// 函数名：digNestedSliceChain
// 参数：
// - data []interface{} — buildNestedSliceChain 产出的切片
// - depth int — 嵌套层数，需与构造时一致
// 返回值：
// - SubDefaultConfig — 最内层结构体
// - bool — 下钻是否成功
// 异常：不触发panic
// 使用示例：
//
//	sub, ok := digNestedSliceChain(chain, 30)
func digNestedSliceChain(data []interface{}, depth int) (SubDefaultConfig, bool) {
	var cur interface{} = data[0]
	for i := 0; i < depth; i++ {
		s, ok := cur.([]interface{})
		if !ok || len(s) == 0 {
			return SubDefaultConfig{}, false
		}
		cur = s[0]
	}
	sub, ok := cur.(SubDefaultConfig)
	return sub, ok
}

// TestCheckAndSetDefault_SliceDepthBoundary 验证切片路径的深度上限精确边界
// 函数名：TestCheckAndSetDefault_SliceDepthBoundary
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：
//   - 切片通过 interface 装下自身构成不含指针的环，只能依靠 maxStructDefaultDepth 兜底，
//     故必须在 setDefaultForSliceField 入口校验深度，否则该环无法收敛
//   - 深度语义与 map 路径统一为「容器嵌套层数」：切片元素按 depth+1 处理，
//     值结构体元素再消耗一层，故最后一层可用深度同样为 max-2
//   - 上限内必须正常填充，超限必须静默截断（不 panic、不栈溢出），两侧都要断言
//
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_SliceDepthBoundary
func TestCheckAndSetDefault_SliceDepthBoundary(t *testing.T) {
	type Cfg struct {
		Data []interface{} `json:"data" default:""`
	}

	// 明显在上限内
	for _, depth := range []int{1, 8, 16} {
		cfg := &Cfg{Data: buildNestedSliceChain(depth)}
		if err := CheckAndSetDefault(cfg); err != nil {
			t.Fatalf("depth=%d unexpected error: %v", depth, err)
		}
		sub, ok := digNestedSliceChain(cfg.Data, depth)
		if !ok {
			t.Fatalf("depth=%d 下钻失败，结构被改变", depth)
		}
		if sub.Host != "127.0.0.1" {
			t.Fatalf("depth=%d 应在上限内被补充默认值: %+v", depth, sub)
		}
	}

	// 明显超出上限：必须被截断且不 panic（同时验证无栈溢出）
	for _, depth := range []int{40, 80, 200} {
		cfg := &Cfg{Data: buildNestedSliceChain(depth)}
		if err := CheckAndSetDefault(cfg); err != nil {
			t.Fatalf("depth=%d unexpected error: %v", depth, err)
		}
		sub, ok := digNestedSliceChain(cfg.Data, depth)
		if !ok {
			t.Fatalf("depth=%d 下钻失败，结构被改变", depth)
		}
		if sub.Host != "" {
			t.Fatalf("depth=%d 超过深度上限后不应再补充默认值: %+v", depth, sub)
		}
	}

	// 紧邻上限的相邻两层必须一升一降（与 map 路径落在同一条边界上）
	lastOK := maxStructDefaultDepth - 2
	firstOver := maxStructDefaultDepth - 1
	for _, depth := range []int{lastOK, firstOver} {
		want := ""
		if depth == lastOK {
			want = "127.0.0.1"
		}
		cfg := &Cfg{Data: buildNestedSliceChain(depth)}
		if err := CheckAndSetDefault(cfg); err != nil {
			t.Fatalf("depth=%d unexpected error: %v", depth, err)
		}
		sub, ok := digNestedSliceChain(cfg.Data, depth)
		if !ok {
			t.Fatalf("depth=%d 下钻失败，结构被改变", depth)
		}
		if sub.Host != want {
			t.Fatalf("depth=%d 边界不符: got %q want %q", depth, sub.Host, want)
		}
	}
}

// ---- 判定矩阵：切片宿主类型识别 ----

// SliceFieldCfg 仅含无标签的 []SubConfig 字段，应随切片能力开放而判定为配置结构体
type SliceFieldCfg struct {
	Items []SubDefaultConfig
}

// SlicePtrFieldCfg 仅含无标签的 []*SubConfig 字段，判定同 SliceFieldCfg
type SlicePtrFieldCfg struct {
	Items []*SubDefaultConfig
}

// SlicePlainFieldCfg 仅含无标签的基础元素切片，不应被判为配置结构体
type SlicePlainFieldCfg struct {
	Items []string
}

// TestIsConfigStructType_SliceFields 验证切片字段参与宿主类型判定
// 函数名：TestIsConfigStructType_SliceFields
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：切片能力开放后，宿主结构体即使只有无标签的 []SubConfig 字段，
// 也必须被判为配置结构体 —— 否则上层类型判定会在进入递归前就把整段切片能力挡掉；
// 同时基础元素切片不得放宽判定（保持判定收紧的既有收益）
// 使用示例：go test ./component/helper/ -run TestIsConfigStructType_SliceFields
func TestIsConfigStructType_SliceFields(t *testing.T) {
	if !isConfigStructType(reflectTypeOf[SliceFieldCfg]()) {
		t.Fatal("仅含 []SubDefaultConfig 字段的宿主应被判定为配置结构体")
	}
	if !isConfigStructType(reflectTypeOf[SlicePtrFieldCfg]()) {
		t.Fatal("仅含 []*SubDefaultConfig 字段的宿主应被判定为配置结构体")
	}
	if isConfigStructType(reflectTypeOf[SlicePlainFieldCfg]()) {
		t.Fatal("仅含 []string 字段的宿主不应被判定为配置结构体")
	}
}

// reflectTypeOf 返回类型参数 T 的反射类型，用于类型判定断言
// 函数名：reflectTypeOf
// 类型参数：T — 目标类型
// 参数：无
// 返回值：reflect.Type — T 的反射类型
// 异常：不触发panic
// 使用示例：
//
//	typ := reflectTypeOf[SubDefaultConfig]()
func reflectTypeOf[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

// TestCheckAndSetDefault_SliceHostWithoutTag 验证仅含切片字段的宿主也能端到端生效
// 函数名：TestCheckAndSetDefault_SliceHostWithoutTag
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：本用例是「判定矩阵」到「实际填充」的端到端串联：
// 宿主自身无 default 标签，仅靠切片元素携带标签进入递归
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_SliceHostWithoutTag
func TestCheckAndSetDefault_SliceHostWithoutTag(t *testing.T) {
	type Cfg struct {
		Items []SubDefaultConfig `json:"items"`
	}

	// 顶层直接调用不经过类型判定，切片元素应被补齐
	cfg := &Cfg{Items: []SubDefaultConfig{{}}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Items[0].Host != "127.0.0.1" || cfg.Items[0].Port != 3306 {
		t.Fatalf("顶层调用下的切片元素未补充默认值: %+v", cfg.Items[0])
	}

	// 作为嵌套字段时依赖宿主类型判定，同样应被补齐
	type Outer struct {
		Cfg SliceFieldCfg `json:"cfg"`
	}
	outer := &Outer{Cfg: SliceFieldCfg{Items: []SubDefaultConfig{{}}}}
	if err := CheckAndSetDefault(outer); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outer.Cfg.Items[0].Host != "127.0.0.1" || outer.Cfg.Items[0].Port != 3306 {
		t.Fatalf("嵌套宿主下的切片元素未补充默认值: %+v", outer.Cfg.Items[0])
	}
}
