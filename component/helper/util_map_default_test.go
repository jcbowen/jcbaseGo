package helper

import (
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// SubDefaultConfig 供 map 默认值测试使用的子配置结构体
type SubDefaultConfig struct {
	Host    string        `json:"host" default:"127.0.0.1"`
	Port    int           `json:"port" default:"3306"`
	Timeout time.Duration `json:"timeout" default:"500ms"`
}

// SelfRefNode 自引用结构体，用于验证递归时的环检测能力
type SelfRefNode struct {
	Name     string                  `json:"name" default:"node"`
	Children map[string]*SelfRefNode `json:"children"`
}

// SelfRefHolder 通过值类型结构体字段参与递归环，map 元素指回 SelfRefWrapper
type SelfRefHolder struct {
	Ref map[string]*SelfRefWrapper `json:"ref"`
}

// SelfRefWrapper 值类型结构体字段 + 自引用 map 的组合，构成含值字段的递归环
type SelfRefWrapper struct {
	Name  string        `json:"name" default:"wrapper"`
	Inner SelfRefHolder `json:"inner"`
}

// TestCheckAndSetDefault_MapStructValue 测试 map 值为结构体时的默认值递归补充
// 函数名：TestCheckAndSetDefault_MapStructValue
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_MapStructValue
func TestCheckAndSetDefault_MapStructValue(t *testing.T) {
	type Cfg struct {
		Clusters map[string]SubDefaultConfig `json:"clusters"`
	}

	cfg := &Cfg{Clusters: map[string]SubDefaultConfig{
		"main":   {},
		"backup": {Host: "192.168.1.10", Port: 3307},
	}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	main := cfg.Clusters["main"]
	if main.Host != "127.0.0.1" || main.Port != 3306 || main.Timeout != 500*time.Millisecond {
		t.Fatalf("map struct element defaults failed: %+v", main)
	}

	backup := cfg.Clusters["backup"]
	if backup.Host != "192.168.1.10" || backup.Port != 3307 || backup.Timeout != 500*time.Millisecond {
		t.Fatalf("map struct element should keep existed values: %+v", backup)
	}
}

// TestCheckAndSetDefault_MapPtrStructValue 测试 map 值为结构体指针时的默认值递归补充
// 函数名：TestCheckAndSetDefault_MapPtrStructValue
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_MapPtrStructValue
func TestCheckAndSetDefault_MapPtrStructValue(t *testing.T) {
	type Cfg struct {
		Clusters map[string]*SubDefaultConfig `json:"clusters"`
	}

	nilItem := (*SubDefaultConfig)(nil)
	cfg := &Cfg{Clusters: map[string]*SubDefaultConfig{
		"main":   {},
		"custom": {Host: "10.0.0.1"},
		"empty":  nilItem,
	}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Clusters["main"].Host != "127.0.0.1" || cfg.Clusters["main"].Port != 3306 {
		t.Fatalf("map ptr element defaults failed: %+v", cfg.Clusters["main"])
	}
	if cfg.Clusters["custom"].Host != "10.0.0.1" || cfg.Clusters["custom"].Port != 3306 {
		t.Fatalf("map ptr element should keep existed values: %+v", cfg.Clusters["custom"])
	}
	// nil 指针不应当被自动实例化
	if cfg.Clusters["empty"] != nil {
		t.Fatalf("nil map ptr element should stay nil")
	}
}

// TestCheckAndSetDefault_MapInterfaceValue 测试 map 值为 interface{} 且运行时为结构体时的默认值补充
// 函数名：TestCheckAndSetDefault_MapInterfaceValue
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_MapInterfaceValue
func TestCheckAndSetDefault_MapInterfaceValue(t *testing.T) {
	type Cfg struct {
		Extras map[string]interface{} `json:"extras"`
	}

	cfg := &Cfg{Extras: map[string]interface{}{
		"db":    SubDefaultConfig{},
		"dbPtr": &SubDefaultConfig{},
		"text":  "plain",
		"num":   10,
	}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	db, ok := cfg.Extras["db"].(SubDefaultConfig)
	if !ok {
		t.Fatalf("map interface struct element type changed: %T", cfg.Extras["db"])
	}
	if db.Host != "127.0.0.1" || db.Port != 3306 {
		t.Fatalf("map interface struct defaults failed: %+v", db)
	}

	dbPtr, ok := cfg.Extras["dbPtr"].(*SubDefaultConfig)
	if !ok {
		t.Fatalf("map interface ptr element type changed: %T", cfg.Extras["dbPtr"])
	}
	if dbPtr.Host != "127.0.0.1" || dbPtr.Port != 3306 {
		t.Fatalf("map interface ptr defaults failed: %+v", dbPtr)
	}

	if cfg.Extras["text"] != "plain" || cfg.Extras["num"] != 10 {
		t.Fatalf("non struct map interface elements should not be changed: %+v", cfg.Extras)
	}
}

// TestCheckAndSetDefault_MapNestedMap 测试嵌套 map 中的结构体默认值补充
// 函数名：TestCheckAndSetDefault_MapNestedMap
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_MapNestedMap
func TestCheckAndSetDefault_MapNestedMap(t *testing.T) {
	type Cfg struct {
		Groups map[string]map[string]SubDefaultConfig `json:"groups"`
	}

	cfg := &Cfg{Groups: map[string]map[string]SubDefaultConfig{
		"shard": {"node1": {}},
	}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	node := cfg.Groups["shard"]["node1"]
	if node.Host != "127.0.0.1" || node.Port != 3306 {
		t.Fatalf("nested map struct defaults failed: %+v", node)
	}
}

// TestCheckAndSetDefault_MapSelfReference 测试自引用结构体不会导致无限递归
// 函数名：TestCheckAndSetDefault_MapSelfReference
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_MapSelfReference
func TestCheckAndSetDefault_MapSelfReference(t *testing.T) {
	type Cfg struct {
		Nodes map[string]*SelfRefNode `json:"nodes"`
	}

	root := &SelfRefNode{}
	child := &SelfRefNode{Name: "child"}
	// 构造 root -> child -> root 的环
	root.Children = map[string]*SelfRefNode{"child": child}
	child.Children = map[string]*SelfRefNode{"root": root}

	cfg := &Cfg{Nodes: map[string]*SelfRefNode{"root": root}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Nodes["root"].Name != "node" {
		t.Fatalf("self reference node defaults failed: %+v", cfg.Nodes["root"])
	}
	if cfg.Nodes["root"].Children["child"].Name != "child" {
		t.Fatalf("self reference child should keep existed value: %+v", cfg.Nodes["root"].Children["child"])
	}
}

// TestCheckAndSetDefault_SelfReferenceViaStructField 测试递归环中包含值类型结构体字段时不会栈溢出
// 函数名：TestCheckAndSetDefault_SelfReferenceViaStructField
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：递归链为 wrapper → 值字段 Inner → map[string]*SelfRefWrapper → wrapper 自身；
//
//	若 struct 字段递归时未传递 visited/depth，则该环会无限递归并触发栈溢出
//
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_SelfReferenceViaStructField
func TestCheckAndSetDefault_SelfReferenceViaStructField(t *testing.T) {
	type Cfg struct {
		Wrapper SelfRefWrapper `json:"wrapper"`
	}

	wrapper := &SelfRefWrapper{}
	wrapper.Inner.Ref = map[string]*SelfRefWrapper{"self": wrapper}

	cfg := &Cfg{Wrapper: *wrapper}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Wrapper.Name != "wrapper" {
		t.Fatalf("self reference wrapper defaults failed: %+v", cfg.Wrapper)
	}
}

// TestCheckAndSetDefault_MapSelfContainingInterface 测试不含指针的递归环（map 通过 interface 装下自身）可被深度上限兜底
// 函数名：TestCheckAndSetDefault_MapSelfContainingInterface
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：该环中没有指针可供 visited 登记，仅能依靠 maxStructDefaultDepth 终止
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_MapSelfContainingInterface
func TestCheckAndSetDefault_MapSelfContainingInterface(t *testing.T) {
	type Cfg struct {
		Data map[string]interface{} `json:"data"`
	}

	cfg := &Cfg{Data: map[string]interface{}{"a": 1}}
	// 构造无指针的自包含环
	cfg.Data["self"] = cfg.Data

	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Data["a"] != 1 {
		t.Fatalf("self containing map should keep existed values: %+v", cfg.Data)
	}
}

// TestCheckAndSetDefault_MapValueNotStruct 测试非结构体值的 map 保持原有处理逻辑
// 函数名：TestCheckAndSetDefault_MapValueNotStruct
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_MapValueNotStruct
func TestCheckAndSetDefault_MapValueNotStruct(t *testing.T) {
	type Cfg struct {
		Labels   map[string]string    `json:"labels" default:"{\"env\":\"dev\"}"`
		Weights  map[string]int       `json:"weights"`
		Times    map[string]time.Time `json:"times"`
		SetLabel map[string]string    `json:"set_label"`
	}

	cfg := &Cfg{SetLabel: map[string]string{"a": "b"}}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Labels == nil || cfg.Labels["env"] != "dev" {
		t.Fatalf("map default tag should still work: %+v", cfg.Labels)
	}
	// 未配置标签的 nil map 仍按原逻辑初始化为空 map
	if cfg.Weights == nil || len(cfg.Weights) != 0 {
		t.Fatalf("nil map should be initialized as empty map: %+v", cfg.Weights)
	}
	if cfg.Times == nil {
		t.Fatalf("time map should be initialized as empty map: %+v", cfg.Times)
	}
	if len(cfg.SetLabel) != 1 || cfg.SetLabel["a"] != "b" {
		t.Fatalf("non empty map should not be changed: %+v", cfg.SetLabel)
	}
}

// TestCheckAndSetDefault_MapStructInsideStruct 测试结构体中的 map 结构体值沿嵌套结构体链递归补充
// 函数名：TestCheckAndSetDefault_MapStructInsideStruct
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：判定条件收紧为「仅 default 标签」后，无标签的中间结构体不再进入递归，
// 因此其内部 nil map 不再被自动初始化为空 map。需要初始化的场景请显式加 default 标签
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_MapStructInsideStruct
func TestCheckAndSetDefault_MapStructInsideStruct(t *testing.T) {
	type Group struct {
		Members map[string]SubDefaultConfig `json:"members"`
	}
	type Cfg struct {
		Group Group `json:"group"`
	}

	// 无 default 标签的中间结构体不参与递归，其 nil map 保持 nil（不再被自动初始化）
	cfg := &Cfg{}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Group.Members != nil {
		t.Fatalf("无标签的中间结构体不应被递归处理: %+v", cfg.Group.Members)
	}

	// 显式声明 default 标签后，中间结构体进入递归，nil map 按标签初始化为空 map
	type TaggedGroup struct {
		Members map[string]SubDefaultConfig `json:"members" default:""`
	}
	type TaggedCfg struct {
		Group TaggedGroup `json:"group"`
	}
	cfgTagged := &TaggedCfg{}
	if err := CheckAndSetDefault(cfgTagged); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfgTagged.Group.Members == nil || len(cfgTagged.Group.Members) != 0 {
		t.Fatalf("带 default 标签的中间结构体应初始化为空 map: %+v", cfgTagged.Group.Members)
	}

	// 非空 map 的元素按 default 标签递归补充：中间结构体需带标签才会进入递归
	type SubGroup struct {
		Members map[string]SubDefaultConfig `json:"members"`
		Name    string                      `json:"name" default:"sub"`
	}
	type SubCfg struct {
		Group SubGroup `json:"group"`
	}
	cfg2 := &SubCfg{Group: SubGroup{Members: map[string]SubDefaultConfig{"m1": {}}}}
	if err := CheckAndSetDefault(cfg2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg2.Group.Members["m1"].Host != "127.0.0.1" || cfg2.Group.Members["m1"].Port != 3306 {
		t.Fatalf("nested struct map defaults failed: %+v", cfg2.Group.Members["m1"])
	}
}

// ---- 配置结构体类型判定矩阵使用的辅助类型（以下类型均不会真正被实例化使用）----

// PredicateTagOnly 仅含带 default 标签字段，应判定为配置结构体
type PredicateTagOnly struct {
	A string `default:"a"`
}

// PredicateSliceOnly 仅含无标签切片字段，收紧判定后不再算配置结构体
// 理由：切片无标签时的「初始化为空切片」只作用于字段自身，递归不会带来额外收益
type PredicateSliceOnly struct {
	S []string
}

// PredicateMapOnly 仅含无标签映射字段，收紧判定后不再算配置结构体（理由同 PredicateSliceOnly）
type PredicateMapOnly struct {
	M map[string]string
}

// PredicateIfaceOnly 仅含无标签 interface 字段，收紧判定后不再算配置结构体
// 理由：真实配置结构体的宿主必然带有 default 标签，判定自会通过；
// 若宿主毫无标签，递归也无任何字段可填充，属行为中性
type PredicateIfaceOnly struct {
	I interface{}
}

// PredicateNestedTag 标签位于内层结构体中，应沿嵌套向内判定为配置结构体
type PredicateNestedTag struct {
	In PredicateTagOnly
}

// PredicateRecursiveWithTag 自引用且带标签，应在环保护下判定为配置结构体
type PredicateRecursiveWithTag struct {
	Name string `default:"n"`
	Next *PredicateRecursiveWithTag
}

// PredicateScalarNoTag 仅含无标签基础字段，递归也不会有任何改写，应判定为非配置结构体
type PredicateScalarNoTag struct {
	A string
	B int
}

// PredicatePtrScalarNoTag 仅含基础类型指针字段，应判定为非配置结构体
type PredicatePtrScalarNoTag struct {
	P *int
}

// PredicateCycleA / PredicateCycleB 互相引用且均无相关字段，用于验证类型判定的环保护不会栈溢出
type PredicateCycleA struct {
	B *PredicateCycleB
}

type PredicateCycleB struct {
	A *PredicateCycleA
}

// TestIsConfigStructType_ConfigStructs 验证各类配置结构体都能被正确识别
// 函数名：TestIsConfigStructType_ConfigStructs
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：判定标准为「递归范围内存在带 default 标签的可导出字段」，
// 即带 default 标签的字段，或包含此类字段的内层结构体 / 结构体指针
// 使用示例：go test ./component/helper/ -run TestIsConfigStructType_ConfigStructs
func TestIsConfigStructType_ConfigStructs(t *testing.T) {
	cases := []reflect.Type{
		reflect.TypeOf(SubDefaultConfig{}),
		reflect.TypeOf(NestedMapCfg{}),
		reflect.TypeOf(SelfRefNode{}),
		reflect.TypeOf(PredicateTagOnly{}),
		reflect.TypeOf(PredicateNestedTag{}),
		reflect.TypeOf(PredicateRecursiveWithTag{}),
	}
	for _, typ := range cases {
		if !isConfigStructType(typ) {
			t.Fatalf("%s 应被判定为配置结构体", typ)
		}
	}
}

// TestIsConfigStructType_NoTagStructs 验证「无 default 标签」的结构体不再被判定为配置结构体
// 函数名：TestIsConfigStructType_NoTagStructs
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：判定收紧后，仅靠 slice / map / interface 字段不再能通过判定。
// 这可避免把 http.Request 这类第三方类型误判为配置结构体，
// 从而杜绝递归进入第三方对象、把其 nil map / nil slice 静默改写为空容器
// 使用示例：go test ./component/helper/ -run TestIsConfigStructType_NoTagStructs
func TestIsConfigStructType_NoTagStructs(t *testing.T) {
	cases := []reflect.Type{
		reflect.TypeOf(PredicateSliceOnly{}),
		reflect.TypeOf(PredicateMapOnly{}),
		reflect.TypeOf(PredicateIfaceOnly{}),
	}
	for _, typ := range cases {
		if isConfigStructType(typ) {
			t.Fatalf("%s 无 default 标签，不应被判定为配置结构体", typ)
		}
	}
}

// TestIsConfigStructType_ThirdPartyTypes 验证第三方类型不会被误判为配置结构体
// 函数名：TestIsConfigStructType_ThirdPartyTypes
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：这是「判定过宽」问题的核心回归保护。
// http.Request 含 Header（map）与 Body（interface）字段但均无 default 标签，
// 判定收紧前会被误判为配置结构体，导致递归进入后把其 nil Header 改写为 map[]
// 使用示例：go test ./component/helper/ -run TestIsConfigStructType_ThirdPartyTypes
func TestIsConfigStructType_ThirdPartyTypes(t *testing.T) {
	cases := []reflect.Type{
		reflect.TypeOf(http.Request{}),
		reflect.TypeOf(http.Response{}),
		reflect.TypeOf(url.URL{}),
		reflect.TypeOf(sync.Mutex{}),
		reflect.TypeOf(atomic.Value{}),
		reflect.TypeOf(time.Time{}),
	}
	for _, typ := range cases {
		if isConfigStructType(typ) {
			t.Fatalf("%s 不应被判定为配置结构体", typ)
		}
	}
}

// TestIsConfigStructType_NonConfigStructs 验证无标签的普通结构体与同步原语会被整体跳过
// 函数名：TestIsConfigStructType_NonConfigStructs
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：
// - 这些类型没有 default 标签，递归不可能改到任何内容；
// 跳过它们可避免对 map 值元素做无意义的「拷贝 → 递归 → SetMapIndex 写回」，
// 例如 map[string]sync.Mutex 不会被静默复制（反射路径绕过了 go vet 的 copylocks）
// - 判定收紧后，其判定依据不再是「有无 slice/map/interface 字段」，而是「有无 default 标签」，
// 故本测试的断言对收紧前后均成立（行为中性）
// 使用示例：go test ./component/helper/ -run TestIsConfigStructType_NonConfigStructs
func TestIsConfigStructType_NonConfigStructs(t *testing.T) {
	cases := []reflect.Type{
		reflect.TypeOf(sync.RWMutex{}),
		reflect.TypeOf(PredicateScalarNoTag{}),
		reflect.TypeOf(PredicatePtrScalarNoTag{}),
		reflect.TypeOf(PredicateCycleA{}),
	}
	for _, typ := range cases {
		if isConfigStructType(typ) {
			t.Fatalf("%s 不应被判定为配置结构体", typ)
		}
	}

	// 映射值类型判定：第三方类型整表跳过，配置结构体照旧处理
	if mapValueMayBeConfigStruct(reflect.TypeOf(map[string]sync.Mutex{}).Elem()) {
		t.Fatalf("map[string]sync.Mutex 不应进入递归")
	}
	if !mapValueMayBeConfigStruct(reflect.TypeOf(map[string]SubDefaultConfig{}).Elem()) {
		t.Fatalf("map[string]SubDefaultConfig 应进入递归")
	}
	// interface 的静态类型无法确定，仍需返回 true，交由运行时的 setDefaultForMapValue 判定
	if !mapValueMayBeConfigStruct(reflect.TypeOf(map[string]interface{}{}).Elem()) {
		t.Fatalf("map[string]interface{} 应返回 true")
	}
}

// ThirdPartyHolder 用于验证第三方对象不会被递归改写的宿主结构体
// 自身带 default 标签以确保判定通过，从而真正走到「第三方字段是否会被递归」这一步
type ThirdPartyHolder struct {
	Name string        `json:"name" default:"holder"`
	Req  *http.Request `json:"req"`
}

// TestCheckAndSetDefault_ThirdPartyNotMutated 验证第三方对象的 nil map 不会被静默改写
// 函数名：TestCheckAndSetDefault_ThirdPartyNotMutated
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：这是「判定过宽」问题的端到端回归保护。
// 判定收紧前，*http.Request 字段会被递归，其 nil Header 会被 setDefaultValue
// 的空标签分支改写为 map[]；收紧后该字段整体跳过，Header 保持 nil。
// 覆盖三条改写路径：结构体指针字段、map 值元素、interface 承载
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_ThirdPartyNotMutated
func TestCheckAndSetDefault_ThirdPartyNotMutated(t *testing.T) {
	// 路径一：结构体指针字段
	req := &http.Request{Method: "GET", URL: &url.URL{Scheme: "http", Host: "example.com"}}
	holder := &ThirdPartyHolder{Req: req}
	if err := CheckAndSetDefault(holder); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if holder.Name != "holder" {
		t.Fatalf("宿主结构体自身默认值应被补充: %+v", holder)
	}
	if req.Header != nil {
		t.Fatalf("第三方对象的 nil Header 不应被改写: %v", req.Header)
	}
	if req.Method != "GET" {
		t.Fatalf("第三方对象既有字段不应被改动: %+v", req.Method)
	}

	// 路径二：map 值元素
	type MapHolder struct {
		Name string                  `json:"name" default:"holder"`
		Reqs map[string]http.Request `json:"reqs"`
	}
	mapHolder := &MapHolder{Reqs: map[string]http.Request{"a": {Method: "GET"}}}
	if err := CheckAndSetDefault(mapHolder); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mapHolder.Reqs["a"].Header != nil {
		t.Fatalf("map 值第三方对象的 nil Header 不应被改写: %v", mapHolder.Reqs["a"].Header)
	}

	// 路径三：interface{} 承载
	type IfaceHolder struct {
		Name string      `json:"name" default:"holder"`
		Any  interface{} `json:"any"`
	}
	ifaceHolder := &IfaceHolder{Any: http.Request{Method: "GET"}}
	if err := CheckAndSetDefault(ifaceHolder); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := ifaceHolder.Any.(http.Request)
	if !ok {
		t.Fatalf("interface 承载类型被改变: %T", ifaceHolder.Any)
	}
	if got.Header != nil {
		t.Fatalf("interface 承载的第三方对象 nil Header 不应被改写: %v", got.Header)
	}
}

// TestCheckAndSetDefault_HostTagKeepsNestedRecursion 验证判定收紧后嵌套递归能力不受影响
// 函数名：TestCheckAndSetDefault_HostTagKeepsNestedRecursion
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：宿主结构体自带 default 标签（真实配置结构体必然如此）时，
// 其内部的 interface / map 字段递归全部照常工作 —— 收紧判定不会削减实际能力
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_HostTagKeepsNestedRecursion
func TestCheckAndSetDefault_HostTagKeepsNestedRecursion(t *testing.T) {
	type Holder struct {
		Name  string                      `json:"name" default:"holder"`
		Any   interface{}                 `json:"any"`
		Items map[string]SubDefaultConfig `json:"items"`
	}

	// interface 承载 map 仍被递归
	holder := &Holder{Any: map[string]SubDefaultConfig{"main": {}}}
	if err := CheckAndSetDefault(holder); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	anyItems, ok := holder.Any.(map[string]SubDefaultConfig)
	if !ok {
		t.Fatalf("interface 承载类型被改变: %T", holder.Any)
	}
	if anyItems["main"].Host != "127.0.0.1" {
		t.Fatalf("interface 承载的 map 元素应被递归补充: %+v", anyItems["main"])
	}

	// interface 承载结构体仍被递归
	holder2 := &Holder{Any: SubDefaultConfig{}}
	if err := CheckAndSetDefault(holder2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	anySub, ok := holder2.Any.(SubDefaultConfig)
	if !ok {
		t.Fatalf("interface 承载类型被改变: %T", holder2.Any)
	}
	if anySub.Host != "127.0.0.1" || anySub.Port != 3306 {
		t.Fatalf("interface 承载的结构体应被递归补充: %+v", anySub)
	}

	// map 字段元素仍被递归
	holder3 := &Holder{Items: map[string]SubDefaultConfig{"main": {}}}
	if err := CheckAndSetDefault(holder3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if holder3.Items["main"].Host != "127.0.0.1" || holder3.Items["main"].Port != 3306 {
		t.Fatalf("map 字段元素应被递归补充: %+v", holder3.Items["main"])
	}
}

// ---- time.Time 字段默认值（default 标签）测试 ----

// TestCheckAndSetDefault_TimeFieldDefault 验证 time.Time 字段的 default 标签可被正确填充
// 函数名：TestCheckAndSetDefault_TimeFieldDefault
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：time.Time 的 Kind 为 Struct，若不特判会被 isNestedStructKind 的 Struct 分支
// 拦截并转入「结构体向内递归」，导致 default 标签被静默忽略。
// 本用例覆盖常见时间字符串格式
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_TimeFieldDefault
func TestCheckAndSetDefault_TimeFieldDefault(t *testing.T) {
	type Cfg struct {
		StartAt  time.Time `json:"start_at" default:"2024-01-01 00:00:00"`
		DateOnly time.Time `json:"date_only" default:"2024-06-15"`
		RFC3339  time.Time `json:"rfc3339" default:"2024-03-01T10:20:30Z"`
	}

	cfg := &Cfg{}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !cfg.StartAt.Equal(wantStart) {
		t.Fatalf("StartAt 默认值不符: got %v want %v", cfg.StartAt, wantStart)
	}
	wantDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	if !cfg.DateOnly.Equal(wantDate) {
		t.Fatalf("DateOnly 默认值不符: got %v want %v", cfg.DateOnly, wantDate)
	}
	wantRFC := time.Date(2024, 3, 1, 10, 20, 30, 0, time.UTC)
	if !cfg.RFC3339.Equal(wantRFC) {
		t.Fatalf("RFC3339 默认值不符: got %v want %v", cfg.RFC3339, wantRFC)
	}
}

// TestCheckAndSetDefault_TimeFieldTimestamp 验证时间戳字符串可作为 time.Time 的默认值
// 函数名：TestCheckAndSetDefault_TimeFieldTimestamp
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：Convert.ToTime 对纯数字字符串会按长度识别为秒/毫秒/纳秒时间戳
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_TimeFieldTimestamp
func TestCheckAndSetDefault_TimeFieldTimestamp(t *testing.T) {
	type Cfg struct {
		SecAt time.Time `json:"sec_at" default:"1704067200"`
	}

	cfg := &Cfg{}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1704067200 秒 = 2024-01-01 00:00:00 UTC
	want := time.Unix(1704067200, 0).UTC()
	if !cfg.SecAt.Equal(want) {
		t.Fatalf("秒时间戳默认值不符: got %v want %v", cfg.SecAt, want)
	}
}

// TestCheckAndSetDefault_TimeFieldInvalidTag 验证非法默认值不会破坏 time.Time 字段
// 函数名：TestCheckAndSetDefault_TimeFieldInvalidTag
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：解析失败（含非法格式字符串）必须静默跳过、不报错、不修改字段，
// 与整型/浮点/time.Duration 的既有容错风格保持一致
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_TimeFieldInvalidTag
func TestCheckAndSetDefault_TimeFieldInvalidTag(t *testing.T) {
	type Cfg struct {
		BadAt time.Time `json:"bad_at" default:"not-a-time"`
	}

	cfg := &Cfg{}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.BadAt.IsZero() {
		t.Fatalf("非法默认值应静默跳过，字段保持零值: %v", cfg.BadAt)
	}
}

// TestCheckAndSetDefault_TimeFieldNotEmptyTag 验证空 default 标签不会改动 time.Time 字段
// 函数名：TestCheckAndSetDefault_TimeFieldNotEmptyTag
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_TimeFieldNotEmptyTag
func TestCheckAndSetDefault_TimeFieldNotEmptyTag(t *testing.T) {
	type Cfg struct {
		EmptyAt time.Time `json:"empty_at" default:""`
		NoTag   time.Time `json:"no_tag"`
	}

	cfg := &Cfg{}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.EmptyAt.IsZero() || !cfg.NoTag.IsZero() {
		t.Fatalf("空标签与无标签的 time.Time 字段都应保持零值: %+v", cfg)
	}
}

// TestCheckAndSetDefault_TimeFieldNonZeroNotOverwritten 验证非零 time.Time 字段不被覆盖
// 函数名：TestCheckAndSetDefault_TimeFieldNonZeroNotOverwritten
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 说明：与其它类型「零值才填」的语义保持一致，已有值不会被 default 标签覆盖
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_TimeFieldNonZeroNotOverwritten
func TestCheckAndSetDefault_TimeFieldNonZeroNotOverwritten(t *testing.T) {
	type Cfg struct {
		SetAt time.Time `json:"set_at" default:"2024-01-01 00:00:00"`
	}

	existed := time.Date(2020, 5, 5, 12, 0, 0, 0, time.UTC)
	cfg := &Cfg{SetAt: existed}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.SetAt.Equal(existed) {
		t.Fatalf("非零 time.Time 字段不应被覆盖: got %v want %v", cfg.SetAt, existed)
	}
}

// TestCheckAndSetDefault_TimeFieldNested 验证嵌套结构体与 map 元素中的 time.Time 也被填充
// 函数名：TestCheckAndSetDefault_TimeFieldNested
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（断言失败由 testing 框架处理）
// 使用示例：go test ./component/helper/ -run TestCheckAndSetDefault_TimeFieldNested
func TestCheckAndSetDefault_TimeFieldNested(t *testing.T) {
	type SubCfg struct {
		StartAt time.Time `json:"start_at" default:"2024-01-01 00:00:00"`
	}
	type Cfg struct {
		Inner    SubCfg            `json:"inner"`
		Items    map[string]SubCfg `json:"items"`
		Any      interface{}       `json:"any"`
		InnerPtr *SubCfg           `json:"inner_ptr"`
	}

	cfg := &Cfg{
		Items:    map[string]SubCfg{"main": {}},
		Any:      SubCfg{},
		InnerPtr: &SubCfg{},
	}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !cfg.Inner.StartAt.Equal(want) {
		t.Fatalf("嵌套结构体 time.Time 未填充: %v", cfg.Inner.StartAt)
	}
	if !cfg.Items["main"].StartAt.Equal(want) {
		t.Fatalf("map 元素 time.Time 未填充: %v", cfg.Items["main"].StartAt)
	}
	anySub, ok := cfg.Any.(SubCfg)
	if !ok {
		t.Fatalf("interface 承载类型被改变: %T", cfg.Any)
	}
	if !anySub.StartAt.Equal(want) {
		t.Fatalf("interface 承载结构体的 time.Time 未填充: %v", anySub.StartAt)
	}
	if !cfg.InnerPtr.StartAt.Equal(want) {
		t.Fatalf("指针字段指向结构体的 time.Time 未填充: %v", cfg.InnerPtr.StartAt)
	}
}
