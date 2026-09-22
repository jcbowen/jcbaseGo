package helper

import (
	"reflect"
	"sync"
	"testing"
)

// 回归测试：自引用/互引用类型环下，类型判定与默认值填充必须收敛，不允许栈溢出
// 背景：2026-09 sl-center stage 崩溃，hasActionableField 的 slice 分支经
// sliceElemMayBeConfigStruct → isConfigStructType 穿越缓存边界，丢弃 visiting/depth，
// 在 T1 → ptr → T2 → []*T1 的类型环下无限互递归直至 fatal error: stack overflow。

// reproChild → *reproParent → []*reproChild → reproChild，构成互引用类型环
type reproChild struct {
	Parent *reproParent
	Note   string `default:"child"`
}

type reproParent struct {
	Children []*reproChild
	Kid      *reproChild
	Name     string `default:"parent"`
}

// 用例1：互引用类型环（T1 → ptr → T2 → []*T1 → T1）必须收敛，不栈溢出
func TestReproCycleTypeNoStackOverflow(t *testing.T) {
	p := &reproParent{Kid: &reproChild{}}
	if err := CheckAndSetDefault(p); err != nil {
		t.Fatalf("环类型判定应收敛: %v", err)
	}
	if p.Name != "parent" {
		t.Fatalf("顶层默认值未填充: %q", p.Name)
	}
	if p.Kid.Note != "child" {
		t.Fatalf("嵌套默认值未填充: %q", p.Kid.Note)
	}
}

// 用例3：并发调用 CheckAndSetDefault（-race 下验证缓存与递归的并发安全）
func TestReproConcurrentCheckAndSetDefault(t *testing.T) {
	p := &reproParent{Kid: &reproChild{}, Children: []*reproChild{{}}}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := CheckAndSetDefault(p); err != nil {
				t.Errorf("并发填充失败: %v", err)
			}
		}()
	}
	wg.Wait()
	if p.Name != "parent" || p.Kid.Note != "child" || p.Children[0].Note != "child" {
		t.Fatal("并发填充结果不正确")
	}
}

// 用例2：自引用类型环（T → []*T），值路径不经过时直接调用类型判定入口验证收敛
type reproSelfNode struct {
	Children []*reproSelfNode
	Name     string `default:"node"`
}

func TestReproSelfCycleNoStackOverflow(t *testing.T) {
	if !isConfigStructType(reflect.TypeOf(reproSelfNode{})) {
		t.Fatal("自引用类型带 default 标签，应判定为配置结构体")
	}

	n := &reproSelfNode{}
	if err := CheckAndSetDefault(n); err != nil {
		t.Fatalf("自引用类型判定应收敛: %v", err)
	}
	if n.Name != "node" {
		t.Fatalf("默认值未填充: %q", n.Name)
	}
}
