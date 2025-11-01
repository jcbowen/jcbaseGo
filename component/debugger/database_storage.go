package debugger

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// DatabaseStorage 数据库存储器实现
// 使用GORM将调试日志保存到数据库中
type DatabaseStorage struct {
	db        *gorm.DB // 数据库连接
	tableName string   // 数据表名
	maxSize   int      // 最大存储条目数
}

// LogEntryModel 数据库模型结构
// 对应数据库中的日志条目表
type LogEntryModel struct {
	ID         string    `gorm:"column:id;type:VARCHAR(64);primaryKey" json:"id"`
	Timestamp  time.Time `gorm:"column:timestamp;type:DATETIME;index" json:"timestamp"`
	Method     string    `gorm:"column:method;type:VARCHAR(10);index" json:"method"`
	URL        string    `gorm:"column:url;type:TEXT" json:"url"`
	StatusCode int       `gorm:"column:status_code;type:INT;index" json:"status_code"`
	Duration   int64     `gorm:"column:duration;type:BIGINT" json:"duration"` // 存储纳秒数
	ClientIP   string    `gorm:"column:client_ip;type:VARCHAR(45)" json:"client_ip"`
	UserAgent  string    `gorm:"column:user_agent;type:TEXT" json:"user_agent"`
	RequestID  string    `gorm:"column:request_id;type:VARCHAR(64);index" json:"request_id"`

	// JSON格式存储的字段
	RequestHeaders  string `gorm:"column:request_headers;type:JSON" json:"request_headers"`
	QueryParams     string `gorm:"column:query_params;type:JSON" json:"query_params"`
	RequestBody     string `gorm:"column:request_body;type:LONGTEXT" json:"request_body"`
	ResponseHeaders string `gorm:"column:response_headers;type:JSON" json:"response_headers"`
	ResponseBody    string `gorm:"column:response_body;type:LONGTEXT" json:"response_body"`
	SessionData     string `gorm:"column:session_data;type:JSON" json:"session_data"`
	Error           string `gorm:"column:error;type:TEXT" json:"error"`

	// 索引字段
	URLHash string `gorm:"column:url_hash;type:VARCHAR(32);index" json:"url_hash"` // URL的MD5哈希，用于快速搜索
}

// TableName 实现自定义表名
func (LogEntryModel) TableName() string {
	return "debug_logs" // 默认表名
}

// NewDatabaseStorage 创建新的数据库存储器
// db: GORM数据库连接
// maxSize: 最大存储条目数（0表示无限制）
// tableName: 数据表名（可选，默认为"debug_logs"）
func NewDatabaseStorage(db *gorm.DB, maxSize int, tableName ...string) (*DatabaseStorage, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库连接不能为空")
	}

	tName := "debug_logs"
	if len(tableName) > 0 && tableName[0] != "" {
		tName = tableName[0]
	}

	ds := &DatabaseStorage{
		db:        db,
		tableName: tName,
		maxSize:   maxSize,
	}

	// 自动迁移表结构
	if err := ds.autoMigrate(); err != nil {
		return nil, fmt.Errorf("自动迁移表结构失败: %v", err)
	}

	return ds, nil
}

// autoMigrate 自动迁移表结构
func (ds *DatabaseStorage) autoMigrate() error {
	// 创建临时模型用于迁移
	model := &LogEntryModel{}

	// 使用Set设置表名
	return ds.db.Table(ds.tableName).AutoMigrate(model)
}

// Save 保存日志条目到数据库
func (ds *DatabaseStorage) Save(entry *LogEntry) error {
	// 检查是否超过最大存储限制
	if ds.maxSize > 0 {
		// 获取当前记录总数
		var count int64
		if err := ds.db.Table(ds.tableName).Count(&count).Error; err != nil {
			return fmt.Errorf("获取记录总数失败: %v", err)
		}

		// 如果超过最大限制，删除最旧的条目
		if count >= int64(ds.maxSize) {
			// 查找最旧的记录ID
			var oldestModel LogEntryModel
			result := ds.db.Table(ds.tableName).Order("timestamp ASC").First(&oldestModel)
			if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
				return fmt.Errorf("查找最旧记录失败: %v", result.Error)
			}

			// 删除最旧的记录
			if result.Error == nil {
				if err := ds.db.Table(ds.tableName).Where("id = ?", oldestModel.ID).Delete(&LogEntryModel{}).Error; err != nil {
					return fmt.Errorf("删除最旧记录失败: %v", err)
				}
			}
		}
	}

	// 转换为数据库模型
	model, err := ds.entryToModel(entry)
	if err != nil {
		return fmt.Errorf("转换日志条目失败: %v", err)
	}

	// 保存到数据库
	result := ds.db.Table(ds.tableName).Create(model)
	if result.Error != nil {
		return fmt.Errorf("保存日志条目失败: %v", result.Error)
	}

	return nil
}

// FindByID 根据ID查找日志条目
func (ds *DatabaseStorage) FindByID(id string) (*LogEntry, error) {
	var model LogEntryModel

	result := ds.db.Table(ds.tableName).Where("id = ?", id).First(&model)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("未找到ID为 %s 的日志条目", id)
		}
		return nil, fmt.Errorf("查询日志条目失败: %v", result.Error)
	}

	// 转换为LogEntry
	entry, err := ds.modelToEntry(&model)
	if err != nil {
		return nil, fmt.Errorf("转换数据库模型失败: %v", err)
	}

	return entry, nil
}

// FindAll 查找所有日志条目，支持分页和过滤
func (ds *DatabaseStorage) FindAll(page, pageSize int, filters map[string]interface{}) ([]*LogEntry, int, error) {
	// 构建查询
	db := ds.db.Table(ds.tableName)

	// 应用过滤器
	db = ds.applyFilters(db, filters)

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("计算总数失败: %v", err)
	}

	// 分页查询
	var models []LogEntryModel
	offset := (page - 1) * pageSize

	result := db.Order("timestamp DESC").Offset(offset).Limit(pageSize).Find(&models)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("查询日志条目失败: %v", result.Error)
	}

	// 转换为LogEntry列表
	entries := make([]*LogEntry, len(models))
	for i, model := range models {
		entry, err := ds.modelToEntry(&model)
		if err != nil {
			return nil, 0, fmt.Errorf("转换数据库模型失败: %v", err)
		}
		entries[i] = entry
	}

	return entries, int(total), nil
}

// Search 搜索日志内容
func (ds *DatabaseStorage) Search(keyword string, page, pageSize int) ([]*LogEntry, int, error) {
	// 构建搜索查询
	db := ds.db.Table(ds.tableName)

	// 使用LIKE进行模糊搜索
	keyword = "%" + strings.ToLower(keyword) + "%"
	db = db.Where("LOWER(url) LIKE ? OR LOWER(request_body) LIKE ? OR LOWER(response_body) LIKE ? OR LOWER(error) LIKE ? OR LOWER(user_agent) LIKE ?",
		keyword, keyword, keyword, keyword, keyword)

	// 计算总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("计算总数失败: %v", err)
	}

	// 分页查询
	var models []LogEntryModel
	offset := (page - 1) * pageSize

	result := db.Order("timestamp DESC").Offset(offset).Limit(pageSize).Find(&models)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("查询日志条目失败: %v", result.Error)
	}

	// 转换为LogEntry列表
	entries := make([]*LogEntry, len(models))
	for i, model := range models {
		entry, err := ds.modelToEntry(&model)
		if err != nil {
			return nil, 0, fmt.Errorf("转换数据库模型失败: %v", err)
		}
		entries[i] = entry
	}

	return entries, int(total), nil
}

// Cleanup 清理过期日志
func (ds *DatabaseStorage) Cleanup(before time.Time) error {
	result := ds.db.Table(ds.tableName).Where("timestamp < ?", before).Delete(&LogEntryModel{})
	if result.Error != nil {
		return fmt.Errorf("清理过期日志失败: %v", result.Error)
	}

	return nil
}

// applyFilters 应用过滤器到查询
func (ds *DatabaseStorage) applyFilters(db *gorm.DB, filters map[string]interface{}) *gorm.DB {
	for key, value := range filters {
		switch key {
		case "method":
			db = db.Where("method = ?", value)
		case "status_code":
			db = db.Where("status_code = ?", value)
		case "url_contains":
			db = db.Where("url LIKE ?", "%"+value.(string)+"%")
		case "start_time":
			db = db.Where("timestamp >= ?", value)
		case "end_time":
			db = db.Where("timestamp <= ?", value)
		case "client_ip":
			db = db.Where("client_ip LIKE ?", "%"+value.(string)+"%")
		case "has_error":
			if value.(bool) {
				db = db.Where("error != ''")
			} else {
				db = db.Where("error = ''")
			}
		case "min_duration":
			db = db.Where("duration >= ?", value.(time.Duration).Nanoseconds())
		case "max_duration":
			db = db.Where("duration <= ?", value.(time.Duration).Nanoseconds())
		}
	}

	return db
}

// entryToModel 将LogEntry转换为数据库模型
func (ds *DatabaseStorage) entryToModel(entry *LogEntry) (*LogEntryModel, error) {
	model := &LogEntryModel{
		ID:           entry.ID,
		Timestamp:    entry.Timestamp,
		Method:       entry.Method,
		URL:          entry.URL,
		StatusCode:   entry.StatusCode,
		Duration:     entry.Duration.Nanoseconds(),
		ClientIP:     entry.ClientIP,
		UserAgent:    entry.UserAgent,
		RequestID:    entry.RequestID,
		RequestBody:  entry.RequestBody,
		ResponseBody: entry.ResponseBody,
		Error:        entry.Error,
	}

	// 转换JSON字段
	if headers, err := json.Marshal(entry.RequestHeaders); err == nil {
		model.RequestHeaders = string(headers)
	}

	if params, err := json.Marshal(entry.QueryParams); err == nil {
		model.QueryParams = string(params)
	}

	if headers, err := json.Marshal(entry.ResponseHeaders); err == nil {
		model.ResponseHeaders = string(headers)
	}

	if sessionData, err := json.Marshal(entry.SessionData); err == nil {
		model.SessionData = string(sessionData)
	}

	// 生成URL哈希（用于快速搜索）
	model.URLHash = ds.generateURLHash(entry.URL)

	return model, nil
}

// modelToEntry 将数据库模型转换为LogEntry
func (ds *DatabaseStorage) modelToEntry(model *LogEntryModel) (*LogEntry, error) {
	entry := &LogEntry{
		ID:           model.ID,
		Timestamp:    model.Timestamp,
		Method:       model.Method,
		URL:          model.URL,
		StatusCode:   model.StatusCode,
		Duration:     time.Duration(model.Duration),
		ClientIP:     model.ClientIP,
		UserAgent:    model.UserAgent,
		RequestID:    model.RequestID,
		RequestBody:  model.RequestBody,
		ResponseBody: model.ResponseBody,
		Error:        model.Error,
	}

	// 解析JSON字段
	if model.RequestHeaders != "" {
		if err := json.Unmarshal([]byte(model.RequestHeaders), &entry.RequestHeaders); err != nil {
			return nil, fmt.Errorf("解析请求头失败: %v", err)
		}
	}

	if model.QueryParams != "" {
		if err := json.Unmarshal([]byte(model.QueryParams), &entry.QueryParams); err != nil {
			return nil, fmt.Errorf("解析查询参数失败: %v", err)
		}
	}

	if model.ResponseHeaders != "" {
		if err := json.Unmarshal([]byte(model.ResponseHeaders), &entry.ResponseHeaders); err != nil {
			return nil, fmt.Errorf("解析响应头失败: %v", err)
		}
	}

	if model.SessionData != "" {
		if err := json.Unmarshal([]byte(model.SessionData), &entry.SessionData); err != nil {
			return nil, fmt.Errorf("解析会话数据失败: %v", err)
		}
	}

	return entry, nil
}

// generateURLHash 生成URL的哈希值（简化实现）
func (ds *DatabaseStorage) generateURLHash(url string) string {
	// 这里使用简单的哈希算法，实际项目中可以使用更复杂的哈希
	// 这里简化实现，使用URL的前32个字符
	if len(url) <= 32 {
		return url
	}
	return url[:32]
}

// GetStats 获取存储统计信息
func (ds *DatabaseStorage) GetStats() (map[string]interface{}, error) {
	// 计算总数
	var total int64
	if err := ds.db.Table(ds.tableName).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("计算总数失败: %v", err)
	}

	// 计算错误率
	var errorCount int64
	if err := ds.db.Table(ds.tableName).Where("error != ''").Count(&errorCount).Error; err != nil {
		return nil, fmt.Errorf("计算错误数失败: %v", err)
	}

	// 计算平均响应时间
	var avgDuration int64
	if err := ds.db.Table(ds.tableName).Select("AVG(duration)").Row().Scan(&avgDuration); err != nil {
		avgDuration = 0
	}

	// 估算存储大小（数据库存储难以精确计算，使用近似估算）
	var estimatedSize int64
	if total > 0 {
		// 获取一个样本条目来估算平均大小
		var sampleModel LogEntryModel
		if err := ds.db.Table(ds.tableName).First(&sampleModel).Error; err == nil {
			// 估算每个条目的平均大小
			// 这里使用一个保守的估算因子，实际大小可能更大
			estimatedSize = int64(len(sampleModel.ID) + len(sampleModel.URL) + len(sampleModel.Method) +
				len(sampleModel.RequestBody) + len(sampleModel.ResponseBody) + len(sampleModel.Error) +
				len(sampleModel.UserAgent) + len(sampleModel.ClientIP) + len(sampleModel.RequestID) +
				len(sampleModel.RequestHeaders) + len(sampleModel.ResponseHeaders) +
				len(sampleModel.QueryParams) + len(sampleModel.SessionData))

			// 乘以总条目数，并考虑数据库索引和元数据的开销
			estimatedSize = estimatedSize * total * 2 // 乘以2考虑数据库开销
		}
	}

	// 统一字段名，移除重复的total_entries字段
	stats := map[string]interface{}{
		"total_requests": total,                                               // 总请求数
		"storage_size":   fmt.Sprintf("%.2f KB", float64(estimatedSize)/1024), // 存储大小（估算显示）
		"max_size":       ds.maxSize,                                          // 最大存储条目数
		"storage_type":   "database",                                          // 存储类型
		"table_name":     ds.tableName,                                        // 数据表名
	}

	// 添加错误统计信息
	if total > 0 {
		stats["error_rate"] = float64(errorCount) / float64(total)
		stats["error_count"] = errorCount
	}

	// 添加平均响应时间
	stats["avg_duration"] = time.Duration(avgDuration)

	return stats, nil
}

// GetMethods 获取所有HTTP方法统计
func (ds *DatabaseStorage) GetMethods() (map[string]int, error) {
	type MethodCount struct {
		Method string
		Count  int
	}

	var results []MethodCount
	if err := ds.db.Table(ds.tableName).
		Select("method, COUNT(*) as count").
		Group("method").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("获取方法统计失败: %v", err)
	}

	methods := make(map[string]int)
	for _, result := range results {
		methods[result.Method] = result.Count
	}

	return methods, nil
}

// GetStatusCodes 获取所有状态码统计
func (ds *DatabaseStorage) GetStatusCodes() (map[int]int, error) {
	type StatusCount struct {
		StatusCode int
		Count      int
	}

	var results []StatusCount
	if err := ds.db.Table(ds.tableName).
		Select("status_code, COUNT(*) as count").
		Group("status_code").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("获取状态码统计失败: %v", err)
	}

	statusCodes := make(map[int]int)
	for _, result := range results {
		statusCodes[result.StatusCode] = result.Count
	}

	return statusCodes, nil
}

// Close 关闭数据库存储器
func (ds *DatabaseStorage) Close() error {
	// GORM连接池会自动管理，这里不需要特殊操作
	return nil
}

// CreateIndexes 创建索引（优化查询性能）
func (ds *DatabaseStorage) CreateIndexes() error {
	// 这里可以添加创建额外索引的逻辑
	// 由于GORM的AutoMigrate已经创建了基本索引，这里主要创建复合索引

	// 示例：创建时间戳和方法的复合索引
	// 实际项目中可以根据查询模式创建合适的索引
	return nil
}
