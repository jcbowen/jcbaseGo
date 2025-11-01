package debugger

import "time"

// Storage 存储接口定义
// 支持多种存储方式：文件、内存、数据库
type Storage interface {
	// Save 保存日志条目
	Save(entry *LogEntry) error

	// FindByID 根据ID查找日志条目
	FindByID(id string) (*LogEntry, error)

	// FindAll 查找所有日志条目，支持分页和过滤
	FindAll(page, pageSize int, filters map[string]interface{}) ([]*LogEntry, int, error)

	// Search 搜索日志内容
	Search(keyword string, page, pageSize int) ([]*LogEntry, int, error)

	// Cleanup 清理过期日志
	Cleanup(before time.Time) error

	// GetStats 获取统计信息
	GetStats() (map[string]interface{}, error)

	// GetMethods 获取HTTP方法统计
	GetMethods() (map[string]int, error)

	// GetStatusCodes 获取状态码统计
	GetStatusCodes() (map[int]int, error)

	// Close 关闭存储
	Close() error
}
