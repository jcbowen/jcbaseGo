package debugger

import (
	"testing"
)

// TestRequiredStorage 测试必须传入存储器的功能
func TestRequiredStorage(t *testing.T) {
	t.Run("必须传入存储器实例", func(t *testing.T) {
		// 创建内存存储器
		memoryStorage, err := NewMemoryStorage(100)
		if err != nil {
			t.Fatalf("创建内存存储器失败: %v", err)
		}

		// 创建调试器实例，传入存储器
		config := &Config{
			Enabled: true,
			Storage: memoryStorage,
		}
		debugger, err := New(config)

		if err != nil {
			t.Fatalf("创建调试器失败: %v", err)
		}

		// 验证存储器是传入的存储器
		if debugger.storage != memoryStorage {
			t.Error("应该使用传入的存储器")
		}

		// 验证存储器不为nil
		if debugger.storage == nil {
			t.Error("存储器不应该为nil")
		}
	})

	t.Run("传入nil存储器应该报错", func(t *testing.T) {
		// 创建调试器实例，传入nil存储器
		config := &Config{
			Enabled: true,
			Storage: nil, // 传入nil存储器
		}
		debugger, err := New(config)

		// 应该返回错误
		if err == nil {
			t.Error("传入nil存储器应该返回错误")
		}

		// 调试器实例应该为nil
		if debugger != nil {
			t.Error("传入nil存储器时调试器实例应该为nil")
		}
	})

	t.Run("使用便捷构造函数创建调试器", func(t *testing.T) {
		// 使用便捷构造函数创建调试器
		debugger, err := NewWithMemoryStorage(100)

		if err != nil {
			t.Fatalf("创建调试器失败: %v", err)
		}

		// 验证存储器不为nil
		if debugger.storage == nil {
			t.Error("存储器不应该为nil")
		}

		// 验证配置正确
		if !debugger.config.Enabled {
			t.Error("调试器应该启用")
		}
	})

	t.Run("使用简单调试器构造函数", func(t *testing.T) {
		// 使用简单调试器构造函数
		debugger, err := NewSimpleDebugger()

		if err != nil {
			t.Fatalf("创建调试器失败: %v", err)
		}

		// 验证存储器不为nil
		if debugger.storage == nil {
			t.Error("存储器不应该为nil")
		}

		// 验证配置正确
		if !debugger.config.Enabled {
			t.Error("调试器应该启用")
		}
	})
}
