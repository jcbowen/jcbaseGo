package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jcbowen/jcbaseGo/component/orm/base"
)

// OldUser 旧版本用户模型（使用字符串类型时间字段）
type OldUser struct {
	base.MysqlBaseModel
	ID        uint   `gorm:"column:id;primaryKey" json:"id"`
	Name      string `gorm:"column:name;size:100" json:"name"`
	Email     string `gorm:"column:email;size:100" json:"email"`
	CreatedAt string `gorm:"column:created_at;type:DATETIME" json:"created_at"`
	UpdatedAt string `gorm:"column:updated_at;type:DATETIME" json:"updated_at"`
	DeletedAt string `gorm:"column:deleted_at;type:DATETIME;default:NULL" json:"deleted_at"`
}

// NewUser 新版本用户模型（使用 time.Time 类型时间字段）
type NewUser struct {
	base.MysqlBaseModel
	ID        uint      `gorm:"column:id;primaryKey" json:"id"`
	Name      string    `gorm:"column:name;size:100" json:"name"`
	Email     string    `gorm:"column:email;size:100" json:"email"`
	CreatedAt time.Time `gorm:"column:created_at;type:DATETIME" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:DATETIME" json:"updated_at"`
	DeletedAt time.Time `gorm:"column:deleted_at;type:DATETIME;default:NULL" json:"deleted_at"`
}

// MigrationHelper 迁移辅助工具
type MigrationHelper struct{}

// ConvertOldToNew 将旧版本用户转换为新版本用户
func (m *MigrationHelper) ConvertOldToNew(oldUser *OldUser) (*NewUser, error) {
	newUser := &NewUser{
		ID:    oldUser.ID,
		Name:  oldUser.Name,
		Email: oldUser.Email,
	}

	// 转换创建时间
	if oldUser.CreatedAt != "" {
		createdAt, err := time.Parse("2006-01-02 15:04:05", oldUser.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("解析创建时间失败: %w", err)
		}
		newUser.CreatedAt = createdAt
	}

	// 转换更新时间
	if oldUser.UpdatedAt != "" {
		updatedAt, err := time.Parse("2006-01-02 15:04:05", oldUser.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("解析更新时间失败: %w", err)
		}
		newUser.UpdatedAt = updatedAt
	}

	// 转换删除时间
	if oldUser.DeletedAt != "" {
		deletedAt, err := time.Parse("2006-01-02 15:04:05", oldUser.DeletedAt)
		if err != nil {
			return nil, fmt.Errorf("解析删除时间失败: %w", err)
		}
		newUser.DeletedAt = deletedAt
	}

	return newUser, nil
}

// ConvertNewToOld 将新版本用户转换为旧版本用户
func (m *MigrationHelper) ConvertNewToOld(newUser *NewUser) *OldUser {
	oldUser := &OldUser{
		ID:    newUser.ID,
		Name:  newUser.Name,
		Email: newUser.Email,
	}

	// 转换创建时间
	if !newUser.CreatedAt.IsZero() {
		oldUser.CreatedAt = newUser.CreatedAt.Format("2006-01-02 15:04:05")
	}

	// 转换更新时间
	if !newUser.UpdatedAt.IsZero() {
		oldUser.UpdatedAt = newUser.UpdatedAt.Format("2006-01-02 15:04:05")
	}

	// 转换删除时间
	if !newUser.DeletedAt.IsZero() {
		oldUser.DeletedAt = newUser.DeletedAt.Format("2006-01-02 15:04:05")
	}

	return oldUser
}

// ValidateTimeField 验证时间字段的有效性
func (m *MigrationHelper) ValidateTimeField(timeStr string) error {
	if timeStr == "" {
		return nil // 空字符串是有效的
	}

	_, err := time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return fmt.Errorf("无效的时间格式: %s", timeStr)
	}

	return nil
}

func main() {
	fmt.Println("=== 时间字段类型迁移示例 ===")

	// 创建迁移辅助工具
	helper := &MigrationHelper{}

	// 创建旧版本用户数据
	fmt.Println("\n1. 创建旧版本用户数据（字符串类型）")
	oldUser := &OldUser{
		ID:        1,
		Name:      "张三",
		Email:     "zhangsan@example.com",
		CreatedAt: "2024-01-01 10:00:00",
		UpdatedAt: "2024-01-01 12:00:00",
		DeletedAt: "", // 未删除
	}

	fmt.Printf("旧版本用户:\n")
	fmt.Printf("  ID: %d\n", oldUser.ID)
	fmt.Printf("  姓名: %s\n", oldUser.Name)
	fmt.Printf("  邮箱: %s\n", oldUser.Email)
	fmt.Printf("  创建时间: %s (类型: %T)\n", oldUser.CreatedAt, oldUser.CreatedAt)
	fmt.Printf("  更新时间: %s (类型: %T)\n", oldUser.UpdatedAt, oldUser.UpdatedAt)
	fmt.Printf("  删除时间: %s (类型: %T)\n", oldUser.DeletedAt, oldUser.DeletedAt)

	// 验证时间字段
	fmt.Println("\n2. 验证时间字段")
	if err := helper.ValidateTimeField(oldUser.CreatedAt); err != nil {
		fmt.Printf("创建时间验证失败: %v\n", err)
	} else {
		fmt.Println("创建时间验证通过")
	}

	if err := helper.ValidateTimeField(oldUser.UpdatedAt); err != nil {
		fmt.Printf("更新时间验证失败: %v\n", err)
	} else {
		fmt.Println("更新时间验证通过")
	}

	// 转换为新版本用户
	fmt.Println("\n3. 转换为新版本用户（time.Time 类型）")
	newUser, err := helper.ConvertOldToNew(oldUser)
	if err != nil {
		fmt.Printf("转换失败: %v\n", err)
		return
	}

	fmt.Printf("新版本用户:\n")
	fmt.Printf("  ID: %d\n", newUser.ID)
	fmt.Printf("  姓名: %s\n", newUser.Name)
	fmt.Printf("  邮箱: %s\n", newUser.Email)
	fmt.Printf("  创建时间: %s (类型: %T)\n", newUser.CreatedAt.Format("2006-01-02 15:04:05"), newUser.CreatedAt)
	fmt.Printf("  更新时间: %s (类型: %T)\n", newUser.UpdatedAt.Format("2006-01-02 15:04:05"), newUser.UpdatedAt)
	fmt.Printf("  删除时间: %s (类型: %T)\n", newUser.DeletedAt.Format("2006-01-02 15:04:05"), newUser.DeletedAt)

	// 演示新版本用户的时间操作
	fmt.Println("\n4. 新版本用户的时间操作")

	// 时间比较
	compareTime := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
	if newUser.CreatedAt.Before(compareTime) {
		fmt.Printf("创建时间 %s 早于 %s\n",
			newUser.CreatedAt.Format("2006-01-02 15:04:05"),
			compareTime.Format("2006-01-02 15:04:05"))
	}

	// 时间计算
	timeDiff := newUser.UpdatedAt.Sub(newUser.CreatedAt)
	fmt.Printf("从创建到更新的时间差: %v\n", timeDiff)

	// 模拟软删除
	fmt.Println("\n5. 模拟软删除")
	newUser.DeletedAt = time.Now()
	fmt.Printf("软删除时间: %s\n", newUser.DeletedAt.Format("2006-01-02 15:04:05"))

	// 转换回旧版本用户
	fmt.Println("\n6. 转换回旧版本用户")
	convertedOldUser := helper.ConvertNewToOld(newUser)

	fmt.Printf("转换后的旧版本用户:\n")
	fmt.Printf("  ID: %d\n", convertedOldUser.ID)
	fmt.Printf("  姓名: %s\n", convertedOldUser.Name)
	fmt.Printf("  邮箱: %s\n", convertedOldUser.Email)
	fmt.Printf("  创建时间: %s (类型: %T)\n", convertedOldUser.CreatedAt, convertedOldUser.CreatedAt)
	fmt.Printf("  更新时间: %s (类型: %T)\n", convertedOldUser.UpdatedAt, convertedOldUser.UpdatedAt)
	fmt.Printf("  删除时间: %s (类型: %T)\n", convertedOldUser.DeletedAt, convertedOldUser.DeletedAt)

	// JSON 序列化对比
	fmt.Println("\n7. JSON 序列化对比")

	// 旧版本 JSON
	oldJSON, err := json.MarshalIndent(oldUser, "", "  ")
	if err != nil {
		fmt.Printf("旧版本 JSON 序列化失败: %v\n", err)
	} else {
		fmt.Printf("旧版本 JSON:\n%s\n", string(oldJSON))
	}

	// 新版本 JSON
	newJSON, err := json.MarshalIndent(newUser, "", "  ")
	if err != nil {
		fmt.Printf("新版本 JSON 序列化失败: %v\n", err)
	} else {
		fmt.Printf("新版本 JSON:\n%s\n", string(newJSON))
	}

	// 批量迁移示例
	fmt.Println("\n8. 批量迁移示例")

	// 创建多个旧版本用户
	oldUsers := []*OldUser{
		{ID: 1, Name: "用户1", Email: "user1@example.com", CreatedAt: "2024-01-01 10:00:00", UpdatedAt: "2024-01-01 12:00:00"},
		{ID: 2, Name: "用户2", Email: "user2@example.com", CreatedAt: "2024-01-02 10:00:00", UpdatedAt: "2024-01-02 12:00:00"},
		{ID: 3, Name: "用户3", Email: "user3@example.com", CreatedAt: "2024-01-03 10:00:00", UpdatedAt: "2024-01-03 12:00:00"},
	}

	// 批量转换为新版本用户
	var newUsers []*NewUser
	for _, oldUser := range oldUsers {
		newUser, err := helper.ConvertOldToNew(oldUser)
		if err != nil {
			fmt.Printf("转换用户 %d 失败: %v\n", oldUser.ID, err)
			continue
		}
		newUsers = append(newUsers, newUser)
	}

	fmt.Printf("成功转换 %d 个用户\n", len(newUsers))
	for _, user := range newUsers {
		fmt.Printf("  用户 %d: %s (创建时间: %s)\n",
			user.ID, user.Name, user.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	fmt.Println("\n=== 迁移示例完成 ===")
	fmt.Println("迁移总结:")
	fmt.Println("1. 旧版本字符串类型完全兼容")
	fmt.Println("2. 新版本 time.Time 类型提供更好的类型安全性")
	fmt.Println("3. 可以逐步迁移，无需一次性修改所有代码")
	fmt.Println("4. 新旧类型可以在同一个项目中并存")
}
