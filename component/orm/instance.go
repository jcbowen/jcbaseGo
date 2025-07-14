package orm

import (
	"gorm.io/gorm"
)

// Instance 定义数据库实例接口
type Instance interface {
	// GetDb 获取数据库连接
	GetDb() *gorm.DB
	// GetConf 获取数据库配置信息，类型：jcbaseGo.DbStruct 或 jcbaseGo.SqlLiteStruct
	GetConf() interface{}
}
