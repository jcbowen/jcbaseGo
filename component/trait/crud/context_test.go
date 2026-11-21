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
