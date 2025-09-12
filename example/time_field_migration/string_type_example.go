package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jcbowen/jcbaseGo/component/orm/base"
)

// StringTimeUser 使用字符串类型时间字段的用户模型
type StringTimeUser struct {
	base.MysqlBaseModel
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	Name      string `gorm:"column:name;size:100" json:"name"`
	Email     string `gorm:"column:email;size:100" json:"email"`
	CreatedAt string `gorm:"column:created_at;type:DATETIME" json:"created_at"`
	UpdatedAt string `gorm:"column:updated_at;type:DATETIME" json:"updated_at"`
	DeletedAt string `gorm:"column:deleted_at;type:DATETIME;default:NULL" json:"deleted_at"`
}

func main() {
	fmt.Println("=== 字符串类型时间字段示例 ===")

	// 创建用户实例
	user := &StringTimeUser{
		Name:  "张三",
		Email: "zhangsan@example.com",
	}

	// 手动设置时间字段（字符串格式）
	now := time.Now().Format("2006-01-02 15:04:05")
	user.CreatedAt = now
	user.UpdatedAt = now

	fmt.Printf("用户信息:\n")
	fmt.Printf("ID: %d\n", user.ID)
	fmt.Printf("姓名: %s\n", user.Name)
	fmt.Printf("邮箱: %s\n", user.Email)
	fmt.Printf("创建时间: %s (类型: %T)\n", user.CreatedAt, user.CreatedAt)
	fmt.Printf("更新时间: %s (类型: %T)\n", user.UpdatedAt, user.UpdatedAt)
	fmt.Printf("删除时间: %s (类型: %T)\n", user.DeletedAt, user.DeletedAt)

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
	compareTime := "2024-01-01 00:00:00"
	if user.CreatedAt > compareTime {
		fmt.Printf("创建时间 %s 晚于 %s\n", user.CreatedAt, compareTime)
	} else {
		fmt.Printf("创建时间 %s 早于或等于 %s\n", user.CreatedAt, compareTime)
	}

	// 时间解析示例
	fmt.Println("\n--- 时间解析示例 ---")
	parsedTime, err := time.Parse("2006-01-02 15:04:05", user.CreatedAt)
	if err != nil {
		fmt.Printf("时间解析失败: %v\n", err)
	} else {
		fmt.Printf("解析后的时间: %s\n", parsedTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("时间戳: %d\n", parsedTime.Unix())
		fmt.Printf("年份: %d\n", parsedTime.Year())
		fmt.Printf("月份: %d\n", parsedTime.Month())
		fmt.Printf("日期: %d\n", parsedTime.Day())
	}

	// 模拟软删除
	fmt.Println("\n--- 软删除示例 ---")
	user.DeletedAt = time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("软删除时间: %s\n", user.DeletedAt)

	// 检查是否已删除
	if user.DeletedAt != "" {
		fmt.Println("用户已被软删除")
	} else {
		fmt.Println("用户未被删除")
	}

	fmt.Println("\n=== 字符串类型示例完成 ===")
}
