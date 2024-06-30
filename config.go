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

// DbStruct 数据库配置
type DbStruct struct {
	DriverName    string `json:"driverName" default:"mysql"`   // 驱动类型
	Protocol      string `json:"protocol" default:"tcp"`       // 协议
	Host          string `json:"host" default:"localhost"`     // 数据库地址
	Port          string `json:"port" default:"3306"`          // 数据库端口号
	Dbname        string `json:"dbname" default:"dbname"`      // 表名称
	Username      string `json:"username" default:"root"`      // 用户名
	Password      string `json:"password" default:""`          // 密码
	Charset       string `json:"charset" default:"utf8mb4"`    // 编码
	TablePrefix   string `json:"tablePrefix" default:"jc_"`    // 表前缀
	ParseTime     string `json:"parseTime" default:"False"`    // 是否开启时间解析
	SingularTable string `json:"singularTable" default:"true"` // 使用单数表名
}

type SqlLiteStruct struct {
	DbFile        string `json:"dbFile" default:"./db/jcbaseGo.db"` // 数据库文件
	TablePrefix   string `json:"tablePrefix" default:"jc_"`         // 表前缀
	SingularTable string `json:"singularTable" default:"true"`      // 使用单数表名
}

// RedisStruct redis配置
type RedisStruct struct {
	Host     string `json:"host" default:"localhost"` // redis地址
	Port     string `json:"port" default:"6379"`      // redis端口号
	Password string `json:"password" default:""`      // redis密码
	Db       string `json:"db" default:"0"`           // redis数据库
}

// MailerStruct 发送邮箱配置
type MailerStruct struct {
	Scheme   string `json:"scheme" default:"smtp"`             // 邮箱协议
	Host     string `json:"host" default:"smtp.qq.com"`        // 邮箱地址
	Port     string `json:"port" default:"465"`                // 邮箱端口号
	Username string `json:"username" default:"example@qq.com"` // 邮箱用户名
	Password string `json:"password" default:"123456"`         // 邮箱密码
}

// AttachmentStruct 附件配置
type AttachmentStruct struct {
	Dir        string `json:"dir" default:"attachment"`   // 附件存储目录
	RemoteType string `json:"remoteType" default:"local"` // 附件存储类型 local 本地存储 oss 阿里云存储 cos 腾讯云存储
}

// OssStruct oss配置
type OssStruct struct {
	AccessKeyId     string `json:"AccessKeyId" default:""`     // 阿里云AccessKeyId
	AccessKeySecret string `json:"AccessKeySecret" default:""` // 阿里云AccessKeySecret
	Endpoint        string `json:"endpoint" default:""`        // 阿里云Oss endpoint
	Bucket          string `json:"bucket" default:""`          // 阿里云Oss bucket
}

// CosStruct cos配置
type CosStruct struct {
	SecretId  string `json:"secretId" default:""`  // 腾讯云Cos SecretId
	SecretKey string `json:"secretKey" default:""` // 腾讯云Cos SecretKey
	Bucket    string `json:"bucket" default:""`    // 腾讯云Cos Bucket
	Region    string `json:"region" default:""`    // 腾讯云Cos Region
	Url       string `json:"url" default:""`       // 腾讯云Cos Url
}

// ProjectStruct 项目配置
type ProjectStruct struct {
	Name string `json:"name" default:"jcbaseGo"` // 项目名称
}

// RepositoryStruct 仓库配置
type RepositoryStruct struct {
	Dir        string `json:"dir" default:"./project/app/"`                            // 本地仓库目录
	Branch     string `json:"branch" default:"master"`                                 // 远程仓库分支
	RemoteName string `json:"remoteName" default:"origin"`                             // 远程仓库名称
	RemoteURL  string `json:"remoteURL" default:"git@github.com:jcbowen/jcbaseGo.git"` // 远程仓库地址
}

// DefaultConfigStruct 默认配置信息结构
// 一般情况下推荐自定义，不想自定义的情况下可以采用默认结构
type DefaultConfigStruct struct {
	Db         DbStruct         `json:"db"`         // 数据库配置信息
	Redis      RedisStruct      `json:"redis"`      // redis配置信息
	Attachment AttachmentStruct `json:"attachment"` // 附件配置信息
	Oss        OssStruct        `json:"oss"`        // oss配置信息
	Cos        CosStruct        `json:"cos"`        // cos配置信息
	Project    ProjectStruct    `json:"project"`    // 项目配置信息
	Repository RepositoryStruct `json:"repository"` // 仓库配置信息
}

// Config 为Config添加默认数据
var Config interface{}

// Option jcbaseGo配置选项
type Option struct {
	ConfigFile  string      `json:"config_file" default:"./config/main.json"` // 配置文件路径
	ConfigData  interface{} `json:"config_data"`                              // 配置信息
	RuntimePath string      `json:"runtime_path" default:"/runtime/"`         // 运行缓存目录
}

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
