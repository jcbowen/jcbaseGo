package jcbaseGo

import (
	"time"
)

// Option jcbaseGo配置选项
type Option struct {
	ConfigFile  string      `json:"config_file" default:"./config/main.json"` // 配置文件路径
	ConfigData  interface{} `json:"config_data"`                              // 配置信息
	RuntimePath string      `json:"runtime_path" default:"/runtime/"`         // 运行缓存目录
}

// SSLStruct ssl配置
type SSLStruct struct {
	CertPath string `json:"cert_path" default:""`
	KeyPath  string `json:"key_path" default:""`
}

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
	TablePrefix   string `json:"tablePrefix" default:""`       // 表前缀
	ParseTime     string `json:"parseTime" default:"False"`    // 是否开启时间解析
	SingularTable bool   `json:"singularTable" default:"true"` // 使用单数表名
	Alias         string `json:"alias" default:"db"`           // 配置信息别名
}

// SqlLiteStruct sqlite配置
type SqlLiteStruct struct {
	DbFile        string `json:"dbFile" default:"./db/jcbaseGo.db"` // 数据库文件
	TablePrefix   string `json:"tablePrefix" default:"jc_"`         // 表前缀
	SingularTable bool   `json:"singularTable" default:"true"`      // 使用单数表名
	Alias         string `json:"alias" default:"main"`              // 配置信息别名
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
	Host     string `json:"host" default:"smtp.qq.com"`        // 邮箱地址
	Port     string `json:"port" default:"465"`                // 邮箱端口号
	Username string `json:"username" default:"example@qq.com"` // 邮箱用户名
	Password string `json:"password" default:"123456"`         // 邮箱密码
	From     string `json:"from" default:"example@qq.com"`     // 发件邮箱
	UseTLS   bool   `json:"useTls" default:"true"`             // 是否使用TLS
	CertPath string `json:"cert_path" default:""`              // 证书文件路径
	KeyPath  string `json:"key_path" default:""`               // 私钥文件路径
	CAPath   string `json:"ca_path" default:""`                // CA证书文件路径
}

// AttachmentStruct 附件配置
type AttachmentStruct struct {
	StorageType string `json:"storage_type" default:"local"`   // 存储类型 local/ftp/sftp/cos/oss
	LocalDir    string `json:"local_dir" default:"attachment"` // 本地附件目录，默认为 attachment

	// 附件访问域名
	// 配置后以配置为准，不配置则初始化中自动赋值，一定以“/”结尾
	// 如果配置了远程附件，则此处为远程附件访问域名，否则为本地附件访问域名
	VisitDomain string `json:"visit_domain" default:"/"`
	// 本地附件访问域名，不管是否配置远程附件
	// 配置后以配置为准，不配置则初始化中自动赋值，一定以“/”结尾
	LocalVisitDomain string `json:"local_visit_domain" default:"/"`
}

// OSSStruct oss配置
type OSSStruct struct {
	AccessKeyId     string `json:"AccessKeyId" default:""`     // 阿里云访问密钥ID
	AccessKeySecret string `json:"AccessKeySecret" default:""` // 阿里云访问密钥Secret
	Endpoint        string `json:"endpoint" default:""`        // 阿里云OSS的Endpoint
	BucketName      string `json:"bucketName" default:""`      // 阿里云OSS存储桶名称

	// 自定义附件访问域名(非平台配置，供程序调用，非必填，一定以“/”结尾)
	CustomizeVisitDomain string `json:"customize_visit_domain" default:""`
}

// COSStruct cos配置
type COSStruct struct {
	SecretId  string `json:"secretId" default:""`  // 腾讯云Cos SecretId
	SecretKey string `json:"secretKey" default:""` // 腾讯云Cos SecretKey
	Bucket    string `json:"bucket" default:""`    // 腾讯云Cos Bucket
	Region    string `json:"region" default:""`    // 腾讯云Cos Region
	Url       string `json:"url" default:""`       // 腾讯云Cos Url

	// 自定义附件访问域名(非平台配置，供程序调用，非必填，一定以“/”结尾)
	CustomizeVisitDomain string `json:"customize_visit_domain" default:""`
}

type FTPStruct struct {
	Address  string        `json:"address" default:""`           // FTP服务器地址
	Username string        `json:"username" default:""`          // FTP登录用户名
	Password string        `json:"password" default:""`          // FTP登录密码
	Timeout  time.Duration `json:"timeout,omitempty" default:""` // 连接超时时间，可选

	// 自定义附件访问域名(非平台配置，供程序调用，非必填，一定以“/”结尾)
	CustomizeVisitDomain string `json:"customize_visit_domain" default:""`
}

// SFTPStruct sftp配置
type SFTPStruct struct {
	Address    string        `json:"address" default:""`           // SFTP服务器地址
	Username   string        `json:"username" default:""`          // SFTP登录用户名
	Password   string        `json:"password" default:""`          // SFTP登录密码
	PrivateKey []byte        `json:"private_key" default:""`       // SFTP登录私钥
	Timeout    time.Duration `json:"timeout,omitempty" default:""` // 连接超时时间，可选

	// 自定义附件访问域名(非平台配置，供程序调用，非必填，一定以“/”结尾)
	CustomizeVisitDomain string `json:"customize_visit_domain" default:""`
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
	Oss        OSSStruct        `json:"oss"`        // oss配置信息
	Cos        COSStruct        `json:"cos"`        // cos配置信息
	Project    ProjectStruct    `json:"project"`    // 项目配置信息
	Repository RepositoryStruct `json:"repository"` // 仓库配置信息
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
