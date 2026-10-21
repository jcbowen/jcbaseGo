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

	// Search 搜索日志内容，支持分页和过滤
	Search(keyword string, page, pageSize int, filters map[string]interface{}) ([]*LogEntry, int, error)

	// Cleanup 清理过期日志
	Cleanup(before time.Time) error

	// GetStats 获取统计信息
	GetStats() (map[string]interface{}, error)

	// GetMethods 获取HTTP方法统计
	GetMethods() (map[string]int, error)

	// GetStatusCodes 获取状态码统计
	GetStatusCodes() (map[int]int, error)

	// GetServices 获取服务标识统计（多服务共享存储时用于区分日志来源）
	GetServices() (map[string]int, error)

	// Close 关闭存储
	Close() error
}

// isCleanupBoundaryValid 判断过期清理时间点是否可用
// 各存储实现的 Cleanup 删除条件都是 timestamp < before，
// before 为零值或不早于当前时间时会命中全部记录（等同于清空日志），必须拒绝
// 参数：
//   - before: 清理时间点
//
// 返回值：
//   - bool: true 表示可以安全执行清理
//
// 使用示例：
//
//	if !isCleanupBoundaryValid(before) { return fmt.Errorf("非法的清理时间点 %v", before) }
func isCleanupBoundaryValid(before time.Time) bool {
	return !before.IsZero() && before.Before(time.Now())
}
