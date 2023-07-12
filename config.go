package jcbaseGo

import (
	"encoding/json"
	"errors"
	"github.com/jcbowen/jcbaseGo/helper"
	"log"
	"os"
	"path/filepath"
)

type DbStruct struct {
	DriverName  string `json:"driverName" default:"mysql"` // 驱动类型
	Protocol    string `json:"protocol" default:"tcp"`     // 协议
	Host        string `json:"host" default:"localhost"`   // 数据库地址
	Port        string `json:"port" default:"3306"`        // 数据库端口号
	Dbname      string `json:"dbname" default:"dbname"`    // 表名称
	Username    string `json:"username" default:"root"`    // 用户名
	Password    string `json:"password" default:""`        // 密码
	Charset     string `json:"charset" default:"utf8mb4"`  // 编码
	TablePrefix string `json:"tablePrefix" default:"jc_"`  // 表前缀
}

type RedisStruct struct {
	Host     string `json:"host" default:"localhost"` // redis地址
	Port     string `json:"port" default:"6379"`      // redis端口号
	Password string `json:"password" default:""`      // redis密码
	Db       string `json:"db" default:"0"`           // redis数据库
}

type AttachmentStruct struct {
	Dir        string `json:"dir" default:"attachment"`   // 附件存储目录
	RemoteType string `json:"remoteType" default:"local"` // 附件存储类型 local 本地存储 oss 阿里云存储 cos 腾讯云存储
}

type OssStruct struct {
	AccessKeyId     string `json:"AccessKeyId" default:""`     // 阿里云AccessKeyId
	AccessKeySecret string `json:"AccessKeySecret" default:""` // 阿里云AccessKeySecret
	Endpoint        string `json:"endpoint" default:""`        // 阿里云Oss endpoint
	Bucket          string `json:"bucket" default:""`          // 阿里云Oss bucket
}

type CosStruct struct {
	SecretId  string `json:"secretId" default:""`  // 腾讯云Cos SecretId
	SecretKey string `json:"secretKey" default:""` // 腾讯云Cos SecretKey
	Bucket    string `json:"bucket" default:""`    // 腾讯云Cos Bucket
	Region    string `json:"region" default:""`    // 腾讯云Cos Region
	Url       string `json:"url" default:""`       // 腾讯云Cos Url
}

type ProjectStruct struct {
	Name string `json:"name" default:"jcbaseGo"` // 项目名称
}

type RepositoryStruct struct {
	Dir        string `json:"dir" default:"./project/app/"`                            // 本地仓库目录
	Branch     string `json:"branch" default:"master"`                                 // 远程仓库分支
	RemoteName string `json:"remoteName" default:"origin"`                             // 远程仓库名称
	RemoteURL  string `json:"remoteURL" default:"git@github.com:jcbowen/jcbaseGo.git"` // 远程仓库地址
}

type ConfigStruct struct {
	Db         DbStruct         `json:"db"`         // 数据库配置信息
	Redis      RedisStruct      `json:"redis"`      // redis配置信息
	Attachment AttachmentStruct `json:"attachment"` // 附件配置信息
	Oss        OssStruct        `json:"oss"`        // oss配置信息
	Cos        CosStruct        `json:"cos"`        // cos配置信息
	Project    ProjectStruct    `json:"project"`    // 项目配置信息
	Repository RepositoryStruct `json:"repository"` // 仓库配置信息
}

// Config 为Config添加默认数据
var Config ConfigStruct

type ConfigOption struct {
	ConfigFile  string       `json:"config_file" default:"./data/config.json"` // 配置文件路径
	ConfigData  ConfigStruct `json:"config_data"`                              // 配置信息
	RuntimePath string       `json:"runtime_path" default:"/runtime/"`         // 运行缓存目录
}

func New(c ConfigOption) *ConfigOption {
	c.checkConfig()
	return &c
}

// checkConfig 将json配置信息初始化到Config中
func (co *ConfigOption) checkConfig() {
	// 判断是否传入了ConfigData，如果传入了ConfigData，则直接使用ConfigData
	if co.ConfigData != (ConfigStruct{}) {
		Config = co.ConfigData
	}
	if err := helper.CheckAndSetDefault(&Config); err != nil {
		log.Panic(err)
	}

	// 为参数添加默认值
	if err := helper.CheckAndSetDefault(co); err != nil {
		log.Panic(err)
	}

	// 获取json配置文件的绝对路径
	fileNameFull, err := filepath.Abs(co.ConfigFile)
	if err != nil {
		log.Panic(err)
	}

	// json配置文件不存在，根据默认配置生成json配置文件
	if !helper.FileExists(fileNameFull) {
		// 如果配置文件不存在，则创建配置文件
		file, _ := json.MarshalIndent(Config, "", " ")
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

	err = json.Unmarshal([]byte(fileDataString), &Config)
	if err != nil {
		log.Panic(err)
	}
}

// GetConfig 获取配置信息
func (co *ConfigOption) GetConfig() *ConfigStruct {
	return &Config
}

// ----- 终结方法 ----- /

func (co *ConfigOption) GetConfigOption() ConfigOption {
	return *co
}

// ------ 弃用函数 ------ /

// Get 获取配置信息(兼容旧的写法)
// Deprecated: 请使用
func (c *ConfigStruct) Get() *ConfigStruct {
	New(ConfigOption{
		ConfigFile: "",
	}).checkConfig()
	return c
}
