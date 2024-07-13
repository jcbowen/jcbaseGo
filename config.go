package jcbaseGo

import (
	"encoding/json"
	"github.com/jcbowen/jcbaseGo/helper"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// New 初始化配置
func New(opt Option) *Option {
	if opt.ConfigData != nil {
		opt.checkConfig()
	}
	return &opt
}

// GetConfig 获取配置信息
func (opt *Option) GetConfig() *interface{} {
	return &opt.ConfigData
}

// ConfigToStruct 将 Option.ConfigData 赋值到自定义结构体中
func (opt *Option) ConfigToStruct(configStruct interface{}) {
	helper.MapToStruct(opt.ConfigData, configStruct)
}

// GetConfigOption 获取配置选项
func (opt *Option) GetConfigOption() Option {
	return *opt
}

// checkConfig 将 json 配置信息初始化到 Config 中
func (opt *Option) checkConfig() {
	if reflect.TypeOf(opt.ConfigData) == nil {
		log.Fatalf("配置信息不能为空")
		return
	}

	opt.initializeConfigWithDefaults()
	fileNameFull := opt.getConfigFilePath()
	opt.createConfigFileIfNotExists(fileNameFull)
	opt.readConfigFile(fileNameFull)
	opt.updateConfigFromFile(fileNameFull)

	Config = opt.ConfigData
}

// PanicIfError 异常处理
// 如果err不为nil，则直接panic，用于省略if判断
func PanicIfError(err interface{}) {
	switch v := err.(type) {
	case error:
		if v != nil {
			log.Panic(v)
		}
	case []error:
		log.Panic(formatErrors(v))
	default:
		// If the type is not error or []error, do nothing
	}
}

// ----- 私有方法 ----- /

func (opt *Option) initializeConfigWithDefaults() {
	if err := helper.CheckAndSetDefault(opt.ConfigData); err != nil {
		log.Fatalf("初始化配置默认值错误: %v", err)
	}

	if err := helper.CheckAndSetDefault(opt); err != nil {
		log.Fatalf("初始化参数默认值错误: %v", err)
	}
}

func (opt *Option) getConfigFilePath() string {
	fileNameFull, err := filepath.Abs(opt.ConfigFile)
	if err != nil {
		log.Fatalf("获取配置文件路径错误: %v", err)
	}
	return fileNameFull
}

func (opt *Option) createConfigFileIfNotExists(fileNameFull string) {
	if !helper.FileExists(fileNameFull) {
		file, _ := json.MarshalIndent(opt.ConfigData, "", " ")
		err := helper.CreateFile(fileNameFull, file, 0755, false)
		if err != nil {
			log.Fatalf("创建配置文件错误: %v", err)
		}
		log.Fatalf("配置文件不存在，已创建默认配置文件，请修改配置文件后重启程序！\n配置文件路径：%s", fileNameFull)
	}
}

func (opt *Option) readConfigFile(fileNameFull string) {
	file, err := os.ReadFile(fileNameFull)
	if err != nil {
		log.Fatalf("读取配置文件错误: %v", err)
	}

	err = json.Unmarshal(file, &opt.ConfigData)
	if err != nil {
		log.Fatalf("解析配置文件错误: %v", err)
	}
}

func (opt *Option) updateConfigFromFile(fileNameFull string) {
	var fileConfig interface{}
	err := json.Unmarshal([]byte(fileNameFull), &fileConfig)
	if err != nil {
		log.Fatalf("解析文件配置错误: %v", err)
	}

	if updated := compareAndUpdateConfig(opt.ConfigData, fileConfig); updated {
		updatedFile, _ := json.MarshalIndent(opt.ConfigData, "", " ")
		err = os.WriteFile(fileNameFull, updatedFile, 0755)
		if err != nil {
			log.Fatalf("更新配置文件错误: %v", err)
		}
	}
}

// compareAndUpdateConfig 比较并更新配置文件
func compareAndUpdateConfig(structConfig, fileConfig interface{}) bool {
	structValue := reflect.ValueOf(structConfig).Elem()
	fileValue := reflect.ValueOf(fileConfig).Elem()
	updated := false

	for i := 0; i < structValue.NumField(); i++ {
		structField := structValue.Field(i)
		fileField := fileValue.FieldByName(structValue.Type().Field(i).Name)

		if !reflect.DeepEqual(structField.Interface(), fileField.Interface()) {
			fileField.Set(structField)
			updated = true
		}
	}

	return updated
}

// formatErrors 将 []error 格式化为单个字符串
func formatErrors(errs []error) string {
	var sb strings.Builder
	for _, err := range errs {
		sb.WriteString(err.Error())
		sb.WriteString("\n")
	}
	return sb.String()
}
