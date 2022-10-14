package common

import (
	"encoding/json"
	"fmt"
	"os"
)

type dbStruct struct {
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

type configStruct struct {
	Db dbStruct `json:"db"` // 数据库配置信息
}

// Config 为Config添加默认数据
var Config = configStruct{
	dbStruct{
		DriverName:  "mysql",
		Protocol:    "tcp",
		Host:        "localhost",
		Port:        "3306",
		Dbname:      "dbname",
		Username:    "root",
		Password:    "123456789",
		Charset:     "utf8",
		TablePrefix: "",
	},
}

// 将json配置信息初始化到Config中
func init() {
	filename := "./data/config.json"
	// 先判断json配置文件是否存在
	if fileExit(filename) {
		// 读取json配置文件
		file, fErr := os.ReadFile(filename)
		if fErr != nil {
			panic(fErr)
		}
		fileDataString := string(file)

		err := json.Unmarshal([]byte(fileDataString), &Config)
		if err != nil {
			fmt.Printf("err was %v", err)
			os.Exit(1)
		}
	}
}

// GetDSN 根据配置信息输出dataSourceName
// @return string username:password@protocol(localhost:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
func (c *configStruct) GetDSN() string {
	dsn := "%s:%s@%s(%s:%s)/%s?charset=%s&parseTime=True&loc=Local"
	return fmt.Sprintf(dsn, Config.Db.Username, Config.Db.Password, Config.Db.Protocol, Config.Db.Host, Config.Db.Port, Config.Db.Dbname, Config.Db.Charset)
}

// fileExit 判断文件是否存在
func fileExit(filename string) bool {
	_, err := os.Lstat(filename)
	return !os.IsNotExist(err)
}
