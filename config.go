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
	DriverName  string `json:"driverName"`  // 驱动类型
	Protocol    string `json:"protocol"`    // 协议
	Host        string `json:"host"`        // 数据库地址
	Port        string `json:"port"`        // 数据库端口号
	Dbname      string `json:"dbname"`      // 表名称
	Username    string `json:"username"`    // 用户名
	Password    string `json:"password"`    // 密码
	Charset     string `json:"charset"`     // 编码
	TablePrefix string `json:"tablePrefix"` // 表前缀
}

type RepositoryStruct struct {
	Dir        string `json:"dir"`        // 本地仓库目录
	Branch     string `json:"branch"`     // 远程仓库分支
	RemoteName string `json:"remoteName"` // 远程仓库名称
	RemoteURL  string `json:"remoteURL"`  // 远程仓库地址
}

type RedisStruct struct {
	Host     string `json:"host"`     // redis地址
	Port     string `json:"port"`     // redis端口号
	Password string `json:"password"` // redis密码
	Db       string `json:"db"`       // redis数据库
}

type ConfigStruct struct {
	Db         DbStruct         `json:"db"`         // 数据库配置信息
	Redis      RedisStruct      `json:"redis"`      // redis配置信息
	Repository RepositoryStruct `json:"repository"` // 仓库配置信息
}

// Config 为Config添加默认数据
var Config = ConfigStruct{
	Db: DbStruct{
		DriverName:  "mysql",
		Protocol:    "tcp",
		Host:        "localhost",
		Port:        "3306",
		Dbname:      "dbname",
		Username:    "root",
		Password:    "",
		Charset:     "utf8mb4",
		TablePrefix: "",
	},
	Redis: RedisStruct{
		Host:     "localhost",
		Port:     "6379",
		Password: "",
		Db:       "0",
	},
	Repository: RepositoryStruct{
		Dir:        "./project/app/", // 本地仓库目录，目录必须以/结尾
		Branch:     "master",
		RemoteName: "origin",
		RemoteURL:  "git@github.com:jcbowen/jcbaseGo.git",
	},
}

type ConfigOption struct {
	ConfigFile  string // 配置文件路径
	RuntimePath string // 运行缓存目录
}

func New(c ConfigOption) *ConfigOption {
	c.checkConfig()
	return &c
}

// checkConfig 将json配置信息初始化到Config中
func (co *ConfigOption) checkConfig() {
	filename := "./data/config.json"
	if co.ConfigFile != "" {
		filename = co.ConfigFile
	} else {
		co.ConfigFile = filename
	}
	fileNameFull, err := filepath.Abs(filename)
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
