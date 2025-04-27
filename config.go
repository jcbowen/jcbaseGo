package jcbaseGo

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-ini/ini"
	"github.com/go-redis/redis/v8"
	"github.com/jcbowen/jcbaseGo/component/helper"
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
// 返回配置数据的指针，调用者需要确保配置数据不为空
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

// checkConfig 将配置信息初始化到 Config 中
func (opt *Option) checkConfig() {
	if reflect.TypeOf(opt.ConfigData) == nil {
		log.Fatalf("配置信息不能为空")
		return
	}

	// 初始化默认配置
	opt.initializeConfigWithDefaults()

	switch opt.ConfigType {
	case ConfigTypeJSON, ConfigTypeINI, ConfigTypeFile:
		// 获取配置文件绝对路径
		fileNameFull := opt.getConfigFilePath()
		// 如果配置文件不存在，则创建
		opt.createConfigFileIfNotExists(fileNameFull)
		// 从文件中读取配置
		opt.readConfigFile(fileNameFull)
		// 配置结构体是有可能更新升级的，所以每次运行之后，应当更新一下配置文件
		opt.updateConfigFile(fileNameFull, true)
	case ConfigTypeCommand:
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

	switch opt.ConfigType {
	case ConfigTypeINI:
		cfg, err := ini.Load(fileNameFull)
		if err != nil {
			log.Fatalf("解析INI配置文件错误: %v", err)
		}
		// 将INI配置转换为map
		for _, section := range cfg.Sections() {
			for _, key := range section.Keys() {
				// 获取结构体字段的类型信息
				val := reflect.ValueOf(opt.ConfigData)
				if val.Kind() == reflect.Ptr {
					val = val.Elem()
				}

				// 处理嵌套结构体
				fieldNames := strings.Split(key.Name(), ".")
				var currentVal reflect.Value = val
				var currentField reflect.StructField
				var found bool

				// 构建完整的字段路径
				var fullPath []string
				for _, fieldName := range fieldNames {
					if currentVal.Kind() != reflect.Struct {
						break
					}

					// 查找字段
					currentField, found = currentVal.Type().FieldByName(fieldName)
					if !found {
						// 尝试通过json或ini标签查找
						for j := 0; j < currentVal.NumField(); j++ {
							field := currentVal.Type().Field(j)
							jsonTag := field.Tag.Get("json")
							iniTag := field.Tag.Get("ini")
							if jsonTag == fieldName || iniTag == fieldName {
								currentField = field
								found = true
								break
							}
						}
					}

					if !found {
						break
					}

					fullPath = append(fullPath, currentField.Name)

					if currentVal.Kind() == reflect.Ptr {
						if currentVal.IsNil() {
							currentVal.Set(reflect.New(currentVal.Type().Elem()))
						}
						currentVal = currentVal.Elem()
					}
					currentVal = currentVal.FieldByName(currentField.Name)
				}

				if found {
					// 根据字段类型进行转换
					switch currentField.Type.Kind() {
					case reflect.Bool:
						// 尝试将字符串转换为bool
						if key.Value() == "true" || key.Value() == "1" {
							currentVal.SetBool(true)
						} else if key.Value() == "false" || key.Value() == "0" {
							currentVal.SetBool(false)
						} else {
							currentVal.SetBool(false)
						}
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						// 尝试将字符串转换为整数
						if i, err := strconv.ParseInt(key.Value(), 10, 64); err == nil {
							currentVal.SetInt(i)
						} else {
							currentVal.SetInt(0)
						}
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						// 尝试将字符串转换为无符号整数
						if i, err := strconv.ParseUint(key.Value(), 10, 64); err == nil {
							currentVal.SetUint(i)
						} else {
							currentVal.SetUint(0)
						}
					case reflect.Float32, reflect.Float64:
						// 尝试将字符串转换为浮点数
						if f, err := strconv.ParseFloat(key.Value(), 64); err == nil {
							currentVal.SetFloat(f)
						} else {
							currentVal.SetFloat(0.0)
						}
					default:
						// 其他类型保持为字符串
						currentVal.SetString(key.Value())
					}
				}
			}
		}
	case ConfigTypeJSON, ConfigTypeFile:
		err = json.Unmarshal(file, &opt.ConfigData)
		if err != nil {
			log.Fatalf("解析JSON配置文件错误: %v", err)
		}
	default:
		log.Fatalf("不支持的配置文件类型")
	}
}

// updateConfigFile 更新配置文件
func (opt *Option) updateConfigFile(fileNameFull string, overwrite bool) {
	var fileData []byte
	var err error

	switch opt.ConfigType {
	case ConfigTypeINI:
		// 将结构体转换为map
		jsonData, err := json.Marshal(opt.ConfigData)
		if err != nil {
			log.Fatalf("转换结构体到JSON错误: %v", err)
		}
		var configMap map[string]interface{}
		if err = json.Unmarshal(jsonData, &configMap); err != nil {
			log.Fatalf("解析JSON到map错误: %v", err)
		}
		// 创建INI文件
		cfg := ini.Empty()
		for sectionName, sectionData := range configMap {
			section, err := cfg.NewSection(sectionName)
			if err != nil {
				log.Fatalf("创建INI节错误: %v", err)
			}
			if sectionMap, ok := sectionData.(map[string]interface{}); ok {
				for key, value := range sectionMap {
					strValue := helper.Convert{Value: value}.ToString()
					_, err = section.NewKey(key, strValue)
					if err != nil {
						log.Fatalf("创建INI键错误: %v", err)
					}
				}
			}
		}
		// 使用 helper.NewFile 创建文件
		var buf bytes.Buffer
		if _, err = cfg.WriteTo(&buf); err != nil {
			log.Fatalf("写入INI缓冲区错误: %v", err)
		}
		err = helper.NewFile(&helper.File{Path: fileNameFull}).CreateFile(buf.Bytes(), overwrite)
	case ConfigTypeJSON, ConfigTypeFile:
		fileData, err = json.MarshalIndent(opt.ConfigData, "", " ")
		if err != nil {
			log.Fatalf("转换JSON错误: %v", err)
		}
		err = helper.NewFile(&helper.File{Path: fileNameFull}).CreateFile(fileData, overwrite)
	default:
		log.Fatalf("不支持的配置文件类型")
	}

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
