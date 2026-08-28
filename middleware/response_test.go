package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

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

func doSuccess(t *testing.T, data interface{}, opts ...Option) map[string]interface{} {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	Success(c, data, opts...)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	return body
}

func TestSuccessErrorEnvelope(t *testing.T) {
	body := doSuccess(t, gin.H{"a": 1})
	if body["code"] != float64(200) {
		t.Errorf("code = %v, want 200", body["code"])
	}
	if body["message"] != "ok" {
		t.Errorf("message = %v, want ok", body["message"])
	}
	if body["data"] == nil {
		t.Error("data should not be nil")
	}
}

func TestSuccessTimeFormatting(t *testing.T) {
	isTime := func(k string) bool { return k == "created" }
	format := func(v interface{}) (string, bool) {
		if t, ok := v.(time.Time); ok && !t.IsZero() {
			return t.UTC().Format("2006-01-02 15:04:05"), true
		}
		return "", false
	}

	utc := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	body := doSuccess(t, gin.H{"created": utc, "name": "x"},
		WithTimeField(isTime), WithFormatTime(format))

	data := body["data"].(map[string]interface{})
	if data["created"] != "2024-01-01 00:00:00" {
		t.Errorf("created = %v, want formatted", data["created"])
	}
	if data["name"] != "x" {
		t.Errorf("name = %v, want x", data["name"])
	}
}

func TestSuccessIDEncryption(t *testing.T) {
	encrypt := func(v uint64) string { return "enc_" + string(rune('0'+v)) }
	isID := func(k string) bool { return k == "id" }

	body := doSuccess(t, gin.H{"id": uint64(5), "name": "x"},
		WithEncryptID(encrypt), WithIDField(isID))

	data := body["data"].(map[string]interface{})
	if data["id"] != "enc_5" {
		t.Errorf("id = %v, want enc_5", data["id"])
	}
	if data["name"] != "x" {
		t.Errorf("name = %v, want x", data["name"])
	}
}

func TestSuccessStructNestedAndSlice(t *testing.T) {
	encrypt := func(v uint64) string { return "E" }
	isID := func(k string) bool { return k == "id" }

	utc := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	data := testOuter{
		Name:  "outer",
		Inner: testInner{ID: 1, Created: utc},
		Items: []testInner{{ID: 2}, {ID: 3}},
		Ptr:   &testInner{ID: 4},
	}

	body := doSuccess(t, data, WithEncryptID(encrypt), WithIDField(isID))
	outer := body["data"].(map[string]interface{})

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

func TestSuccessSkipField(t *testing.T) {
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
	body := doSuccess(t, s{baseModel: baseModel{ID: 9}, Keep: "shown"}, WithSkipField(skip))
	data := body["data"].(map[string]interface{})
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

func TestSuccessNilData(t *testing.T) {
	body := doSuccess(t, nil)
	if body["data"] != nil {
		t.Errorf("data = %v, want nil (omitempty)", body["data"])
	}
}
