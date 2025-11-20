package helper

import (
    "reflect"
    "testing"
    "time"
)

// TestCheckAndSetDefault 测试 CheckAndSetDefault
// 函数名：TestCheckAndSetDefault
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无（测试用例断言失败时由 testing 框架处理）
// 使用示例：go test ./...
func TestCheckAndSetDefault(t *testing.T) {
	type Inner struct {
		Label string        `default:"inner"`
		Age   int           `default:"18"`
		Delay time.Duration `default:"150ms"`
	}

	type Cfg struct {
		// 基本类型
		Name    string        `default:"app"`
		Enabled bool          `default:"true"`
		Port    int           `default:"8080"`
		Rate    float64       `default:"3.14"`
		Timeout time.Duration `default:"300ms"`

		// 嵌套结构体（递归）
		Inner Inner

		// 指针嵌套（顶层一层指针可设置）
		PInner *Inner
	}

	// 空值应被默认值填充
	cfg := &Cfg{}
	if err := CheckAndSetDefault(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name != "app" {
		t.Fatalf("Name default failed: %q", cfg.Name)
	}
	if cfg.Enabled != true {
		t.Fatalf("Enabled default failed: %v", cfg.Enabled)
	}
	if cfg.Port != 8080 {
		t.Fatalf("Port default failed: %d", cfg.Port)
	}
	if cfg.Rate != 3.14 {
		t.Fatalf("Rate default failed: %v", cfg.Rate)
	}
	if cfg.Timeout != 300*time.Millisecond {
		t.Fatalf("Timeout default failed: %v", cfg.Timeout)
	}
	if cfg.Inner.Label != "inner" || cfg.Inner.Age != 18 || cfg.Inner.Delay != 150*time.Millisecond {
		t.Fatalf("Inner defaults failed: %+v", cfg.Inner)
	}

	// 非空值不应被覆盖
	cfg2 := &Cfg{Name: "custom", Enabled: true, Port: 9090, Rate: 1.23, Timeout: 2 * time.Second}
	cfg2.Inner = Inner{Label: "x", Age: 20, Delay: 50 * time.Millisecond}
	if err := CheckAndSetDefault(cfg2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg2.Name != "custom" || !cfg2.Enabled || cfg2.Port != 9090 || cfg2.Rate != 1.23 || cfg2.Timeout != 2*time.Second {
		t.Fatalf("non-empty overwrite happened: %+v", cfg2)
	}
	if cfg2.Inner.Label != "x" || cfg2.Inner.Age != 20 || cfg2.Inner.Delay != 50*time.Millisecond {
		t.Fatalf("non-empty overwrite in inner happened: %+v", cfg2.Inner)
	}

	// 指针字段：库函数不处理结构体指针字段的默认值，保持零值
	cfg3 := &Cfg{PInner: &Inner{}}
	if err := CheckAndSetDefault(cfg3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg3.PInner == nil {
		t.Fatalf("PInner should not be nil after processing")
	}
	if cfg3.PInner.Label != "" || cfg3.PInner.Age != 0 || cfg3.PInner.Delay != 0 {
		t.Fatalf("pointer inner should remain zero values: %+v", cfg3.PInner)
	}

	// 非结构体输入应当静默返回
	var x int
	if err := CheckAndSetDefault(x); err != nil { // 非指针非结构体
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCheckAndSetDefault_ErrorTolerance 测试错误容忍：非法标签不应panic，保持原值
// 函数名：TestCheckAndSetDefault_ErrorTolerance
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无
// 使用示例：go test ./...
func TestCheckAndSetDefault_ErrorTolerance(t *testing.T) {
	type Bad struct {
		N   int           `default:"not-a-number"`
		Dur time.Duration `default:"not-a-duration"`
	}
	b := &Bad{}
	if err := CheckAndSetDefault(b); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.N != 0 || b.Dur != 0 {
		t.Fatalf("bad tags should not change values: %+v", b)
	}
}

// TestCheckAndSetDefaultWithPreserveTag 测试 CheckAndSetDefaultWithPreserveTag
// 函数名：TestCheckAndSetDefaultWithPreserveTag
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无
// 使用示例：go test ./...
func TestCheckAndSetDefaultWithPreserveTag(t *testing.T) {
	type Inner struct {
		Label string        `default:"inner" preserve:"true"`
		Delay time.Duration `default:"100ms" preserve:"true"`
	}
	type Cfg struct {
		Name         string  `default:"app"`
		NameKeep     string  `default:"app" preserve:"true"`
		Enabled      bool    `default:"true" preserve:"true"`
		EnabledPlain bool    `default:"true"`
		Port         int     `default:"8080" preserve:"true"`
		Rate         float64 `default:"3.14"`
		Inner        Inner
	}

	// 预置值与默认不同，标记 preserve 的字段应保留
	cfg := &Cfg{
		NameKeep: "",    // 与默认不等，需保留为空字符串
		Enabled:  false, // 与默认 true 不等，需要保留为 false
		Port:     0,     // 与默认 8080 不等，需要保留为 0
		Inner:    Inner{Label: "x", Delay: 250 * time.Millisecond},
	}
	if err := CheckAndSetDefaultWithPreserveTag(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 未标记 preserve 的字段应按默认填充
	if cfg.Name != "app" {
		t.Fatalf("Name default failed: %q", cfg.Name)
	}
	if cfg.Rate != 3.14 {
		t.Fatalf("Rate default failed: %v", cfg.Rate)
	}
	if cfg.EnabledPlain != true {
		t.Fatalf("EnabledPlain default failed: %v", cfg.EnabledPlain)
	}

	// preserve 字段保留预置值（不依赖显式值判定）
	if cfg.NameKeep != "" {
		t.Fatalf("NameKeep should be preserved empty: %q", cfg.NameKeep)
	}
	if cfg.Enabled != false {
		t.Fatalf("Enabled should be preserved false: %v", cfg.Enabled)
	}
	if cfg.Port != 0 {
		t.Fatalf("Port should be preserved zero: %d", cfg.Port)
	}
	if cfg.Inner.Label != "x" || cfg.Inner.Delay != 250*time.Millisecond {
		t.Fatalf("Inner preserve failed: %+v", cfg.Inner)
	}

	// 当值等于默认时，不需要保留（将维持默认）
	cfg2 := &Cfg{ // 预设等于默认
		NameKeep: "app",
		Enabled:  true,
		Port:     8080,
	}
	if err := CheckAndSetDefaultWithPreserveTag(cfg2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg2.NameKeep != "app" || cfg2.Enabled != true || cfg2.Port != 8080 {
		t.Fatalf("equal-to-default values should remain default: %+v", cfg2)
	}

	// 顶层为 nil 指针时静默返回
	var nilCfg *Cfg
	if err := CheckAndSetDefaultWithPreserveTag(nilCfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestCheckAndSetDefaultWithPreserveTag_InterfaceAndPointers 接口与指针类型快照兼容性
// 函数名：TestCheckAndSetDefaultWithPreserveTag_InterfaceAndPointers
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无
// 使用示例：go test ./...
func TestCheckAndSetDefaultWithPreserveTag_InterfaceAndPointers(t *testing.T) {
	type S struct {
		Val string `default:"x" preserve:"true"`
	}
	type C struct {
		Any   interface{} `default:"" preserve:"true"`
		PS    *S          `default:"" preserve:"true"`
		Plain string      `default:"p"`
	}

	s := &S{Val: "custom"}
	c := &C{Any: s, PS: s}
	if err := CheckAndSetDefaultWithPreserveTag(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Any == nil || c.PS == nil {
		t.Fatalf("interface/pointer should be preserved non-nil")
	}
	if vs := c.PS.Val; vs != "custom" {
		t.Fatalf("pointer struct value should be preserved: %q", vs)
	}
	if c.Plain != "p" {
		t.Fatalf("plain default failed: %q", c.Plain)
	}

	// nil 值也应保留（与默认不等）
	c2 := &C{Any: nil, PS: nil}
	if err := CheckAndSetDefaultWithPreserveTag(c2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c2.Any != nil || c2.PS != nil {
		t.Fatalf("nil interface/pointer should be preserved")
	}
}

// Test_buildDefaultValueForType_equalToDefault 辅助函数等价性校验
// 函数名：Test_buildDefaultValueForType_equalToDefault
// 参数：t *testing.T — 测试框架上下文
// 返回值：无
// 异常：无
// 使用示例：go test ./...
func Test_buildDefaultValueForType_equalToDefault(t *testing.T) {
	typ := reflect.TypeOf(0)
	def, ok := buildDefaultValueForType(typ, "123")
	if !ok || !def.IsValid() {
		t.Fatalf("buildDefaultValueForType failed")
	}
	if !equalToDefault(reflect.ValueOf(123), def) {
		t.Fatalf("equalToDefault failed for int")
	}

	dtyp := reflect.TypeOf(time.Duration(0))
	ddef, ok := buildDefaultValueForType(dtyp, "250ms")
	if !ok || !equalToDefault(reflect.ValueOf(250*time.Millisecond), ddef) {
		t.Fatalf("equalToDefault failed for duration")
	}
}

// 命名整型类型默认值赋值（仅内部命名类型，避免引入循环依赖）
func TestCheckAndSetDefault_NamedIntTypes_Internal(t *testing.T) {
    type myInt int
    type C1 struct {
        Level myInt `default:"2"`
    }
    c1 := &C1{}
    if err := CheckAndSetDefault(c1); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if c1.Level != myInt(2) {
        t.Fatalf("expected myInt Level=2, got %v", c1.Level)
    }
}
