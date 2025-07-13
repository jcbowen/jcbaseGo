package ExampleRedis

import (
	"fmt"
	"log"
	"time"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/redis"
)

// User 用户结构体示例
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	fmt.Println("=== Redis Cache 使用示例 ===")

	// 1. 初始化Redis连接
	conf := jcbaseGo.RedisStruct{
		Host:     "127.0.0.1",
		Port:     "6379",
		Password: "",
		Db:       "0",
	}

	redisInstance := redis.New(conf)
	fmt.Println("✓ Redis连接初始化成功")

	// 2. 创建缓存管理器 - 基础用法
	cache, err := redis.NewCache(redisInstance)
	if err != nil {
		log.Fatalf("创建缓存管理器失败: %v", err)
	}
	fmt.Println("✓ 基础缓存管理器创建成功")

	// 3. 创建带配置的缓存管理器
	userCache, err := redis.NewCache(
		redisInstance,
		redis.WithPrefix("user"),
		redis.WithExpire(30*time.Minute),
	)
	if err != nil {
		log.Fatalf("创建用户缓存管理器失败: %v", err)
	}
	fmt.Println("✓ 用户缓存管理器创建成功")

	// 4. 基本缓存操作
	fmt.Println("\n--- 基本缓存操作 ---")

	// 设置字符串缓存
	err = cache.Set("greeting", "Hello, World!", 5*time.Minute)
	if err != nil {
		log.Printf("设置字符串缓存失败: %v", err)
	} else {
		fmt.Println("✓ 设置字符串缓存成功")
	}

	// 获取字符串缓存
	greeting, err := cache.GetString("greeting")
	if err != nil {
		log.Printf("获取字符串缓存失败: %v", err)
	} else {
		fmt.Printf("✓ 获取字符串缓存: %s\n", greeting)
	}

	// 获取不存在的键（带默认值）
	notFound, err := cache.GetString("not_exists", "默认值")
	if err != nil {
		log.Printf("获取不存在的键失败: %v", err)
	} else {
		fmt.Printf("✓ 获取不存在的键（带默认值）: %s\n", notFound)
	}

	// 5. 结构体缓存操作
	fmt.Println("\n--- 结构体缓存操作 ---")

	// 创建用户数据
	user := &User{
		ID:   1,
		Name: "张三",
		Age:  25,
	}

	// 设置结构体缓存
	err = userCache.Set("user:1", user)
	if err != nil {
		log.Printf("设置用户缓存失败: %v", err)
	} else {
		fmt.Println("✓ 设置用户缓存成功")
	}

	// 获取结构体缓存
	var retrievedUser User
	err = userCache.GetStruct("user:1", &retrievedUser)
	if err != nil {
		log.Printf("获取用户缓存失败: %v", err)
	} else {
		fmt.Printf("✓ 获取用户缓存: ID=%d, Name=%s, Age=%d\n",
			retrievedUser.ID, retrievedUser.Name, retrievedUser.Age)
	}

	// 获取不存在的结构体（带默认值）
	var defaultUser User
	defaultValue := User{ID: 0, Name: "默认用户", Age: 0}
	err = userCache.GetStruct("user:999", &defaultUser, defaultValue)
	if err != nil {
		log.Printf("获取不存在的用户缓存失败: %v", err)
	} else {
		fmt.Printf("✓ 获取不存在的用户缓存（带默认值）: ID=%d, Name=%s, Age=%d\n",
			defaultUser.ID, defaultUser.Name, defaultUser.Age)
	}

	// 6. 缓存键管理
	fmt.Println("\n--- 缓存键管理 ---")

	// 检查键是否存在
	exists, err := cache.Exists("greeting")
	if err != nil {
		log.Printf("检查键存在性失败: %v", err)
	} else {
		fmt.Printf("✓ 键 'greeting' 存在: %t\n", exists)
	}

	// 获取所有用户缓存键
	keys, err := userCache.GetKeys("*")
	if err != nil {
		log.Printf("获取用户缓存键失败: %v", err)
	} else {
		fmt.Printf("✓ 用户缓存键列表: %v\n", keys)
	}

	// 7. 缓存统计
	fmt.Println("\n--- 缓存统计 ---")

	stats, err := userCache.GetStats()
	if err != nil {
		log.Printf("获取缓存统计失败: %v", err)
	} else {
		fmt.Printf("✓ 用户缓存统计: %+v\n", stats)
	}

	// 8. 缓存清理
	fmt.Println("\n--- 缓存清理 ---")

	// 删除单个键
	err = cache.Del("greeting")
	if err != nil {
		log.Printf("删除键失败: %v", err)
	} else {
		fmt.Println("✓ 删除键 'greeting' 成功")
	}

	// 验证删除结果
	exists, err = cache.Exists("greeting")
	if err != nil {
		log.Printf("检查删除结果失败: %v", err)
	} else {
		fmt.Printf("✓ 键 'greeting' 删除后存在: %t\n", exists)
	}

	// 清理所有用户缓存
	err = userCache.ClearAll()
	if err != nil {
		log.Printf("清理用户缓存失败: %v", err)
	} else {
		fmt.Println("✓ 清理所有用户缓存成功")
	}

	// 验证清理结果
	stats, err = userCache.GetStats()
	if err != nil {
		log.Printf("获取清理后统计失败: %v", err)
	} else {
		fmt.Printf("✓ 清理后用户缓存统计: %+v\n", stats)
	}

	fmt.Println("\n=== 示例完成 ===")
}
