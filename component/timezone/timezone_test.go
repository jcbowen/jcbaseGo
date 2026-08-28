package timezone

import (
	"testing"
	"time"
)

func TestUTCStringToShanghai(t *testing.T) {
	// 2024-01-01 00:00:00 UTC = 2024-01-01 08:00:00 北京时间
	if got := UTCStringToShanghai("2024-01-01 00:00:00"); got != "2024-01-01 08:00:00" {
		t.Errorf("UTCStringToShanghai = %q, want %q", got, "2024-01-01 08:00:00")
	}
	if got := UTCStringToShanghai(""); got != "" {
		t.Errorf("UTCStringToShanghai(\"\") = %q, want empty", got)
	}
	if got := UTCStringToShanghai("not-a-time"); got != "not-a-time" {
		t.Errorf("UTCStringToShanghai(invalid) = %q, want original", got)
	}
}

func TestFormatShanghai(t *testing.T) {
	utc := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if got := FormatShanghai(utc); got != "2024-01-01 08:00:00" {
		t.Errorf("FormatShanghai = %q, want %q", got, "2024-01-01 08:00:00")
	}
	if got := FormatShanghai(time.Time{}); got != "" {
		t.Errorf("FormatShanghai(zero) = %q, want empty", got)
	}
}

func TestParseUTCString(t *testing.T) {
	got, err := ParseUTCString("2024-01-01 00:00:00")
	if err != nil {
		t.Fatalf("ParseUTCString returned error: %v", err)
	}
	want := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("ParseUTCString = %v, want %v", got, want)
	}
	if _, err := ParseUTCString("garbage"); err == nil {
		t.Error("ParseUTCString(\"garbage\") expected error, got nil")
	}
}

func TestFormatter(t *testing.T) {
	f := NewFormatter("sent_at")
	if !f.IsTimeField("sent_at") {
		t.Error("IsTimeField(\"sent_at\") = false, want true")
	}
	if f.IsTimeField("created_at") {
		t.Error("IsTimeField(\"created_at\") = true, want false before AddTimeFields")
	}
	f.AddTimeFields("created_at")
	if !f.IsTimeField("created_at") {
		t.Error("IsTimeField(\"created_at\") = false, want true after AddTimeFields")
	}

	// Formatter 实例互不影响
	f2 := NewFormatter()
	if f2.IsTimeField("sent_at") {
		t.Error("new Formatter should not inherit timeFields from another instance")
	}

	// FormatValue: string
	if got, ok := f.FormatValue("2024-01-01 00:00:00"); !ok || got != "2024-01-01 08:00:00" {
		t.Errorf("FormatValue(string) = %q, %v", got, ok)
	}
	// FormatValue: time.Time
	utc := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if got, ok := f.FormatValue(utc); !ok || got != "2024-01-01 08:00:00" {
		t.Errorf("FormatValue(time.Time) = %q, %v", got, ok)
	}
	// FormatValue: zero time
	if _, ok := f.FormatValue(time.Time{}); ok {
		t.Error("FormatValue(zero time) should return ok=false")
	}
}

// TestFormatterCustomParseLayouts 验证 Formatter 支持自定义时间解析 layout。
func TestFormatterCustomParseLayouts(t *testing.T) {
	f := NewFormatter()
	f.ParseLayouts = []string{"2006/01/02 15:04:05"}

	got, ok := f.FormatValue("2024/01/01 00:00:00")
	if !ok {
		t.Fatal("FormatValue with custom layout should succeed")
	}
	if got != "2024-01-01 08:00:00" {
		t.Errorf("FormatValue = %q, want %q", got, "2024-01-01 08:00:00")
	}

	// 未配置自定义 layout 时，默认 layout 应仍能解析标准格式
	f2 := NewFormatter()
	got2, ok2 := f2.FormatValue("2024-01-01 00:00:00")
	if !ok2 || got2 != "2024-01-01 08:00:00" {
		t.Errorf("FormatValue with default layout = %q, %v", got2, ok2)
	}
}
