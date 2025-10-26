package message

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Config 消息配置结构
type Config struct {
	TemplatePath    string `json:"template_path"`    // 模板文件路径（可选）
	TemplateContent string `json:"template_content"` // 模板内容（优先使用）
	AutoRedirect    bool   `json:"auto_redirect"`    // 是否自动跳转
	DefaultType     string `json:"default_type"`     // 默认消息类型
	TitlePrefix     string `json:"title_prefix"`     // 标题前缀
	TitleSuffix     string `json:"title_suffix"`     // 标题后缀
}

// ConfigManager 配置管理器
type ConfigManager struct {
	config     *Config
	configPath string
	mutex      sync.RWMutex
}

// NewMessageConfigManager 创建配置管理器
func NewMessageConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
		config: &Config{
			AutoRedirect: true,
			DefaultType:  "info",
			TitlePrefix:  "",
			TitleSuffix:  "",
		},
	}
}

// LoadConfig 加载配置
func (m *ConfigManager) LoadConfig() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 如果配置文件不存在，使用默认配置
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	m.config = &config
	return nil
}

// SaveConfig 保存配置
func (m *ConfigManager) SaveConfig() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 确保目录存在
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// GetConfig 获取配置
func (m *ConfigManager) GetConfig() *Config {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 返回配置的副本
	configCopy := *m.config
	return &configCopy
}

// UpdateConfig 更新配置
func (m *ConfigManager) UpdateConfig(updates map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 将配置转换为map以便更新
	configMap := make(map[string]interface{})
	data, _ := json.Marshal(m.config)
	_ = json.Unmarshal(data, &configMap)

	// 应用更新
	for key, value := range updates {
		configMap[key] = value
	}

	// 转换回结构体
	data, err := json.Marshal(configMap)
	if err != nil {
		return err
	}

	var newConfig Config
	if err := json.Unmarshal(data, &newConfig); err != nil {
		return err
	}

	m.config = &newConfig
	return nil
}

// ApplyConfig 应用配置到渲染器
func (m *ConfigManager) ApplyConfig() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 优先使用模板内容
	if m.config.TemplateContent != "" {
		return GetDefaultRenderer().SetTemplate(m.config.TemplateContent)
	}

	// 如果设置了模板文件路径，从文件加载
	if m.config.TemplatePath != "" {
		content, err := os.ReadFile(m.config.TemplatePath)
		if err != nil {
			return err
		}
		return GetDefaultRenderer().SetTemplate(string(content))
	}

	// 如果都没有设置，保持默认模板
	return nil
}

// 全局配置管理器实例
var globalConfigManager *ConfigManager

func init() {
	// 默认配置文件路径
	globalConfigManager = NewMessageConfigManager("./config/message.json")

	// 尝试加载配置
	_ = globalConfigManager.LoadConfig()

	// 应用配置
	_ = globalConfigManager.ApplyConfig()
}

// GetGlobalConfigManager 获取全局配置管理器
func GetGlobalConfigManager() *ConfigManager {
	return globalConfigManager
}

// SetGlobalConfigPath 设置全局配置文件路径
func SetGlobalConfigPath(path string) {
	globalConfigManager = NewMessageConfigManager(path)
	_ = globalConfigManager.LoadConfig()
	_ = globalConfigManager.ApplyConfig()
}

// ConfigureMessage 配置消息系统
func ConfigureMessage(options ...func(*Config)) {
	config := GetGlobalConfigManager().GetConfig()

	for _, option := range options {
		option(config)
	}

	// 保存并应用配置
	_ = GetGlobalConfigManager().UpdateConfig(map[string]interface{}{
		"template_content": config.TemplateContent,
		"template_path":    config.TemplatePath,
		"auto_redirect":    config.AutoRedirect,
		"default_type":     config.DefaultType,
		"title_prefix":     config.TitlePrefix,
		"title_suffix":     config.TitleSuffix,
	})

	_ = GetGlobalConfigManager().SaveConfig()
	_ = GetGlobalConfigManager().ApplyConfig()
}

// 配置选项函数

// WithTemplateContent 设置模板内容
func WithTemplateContent(content string) func(*Config) {
	return func(c *Config) {
		c.TemplateContent = content
	}
}

// WithTemplatePath 设置模板文件路径
func WithTemplatePath(path string) func(*Config) {
	return func(c *Config) {
		c.TemplatePath = path
	}
}

// WithMessageAutoRedirect 设置自动跳转
func WithMessageAutoRedirect(auto bool) func(*Config) {
	return func(c *Config) {
		c.AutoRedirect = auto
	}
}

// WithDefaultMessageType 设置默认消息类型
func WithDefaultMessageType(msgType string) func(*Config) {
	return func(c *Config) {
		c.DefaultType = msgType
	}
}

// WithMessageTitlePrefix 设置标题前缀
func WithMessageTitlePrefix(prefix string) func(*Config) {
	return func(c *Config) {
		c.TitlePrefix = prefix
	}
}

// WithMessageTitleSuffix 设置标题后缀
func WithMessageTitleSuffix(suffix string) func(*Config) {
	return func(c *Config) {
		c.TitleSuffix = suffix
	}
}
