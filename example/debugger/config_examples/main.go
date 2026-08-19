package main

import (
	"fmt"

	"github.com/jcbowen/jcbaseGo/component/debugger"
)

// ConfigExamples 演示不同的debugger配置方式
// 这个示例展示了各种构造函数和配置选项
func main() {
	fmt.Println("=== Debugger 配置示例 ===")

	// 示例1：使用便捷构造函数（推荐）
	fmt.Println("\n1. 使用便捷构造函数（推荐）")
	debugger1, err := debugger.NewSimpleDebugger()
	if err != nil {
		fmt.Printf("创建失败: %v\n", err)
	} else {
		fmt.Println("✓ 创建成功 - 使用默认配置")
		_ = debugger1
	}

	// 示例2：使用内存存储器
	fmt.Println("\n2. 使用内存存储器")
	debugger2, err := debugger.NewWithMemoryStorage(200)
	if err != nil {
		fmt.Printf("创建失败: %v\n", err)
	} else {
		fmt.Println("✓ 创建成功 - 内存存储，最大200条记录")
		_ = debugger2
	}

	// 示例3：使用文件存储器
	fmt.Println("\n3. 使用文件存储器")
	debugger3, err := debugger.NewWithFileStorage("/tmp/debug_logs", 1000)
	if err != nil {
		fmt.Printf("创建失败: %v\n", err)
	} else {
		fmt.Println("✓ 创建成功 - 文件存储，路径: /tmp/debug_logs")
		_ = debugger3
	}

	// 示例4：使用自定义存储器
	fmt.Println("\n4. 使用自定义存储器")
	customStorage, err := debugger.NewMemoryStorage(150)
	if err != nil {
		fmt.Printf("创建自定义存储器失败: %v\n", err)
	} else {
		debugger4, err := debugger.NewWithCustomStorage(customStorage)
		if err != nil {
			fmt.Printf("创建调试器失败: %v\n", err)
		} else {
			fmt.Println("✓ 创建成功 - 使用自定义存储器")
			_ = debugger4
		}
	}

	// 示例5：使用生产环境配置
	fmt.Println("\n5. 使用生产环境配置")
	debugger5, err := debugger.NewProductionDebugger("/var/log/debug_logs")
	if err != nil {
		fmt.Printf("创建失败: %v\n", err)
	} else {
		fmt.Println("✓ 创建成功 - 生产环境配置")
		_ = debugger5
	}

	// 示例6：手动配置所有选项
	fmt.Println("\n6. 手动配置所有选项")
	fileStorage, err := debugger.NewFileStorage("./logs", 500)
	if err != nil {
		fmt.Printf("创建文件存储器失败: %v\n", err)
	} else {
		config := &debugger.Config{
			Enabled:         true,
			Storage:         fileStorage,
			MaxRecords:      500,
			RetentionPeriod: 168 * 60 * 60 * 1000000000, // 168小时（纳秒）
			SampleRate:      0.8,                        // 80%采样率
			SkipPaths:       []string{"/health", "/metrics"},
			SkipMethods:     []string{"OPTIONS", "HEAD"},
			MaxBodySize:     1024, // 1MB（单位：KB）
		}

		debugger6, err := debugger.New(config)
		if err != nil {
			fmt.Printf("创建失败: %v\n", err)
		} else {
			fmt.Println("✓ 创建成功 - 完全自定义配置")
			_ = debugger6
		}
	}

	fmt.Println("\n=== 配置示例完成 ===")
	fmt.Println("这些示例展示了debugger组件的各种配置方式")
	fmt.Println("可以根据实际需求选择合适的配置方法")

	// 演示配置选项说明
	demonstrateConfigOptions()
}

// 辅助函数：演示配置选项的用法
func demonstrateConfigOptions() {
	fmt.Println("\n=== 配置选项说明 ===")

	// 采样率配置示例
	fmt.Println("采样率配置:")
	fmt.Println("- SampleRate: 1.0  - 记录所有请求（100%）")
	fmt.Println("- SampleRate: 0.5  - 记录一半请求（50%）")
	fmt.Println("- SampleRate: 0.1  - 记录十分之一请求（10%）")

	// 保留时间配置示例
	fmt.Println("\n保留时间配置:")
	fmt.Println("- RetentionPeriod: 24h  - 保留24小时")
	fmt.Println("- RetentionPeriod: 168h - 保留7天")
	fmt.Println("- RetentionPeriod: 720h - 保留30天")

	// 跳过路径配置示例
	fmt.Println("\n跳过路径配置:")
	fmt.Println("- SkipPaths: []string{\"/static\", \"/favicon.ico\"}")
	fmt.Println("- 这些路径的请求不会被记录")

	// 跳过方法配置示例
	fmt.Println("\n跳过方法配置:")
	fmt.Println("- SkipMethods: []string{\"OPTIONS\", \"HEAD\"}")
	fmt.Println("- 这些HTTP方法的请求不会被记录")

	// 最大请求体大小配置
	fmt.Println("\n最大请求体大小配置:")
	fmt.Println("- MaxBodySize: 1024  - 1MB（单位：KB）")
	fmt.Println("- MaxBodySize: 2048  - 2MB")
	fmt.Println("- MaxBodySize: 5120  - 5MB")
}
