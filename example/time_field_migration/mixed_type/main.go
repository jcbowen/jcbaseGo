package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jcbowen/jcbaseGo/component/orm/base"
)

// MixedTypeUser 混合使用字符串和时间类型的用户模型
type MixedTypeUser struct {
	base.MysqlBaseModel
	ID        uint      `gorm:"column:id;primaryKey" json:"id"`
	Name      string    `gorm:"column:name;size:100" json:"name"`
	Email     string    `gorm:"column:email;size:100" json:"email"`
	CreatedAt time.Time `gorm:"column:created_at;type:DATETIME" json:"created_at"`              // 时间类型
	UpdatedAt string    `gorm:"column:updated_at;type:DATETIME" json:"updated_at"`              // 字符串类型
	DeletedAt time.Time `gorm:"column:deleted_at;type:DATETIME;default:NULL" json:"deleted_at"` // 时间类型
}

func main() {
	fmt.Println("=== 混合类型时间字段示例 ===")
	fmt.Println("演示同一个模型中同时使用字符串和时间类型字段")

	// 创建用户实例
	user := &MixedTypeUser{
		Name:  "王五",
		Email: "wangwu@example.com",
	}

	// 设置时间字段
	now := time.Now()
	user.CreatedAt = now                               // time.Time 类型
	user.UpdatedAt = now.Format("2006-01-02 15:04:05") // 字符串类型

	fmt.Printf("用户信息:\n")
	fmt.Printf("ID: %d\n", user.ID)
	fmt.Printf("姓名: %s\n", user.Name)
	fmt.Printf("邮箱: %s\n", user.Email)
	fmt.Printf("创建时间: %s (类型: %T)\n", user.CreatedAt.Format("2006-01-02 15:04:05"), user.CreatedAt)
	fmt.Printf("更新时间: %s (类型: %T)\n", user.UpdatedAt, user.UpdatedAt)
	fmt.Printf("删除时间: %s (类型: %T)\n", user.DeletedAt.Format("2006-01-02 15:04:05"), user.DeletedAt)

	// JSON 序列化
	fmt.Println("\n--- JSON 序列化 ---")
	jsonData, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		fmt.Printf("JSON 序列化失败: %v\n", err)
	} else {
		fmt.Printf("JSON 数据:\n%s\n", string(jsonData))
	}

	// 时间比较示例
	fmt.Println("\n--- 时间比较示例 ---")

	// 比较 time.Time 类型字段
	compareTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if user.CreatedAt.After(compareTime) {
		fmt.Printf("创建时间 %s 晚于 %s\n",
			user.CreatedAt.Format("2006-01-02 15:04:05"),
			compareTime.Format("2006-01-02 15:04:05"))
	}

	// 比较字符串类型字段
	compareTimeStr := "2024-01-01 00:00:00"
	if user.UpdatedAt > compareTimeStr {
		fmt.Printf("更新时间 %s 晚于 %s\n", user.UpdatedAt, compareTimeStr)
	}

	// 类型转换示例
	fmt.Println("\n--- 类型转换示例 ---")

	// 将字符串时间转换为 time.Time
	parsedTime, err := time.Parse("2006-01-02 15:04:05", user.UpdatedAt)
	if err != nil {
		fmt.Printf("字符串时间解析失败: %v\n", err)
	} else {
		fmt.Printf("解析后的更新时间: %s\n", parsedTime.Format("2006-01-02 15:04:05"))

		// 现在可以进行比较
		if user.CreatedAt.After(parsedTime) {
			fmt.Println("创建时间晚于更新时间")
		} else {
			fmt.Println("创建时间早于或等于更新时间")
		}
	}

	// 将 time.Time 转换为字符串
	createdAtStr := user.CreatedAt.Format("2006-01-02 15:04:05")
	fmt.Printf("创建时间字符串: %s\n", createdAtStr)

	// 时间计算示例
	fmt.Println("\n--- 时间计算示例 ---")

	// 计算两个 time.Time 字段的差值
	if !user.DeletedAt.IsZero() {
		timeDiff := user.DeletedAt.Sub(user.CreatedAt)
		fmt.Printf("从创建到删除的时间: %v\n", timeDiff)
	}

	// 计算从创建时间到现在的时间
	timeSinceCreation := time.Since(user.CreatedAt)
	fmt.Printf("距离创建时间: %v\n", timeSinceCreation)

	// 模拟软删除
	fmt.Println("\n--- 软删除示例 ---")
	user.DeletedAt = time.Now()
	fmt.Printf("软删除时间: %s\n", user.DeletedAt.Format("2006-01-02 15:04:05"))

	// 检查是否已删除
	if !user.DeletedAt.IsZero() {
		fmt.Println("用户已被软删除")
	} else {
		fmt.Println("用户未被删除")
	}

	// 时间字段统一处理示例
	fmt.Println("\n--- 时间字段统一处理示例 ---")

	// 获取所有时间字段的字符串表示
	timeFields := map[string]string{
		"CreatedAt": user.CreatedAt.Format("2006-01-02 15:04:05"),
		"UpdatedAt": user.UpdatedAt,
		"DeletedAt": user.DeletedAt.Format("2006-01-02 15:04:05"),
	}

	fmt.Println("所有时间字段的字符串表示:")
	for fieldName, timeStr := range timeFields {
		fmt.Printf("  %s: %s\n", fieldName, timeStr)
	}

	// 获取所有时间字段的 time.Time 表示
	fmt.Println("\n所有时间字段的 time.Time 表示:")
	createdAtTime := user.CreatedAt
	updatedAtTime, _ := time.Parse("2006-01-02 15:04:05", user.UpdatedAt)
	deletedAtTime := user.DeletedAt

	fmt.Printf("  CreatedAt: %s\n", createdAtTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("  UpdatedAt: %s\n", updatedAtTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("  DeletedAt: %s\n", deletedAtTime.Format("2006-01-02 15:04:05"))

	fmt.Println("\n=== 混合类型示例完成 ===")
}
