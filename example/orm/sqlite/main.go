package main

import (
	"fmt"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/orm/sqlite"
	"gorm.io/gorm"
)

// User 用户模型
// 定义用户表结构，包含基本用户信息和时间戳字段
// 注意：SQLite使用字符串类型的时间戳，而非time.Time类型
// 这是因为SQLite的日期时间处理方式与MySQL不同
// 使用字符串类型可以避免时区转换问题
// 参数：
//   - ID: 用户ID，主键
//   - Name: 用户名，最大长度100，非空
//   - Email: 邮箱，唯一索引
//   - Age: 年龄
//   - Status: 状态，默认值1
//   - CreateAt: 创建时间，自动生成
//   - UpdateAt: 更新时间，自动更新
type User struct {
	ID       int    `json:"id" gorm:"primaryKey"`              // 用户ID，主键
	Name     string `json:"name" gorm:"size:100;not null"`     // 用户名，最大长度100，非空
	Email    string `json:"email" gorm:"size:100;uniqueIndex"` // 邮箱，唯一索引
	Age      int    `json:"age"`                               // 年龄
	Status   int    `json:"status" gorm:"default:1"`           // 状态，默认值1
	CreateAt string `json:"create_at" gorm:"autoCreateTime"`   // 创建时间，自动生成
	UpdateAt string `json:"update_at" gorm:"autoUpdateTime"`   // 更新时间，自动更新
}

// TableName 设置表名
// 返回：
//   - string: 数据库表名
func (User) TableName() string {
	return "users"
}

// main 主函数
// 演示SQLite ORM的基本使用方法，包括连接、迁移、CRUD、查询和事务操作
// SQLite与MySQL的主要区别：
// 1. 连接配置使用文件路径而非网络地址
// 2. 日期时间使用字符串类型而非time.Time
// 3. 连接池配置更简单（通常只需要1个连接）
// 4. 无需网络连接，适合本地开发和测试
func main() {
	fmt.Println("=== SQLite ORM 使用示例 ===\n")

	// 1. 连接数据库
	fmt.Println("1. 连接SQLite数据库:")
	db, err := connectDatabase()
	if err != nil {
		fmt.Printf("连接SQLite数据库失败: %v\n", err)
		return
	}

	// 2. 自动迁移表结构
	fmt.Println("\n2. 自动迁移表结构:")
	if err := migrateTable(db); err != nil {
		fmt.Printf("表结构迁移失败: %v\n", err)
		return
	}

	// 3. 基本 CRUD 操作
	fmt.Println("\n3. 基本 CRUD 操作:")
	if err := basicCRUD(db); err != nil {
		fmt.Printf("基本CRUD操作失败: %v\n", err)
		return
	}

	// 4. 查询操作
	fmt.Println("\n4. 查询操作:")
	if err := queryOperations(db); err != nil {
		fmt.Printf("查询操作失败: %v\n", err)
		return
	}

	// 5. 事务操作
	fmt.Println("\n5. 事务操作:")
	if err := transactionOperations(db); err != nil {
		fmt.Printf("事务操作失败: %v\n", err)
		return
	}

	fmt.Println("\n=== 所有操作执行完成 ===")
}

// connectDatabase 连接SQLite数据库
// 创建SQLite数据库连接并返回实例
// SQLite使用文件路径连接，无需网络配置
// 返回：
//   - *sqlite.Instance: SQLite数据库实例
//   - error: 连接失败时的错误信息
func connectDatabase() (*sqlite.Instance, error) {
	// 创建 SQLite 连接配置
	// SQLite使用文件路径而非网络地址
	config := jcbaseGo.SqlLiteStruct{
		DbFile:        "./test_sqlite.db", // SQLite数据库文件路径
		TablePrefix:   "",                 // 表名前缀
		SingularTable: false,              // 是否使用单数表名
	}

	// 连接数据库
	db := sqlite.New(config)

	// 检查是否有错误
	if len(db.Error()) > 0 {
		return nil, fmt.Errorf("连接SQLite数据库失败: %v", db.Error())
	}

	fmt.Println("SQLite数据库连接成功")
	return db, nil
}

// migrateTable 自动迁移表结构
// 根据User模型自动创建或更新数据库表结构
// SQLite的迁移与MySQL类似，但数据类型支持略有不同
// 参数：
//   - db: SQLite数据库实例
//
// 返回：
//   - error: 迁移失败时的错误信息
func migrateTable(db *sqlite.Instance) error {
	gormDB := db.GetDb()
	if gormDB == nil {
		return fmt.Errorf("SQLite数据库连接为空")
	}

	// 自动迁移 User 表
	err := gormDB.AutoMigrate(&User{})
	if err != nil {
		return fmt.Errorf("自动迁移失败: %v", err)
	}

	fmt.Println("SQLite表结构迁移成功")
	return nil
}

// basicCRUD 基本 CRUD 操作
// 演示创建、查询、更新、删除操作
// SQLite的CRUD操作与MySQL完全兼容
// 参数：
//   - db: SQLite数据库实例
//
// 返回：
//   - error: 操作失败时的错误信息
func basicCRUD(db *sqlite.Instance) error {
	gormDB := db.GetDb()
	if gormDB == nil {
		return fmt.Errorf("SQLite数据库连接为空")
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
		return fmt.Errorf("创建用户失败: %v", err)
	}
	fmt.Printf("创建用户成功，ID: %d\n", user.ID)

	// 查询用户
	var foundUser User
	err = gormDB.First(&foundUser, user.ID).Error
	if err != nil {
		return fmt.Errorf("查询用户失败: %v", err)
	}
	fmt.Printf("查询用户成功: %+v\n", foundUser)

	// 更新用户
	updates := map[string]interface{}{
		"age":    26,
		"status": 2,
	}
	err = gormDB.Model(&foundUser).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("更新用户失败: %v", err)
	}
	fmt.Println("更新用户成功")

	// 删除用户
	err = gormDB.Delete(&foundUser).Error
	if err != nil {
		return fmt.Errorf("删除用户失败: %v", err)
	}
	fmt.Println("删除用户成功")

	return nil
}

// queryOperations 查询操作
// 演示批量创建、条件查询、排序、分页、统计等查询操作
// SQLite支持大部分标准SQL查询语法
// 参数：
//   - db: SQLite数据库实例
//
// 返回：
//   - error: 操作失败时的错误信息
func queryOperations(db *sqlite.Instance) error {
	gormDB := db.GetDb()
	if gormDB == nil {
		return fmt.Errorf("SQLite数据库连接为空")
	}

	// 批量创建用户
	users := []User{
		{Name: "李四", Email: "lisi@example.com", Age: 30, Status: 1},
		{Name: "王五", Email: "wangwu@example.com", Age: 28, Status: 1},
		{Name: "赵六", Email: "zhaoliu@example.com", Age: 35, Status: 2},
		{Name: "钱七", Email: "qianqi@example.com", Age: 22, Status: 1},
	}

	err := gormDB.Create(&users).Error
	if err != nil {
		return fmt.Errorf("批量创建用户失败: %v", err)
	}
	fmt.Println("批量创建用户成功")

	// 查询所有用户
	var allUsers []User
	err = gormDB.Find(&allUsers).Error
	if err != nil {
		return fmt.Errorf("查询所有用户失败: %v", err)
	}
	fmt.Printf("所有用户数量: %d\n", len(allUsers))

	// 条件查询
	var activeUsers []User
	err = gormDB.Where("status = ?", 1).Find(&activeUsers).Error
	if err != nil {
		return fmt.Errorf("条件查询失败: %v", err)
	}
	fmt.Printf("活跃用户数量: %d\n", len(activeUsers))

	// 排序查询
	var sortedUsers []User
	err = gormDB.Order("age desc").Find(&sortedUsers).Error
	if err != nil {
		return fmt.Errorf("排序查询失败: %v", err)
	}
	fmt.Printf("按年龄降序排列的用户数量: %d\n", len(sortedUsers))

	// 分页查询
	var pagedUsers []User
	err = gormDB.Offset(0).Limit(2).Find(&pagedUsers).Error
	if err != nil {
		return fmt.Errorf("分页查询失败: %v", err)
	}
	fmt.Printf("分页查询结果数量: %d\n", len(pagedUsers))

	// 统计查询
	var count int64
	err = gormDB.Model(&User{}).Where("age > ?", 25).Count(&count).Error
	if err != nil {
		return fmt.Errorf("统计查询失败: %v", err)
	}
	fmt.Printf("年龄大于25的用户数量: %d\n", count)

	return nil
}

// transactionOperations 事务操作
// 演示数据库事务操作，确保数据一致性
// SQLite支持完整的事务操作，与MySQL兼容
// 参数：
//   - db: SQLite数据库实例
//
// 返回：
//   - error: 事务操作失败时的错误信息
func transactionOperations(db *sqlite.Instance) error {
	gormDB := db.GetDb()
	if gormDB == nil {
		return fmt.Errorf("SQLite数据库连接为空")
	}

	// 事务操作
	err := gormDB.Transaction(func(tx *gorm.DB) error {
		// 创建用户1
		user1 := User{Name: "事务用户1", Email: "tx1@example.com", Age: 30, Status: 1}
		if err := tx.Create(&user1).Error; err != nil {
			return fmt.Errorf("创建用户1失败: %v", err)
		}

		// 创建用户2
		user2 := User{Name: "事务用户2", Email: "tx2@example.com", Age: 25, Status: 1}
		if err := tx.Create(&user2).Error; err != nil {
			return fmt.Errorf("创建用户2失败: %v", err)
		}

		// 模拟业务逻辑错误，触发回滚（注释掉，实际使用时根据需要开启）
		// return fmt.Errorf("模拟业务错误")

		fmt.Println("事务内操作执行成功")
		return nil
	})

	if err != nil {
		return fmt.Errorf("事务操作失败: %v", err)
	}

	fmt.Println("事务操作成功")
	return nil
}

// SQLite使用注意事项：
// 1. 文件路径：SQLite数据库文件路径可以是相对路径或绝对路径
// 2. 并发访问：SQLite支持并发读取，但写入时需要锁定
// 3. 数据类型：SQLite使用动态类型系统，但GORM会进行类型映射
// 4. 性能优化：对于大量数据操作，建议使用事务批量处理
// 5. 迁移兼容：SQLite的迁移语法与MySQL略有不同，但GORM会自动处理
