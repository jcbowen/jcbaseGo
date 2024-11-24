package jcbaseGo

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
)

// Config 实例化后配置信息将储存在此全局变量中
var Config interface{}

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

	// 初始化默认配置
	opt.initializeConfigWithDefaults()

	switch opt.ConfigType {
	case ConfigTypeFile: // json文件
		// 获取配置文件绝对路径
		fileNameFull := opt.getConfigFilePath()
		// 如果配置文件不存在，则创建
		opt.createConfigFileIfNotExists(fileNameFull)
		// 从文件中读取配置
		opt.readConfigFile(fileNameFull)
		// 配置结构体是有可能更新升级的，所以每次运行之后，应当更新一下配置文件
		opt.updateConfigFile(fileNameFull, true)
	case ConfigTypeCommand: // 命令行json
		// 执行脚本并获取JSON输出
		cmd := exec.Command("sh", "-c", opt.ConfigSource)

		// 获取标准输出和错误输出
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			log.Fatalf("执行脚本错误: %v\n错误输出: %s", err, stderr.String())
			return
		}

		output := stdout.Bytes()
		//log.Println("标准输出:", string(output))
		//log.Println("错误输出:", stderr.String())

		jsonStartIndex := bytes.Index(output, []byte("{"))
		if jsonStartIndex == -1 {
			log.Fatalf("输出中未找到JSON数据: %s", string(output))
			return
		}

		// 截取可能的JSON部分
		pureJSON := output[jsonStartIndex:]
		if err = json.Unmarshal(pureJSON, &opt.ConfigData); err != nil {
			log.Fatalf("JSON解析错误: %v\n原始数据: %s", err, pureJSON)
			return
		}
	default:
		log.Panic("错误的配置类型")
	}

	// 将配置信息写入全局变量
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
	case redis.Error:
		if !errors.Is(v, redis.Nil) {
			log.Panic(v)
		}
	default:
		// If the type is not error , []error or redis.Error, do nothing
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
	fileNameFull, err := filepath.Abs(opt.ConfigSource)
	if err != nil {
		log.Fatalf("获取配置文件路径错误: %v", err)
	}
	return fileNameFull
}

func (opt *Option) createConfigFileIfNotExists(fileNameFull string) {
	if !helper.NewFile(&helper.File{Path: fileNameFull}).Exists() {
		opt.updateConfigFile(fileNameFull, false)
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

// updateConfigFile 更新配置文件
func (opt *Option) updateConfigFile(fileNameFull string, overwrite bool) {
	fileData, _ := json.MarshalIndent(opt.ConfigData, "", " ")
	err := helper.NewFile(&helper.File{Path: fileNameFull}).CreateFile(fileData, overwrite)
	if err != nil {
		log.Fatalf("更新配置文件出错: %v", err)
	}
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
