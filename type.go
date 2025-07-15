package jcbaseGo

import (
	"time"
)

const (
	ConfigTypeJSON    = "json"    // 配置类型：json文件
	ConfigTypeINI     = "ini"     // 配置类型：ini文件
	ConfigTypeCommand = "command" // 配置类型：通过命令行传递的字符串json

	// ConfigTypeFile 文件类型，根据文件后缀自动识别是json还是ini文件
	// Deprecated: 推荐使用ConfigTypeJSON或者ConfigTypeINI，保留仅为了兼容旧版
	ConfigTypeFile = "file"
)

// Option jcbaseGo配置选项
type Option struct {
	ConfigType   string      `json:"config_type" ini:"config_type" default:"file"`                // 配置类型，仅支持：json、ini、command
	ConfigSource string      `json:"config_source" ini:"config_source" default:"./data/conf.ini"` // 配置源（json文件/ini文件/命令行）
	ConfigData   interface{} `json:"config_data" ini:"config_data"`                               // 配置信息
	RuntimePath  string      `json:"runtime_path" ini:"runtime_path" default:"./data/runtime/"`   // 运行缓存目录，默认在data目录下
}

// SSLStruct ssl配置
type SSLStruct = WebServer

// WebServer web服务配置
type WebServer struct {
	Port      int    `json:"port" ini:"port" default:"8080"`              // web服务端口号
	EnableSSL bool   `json:"enable_ssl" ini:"enable_ssl" default:"false"` // 是否启用ssl
	CertPath  string `json:"cert_path" ini:"cert_path" default:""`        // ssl证书路径
	KeyPath   string `json:"key_path" ini:"key_path" default:""`          // ssl密钥路径
}

// DbStruct 数据库配置
type DbStruct struct {
	DriverName                               string `json:"driverName" ini:"driverName" default:"mysql"`                                                             // 驱动类型
	Protocol                                 string `json:"protocol" ini:"protocol" default:"tcp"`                                                                   // 协议
	Host                                     string `json:"host" ini:"host" default:"localhost"`                                                                     // 数据库地址
	Port                                     string `json:"port" ini:"port" default:"3306"`                                                                          // 数据库端口号
	Dbname                                   string `json:"dbname" ini:"dbname" default:"dbname"`                                                                    // 表名称
	Username                                 string `json:"username" ini:"username" default:"root"`                                                                  // 用户名
	Password                                 string `json:"password" ini:"password" default:""`                                                                      // 密码
	Charset                                  string `json:"charset" ini:"charset" default:"utf8mb4"`                                                                 // 编码
	TablePrefix                              string `json:"tablePrefix,omitempty" ini:"tablePrefix,omitempty" default:""`                                            // 表前缀
	ParseTime                                string `json:"parseTime" ini:"parseTime" default:"False"`                                                               // 是否开启时间解析
	SingularTable                            bool   `json:"singularTable" ini:"singularTable" default:"true"`                                                        // 使用单数表名
	DisableForeignKeyConstraintWhenMigrating bool   `json:"disableForeignKeyConstraintWhenMigrating" ini:"disableForeignKeyConstraintWhenMigrating" default:"false"` // 是否禁用外键约束
}

// SqlLiteStruct sqlite配置
type SqlLiteStruct struct {
	DbFile                                   string `json:"dbFile" ini:"dbFile" default:"./data/db/jcbaseGo.db"`                                                     // 数据库文件，默认在data目录下
	TablePrefix                              string `json:"tablePrefix" ini:"tablePrefix" default:"jc_"`                                                             // 表前缀
	SingularTable                            bool   `json:"singularTable" ini:"singularTable" default:"true"`                                                        // 使用单数表名
	DisableForeignKeyConstraintWhenMigrating bool   `json:"disableForeignKeyConstraintWhenMigrating" ini:"disableForeignKeyConstraintWhenMigrating" default:"false"` // 是否禁用外键约束
}

// RedisStruct redis配置
type RedisStruct struct {
	Host     string `json:"host" ini:"host" default:"localhost"`                    // redis地址
	Port     string `json:"port" ini:"port" default:"6379"`                         // redis端口号
	Password string `json:"password,omitempty" ini:"password,omitempty" default:""` // redis密码
	Db       string `json:"db,omitempty" ini:"db,omitempty" default:"0"`            // redis数据库
}

// MailerStruct 发送邮箱配置
type MailerStruct struct {
	Host     string `json:"host" ini:"host" default:"smtp.qq.com"`            // 邮箱地址
	Port     string `json:"port" ini:"port" default:"465"`                    // 邮箱端口号
	Username string `json:"username" ini:"username" default:"example@qq.com"` // 邮箱用户名
	Password string `json:"password" ini:"password" default:"123456"`         // 邮箱密码
	From     string `json:"from" ini:"from" default:"example@qq.com"`         // 发件邮箱
	UseTLS   bool   `json:"useTls" ini:"useTls" default:"true"`               // 是否使用TLS
	CertPath string `json:"cert_path" ini:"cert_path" default:""`             // 证书文件路径
	KeyPath  string `json:"key_path" ini:"key_path" default:""`               // 私钥文件路径
	CAPath   string `json:"ca_path" ini:"ca_path" default:""`                 // CA证书文件路径
}

// AttachmentStruct 附件配置
type AttachmentStruct struct {
	StorageType string `json:"storage_type" ini:"storage_type" default:"local"` // 存储类型 local/ftp/sftp/cos/oss
	LocalDir    string `json:"local_dir" ini:"local_dir" default:"attachment"`  // 本地附件目录，默认为 attachment

	// 附件访问域名
	// 配置后以配置为准，不配置则初始化中自动赋值，一定以"/"结尾
	// 如果配置了远程附件，则此处为远程附件访问域名，否则为本地附件访问域名
	VisitDomain string `json:"visit_domain" ini:"visit_domain" default:"/"`
	// 本地附件访问域名，不管是否配置远程附件
	// 配置后以配置为准，不配置则初始化中自动赋值，一定以"/"结尾
	LocalVisitDomain string `json:"local_visit_domain" ini:"local_visit_domain" default:"/"`
}

// OSSStruct oss配置
type OSSStruct struct {
	AccessKeyId     string `json:"AccessKeyId" ini:"AccessKeyId" default:""`         // 阿里云访问密钥ID
	AccessKeySecret string `json:"AccessKeySecret" ini:"AccessKeySecret" default:""` // 阿里云访问密钥Secret
	Endpoint        string `json:"endpoint" ini:"endpoint" default:""`               // 阿里云OSS的Endpoint
	BucketName      string `json:"bucketName" ini:"bucketName" default:""`           // 阿里云OSS存储桶名称

	// 自定义附件访问域名(非平台配置，供程序调用，非必填，一定以"/"结尾)
	CustomizeVisitDomain string `json:"customize_visit_domain,omitempty" ini:"customize_visit_domain,omitempty" default:""`
}

// COSStruct cos配置
type COSStruct struct {
	SecretId  string `json:"secretId" ini:"secretId" default:""`   // 腾讯云Cos SecretId
	SecretKey string `json:"secretKey" ini:"secretKey" default:""` // 腾讯云Cos SecretKey
	Bucket    string `json:"bucket" ini:"bucket" default:""`       // 腾讯云Cos Bucket
	Region    string `json:"region" ini:"region" default:""`       // 腾讯云Cos Region
	Url       string `json:"url" ini:"url" default:""`             // 腾讯云Cos Url

	// 自定义附件访问域名(非平台配置，供程序调用，非必填，一定以"/"结尾)
	CustomizeVisitDomain string `json:"customize_visit_domain,omitempty" ini:"customize_visit_domain,omitempty" default:""`
}

type FTPStruct struct {
	Address  string        `json:"address" ini:"address" default:""`                     // FTP服务器地址
	Username string        `json:"username" ini:"username" default:""`                   // FTP登录用户名
	Password string        `json:"password" ini:"password" default:""`                   // FTP登录密码
	Timeout  time.Duration `json:"timeout,omitempty" ini:"timeout,omitempty" default:""` // 连接超时时间，可选

	// 自定义附件访问域名(非平台配置，供程序调用，非必填，一定以"/"结尾)
	CustomizeVisitDomain string `json:"customize_visit_domain,omitempty" ini:"customize_visit_domain,omitempty" default:""`
}

// SFTPStruct sftp配置
type SFTPStruct struct {
	Address    string        `json:"address" ini:"address" default:""`                     // SFTP服务器地址
	Username   string        `json:"username" ini:"username" default:""`                   // SFTP登录用户名
	Password   string        `json:"password" ini:"password" default:""`                   // SFTP登录密码
	PrivateKey []byte        `json:"private_key" ini:"private_key" default:""`             // SFTP登录私钥
	Timeout    time.Duration `json:"timeout,omitempty" ini:"timeout,omitempty" default:""` // 连接超时时间，可选

	// 自定义附件访问域名(非平台配置，供程序调用，非必填，一定以"/"结尾)
	CustomizeVisitDomain string `json:"customize_visit_domain,omitempty" ini:"customize_visit_domain,omitempty" default:""`
}

// ProjectStruct 项目配置
type ProjectStruct struct {
	Name string `json:"name" ini:"name" default:"jcbaseGo"` // 项目名称
}

// RepositoryStruct 仓库配置
type RepositoryStruct struct {
	Dir        string `json:"dir" ini:"dir" default:"./project/app/"`                                  // 本地仓库目录
	Branch     string `json:"branch" ini:"branch" default:"master"`                                    // 远程仓库分支
	RemoteName string `json:"remoteName" ini:"remoteName" default:"origin"`                            // 远程仓库名称
	RemoteURL  string `json:"remoteURL" ini:"remoteURL" default:"git@github.com:jcbowen/jcbaseGo.git"` // 远程仓库地址
}

// DefaultConfigStruct 默认配置信息结构
// 一般情况下推荐自定义，不想自定义的情况下可以采用默认结构
type DefaultConfigStruct struct {
	Db         DbStruct         `json:"db" ini:"db"`                 // 数据库配置信息
	Redis      RedisStruct      `json:"redis" ini:"redis"`           // redis配置信息
	Attachment AttachmentStruct `json:"attachment" ini:"attachment"` // 附件配置信息
	Oss        OSSStruct        `json:"oss" ini:"oss"`               // oss配置信息
	Cos        COSStruct        `json:"cos" ini:"cos"`               // cos配置信息
	Project    ProjectStruct    `json:"project" ini:"project"`       // 项目配置信息
	Repository RepositoryStruct `json:"repository" ini:"repository"` // 仓库配置信息
}

// ListData 分页查询数据输出
type ListData struct {
	List     interface{} `json:"list"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// Result 响应结构
type Result struct {
	Code    int         `json:"code" default:"200"`
	Message string      `json:"message" default:"success"`
	Data    interface{} `json:"data,omitempty"`
	Total   *int        `json:"total,omitempty"`
}
