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
