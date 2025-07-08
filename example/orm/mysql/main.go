package main

import (
	"fmt"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/orm/mysql"
)

// User 用户模型
type User struct {
	ID       int    `json:"id" gorm:"primaryKey"`
	Name     string `json:"name" gorm:"size:100;not null"`
	Email    string `json:"email" gorm:"size:100;uniqueIndex"`
	Age      int    `json:"age"`
	Status   int    `json:"status" gorm:"default:1"`
	CreateAt string `json:"create_at" gorm:"autoCreateTime"`
	UpdateAt string `json:"update_at" gorm:"autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}

func main() {
	fmt.Println("=== MySQL ORM 使用示例 ===\n")

	// 1. 连接数据库
	fmt.Println("1. 连接数据库:")
	db, err := connectDatabase()
	if err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		return
	}

	// 2. 自动迁移表结构
	fmt.Println("\n2. 自动迁移表结构:")
	migrateTable(db)

	// 3. 基本 CRUD 操作
	fmt.Println("\n3. 基本 CRUD 操作:")
	basicCRUD(db)

	// 4. 查询操作
	fmt.Println("\n4. 查询操作:")
	queryOperations(db)

	// 5. 事务操作
	fmt.Println("\n5. 事务操作:")
	transactionOperations(db)
}

// connectDatabase 连接数据库
func connectDatabase() (*mysql.Instance, error) {
	// 创建 MySQL 连接配置
	config := jcbaseGo.DbStruct{
		Host:        "localhost",
		Port:        "3306",
		Username:    "root",
		Password:    "password",
		Dbname:      "jcbase_test",
		Charset:     "utf8mb4",
		Protocol:    "tcp",
		ParseTime:   "True",
		TablePrefix: "",
		Alias:       "default",
	}

	// 连接数据库
	db := mysql.New(config)

	// 检查是否有错误
	if len(db.Error()) > 0 {
		return nil, fmt.Errorf("连接数据库失败: %v", db.Error())
	}

	fmt.Println("数据库连接成功")
	return db, nil
}

// migrateTable 自动迁移表结构
func migrateTable(db *mysql.Instance) {
	gormDB := db.GetDb()
	if gormDB == nil {
		fmt.Println("数据库连接为空")
		return
	}

	// 自动迁移 User 表
	err := gormDB.AutoMigrate(&User{})
	if err != nil {
		fmt.Printf("自动迁移失败: %v\n", err)
		return
	}
	fmt.Println("表结构迁移成功")
}

// basicCRUD 基本 CRUD 操作
func basicCRUD(db *mysql.Instance) {
	gormDB := db.GetDb()
	if gormDB == nil {
		fmt.Println("数据库连接为空")
		return
	}

	// 创建用户
	user := User{
		Name:   "张三",
		Email:  "zhangsan@example.com",
		Age:    25,
		Status: 1,
	}

	err := gormDB.Create(&user).Error
	if err != nil {
		fmt.Printf("创建用户失败: %v\n", err)
		return
	}
	fmt.Printf("创建用户成功，ID: %d\n", user.ID)

	// 查询用户
	var foundUser User
	err = gormDB.First(&foundUser, user.ID).Error
	if err != nil {
		fmt.Printf("查询用户失败: %v\n", err)
		return
	}
	fmt.Printf("查询用户成功: %+v\n", foundUser)

	// 更新用户
	updates := map[string]interface{}{
		"age":    26,
		"status": 2,
	}
	err = gormDB.Model(&foundUser).Updates(updates).Error
	if err != nil {
		fmt.Printf("更新用户失败: %v\n", err)
		return
	}
	fmt.Println("更新用户成功")

	// 删除用户
	err = gormDB.Delete(&foundUser).Error
	if err != nil {
		fmt.Printf("删除用户失败: %v\n", err)
		return
	}
	fmt.Println("删除用户成功")
}

// queryOperations 查询操作
func queryOperations(db *mysql.Instance) {
	gormDB := db.GetDb()
	if gormDB == nil {
		fmt.Println("数据库连接为空")
		return
	}

	// 创建测试数据
	users := []User{
		{Name: "李四", Email: "lisi@example.com", Age: 30, Status: 1},
		{Name: "王五", Email: "wangwu@example.com", Age: 28, Status: 1},
		{Name: "赵六", Email: "zhaoliu@example.com", Age: 35, Status: 0},
	}

	for _, user := range users {
		gormDB.Create(&user)
	}

	// 查询所有用户
	var allUsers []User
	err := gormDB.Find(&allUsers).Error
	if err != nil {
		fmt.Printf("查询所有用户失败: %v\n", err)
		return
	}
	fmt.Printf("所有用户数量: %d\n", len(allUsers))

	// 条件查询
	var activeUsers []User
	err = gormDB.Where("status = ?", 1).Find(&activeUsers).Error
	if err != nil {
		fmt.Printf("条件查询失败: %v\n", err)
		return
	}
	fmt.Printf("活跃用户数量: %d\n", len(activeUsers))

	// 排序查询
	var sortedUsers []User
	err = gormDB.Order("age desc").Find(&sortedUsers).Error
	if err != nil {
		fmt.Printf("排序查询失败: %v\n", err)
		return
	}
	fmt.Printf("按年龄排序的用户数量: %d\n", len(sortedUsers))

	// 分页查询
	var pagedUsers []User
	err = gormDB.Limit(2).Offset(0).Find(&pagedUsers).Error
	if err != nil {
		fmt.Printf("分页查询失败: %v\n", err)
		return
	}
	fmt.Printf("分页查询结果数量: %d\n", len(pagedUsers))

	// 统计查询
	var count int64
	err = gormDB.Model(&User{}).Count(&count).Error
	if err != nil {
		fmt.Printf("统计查询失败: %v\n", err)
		return
	}
	fmt.Printf("用户总数: %d\n", count)
}

// transactionOperations 事务操作
func transactionOperations(db *mysql.Instance) {
	gormDB := db.GetDb()
	if gormDB == nil {
		fmt.Println("数据库连接为空")
		return
	}

	// 开始事务
	tx := gormDB.Begin()
	if tx.Error != nil {
		fmt.Printf("开始事务失败: %v\n", tx.Error)
		return
	}

	// 在事务中创建用户
	user1 := User{
		Name:   "事务用户1",
		Email:  "tx1@example.com",
		Age:    25,
		Status: 1,
	}

	err := tx.Create(&user1).Error
	if err != nil {
		fmt.Printf("事务中创建用户1失败: %v\n", err)
		tx.Rollback()
		return
	}

	user2 := User{
		Name:   "事务用户2",
		Email:  "tx2@example.com",
		Age:    30,
		Status: 1,
	}

	err = tx.Create(&user2).Error
	if err != nil {
		fmt.Printf("事务中创建用户2失败: %v\n", err)
		tx.Rollback()
		return
	}

	// 提交事务
	err = tx.Commit().Error
	if err != nil {
		fmt.Printf("提交事务失败: %v\n", err)
		return
	}

	fmt.Println("事务操作成功完成")

	// 清理测试数据
	gormDB.Where("email LIKE ?", "tx%@example.com").Delete(&User{})
}
