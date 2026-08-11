package jcbaseGo

import (
	"strings"
	"testing"
)

// TestReloadConfigWithNilStructPointer 验证当 ConfigData 为 nil 的结构体指针时，
// ReloadConfig 应能自动分配实例并成功加载，而不是返回"配置信息不能为空"错误。
//
// 参数：
//   - t: 测试上下文
func TestReloadConfigWithNilStructPointer(t *testing.T) {
	// 构造一个 nil 的结构体指针；赋值给 interface{} 后，interface 的动态类型
	// 仍为 *DefaultConfigStruct，因此 reflect.TypeOf 可以推断出结构体类型，
	// loadConfig/resetConfigData 应能自动分配实例。
	var cfg *DefaultConfigStruct

	opt := &Option{
		ConfigType:   ConfigTypeJSON,
		ConfigSource: "./data/conf.json",
		ConfigData:   cfg,
	}

	// 仅验证 resetConfigData 不会错误地返回"配置信息不能为空"。
	// 由于缺少真实配置文件，后续步骤可能返回其他错误，但不属于本次修复范围。
	err := opt.resetConfigData()
	if err != nil {
		t.Fatalf("resetConfigData 对 nil 结构体指针不应返回错误，实际: %v", err)
	}

	// 验证 ConfigData 已被自动分配为非 nil 实例
	if opt.ConfigData == nil {
		t.Fatal("resetConfigData 应自动为 nil 结构体指针分配实例")
	}
}

// TestReloadConfigWithPureNilInterface 验证当 ConfigData 为完全 nil 的 interface{} 时，
// resetConfigData 仍应返回错误，因为无法推断要分配的结构体类型。
//
// 参数：
//   - t: 测试上下文
func TestReloadConfigWithPureNilInterface(t *testing.T) {
	opt := &Option{
		ConfigType:   ConfigTypeJSON,
		ConfigSource: "./data/conf.json",
		ConfigData:   nil,
	}

	err := opt.resetConfigData()
	if err == nil {
		t.Fatal("ConfigData 为完全 nil 的 interface{} 时，resetConfigData 应返回错误")
	}
	if !strings.Contains(err.Error(), "配置信息不能为空") {
		t.Fatalf("期望返回'配置信息不能为空'错误，实际得到: %v", err)
	}
}
