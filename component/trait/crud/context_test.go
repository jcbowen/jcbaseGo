package crud

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/jcbowen/jcbaseGo/component/serializer"
)

type legacyCustomData struct {
	Value string `json:"value"`
}

// MarshalJSON 使 legacyCustomData 在 json 序列化时输出带前缀的值。
func (d legacyCustomData) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"value": "legacy:" + d.Value})
}

func TestResultDefaultLegacyMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	ctx := NewContext(&NewContextOpt{GinContext: c})
	ctx.Result(200, "ok", legacyCustomData{Value: "hello"})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data should be map, got %T", body["data"])
	}
	// 兼容模式尊重 MarshalJSON，应输出 legacy:hello
	if data["value"] != "legacy:hello" {
		t.Errorf("value = %v, want legacy:hello", data["value"])
	}
}

func TestResultWithSerializerOptsUsesNewMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	ctx := NewContext(&NewContextOpt{GinContext: c})
	ctx.SerializerOpts = []serializer.Option{serializer.WithLegacyMode(false)}

	ctx.Result(200, "ok", legacyCustomData{Value: "hello"})

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data should be map, got %T", body["data"])
	}
	// 新序列化器不调用 MarshalJSON，应输出原始值 hello
	if data["value"] != "hello" {
		t.Errorf("value = %v, want hello", data["value"])
	}
}

// TestResultZeroValueDataPreserved 锁定 Result 对「显式传入的零值数据」如实输出。
//
// 既有行为（已修正）：Result 用 helper.IsEmptyValue 判空，全零 struct 会被
// 递归判定为零值而吞成 data:{}，破坏前端类型契约（detail 场景拿不到结构体形状）。
// 新语义：只有 nil / nil 指针视为未提供数据；零值 struct、false、""、空 slice
// 都是显式传入的数据，如实序列化输出。
func TestResultZeroValueDataPreserved(t *testing.T) {
	type zeroRecord struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	tests := []struct {
		name    string
		data    any
		wantData any // 期望的 data 字段值（JSON 反序列化后）
	}{
		{"全零struct保留字段形状", zeroRecord{}, map[string]interface{}{"id": float64(0), "name": ""}},
		{"bool false放行", false, false},
		{"空字符串放行", "", ""},
		{"空slice输出数组", []any{}, []interface{}{}},
		{"空map输出对象", map[string]any{}, map[string]interface{}{}},
		{"未传数据输出空对象", nil, map[string]interface{}{}},
		{"nil指针视为未提供", (*zeroRecord)(nil), map[string]interface{}{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			ctx := NewContext(&NewContextOpt{GinContext: c})
			if tt.data == nil {
				// 未传数据的场景：Result 只带 code/msg
				ctx.Result(200, "ok")
			} else {
				ctx.Result(200, "ok", tt.data)
			}

			var body map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("unmarshal body: %v", err)
			}

			// 用 JSON 再往返一次做深度比较，规避类型断言的繁琐
			want, _ := json.Marshal(tt.wantData)
			got, _ := json.Marshal(body["data"])
			if string(got) != string(want) {
				t.Errorf("data = %s, want %s", got, want)
			}
		})
	}
}

// TestIsEmptyForResponse 直接覆盖判空函数的分支语义。
func TestIsEmptyForResponse(t *testing.T) {
	type s struct{ A int }

	cases := []struct {
		name string
		in   any
		want bool
	}{
		{"nil", nil, true},
		{"nil指针", (*s)(nil), true},
		{"零值struct", s{}, false},
		{"有值struct", s{A: 1}, false},
		{"false", false, false},
		{"零", 0, false},
		{"空串", "", false},
		{"空map", map[string]any{}, false},
		{"空slice", []any{}, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isEmptyForResponse(c.in); got != c.want {
				t.Errorf("isEmptyForResponse(%v) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

// TestNewContextWiresSerializerOpts 锁定 NewContext 必须透传 SerializerOpts。
//
// 历史缺陷：构造器仅复制 ActionName/Debug/GinContext，调用方注入的
// ID 加密 / 时间格式化策略被丢弃，Result 永远走 LegacyMode。
// 本用例通过 NewContextOpt（即 Context 别名）注入 opts，断言新序列化路径生效。
func TestNewContextWiresSerializerOpts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	ctx := NewContext(&NewContextOpt{
		GinContext:     c,
		SerializerOpts: []serializer.Option{serializer.WithLegacyMode(false)},
	})
	if len(ctx.SerializerOpts) == 0 {
		t.Fatal("NewContext 丢弃了 SerializerOpts——ID 加密/时间格式化将全部失效")
	}

	ctx.Result(200, "ok", legacyCustomData{Value: "hello"})

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	data := body["data"].(map[string]interface{})
	// 新序列化器不调用 MarshalJSON；若 opts 被丢弃则走 legacy 输出 legacy:hello
	if data["value"] != "hello" {
		t.Errorf("value = %v, want hello（SerializerOpts 未透传，仍走 legacy）", data["value"])
	}
}
