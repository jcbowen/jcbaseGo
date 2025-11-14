package orm

import (
	"github.com/jcbowen/jcbaseGo/component/debugger"
	"gorm.io/gorm"
)

// Instance 定义数据库实例接口
type Instance interface {
	// GetDb 获取数据库连接
	GetDb() *gorm.DB
	// GetConf 获取数据库配置信息，类型：jcbaseGo.DbStruct 或 jcbaseGo.SqlLiteStruct
	GetConf() interface{}
	// SetDebuggerLogger 设置debugger日志记录器
	SetDebuggerLogger(debuggerLogger debugger.LoggerInterface)
	// GetDebuggerLogger 获取debugger日志记录器
	GetDebuggerLogger() debugger.LoggerInterface
}
