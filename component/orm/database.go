package orm

import (
	"gorm.io/gorm"
)

// Database 定义数据库接口
type Database interface {
	// GetDb 获取数据库连接
	GetDb() *gorm.DB
}

// DatabaseInstance 数据库实例结构体
type DatabaseInstance struct {
	Database
}

// NewDatabaseInstance 创建新的数据库实例
func NewDatabaseInstance(db Database) *DatabaseInstance {
	return &DatabaseInstance{
		Database: db,
	}
}
