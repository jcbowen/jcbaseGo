package jcbaseGo

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
	TablePrefix   string `json:"tablePrefix" default:"jc_"`    // 表前缀
	ParseTime     string `json:"parseTime" default:"False"`    // 是否开启时间解析
	SingularTable string `json:"singularTable" default:"true"` // 使用单数表名
}

// SqlLiteStruct sqlite配置
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

// MysqlModel gorm基础模型
type MysqlModel struct {
	Id        uint   `gorm:"column:id;type:INT(11) UNSIGNED;primaryKey;autoIncrement" json:"id"`
	UpdatedAt string `gorm:"column:updated_at;type:DATETIME;default:NULL;comment:更新时间" json:"updated_at"`       // 更新时间
	CreatedAt string `gorm:"column:created_at;type:DATETIME;default:NULL;comment:创建时间" json:"created_at"`       // 创建时间
	DeletedAt string `gorm:"column:deleted_at;type:DATETIME;index;default:NULL;comment:删除时间" json:"deleted_at"` // 删除时间
}

// SQLLiteModel gorm基础模型
type SQLLiteModel struct {
	Id        uint   `gorm:"column:id;type:INTEGER;primaryKey;autoIncrement" json:"id"`
	UpdatedAt string `gorm:"column:updated_at;type:STRING;default:NULL" json:"updated_at"`       // 更新时间
	CreatedAt string `gorm:"column:created_at;type:STRING;default:NULL" json:"created_at"`       // 创建时间
	DeletedAt string `gorm:"column:deleted_at;type:STRING;index;default:NULL" json:"deleted_at"` // 删除时间
}
