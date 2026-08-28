// Package timezone 提供"数据库存 UTC、前端展示北京时间"约定的通用时区工具。
//
// 多个中国区项目共用同一约定：时间以 UTC 写入数据库，对外展示时转换为
// 北京时间（Asia/Shanghai）。本包封装该约定下的格式化与解析能力，
// 并通过 Formatter 支持调用方注入业务专属的时间字段集合，避免全局可变状态。
package timezone

import (
	"fmt"
	"reflect"
	"time"
)

// DateTimeFormat 标准日期时间格式
const DateTimeFormat = "2006-01-02 15:04:05"

// parseLayouts 时间字符串解析 layout 顺序。
// 优先按 UTC 解析（符合 UTC 存库约定），失败后回退到服务器本地时区，
// 兼容历史遗留的本地时间数据。
var parseLayouts = []string{
	time.RFC3339,
	time.RFC3339Nano,
	DateTimeFormat,
	"2006-01-02 15:04:05.000000",
	"2006-01-02 15:04:05.000",
	"2006-01-02T15:04:05",
	"2006-01-02",
}

// Shanghai 北京时间时区。系统缺失 tzdata 时回退为固定东八区偏移。
var Shanghai *time.Location

func init() {
	var err error
	Shanghai, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		Shanghai = time.FixedZone("Asia/Shanghai", 8*60*60)
	}
}

// NowUTCString 返回当前 UTC 时间的标准字符串表示，用于写入数据库的时间字段。
//
// 返回值:
//   - string: UTC 时间字符串，格式为 2006-01-02 15:04:05
func NowUTCString() string {
	return time.Now().UTC().Format(DateTimeFormat)
}

// UTCStringToShanghai 将 UTC 时间字符串转换为北京时间字符串。
// 输入为空或解析失败时原样返回，避免覆盖非标准字段。
//
// 参数:
//   - s: UTC 时间字符串
//
// 返回值:
//   - string: 北京时间字符串；解析失败时返回原值
func UTCStringToShanghai(s string) string {
	if s == "" {
		return s
	}
	t, err := time.ParseInLocation(DateTimeFormat, s, time.UTC)
	if err != nil {
		return s
	}
	return t.In(Shanghai).Format(DateTimeFormat)
}

// FormatShanghai 将 time.Time（按 UTC 解析）格式化为北京时间字符串。
//
// 参数:
//   - t: 时间值
//
// 返回值:
//   - string: 北京时间字符串；零值返回空字符串
func FormatShanghai(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().In(Shanghai).Format(DateTimeFormat)
}

// ParseUTCString 将时间字符串解析为 time.Time。
// 优先按 UTC 解析，失败时回退到服务器本地时区，兼容历史本地时间数据。
//
// 参数:
//   - s: 时间字符串
//
// 返回值:
//   - time.Time: 解析后的时间
//   - error: 所有 layout 均解析失败时返回错误
func ParseUTCString(s string) (time.Time, error) {
	for _, layout := range parseLayouts {
		if t, err := time.ParseInLocation(layout, s, time.UTC); err == nil {
			return t, nil
		}
	}
	for _, layout := range parseLayouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析时间字符串: %s", s)
}

// Formatter 时区格式化器。时间字段集合与解析 layout 由调用方注入，避免全局可变状态。
type Formatter struct {
	Location     *time.Location
	ParseLayouts []string
	timeFields   map[string]bool
}

// NewFormatter 创建 Formatter；extraFields 为业务追加的时间字段名。
func NewFormatter(extraFields ...string) *Formatter {
	f := &Formatter{
		Location:   Shanghai,
		timeFields: make(map[string]bool),
	}
	f.AddTimeFields(extraFields...)
	return f
}

// parseLayoutsOrDefault 返回当前 Formatter 配置的解析 layout；未配置时使用默认 layout。
func (f *Formatter) parseLayoutsOrDefault() []string {
	if f != nil && len(f.ParseLayouts) > 0 {
		return f.ParseLayouts
	}
	return parseLayouts
}

// parseString 使用当前 Formatter 配置的 layout 解析时间字符串。
// 优先按 UTC 解析，失败时回退到服务器本地时区。
func (f *Formatter) parseString(s string) (time.Time, bool) {
	layouts := f.parseLayoutsOrDefault()
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.UTC); err == nil {
			return t, true
		}
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// AddTimeFields 向当前 Formatter 追加时间字段名（仅修改当前实例，不影响其他实例）。
func (f *Formatter) AddTimeFields(keys ...string) {
	if f.timeFields == nil {
		f.timeFields = make(map[string]bool)
	}
	for _, k := range keys {
		f.timeFields[k] = true
	}
}

// IsTimeField 判断字段名是否需做时区格式化。
func (f *Formatter) IsTimeField(key string) bool {
	if f == nil || f.timeFields == nil {
		return false
	}
	return f.timeFields[key]
}

// loc 返回当前 Formatter 的时区，未设置时回退为 Shanghai。
func (f *Formatter) loc() *time.Location {
	if f != nil && f.Location != nil {
		return f.Location
	}
	return Shanghai
}

// ToShanghaiString 将 UTC 时间字符串转为北京时间字符串；空/解析失败原样返回。
func (f *Formatter) ToShanghaiString(s string) string {
	if s == "" {
		return s
	}
	t, err := time.ParseInLocation(DateTimeFormat, s, time.UTC)
	if err != nil {
		return s
	}
	return t.In(f.loc()).Format(DateTimeFormat)
}

// FormatShanghai 将 time.Time 格式化为北京时间字符串；零值返回空字符串。
func (f *Formatter) FormatShanghai(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().In(f.loc()).Format(DateTimeFormat)
}

// FormatValue 处理单个时间值（time.Time 或时间字符串），统一转换为北京时间字符串。
// 供响应中间件的 Options.FormatTime 注入使用。
//
// 参数:
//   - val: 原始字段值，支持 time.Time 或时间字符串
//
// 返回值:
//   - string: 格式化后的北京时间字符串
//   - bool: 是否成功识别并格式化
func (f *Formatter) FormatValue(val interface{}) (string, bool) {
	switch v := val.(type) {
	case time.Time:
		if v.IsZero() {
			return "", false
		}
		return v.UTC().In(f.loc()).Format(DateTimeFormat), true
	case string:
		if v == "" || v == "0000-00-00 00:00:00" {
			return "", false
		}
		if t, ok := f.parseString(v); ok {
			return t.UTC().In(f.loc()).Format(DateTimeFormat), true
		}
		return v, false
	}

	// reflect 兜底：处理被包装成 interface 的 time.Time
	if reflect.TypeOf(val) != nil && reflect.TypeOf(val) == reflect.TypeOf(time.Time{}) {
		t := reflect.ValueOf(val).Interface().(time.Time)
		if t.IsZero() {
			return "", false
		}
		return t.UTC().In(f.loc()).Format(DateTimeFormat), true
	}
	return "", false
}
