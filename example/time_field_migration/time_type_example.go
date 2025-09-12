package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jcbowen/jcbaseGo/component/orm/base"
)

// TimeTypeUser 使用 time.Time 类型时间字段的用户模型
type TimeTypeUser struct {
	base.MysqlBaseModel
	ID        uint      `gorm:"column:id;primaryKey" json:"id"`
	Name      string    `gorm:"column:name;size:100" json:"name"`
	Email     string    `gorm:"column:email;size:100" json:"email"`
	CreatedAt time.Time `gorm:"column:created_at;type:DATETIME" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:DATETIME" json:"updated_at"`
	DeletedAt time.Time `gorm:"column:deleted_at;type:DATETIME;default:NULL" json:"deleted_at"`
}

func main() {
	fmt.Println("=== time.Time 类型时间字段示例 ===")

	// 创建用户实例
	user := &TimeTypeUser{
		Name:  "李四",
		Email: "lisi@example.com",
	}

	// 设置时间字段（time.Time 类型）
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	fmt.Printf("用户信息:\n")
	fmt.Printf("ID: %d\n", user.ID)
	fmt.Printf("姓名: %s\n", user.Name)
	fmt.Printf("邮箱: %s\n", user.Email)
	fmt.Printf("创建时间: %s (类型: %T)\n", user.CreatedAt.Format("2006-01-02 15:04:05"), user.CreatedAt)
	fmt.Printf("更新时间: %s (类型: %T)\n", user.UpdatedAt.Format("2006-01-02 15:04:05"), user.UpdatedAt)
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
	compareTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if user.CreatedAt.After(compareTime) {
		fmt.Printf("创建时间 %s 晚于 %s\n",
			user.CreatedAt.Format("2006-01-02 15:04:05"),
			compareTime.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("创建时间 %s 早于或等于 %s\n",
			user.CreatedAt.Format("2006-01-02 15:04:05"),
			compareTime.Format("2006-01-02 15:04:05"))
	}

	// 时间计算示例
	fmt.Println("\n--- 时间计算示例 ---")
	timeDiff := time.Since(user.CreatedAt)
	fmt.Printf("距离创建时间: %v\n", timeDiff)
	fmt.Printf("距离创建时间（小时）: %.2f\n", timeDiff.Hours())
	fmt.Printf("距离创建时间（分钟）: %.2f\n", timeDiff.Minutes())

	// 时间格式化示例
	fmt.Println("\n--- 时间格式化示例 ---")
	fmt.Printf("RFC3339 格式: %s\n", user.CreatedAt.Format(time.RFC3339))
	fmt.Printf("RFC3339Nano 格式: %s\n", user.CreatedAt.Format(time.RFC3339Nano))
	fmt.Printf("自定义格式: %s\n", user.CreatedAt.Format("2006年01月02日 15:04:05"))
	fmt.Printf("仅日期: %s\n", user.CreatedAt.Format("2006-01-02"))
	fmt.Printf("仅时间: %s\n", user.CreatedAt.Format("15:04:05"))

	// 时间戳示例
	fmt.Println("\n--- 时间戳示例 ---")
	fmt.Printf("Unix 时间戳: %d\n", user.CreatedAt.Unix())
	fmt.Printf("Unix 纳秒时间戳: %d\n", user.CreatedAt.UnixNano())
	fmt.Printf("毫秒时间戳: %d\n", user.CreatedAt.UnixMilli())
	fmt.Printf("微秒时间戳: %d\n", user.CreatedAt.UnixMicro())

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

	// 时区处理示例
	fmt.Println("\n--- 时区处理示例 ---")
	utcTime := user.CreatedAt.UTC()
	localTime := user.CreatedAt.Local()
	beijingTime := user.CreatedAt.In(time.FixedZone("CST", 8*3600))

	fmt.Printf("UTC 时间: %s\n", utcTime.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("本地时间: %s\n", localTime.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("北京时间: %s\n", beijingTime.Format("2006-01-02 15:04:05 MST"))

	fmt.Println("\n=== time.Time 类型示例完成 ===")
}
