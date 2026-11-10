package serializer

import (
	"encoding/json"
	"testing"
	"time"
)

// timeSerializeFixture 模拟时间字段形态：
// *time.Time+ nil *time.Time。
type timeSerializeFixture struct {
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	GraceUntil *time.Time `json:"grace_until,omitempty"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DaysLeft   int        `json:"days_left"`
}

// TestProcessTimePtrKeepsRFC3339 回归：非 nil *time.Time 必须输出 RFC3339 字符串，
// 而不是被反射递归折叠成 {}（2026-11-10 线上 expires_at:{} 根因）。
func TestProcessTimePtrKeepsRFC3339(t *testing.T) {
	at := time.Date(2026, 11, 10, 12, 30, 0, 0, time.Local)
	in := timeSerializeFixture{ExpiresAt: &at, UpdatedAt: at, DaysLeft: 30}

	out, err := json.Marshal(Process(in))
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}
	got := string(out)

	if want := `"expires_at":"2026-11-10T12:30:00+08:00"`; !contains(got, want) {
		t.Fatalf("*time.Time 应输出 RFC3339 字符串\n got: %s\nwant 含: %s", got, want)
	}
	if contains(got, `"expires_at":{}`) || contains(got, `"updated_at":{}`) {
		t.Fatalf("时间字段不得折叠为空对象: %s", got)
	}
}

// TestProcessNilTimePtrOutputsNull nil *time.Time 保持 null（买断 expires_at 语义）。
func TestProcessNilTimePtrOutputsNull(t *testing.T) {
	in := timeSerializeFixture{UpdatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
	out, _ := json.Marshal(Process(in))
	got := string(out)
	if want := `"grace_until":null`; !contains(got, want) {
		t.Fatalf("nil *time.Time 应输出 null: %s", got)
	}
	if want := `"expires_at":null`; !contains(got, want) {
		t.Fatalf("反射路径忽略 omitempty，nil 指针应显式输出 null（保持既有行为）: %s", got)
	}
}

// TestProcessTimeInMapAndSlice map/slice 容器内的时间字段同样透传。
func TestProcessTimeInMapAndSlice(t *testing.T) {
	at := time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC)
	in := map[string]interface{}{
		"expires_at": &at,
		"plain":      at,
		"list":       []interface{}{&at, "x"},
	}
	out, _ := json.Marshal(Process(in))
	got := string(out)
	if contains(got, `"expires_at":{}`) || contains(got, `"plain":{}`) {
		t.Fatalf("容器内时间不得折叠: %s", got)
	}
	if !contains(got, `"2026-06-01T08:00:00Z"`) {
		t.Fatalf("时间应输出 RFC3339: %s", got)
	}
}

// TestProcessTimeFormattingOptStillWins 注入 WithTimeField+WithFormatTime 时，
// 时间格式化优先级高于透传（既有能力不得被特判破坏）。
func TestProcessTimeFormattingOptStillWins(t *testing.T) {
	at := time.Date(2026, 11, 10, 12, 30, 0, 0, time.Local)
	in := struct {
		ExpiresAt *time.Time `json:"expires_at"`
	}{ExpiresAt: &at}

	out, _ := json.Marshal(Process(in,
		WithTimeField(func(name string) bool { return name == "expires_at" }),
		WithFormatTime(func(v interface{}) (string, bool) {
			if ts, ok := v.(time.Time); ok {
				return ts.Format("2006-01-02 15:04"), true
			}
			return "", false
		}),
	))
	if want := `"expires_at":"2026-11-10 12:30"`; !contains(string(out), want) {
		t.Fatalf("格式化 opt 应生效: got %s want 含 %s", out, want)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
