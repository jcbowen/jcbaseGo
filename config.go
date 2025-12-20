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

	"github.com/go-redis/redis/v8"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"gopkg.in/ini.v1"
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

// ReplaceConfigNode 替换指定节点的配置信息
// 参数：
//   - nodePath: 节点路径，如 "title" 或 "db.host"
//   - newValue: 新的配置值，类型必须与原节点类型匹配
//   - condition: 条件值，如果提供则只有当原节点值等于条件值时才替换
//
// 返回：
//   - bool: 是否成功替换
//
// 使用示例：
//
//	// 替换所有title节点为ok
//	opt.ReplaceConfigNode("title", "ok", nil)
//	// 仅当title=test时替换为ok
//	opt.ReplaceConfigNode("title", "ok", "test")
func (opt *Option) ReplaceConfigNode(nodePath string, newValue interface{}, condition interface{}) bool {
	// 检查配置数据是否为空
	if opt.ConfigData == nil {
		log.Printf("配置数据为空，无法替换节点: %s", nodePath)
		return false
	}

	// 按点号分割节点路径
	nodeNames := strings.Split(nodePath, ".")
	if len(nodeNames) == 0 {
		log.Printf("节点路径为空，无法替换")
		return false
	}

	// 获取配置数据的反射值
	val := reflect.ValueOf(opt.ConfigData)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 遍历节点路径，找到目标节点
	var currentVal reflect.Value = val
	for i := 0; i < len(nodeNames); i++ {
		// 如果当前值不是结构体，则无法继续查找
		if currentVal.Kind() != reflect.Struct {
			log.Printf("节点路径 %s 无效，当前值不是结构体: %v", strings.Join(nodeNames[:i+1], "."), currentVal.Kind())
			return false
		}

		// 查找字段（先按字段名，再按标签）
		fieldName := nodeNames[i]
		_, found := currentVal.Type().FieldByName(fieldName)
		if !found {
			// 尝试通过标签查找
			for j := 0; j < currentVal.NumField(); j++ {
				f := currentVal.Type().Field(j)
				jsonTag := parseTagName(f.Tag.Get("json"))
				iniTag := parseTagName(f.Tag.Get("ini"))
				if jsonTag == fieldName || iniTag == fieldName {
					fieldName = f.Name
					found = true
					break
				}
			}
		}

		if !found {
			log.Printf("未找到节点: %s", strings.Join(nodeNames[:i+1], "."))
			return false
		}

		// 获取字段值
		fieldVal := currentVal.FieldByName(fieldName)
		if fieldVal.Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				// 如果指针为空且不是最后一个节点，则创建新实例
				if i < len(nodeNames)-1 {
					newVal := reflect.New(fieldVal.Type().Elem())
					fieldVal.Set(newVal)
					fieldVal = fieldVal.Elem()
				} else {
					log.Printf("节点 %s 是 nil 指针，无法替换", nodePath)
					return false
				}
			} else {
				fieldVal = fieldVal.Elem()
			}
		}

		// 如果是最后一个节点，执行替换
		if i == len(nodeNames)-1 {
			return opt.replaceNodeValue(fieldVal, newValue, condition)
		}

		// 否则继续遍历下一个节点
		currentVal = fieldVal
	}

	return false
}

// applyConfigReplaceRules 应用配置替换规则
// 遍历ConfigReplaceRules切片，对每个规则调用ReplaceConfigNode方法执行配置替换
func (opt *Option) applyConfigReplaceRules() {
	if len(opt.ConfigReplaceRules) == 0 {
		return
	}

	for _, rule := range opt.ConfigReplaceRules {
		opt.ReplaceConfigNode(rule.NodePath, rule.NewValue, rule.Condition)
	}
}

// replaceNodeValue 替换节点值，处理类型检查和条件判断
func (opt *Option) replaceNodeValue(fieldVal reflect.Value, newValue interface{}, condition interface{}) bool {
	// 检查字段是否可设置
	if !fieldVal.CanSet() {
		log.Printf("字段不可设置，无法替换")
		return false
	}

	// 获取字段当前值
	currentVal := fieldVal.Interface()

	// 检查条件
	if condition != nil {
		// 检查条件类型是否匹配
		if reflect.TypeOf(condition) != reflect.TypeOf(currentVal) {
			log.Printf("条件类型不匹配，期望 %T，实际 %T", currentVal, condition)
			return false
		}

		// 检查当前值是否等于条件值
		if !reflect.DeepEqual(currentVal, condition) {
			return false
		}
	}

	// 检查新值类型是否匹配
	if reflect.TypeOf(newValue) != reflect.TypeOf(currentVal) {
		log.Printf("新值类型不匹配，期望 %T，实际 %T", currentVal, newValue)
		return false
	}

	// 替换值
	fieldVal.Set(reflect.ValueOf(newValue))
	return true
}

// checkConfig 将配置信息初始化到 Config 中
func (opt *Option) checkConfig() {
	if reflect.TypeOf(opt.ConfigData) == nil {
		log.Fatalf("配置信息不能为空")
		return
	}

	// 初始化默认配置
	opt.initializeConfigWithDefaults()

	// 如果是文件类型配置，根据文件后缀名判断配置类型
	if opt.ConfigType == ConfigTypeFile {
		ext := filepath.Ext(opt.ConfigSource)
		switch strings.ToLower(ext) {
		case ".json":
			opt.ConfigType = ConfigTypeJSON
		case ".ini":
			opt.ConfigType = ConfigTypeINI
		default:
			log.Fatalf("不支持的配置文件类型: %s", ext)
		}
	}

	switch opt.ConfigType {
	case ConfigTypeJSON, ConfigTypeINI, ConfigTypeFile:
		// 获取配置文件绝对路径
		fileNameFull := opt.getConfigFilePath()
		// 如果配置文件不存在，则创建
		opt.createConfigFileIfNotExists(fileNameFull)
		// 从文件中读取配置
		opt.readConfigFile(fileNameFull)
		// 执行配置替换规则
		opt.applyConfigReplaceRules()
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
		// 执行配置替换规则
		opt.applyConfigReplaceRules()
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
		log.Printf("配置文件不存在，已创建默认配置文件，请修改配置文件后重启程序！\n配置文件路径：%s", fileNameFull)
	}
}

// readConfigFile 读取配置文件并解析到结构体中
// 支持 INI 和 JSON 两种格式
// INI 格式支持多级嵌套，使用点号(.)分隔，第一级为节名，后续为字段名
// 例如：[Database] db.name = test 会被解析到 Database 结构体的 DB 字段的 Name 属性
func (opt *Option) readConfigFile(fileNameFull string) {
	// 读取配置文件内容
	file, err := os.ReadFile(fileNameFull)
	if err != nil {
		log.Fatalf("读取配置文件错误: %v", err)
	}

	switch opt.ConfigType {
	case ConfigTypeINI:
		// 加载 INI 配置文件
		cfg, err := ini.Load(fileNameFull)
		if err != nil {
			log.Fatalf("解析INI配置文件错误: %v", err)
		}
		// 遍历所有节（第一级标题）
		for _, section := range cfg.Sections() {
			// 获取配置结构体的反射值
			val := reflect.ValueOf(opt.ConfigData)
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}

			// 查找与节名匹配的字段（第一级结构体字段）
			var sectionField reflect.StructField
			var found bool
			for i := 0; i < val.NumField(); i++ {
				field := val.Type().Field(i)
				jsonTag := parseTagName(field.Tag.Get("json"))
				iniTag := parseTagName(field.Tag.Get("ini"))
				if jsonTag == section.Name() || iniTag == section.Name() {
					sectionField = field
					found = true
					break
				}
			}

			if !found {
				continue
			}

			// 获取节对应的字段值（第一级结构体实例）
			sectionVal := val.FieldByName(sectionField.Name)
			if sectionVal.Kind() == reflect.Ptr {
				if sectionVal.IsNil() {
					sectionVal.Set(reflect.New(sectionVal.Type().Elem()))
				}
				sectionVal = sectionVal.Elem()
			}

			// 处理节内的每个键值对
			for _, key := range section.Keys() {
				// 将键名按点号分割，处理嵌套结构
				fieldNames := strings.Split(key.Name(), ".")
				var currentVal reflect.Value = sectionVal
				var currentField reflect.StructField
				var found bool

				// 遍历字段路径，处理嵌套结构体
				for i := 0; i < len(fieldNames); i++ {
					// 如果当前值不是结构体，则退出循环
					if currentVal.Kind() != reflect.Struct {
						break
					}

					// 尝试通过字段名查找
					currentField, found = currentVal.Type().FieldByName(fieldNames[i])
					if !found {
						// 如果通过字段名没找到，尝试通过标签查找
						for j := 0; j < currentVal.NumField(); j++ {
							field := currentVal.Type().Field(j)
							jsonTag := parseTagName(field.Tag.Get("json"))
							iniTag := parseTagName(field.Tag.Get("ini"))
							if jsonTag == fieldNames[i] || iniTag == fieldNames[i] {
								currentField = field
								found = true
								break
							}
						}
					}

					if !found {
						break
					}

					// 处理指针类型的字段
					if currentVal.Kind() == reflect.Ptr {
						if currentVal.IsNil() {
							currentVal.Set(reflect.New(currentVal.Type().Elem()))
						}
						currentVal = currentVal.Elem()
					}
					// 获取下一个层级的字段值
					currentVal = currentVal.FieldByName(currentField.Name)
				}

				// 如果找到了对应的字段，设置其值
				if found {
					// 根据字段类型进行相应的类型转换
					switch currentField.Type.Kind() {
					case reflect.Bool:
						// 布尔类型转换
						if key.Value() == "true" || key.Value() == "1" {
							currentVal.SetBool(true)
						} else if key.Value() == "false" || key.Value() == "0" {
							currentVal.SetBool(false)
						} else {
							log.Fatalf("\n配置错误：[%s] %s = %s\n期望类型：布尔值\n实际值有误，无法转换为布尔类型。\n请检查并修正该配置项的值后重启程序。\n", section.Name(), key.Name(), key.Value())
						}
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						// 整数类型转换
						if i, err := strconv.ParseInt(key.Value(), 10, 64); err == nil {
							currentVal.SetInt(i)
						} else {
							log.Fatalf("\n配置错误：[%s] %s = %s\n期望类型：整数\n实际值有误，无法转换（错误信息：%v）\n请检查并修正该配置项的值后重启程序。\n", section.Name(), key.Name(), key.Value(), err)
						}
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						// 无符号整数类型转换
						if i, err := strconv.ParseUint(key.Value(), 10, 64); err == nil {
							currentVal.SetUint(i)
						} else {
							log.Fatalf("\n配置错误：[%s] %s = %s\n期望类型：无符号整数\n实际值有误，无法转换（错误信息：%v）\n请检查并修正该配置项的值后重启程序。\n", section.Name(), key.Name(), key.Value(), err)
						}
					case reflect.Float32, reflect.Float64:
						// 浮点数类型转换
						if f, err := strconv.ParseFloat(key.Value(), 64); err == nil {
							currentVal.SetFloat(f)
						} else {
							log.Fatalf("\n配置错误：[%s] %s = %s\n期望类型：浮点数\n实际值有误，无法转换（错误信息：%v）\n请检查并修正该配置项的值后重启程序。\n", section.Name(), key.Name(), key.Value(), err)
						}
					default:
						// 其他类型（如字符串）直接设置
						currentVal.SetString(key.Value())
					}
				}
			}
		}
	case ConfigTypeJSON:
		// JSON 格式直接解析到结构体
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
		// 创建INI文件
		cfg := ini.Empty()
		// 递归处理结构体，生成INI配置
		opt.processStructToINI(cfg, opt.ConfigData, "")
		// 使用 helper.NewFile 创建文件
		var buf bytes.Buffer
		if _, err = cfg.WriteTo(&buf); err != nil {
			log.Fatalf("写入INI缓冲区错误: %v", err)
		}
		err = helper.NewFile(&helper.File{Path: fileNameFull}).CreateFile(buf.Bytes(), overwrite)
	case ConfigTypeJSON:
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

// parseTagName 解析结构体标签，提取字段名部分
// 参数：
//   - tag: 完整的标签字符串，如 "fieldname,omitempty"
//
// 返回：
//   - string: 提取的字段名，如 "fieldname"
func parseTagName(tag string) string {
	// 如果标签为空，直接返回
	if tag == "" {
		return ""
	}

	// 按逗号分割，取第一部分作为字段名
	parts := strings.Split(tag, ",")
	return parts[0]
}

// hasOmitEmpty 检查标签是否包含omitempty选项
// 参数：
//   - tag: 完整的标签字符串，如 "fieldname,omitempty"
//
// 返回：
//   - bool: 是否包含omitempty选项
func hasOmitEmpty(tag string) bool {
	// 如果标签为空，直接返回false
	if tag == "" {
		return false
	}

	// 按逗号分割，检查是否包含omitempty
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		if strings.TrimSpace(part) == "omitempty" {
			return true
		}
	}
	return false
}

// isZeroValue 检查反射值是否为零值
// 参数：
//   - val: 反射值
//
// 返回：
//   - bool: 是否为零值
func isZeroValue(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.Bool:
		return !val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.String:
		return val.String() == ""
	case reflect.Ptr, reflect.Interface:
		return val.IsNil()
	case reflect.Slice, reflect.Array, reflect.Map:
		return val.Len() == 0
	default:
		// 对于其他类型，使用reflect的Zero方法比较
		return reflect.DeepEqual(val.Interface(), reflect.Zero(val.Type()).Interface())
	}
}

// processStructToINI 递归处理结构体，生成INI配置
// 参数：
//   - cfg: INI配置文件对象
//   - data: 要处理的数据（结构体或指针）
//   - prefix: 当前处理的路径前缀
func (opt *Option) processStructToINI(cfg *ini.File, data interface{}, prefix string) {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
		val = val.Elem()
	}

	// 如果不是结构体，直接返回
	if val.Kind() != reflect.Struct {
		return
	}

	// 遍历结构体的所有字段
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fieldVal := val.Field(i)

		// 获取字段的标签并正确解析
		jsonTag := parseTagName(field.Tag.Get("json"))
		iniTag := parseTagName(field.Tag.Get("ini"))

		// 确定字段名（优先使用ini标签，其次json标签，最后字段名）
		fieldName := field.Name
		if iniTag != "" {
			fieldName = iniTag
		} else if jsonTag != "" {
			fieldName = jsonTag
		}

		// 处理指针类型字段
		if fieldVal.Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				continue
			}
			fieldVal = fieldVal.Elem()
		}

		// 检查omitempty标签和零值
		// 优先检查ini标签，其次检查json标签
		hasOmitEmptyTag := hasOmitEmpty(field.Tag.Get("ini")) || hasOmitEmpty(field.Tag.Get("json"))
		if hasOmitEmptyTag && isZeroValue(fieldVal) {
			// 如果有omitempty标签且值为零值，则跳过该字段
			continue
		}

		// 构建当前字段的完整路径
		currentPath := fieldName
		if prefix != "" {
			currentPath = prefix + "." + fieldName
		}

		// 根据字段类型进行处理
		switch fieldVal.Kind() {
		case reflect.Struct:
			// 如果是结构体，递归处理
			opt.processStructToINI(cfg, fieldVal.Interface(), currentPath)
		case reflect.Map, reflect.Slice, reflect.Array:
			// 如果是复杂类型，转换为JSON字符串
			jsonData, err := json.Marshal(fieldVal.Interface())
			if err != nil {
				log.Printf("序列化字段 %s 失败: %v", currentPath, err)
				continue
			}
			opt.addINIKey(cfg, currentPath, string(jsonData))
		default:
			// 基本类型，直接转换为字符串
			strValue := opt.valueToString(fieldVal)
			opt.addINIKey(cfg, currentPath, strValue)
		}
	}
}

// addINIKey 添加INI键值对
// 参数：
//   - cfg: INI配置文件对象
//   - key: 键名（可能包含点号分隔的路径）
//   - value: 值
func (opt *Option) addINIKey(cfg *ini.File, key, value string) {
	// 如果键名包含点号，需要分割为节名和键名
	if strings.Contains(key, ".") {
		parts := strings.SplitN(key, ".", 2)
		sectionName := parts[0]
		keyName := parts[1]

		// 获取或创建节
		section := cfg.Section(sectionName)
		if section == nil {
			var err error
			section, err = cfg.NewSection(sectionName)
			if err != nil {
				log.Printf("创建INI节 %s 失败: %v", sectionName, err)
				return
			}
		}

		// 添加键值对
		_, err := section.NewKey(keyName, value)
		if err != nil {
			log.Printf("添加INI键 %s 失败: %v", key, err)
		}
	} else {
		// 如果没有点号，使用默认节
		section := cfg.Section("")
		_, err := section.NewKey(key, value)
		if err != nil {
			log.Printf("添加INI键 %s 失败: %v", key, err)
		}
	}
}

// valueToString 将反射值转换为字符串
// 参数：
//   - val: 反射值
//
// 返回：
//   - string: 转换后的字符串
func (opt *Option) valueToString(val reflect.Value) string {
	switch val.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(val.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'f', -1, 64)
	case reflect.String:
		return val.String()
	default:
		// 对于其他类型，尝试使用 helper.Convert
		return helper.Convert{Value: val.Interface()}.ToString()
	}
}
