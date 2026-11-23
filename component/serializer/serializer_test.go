package serializer

import (
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type testInner struct {
	ID      uint64    `json:"id"`
	Created time.Time `json:"created"`
}

type testOuter struct {
	Name  string      `json:"name"`
	Inner testInner   `json:"inner"`
	Items []testInner `json:"items"`
	Ptr   *testInner  `json:"ptr"`
}

func TestProcessTimeFormatting(t *testing.T) {
	isTime := func(k string) bool { return k == "created" }
	format := func(v interface{}) (string, bool) {
		if t, ok := v.(time.Time); ok && !t.IsZero() {
			return t.UTC().Format("2006-01-02 15:04:05"), true
		}
		return "", false
	}

	utc := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	result := Process(map[string]interface{}{
		"created": utc,
		"name":    "x",
	}, WithTimeField(isTime), WithFormatTime(format))

	data, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if data["created"] != "2024-01-01 00:00:00" {
		t.Errorf("created = %v, want formatted", data["created"])
	}
	if data["name"] != "x" {
		t.Errorf("name = %v, want x", data["name"])
	}
}

func TestProcessIDEncryption(t *testing.T) {
	encrypt := func(v uint64) string { return "enc_" + string(rune('0'+v)) }
	isID := func(k string) bool { return k == "id" }

	result := Process(map[string]interface{}{
		"id":   uint64(5),
		"name": "x",
	}, WithEncryptID(encrypt), WithIDField(isID))

	data, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if data["id"] != "enc_5" {
		t.Errorf("id = %v, want enc_5", data["id"])
	}
	if data["name"] != "x" {
		t.Errorf("name = %v, want x", data["name"])
	}
}

// testSignedID 有符号整型命名类型（模拟业务侧 int32 主键）。
type testSignedID int32

// testNamedUint64 无符号整型命名类型（模拟 iMe utils.EncryptedID）。
type testNamedUint64 uint64

// testSignedStruct 有符号整型字段的结构体。
type testSignedStruct struct {
	ID   testSignedID  `json:"id"`
	Big  int64         `json:"big"`
	Raw  int32         `json:"raw"`
	Ptr  *testSignedID `json:"ptr"`
	Bad  int32         `json:"bad"`  // 负数不加密
	Keep int32         `json:"keep"` // 非 ID 字段不加密
}

func TestProcessIDEncryptionSignedIntegers(t *testing.T) {
	encrypt := func(v uint64) string { return "E" }
	isID := func(k string) bool { return k == "id" || k == "big" || k == "raw" || k == "named" || k == "ptr" }

	// map：内置 int32 / int64 / 命名 uint64 / 负数
	result := Process(map[string]interface{}{
		"id":    int32(7),
		"big":   int64(1 << 40),
		"named": testNamedUint64(9),
		"neg":   int32(-3),
		"other": int32(7),
	}, WithEncryptID(encrypt), WithIDField(isID))

	data, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if data["id"] != "E" {
		t.Errorf("id(int32) = %v, want E", data["id"])
	}
	if data["big"] != "E" {
		t.Errorf("big(int64) = %v, want E", data["big"])
	}
	if data["named"] != "E" {
		t.Errorf("named(uint64) = %v, want E", data["named"])
	}
	if data["neg"] != int32(-3) {
		t.Errorf("neg = %v, want -3（负数不加密）", data["neg"])
	}
	if data["other"] != int32(7) {
		t.Errorf("other = %v, want 7（非 ID 字段不加密）", data["other"])
	}

	// 结构体：命名 int32 字段、指针字段、嵌套值
	neg := testSignedID(-1)
	s := testSignedStruct{ID: 5, Big: 1, Raw: 6, Ptr: &neg, Bad: -2, Keep: 8}
	sr := Process(s, WithEncryptID(encrypt), WithIDField(isID))
	m, ok := sr.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", sr)
	}
	if m["id"] != "E" {
		t.Errorf("struct id = %v, want E", m["id"])
	}
	if m["big"] != "E" || m["raw"] != "E" {
		t.Errorf("big = %v, raw = %v, want E", m["big"], m["raw"])
	}
	if m["ptr"] == nil {
		t.Fatalf("ptr should not be nil")
	}
	if got := *(m["ptr"].(*testSignedID)); got != testSignedID(-1) {
		t.Errorf("ptr = %v, want -1（负数不加密，指针透传）", got)
	}
	if m["bad"] != int32(-2) {
		t.Errorf("bad = %v, want -2", m["bad"])
	}
	if m["keep"] != int32(8) {
		t.Errorf("keep = %v, want 8", m["keep"])
	}
}

func TestProcessStructNestedAndSlice(t *testing.T) {
	encrypt := func(v uint64) string { return "E" }
	isID := func(k string) bool { return k == "id" }

	utc := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	data := testOuter{
		Name:  "outer",
		Inner: testInner{ID: 1, Created: utc},
		Items: []testInner{{ID: 2}, {ID: 3}},
		Ptr:   &testInner{ID: 4},
	}

	result := Process(data, WithEncryptID(encrypt), WithIDField(isID))
	outer, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}

	if outer["name"] != "outer" {
		t.Errorf("name = %v", outer["name"])
	}
	inner := outer["inner"].(map[string]interface{})
	if inner["id"] != "E" {
		t.Errorf("inner.id = %v, want E", inner["id"])
	}
	items := outer["items"].([]interface{})
	if len(items) != 2 {
		t.Fatalf("items len = %d, want 2", len(items))
	}
	if items[0].(map[string]interface{})["id"] != "E" {
		t.Errorf("items[0].id = %v, want E", items[0])
	}
	ptr := outer["ptr"].(map[string]interface{})
	if ptr["id"] != "E" {
		t.Errorf("ptr.id = %v, want E", ptr["id"])
	}
}

func TestProcessSkipField(t *testing.T) {
	type baseModel struct {
		ID uint64 `json:"id"`
	}
	skip := func(f reflect.StructField) bool {
		return f.Name == "baseModel"
	}
	type s struct {
		baseModel
		Keep string `json:"keep"`
	}
	result := Process(s{baseModel: baseModel{ID: 9}, Keep: "shown"}, WithSkipField(skip))
	data, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if _, ok := data["baseModel"]; ok {
		t.Error("baseModel should be skipped")
	}
	if _, ok := data["id"]; ok {
		t.Error("id should be skipped as part of skipped base model")
	}
	if data["keep"] != "shown" {
		t.Errorf("keep = %v, want shown", data["keep"])
	}
}

func TestProcessNilData(t *testing.T) {
	if Process(nil) != nil {
		t.Error("Process(nil) should return nil")
	}
}

type selfRef struct {
	ID   uint64   `json:"id"`
	Next *selfRef `json:"next"`
}

// TestProcessSelfReference 验证自引用结构体不会导致栈溢出，
// 且循环引用位置被正确处理为 nil。
func TestProcessSelfReference(t *testing.T) {
	node := &selfRef{ID: 1}
	node.Next = node

	result := Process(node)

	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["id"] != uint64(1) {
		t.Errorf("id = %v, want 1", m["id"])
	}
	if m["next"] != nil {
		t.Errorf("next should be nil for circular reference, got %v", m["next"])
	}
}

type deepNode struct {
	Child *deepNode `json:"child"`
}

// TestProcessMapCycle 验证 map 循环引用在默认 MaxDepth 下不会栈溢出。
func TestProcessMapCycle(t *testing.T) {
	m := map[string]interface{}{"id": 1}
	m["self"] = m

	result := Process(m)

	root, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if root["id"] != 1 {
		t.Errorf("id = %v, want 1", root["id"])
	}
}

// TestProcessLegacyMode 验证兼容模式走 json.Marshal/Unmarshal 路径。
func TestProcessLegacyMode(t *testing.T) {
	utc := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	data := testOuter{
		Name:  "outer",
		Inner: testInner{ID: 1, Created: utc},
		Items: []testInner{{ID: 2}, {ID: 3}},
		Ptr:   &testInner{ID: 4},
	}

	// 兼容模式下仍使用 json 标签转换字段名
	result := Process(data, WithLegacyMode(true))
	outer, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if outer["name"] != "outer" {
		t.Errorf("name = %v, want outer", outer["name"])
	}

	// 内嵌结构体应被 json 序列化器展开为 map
	inner, ok := outer["inner"].(map[string]interface{})
	if !ok {
		t.Fatalf("inner should be map, got %T", outer["inner"])
	}
	if inner["id"] != float64(1) {
		t.Errorf("inner.id = %v, want 1", inner["id"])
	}

	// 嵌套在结构体中的切片，经过 json 往返后元素为 map
	items, ok := outer["items"].([]interface{})
	if !ok {
		t.Fatalf("items should be []interface{}, got %T", outer["items"])
	}
	if len(items) != 2 {
		t.Fatalf("items len = %d, want 2", len(items))
	}
	if _, ok := items[0].(map[string]interface{}); !ok {
		t.Errorf("items[0] type = %T, want map", items[0])
	}

	// 指针应被正确解引用
	ptr, ok := outer["ptr"].(map[string]interface{})
	if !ok {
		t.Fatalf("ptr should be map, got %T", outer["ptr"])
	}
	if ptr["id"] != float64(4) {
		t.Errorf("ptr.id = %v, want 4", ptr["id"])
	}
}

// TestProcessLegacyModeTopLevelSlice 验证兼容模式下顶层切片不递归转换元素。
func TestProcessLegacyModeTopLevelSlice(t *testing.T) {
	data := []testInner{{ID: 1}, {ID: 2}}

	result := Process(data, WithLegacyMode(true))
	items, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", result)
	}
	if len(items) != 2 {
		t.Fatalf("items len = %d, want 2", len(items))
	}
	// 旧 convertToInterfaceSlice 行为：保留原始结构体
	if _, ok := items[0].(testInner); !ok {
		t.Errorf("items[0] type = %T, want testInner", items[0])
	}
}

// TestProcessLegacyModeNilPointer 验证兼容模式对 nil 指针返回 nil 而不是 panic。
func TestProcessLegacyModeNilPointer(t *testing.T) {
	var ptr *testInner
	result := Process(ptr, WithLegacyMode(true))
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

// TestProcessLegacyModePrimitive 验证兼容模式对基础类型直接返回原值，不 panic。
func TestProcessLegacyModePrimitive(t *testing.T) {
	cases := []interface{}{1, 1.5, true, false, "hello"}
	for _, c := range cases {
		result := Process(c, WithLegacyMode(true))
		if result != c {
			t.Errorf("Process(%v) = %v, want %v", c, result, c)
		}
	}
}

// TestProcessLegacyModeMapPointer 验证兼容模式能正确处理 map 指针。
func TestProcessLegacyModeMapPointer(t *testing.T) {
	// *gin.H
	h := gin.H{"key": "value"}
	result := Process(&h, WithLegacyMode(true))
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["key"] != "value" {
		t.Errorf("key = %v, want value", m["key"])
	}

	// *map[string]interface{}
	raw := map[string]interface{}{"id": 1}
	result = Process(&raw, WithLegacyMode(true))
	m, ok = result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["id"] != 1 {
		t.Errorf("id = %v, want 1", m["id"])
	}

	// **gin.H
	ph := &h
	result = Process(&ph, WithLegacyMode(true))
	m, ok = result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["key"] != "value" {
		t.Errorf("key = %v, want value", m["key"])
	}
}

// TestProcessLegacyModeUnserializableStruct 验证兼容模式下结构体 json 失败时返回原值，不 panic。
func TestProcessLegacyModeUnserializableStruct(t *testing.T) {
	// 使用一个无法被 json 序列化的通道类型来触发 Marshal 错误
	unserializable := struct {
		Ch chan int `json:"ch"`
	}{Ch: make(chan int)}

	result := Process(unserializable, WithLegacyMode(true))
	if result != unserializable {
		t.Errorf("expected original data on marshal error, got %v", result)
	}
}

// TestProcessLegacyModeCustomMarshal 验证兼容模式尊重自定义 MarshalJSON。
func TestProcessLegacyModeCustomMarshal(t *testing.T) {
	type custom struct {
		Value string `json:"value"`
	}

	original := custom{Value: "hello"}

	result := Process(original, WithLegacyMode(true))
	data, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if data["value"] != "hello" {
		t.Errorf("value = %v, want hello", data["value"])
	}
}

// TestProcessLegacyModeIgnoresOptions 验证兼容模式下时间/ID 等选项不会生效。
func TestProcessLegacyModeIgnoresOptions(t *testing.T) {
	encrypt := func(v uint64) string { return "enc_" + string(rune('0'+v)) }
	isID := func(k string) bool { return k == "id" }
	isTime := func(k string) bool { return k == "created" }
	format := func(v interface{}) (string, bool) { return "formatted", true }

	utc := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	result := Process(testInner{ID: 5, Created: utc},
		WithLegacyMode(true),
		WithEncryptID(encrypt), WithIDField(isID),
		WithTimeField(isTime), WithFormatTime(format),
	)

	data, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if data["id"] != float64(5) {
		t.Errorf("id should not be encrypted in legacy mode, got %v", data["id"])
	}
	if data["created"] == "formatted" {
		t.Error("created should not be formatted in legacy mode")
	}
}

// TestProcessMaxDepth 验证超出 MaxDepth 的层级返回原始数据而不是 nil。
func TestProcessMaxDepth(t *testing.T) {
	root := &deepNode{}
	cur := root
	for i := 0; i < 10; i++ {
		cur.Child = &deepNode{}
		cur = cur.Child
	}

	// MaxDepth=3 时，第 3 层之后的 child 应返回原始 *deepNode，而不是 nil
	result := Process(root, WithMaxDepth(3))

	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map at root, got %T", result)
	}
	child1, ok := m["child"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected child map at depth 1, got %T", m["child"])
	}
	child2, ok := child1["child"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected child map at depth 2, got %T", child1["child"])
	}
	if child2["child"] == nil {
		t.Fatal("expected child beyond max depth to be original data, got nil")
	}
	if _, ok := child2["child"].(*deepNode); !ok {
		t.Errorf("expected child beyond max depth to be *deepNode, got %T", child2["child"])
	}
}

// ---------- 匿名嵌入 struct 扁平化（对齐 encoding/json 字段提升语义） ----------

// testEmbeddedGoods 模拟业务侧匿名嵌入的商品行（如 merchant 项目 model.MirrorGoods）。
type testEmbeddedGoods struct {
	Title string  `json:"title"`
	Price float64 `json:"price"`
	Stock int32   `json:"stock"`
}

// testEmbedOuter 顶层视图：匿名嵌入业务结构 + 直接声明字段（模拟 GoodsAuditListItem）。
type testEmbedOuter struct {
	testEmbeddedGoods

	MerchantName string `json:"merchant_name"`
	AuditStatus  int8   `json:"audit_status"`
}

// TestProcessAnonymousEmbeddedFlatten 验证匿名嵌入 struct 被扁平化展开，
// 不再以类型名为 key 输出嵌套对象。
func TestProcessAnonymousEmbeddedFlatten(t *testing.T) {
	data := testEmbedOuter{
		testEmbeddedGoods: testEmbeddedGoods{Title: "古树白茶", Price: 198, Stock: 200},
		MerchantName:      "平台自营店",
		AuditStatus:       1,
	}

	result := Process(data)
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}

	if _, exists := m["testEmbedOuter"]; exists {
		t.Error("outer type name key should not appear")
	}
	if _, exists := m["testEmbeddedGoods"]; exists {
		t.Error("embedded type name key should not appear (flattened)")
	}
	if m["title"] != "古树白茶" {
		t.Errorf("title = %v, want 古树白茶", m["title"])
	}
	if m["price"] != float64(198) {
		t.Errorf("price = %v, want 198", m["price"])
	}
	if m["stock"] != int32(200) {
		t.Errorf("stock = %v, want 200", m["stock"])
	}
	if m["merchant_name"] != "平台自营店" {
		t.Errorf("merchant_name = %v, want 平台自营店", m["merchant_name"])
	}
	if m["audit_status"] != int8(1) {
		t.Errorf("audit_status = %v, want 1", m["audit_status"])
	}
}

// testEmbedConflict 外层直接声明字段与嵌入字段同名（浅层优先语义）。
type testEmbedConflict struct {
	testEmbeddedGoods
	Title string `json:"title"`
}

// TestProcessAnonymousEmbeddedConflictOuterWins 验证同名字段冲突时外层优先，
// 与 encoding/json 浅层覆盖深层的行为一致。
func TestProcessAnonymousEmbeddedConflictOuterWins(t *testing.T) {
	data := testEmbedConflict{
		testEmbeddedGoods: testEmbeddedGoods{Title: "from-embedded", Price: 1},
		Title:             "from-outer",
	}

	result := Process(data)
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["title"] != "from-outer" {
		t.Errorf("title = %v, want from-outer（外层浅层优先）", m["title"])
	}
	if m["price"] != float64(1) {
		t.Errorf("price = %v, want 1（嵌入内非冲突字段正常提升）", m["price"])
	}
}

// testEmbedNilPtr nil 指针匿名嵌入：无可展开内容，输出中不应出现任何相关键。
type testEmbedNilPtr struct {
	*testEmbeddedGoods
	Keep string `json:"keep"`
}

// TestProcessAnonymousEmbeddedNilPointer 验证 nil 指针嵌入整体跳过（对齐 encoding/json）。
func TestProcessAnonymousEmbeddedNilPointer(t *testing.T) {
	result := Process(testEmbedNilPtr{Keep: "shown"})
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if _, exists := m["testEmbeddedGoods"]; exists {
		t.Error("nil pointer embedded type name key should not appear")
	}
	if _, exists := m["title"]; exists {
		t.Error("title should not appear for nil pointer embedded struct")
	}
	if m["keep"] != "shown" {
		t.Errorf("keep = %v, want shown", m["keep"])
	}
}

// testEmbedDash 带 json:"-" 的匿名嵌入：整体跳过。
type testEmbedDash struct {
	testEmbeddedGoods `json:"-"`
	Keep              string `json:"keep"`
}

// TestProcessAnonymousEmbeddedDashTag 验证 json:"-" 嵌入不展开也不输出。
func TestProcessAnonymousEmbeddedDashTag(t *testing.T) {
	result := Process(testEmbedDash{
		testEmbeddedGoods: testEmbeddedGoods{Title: "hidden"},
		Keep:              "shown",
	})
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if _, exists := m["title"]; exists {
		t.Error("title should not appear for dash-tagged embedded struct")
	}
	if _, exists := m["testEmbeddedGoods"]; exists {
		t.Error("embedded type name key should not appear")
	}
	if m["keep"] != "shown" {
		t.Errorf("keep = %v, want shown", m["keep"])
	}
}

// testEmbedL1/L2/L3 多级匿名嵌入链。
type testEmbedL1 struct {
	Deep string `json:"deep"`
}

type testEmbedL2 struct {
	testEmbedL1
}

type testEmbedL3 struct {
	testEmbedL2
	Keep string `json:"keep"`
}

// TestProcessAnonymousEmbeddedMultiLevel 验证多级匿名嵌入全部递归扁平化。
func TestProcessAnonymousEmbeddedMultiLevel(t *testing.T) {
	result := Process(testEmbedL3{
		testEmbedL2: testEmbedL2{testEmbedL1{Deep: "v"}},
		Keep:        "k",
	})
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	for _, name := range []string{"testEmbedL1", "testEmbedL2"} {
		if _, exists := m[name]; exists {
			t.Errorf("nested embedded type name key %q should not appear", name)
		}
	}
	if m["deep"] != "v" || m["keep"] != "k" {
		t.Errorf("deep = %v, keep = %v, want v/k", m["deep"], m["keep"])
	}
}

// testEmbedSliceElement 顶层切片中的匿名嵌入行。
type testEmbedSliceRow struct {
	testEmbeddedGoods
	Flag string `json:"flag"`
}

// TestProcessAnonymousEmbeddedInSlice 验证切片元素的匿名嵌入同样被扁平化。
func TestProcessAnonymousEmbeddedInSlice(t *testing.T) {
	result := Process([]testEmbedSliceRow{
		{testEmbeddedGoods: testEmbeddedGoods{Title: "a"}, Flag: "f1"},
		{testEmbeddedGoods: testEmbeddedGoods{Title: "b"}, Flag: "f2"},
	})
	items, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", result)
	}
	if len(items) != 2 {
		t.Fatalf("items len = %d, want 2", len(items))
	}
	first, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatalf("items[0] should be map, got %T", items[0])
	}
	if _, exists := first["testEmbeddedGoods"]; exists {
		t.Error("embedded type name key should not appear in slice element")
	}
	if first["title"] != "a" || first["flag"] != "f1" {
		t.Errorf("items[0] = %v, want title=a flag=f1", first)
	}
}
