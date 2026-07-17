package crud

import (
	"reflect"
	"sync"
	"testing"
)

// 用于测试 gorm 标签解析的嵌套结构体
type auditInfo struct {
	CreatedBy uint `gorm:"column:created_by" json:"created_by"`
	UpdatedBy uint `gorm:"column:updated_by" json:"updated_by"`
}

// 测试模型，包含各种 gorm 忽略标签
type testUpdateFilterModel struct {
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	Username  string `gorm:"column:username" json:"username"`
	Password  string `gorm:"column:password;-:update" json:"password"`
	TempAll   string `gorm:"column:temp_all;-:all" json:"temp_all"`
	TempOnly  string `gorm:"column:temp_only" json:"temp_only"`
}

// 测试匿名嵌入字段的 gorm 标签解析
type testEmbeddedModel struct {
	auditInfo
	ID uint `gorm:"column:id;primaryKey" json:"id"`
}

// TestBuildGormTagMap 验证列名到 gorm 标签的映射构建正确
func TestBuildGormTagMap(t *testing.T) {
	m := buildGormTagMap(reflect.TypeOf(testUpdateFilterModel{}))

	if _, ok := m["id"]; !ok {
		t.Errorf("应解析出 id 字段，实际映射 %v", m)
	}
	if tag, ok := m["password"]; !ok || tag != "column:password;-:update" {
		t.Errorf("password 字段的 gorm 标签应为 column:password;-:update，实际 %v", tag)
	}
	if tag, ok := m["temp_all"]; !ok || tag != "column:temp_all;-:all" {
		t.Errorf("temp_all 字段的 gorm 标签应为 column:temp_all;-:all，实际 %v", tag)
	}
}

// TestBuildGormTagMapEmbedded 验证匿名嵌入字段的 gorm 标签被递归解析
func TestBuildGormTagMapEmbedded(t *testing.T) {
	m := buildGormTagMap(reflect.TypeOf(testEmbeddedModel{}))

	if tag, ok := m["created_by"]; !ok || tag != "column:created_by" {
		t.Errorf("嵌入字段 created_by 的 gorm 标签应为 column:created_by，实际 %v", tag)
	}
	if tag, ok := m["updated_by"]; !ok || tag != "column:updated_by" {
		t.Errorf("嵌入字段 updated_by 的 gorm 标签应为 column:updated_by，实际 %v", tag)
	}
}

// TestIsGormUpdateIgnoredTag 验证 gorm 标签更新忽略判断
func TestIsGormUpdateIgnoredTag(t *testing.T) {
	cases := []struct {
		tag      string
		expected bool
	}{
		{"-", true},
		{"-:all", true},
		{"-:update", true},
		{"column:password;-:update", true},
		{"column:password;-:all", true},
		{"column:password;-", true},
		{"-:create", false},
		{"-:migrate", false},
		{"column:username", false},
		{"column:temp;size:50", false},
	}

	for _, c := range cases {
		got := isGormUpdateIgnoredTag(c.tag)
		if got != c.expected {
			t.Errorf("标签 %q 期望 %v，实际 %v", c.tag, c.expected, got)
		}
	}
}

// TestTraitIsGormUpdateIgnored 验证 Trait 能正确判断字段是否在更新时被忽略
func TestTraitIsGormUpdateIgnored(t *testing.T) {
	trait := &Trait{
		Model: &testUpdateFilterModel{},
	}

	if !trait.isGormUpdateIgnored("password") {
		t.Error("password 字段标记为 -:update，应被判断为更新忽略")
	}
	if !trait.isGormUpdateIgnored("temp_all") {
		t.Error("temp_all 字段标记为 -:all，应被判断为更新忽略")
	}
	if trait.isGormUpdateIgnored("username") {
		t.Error("username 字段未标记忽略，不应被判断为更新忽略")
	}
	if trait.isGormUpdateIgnored("temp_only") {
		t.Error("temp_only 字段未标记忽略，不应被判断为更新忽略")
	}
}

// TestSetValueCheckFieldRejectsUpdateIgnored 验证 SetValueCheckField 拒绝 -:update 字段
func TestSetValueCheckFieldRejectsUpdateIgnored(t *testing.T) {
	trait := &Trait{
		Model:       &testUpdateFilterModel{},
		ModelFields: []string{"id", "username", "password", "temp_all", "temp_only"},
	}

	if err := trait.SetValueCheckField(nil, "password"); err == nil {
		t.Error("SetValueCheckField 对 -:update 字段应返回错误")
	}
	if err := trait.SetValueCheckField(nil, "temp_all"); err == nil {
		t.Error("SetValueCheckField 对 -:all 字段应返回错误")
	}
	if err := trait.SetValueCheckField(nil, "username"); err != nil {
		t.Errorf("SetValueCheckField 对普通字段不应返回错误，实际 %v", err)
	}
}

// TestGetGormTagByColumnCache 验证 getGormTagByColumn 的缓存机制
func TestGetGormTagByColumnCache(t *testing.T) {
	trait := &Trait{
		Model: &testUpdateFilterModel{},
	}

	// 首次调用，写入缓存
	first := trait.getGormTagByColumn("password")
	if first != "column:password;-:update" {
		t.Errorf("首次调用期望 column:password;-:update，实际 %v", first)
	}

	// 再次调用，应从缓存读取，结果一致
	second := trait.getGormTagByColumn("password")
	if second != first {
		t.Errorf("缓存命中结果应一致，期望 %v，实际 %v", first, second)
	}

	// 验证缓存中确实存在该类型
	modelType := reflect.TypeOf(trait.Model).Elem()
	if cached, ok := gormTagMapCache.Load(modelType); !ok {
		t.Error("缓存中应存在模型类型的映射")
	} else if m, ok := cached.(gormFieldTagMap); !ok {
		t.Errorf("缓存值类型应为 gormFieldTagMap，实际 %T", cached)
	} else if m["password"] != "column:password;-:update" {
		t.Errorf("缓存中 password 标签应为 column:password;-:update，实际 %v", m["password"])
	}
}

// TestGetGormTagByColumnConcurrent 验证并发调用 getGormTagByColumn 的安全性
func TestGetGormTagByColumnConcurrent(t *testing.T) {
	trait := &Trait{
		Model: &testUpdateFilterModel{},
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tag := trait.getGormTagByColumn("password")
			if tag != "column:password;-:update" {
				t.Errorf("并发读取标签错误，期望 column:password;-:update，实际 %v", tag)
			}
		}()
	}
	wg.Wait()
}
