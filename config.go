package jcbaseGo

import (
	"encoding/json"
	"errors"
	"github.com/jcbowen/jcbaseGo/helper"
	"log"
	"os"
	"path/filepath"
	"reflect"
)

// New 初始化配置
func New(opt Option) *Option {
	if opt.ConfigData != nil {
		opt.checkConfig()
	}
	return &opt
}

// checkConfig 将json配置信息初始化到Config中
func (opt *Option) checkConfig() {
	if reflect.TypeOf(opt.ConfigData) == nil {
		log.Panic("配置信息不能为空")
		return
	}

	// 为Config添加默认值
	if err := helper.CheckAndSetDefault(opt.ConfigData); err != nil {
		log.Panic(err)
	}

	// 为参数添加默认值
	if err := helper.CheckAndSetDefault(*opt); err != nil {
		log.Panic(err)
	}

	// 获取json配置文件的绝对路径
	fileNameFull, err := filepath.Abs(opt.ConfigFile)
	if err != nil {
		log.Panic(err)
	}

	// json配置文件不存在，根据默认配置生成json配置文件
	if !helper.FileExists(fileNameFull) {
		// 如果配置文件不存在，则创建配置文件
		file, _ := json.MarshalIndent(opt.ConfigData, "", " ")
		err := helper.CreateFile(fileNameFull, file, 0755, false)
		if err != nil {
			log.Panic(err)
		}
		err = errors.New("配置文件不存在，已创建默认配置文件，请修改配置文件后重启程序！\n配置文件路径：" + fileNameFull)
		log.Panic(err)
	}

	// 读取json配置文件
	file, fErr := os.ReadFile(fileNameFull)
	if fErr != nil {
		log.Panic(fErr)
	}
	fileDataString := string(file)

	err = json.Unmarshal([]byte(fileDataString), &opt.ConfigData)
	if err != nil {
		log.Panic(err)
	}

	Config = opt.ConfigData
}

// GetConfig 获取配置信息
func (opt *Option) GetConfig() *interface{} {
	return &opt.ConfigData
}

// GetConfigStruct 将Option.ConfigData赋值到自定义结构体中
func (opt *Option) ConfigToStruct(configStruct interface{}) {
	// 由于opt.ConfigData是interface，在json解析后会变为map，所以这里需要进行类型转换
	helper.MapToStruct(opt.ConfigData, configStruct)
}

// ----- 终结方法 ----- /

func (opt *Option) GetConfigOption() Option {
	return *opt
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
		for _, err := range v {
			if err != nil {
				log.Panic(err)
			}
		}
	default:
		// If the type is not error or []error, do nothing
	}
}

// ------ 弃用函数 ------ /

// ConfigStruct 配置信息
// Deprecated: 已经弃用，请自定义数据配置结构
type ConfigStruct = DefaultConfigStruct

// Get 获取配置信息(兼容旧的写法)
// Deprecated: 请使用
func (c *DefaultConfigStruct) Get() *DefaultConfigStruct {
	New(Option{
		ConfigFile: "",
	}).checkConfig()
	return c
}
