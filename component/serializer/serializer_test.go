package serializer

import (
	"reflect"
	"testing"
	"time"
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

// TestProcessMaxDepth 验证超过 MaxDepth 的嵌套结构体会被截断，
// 不会导致栈溢出，且截断位置返回原始数据以避免子树丢失。
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
