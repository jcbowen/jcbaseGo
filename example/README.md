# jcbaseGo 使用示例

本目录包含了 jcbaseGo 各个组件的使用示例，按照功能模块进行分类。

## 目录结构

```
example/
├── README.md                    # 本说明文件
├── attachment/                  # 附件管理组件示例
│   ├── ftp/                    # FTP 上传示例
│   │   └── main.go
│   ├── upload/                 # 本地文件上传示例
│   │   └── main.go
│   ├── oss/                    # 阿里云 OSS 示例（待补充）
│   │   └── .gitkeep
│   ├── cos/                    # 腾讯云 COS 示例（待补充）
│   │   └── .gitkeep
│   ├── sftp/                   # SFTP 示例（待补充）
│   │   └── .gitkeep
│   └── local/                  # 本地存储示例（待补充）
│       └── .gitkeep
├── helper/                     # 工具函数示例
│   ├── convert/                # 类型转换工具
│   │   └── main.go
│   ├── string/                 # 字符串处理工具
│   │   └── main.go
│   ├── json/                   # JSON 处理工具
│   │   ├── example.json
│   │   └── ExampleJson.go
│   ├── file/                   # 文件操作工具（待补充）
│   │   └── .gitkeep
│   ├── money/                  # 金额处理工具（待补充）
│   │   └── .gitkeep
│   ├── ssh/                    # SSH 工具（待补充）
│   │   └── .gitkeep
│   └── util/                   # 通用工具（待补充）
│       └── .gitkeep
├── mailer/                     # 邮件发送组件示例
│   └── main.go
├── orm/                        # 数据库 ORM 示例
│   ├── mysql/                  # MySQL 数据库示例
│   │   └── main.go
│   └── sqlite/                 # SQLite 数据库示例（待补充）
│       └── .gitkeep
├── redis/                      # Redis 缓存组件示例
│   └── main.go
├── security/                   # 安全组件示例
│   ├── sm4/                    # SM4 加密示例
│   │   └── main.go
│   └── aes/                    # AES 加密示例
│       └── main.go
└── validator/                  # 数据验证组件示例
    └── main.go
```

## 运行示例

### 1. 安全组件示例

```bash
# 运行 SM4 加密示例
go run example/security/sm4/main.go

# 运行 AES 加密示例
go run example/security/aes/main.go
```

### 2. 工具函数示例

```bash
# 运行类型转换示例
go run example/helper/convert/main.go

# 运行字符串处理示例
go run example/helper/string/main.go

# 运行 JSON 处理示例
go run example/helper/json/ExampleJson.go
```

### 3. 数据库 ORM 示例

```bash
# 运行 MySQL ORM 示例（需要配置数据库连接）
go run example/orm/mysql/main.go
```

### 4. 其他组件示例

```bash
# 运行邮件发送示例
go run example/mailer/main.go

# 运行 Redis 缓存示例
go run example/redis/main.go

# 运行数据验证示例
go run example/validator/main.go

# 运行附件上传示例
go run example/attachment/upload/main.go

# 运行 FTP 上传示例
go run example/attachment/ftp/main.go
```

## 示例说明

### 安全组件 (security/)
- **SM4**: 国密 SM4 对称加密算法，支持 CBC 和 GCM 模式
- **AES**: AES 对称加密算法，支持 16/24/32 字节密钥

### 工具函数 (helper/)
- **convert**: 类型转换工具，支持各种数据类型之间的转换
- **string**: 字符串处理工具，包含截取、替换、分割等功能
- **json**: JSON 处理工具，支持 JSON 序列化和反序列化

### 数据库 ORM (orm/)
- **mysql**: MySQL 数据库操作示例，包含 CRUD、事务等操作
- **sqlite**: SQLite 数据库操作示例（待补充）

### 其他组件
- **mailer**: 邮件发送功能示例
- **redis**: Redis 缓存操作示例
- **validator**: 数据验证功能示例
- **attachment**: 文件上传和管理示例

## 注意事项

1. **数据库示例**: 运行 MySQL 示例前需要先配置数据库连接信息
2. **邮件示例**: 运行邮件示例前需要配置 SMTP 服务器信息
3. **Redis 示例**: 运行 Redis 示例前需要确保 Redis 服务正在运行
4. **附件示例**: 运行附件上传示例前需要配置相应的存储服务

## 关于 .gitkeep 文件

项目中包含一些 `.gitkeep` 文件，这些文件的作用是：

- **保留空目录**: Git 不会跟踪空目录，使用 `.gitkeep` 文件可以保留目录结构
- **占位符**: 标记该目录为预留位置，后续会添加相应的示例代码
- **目录结构**: 确保项目的目录结构在 Git 仓库中保持一致

当向这些目录添加实际内容时，可以删除对应的 `.gitkeep` 文件。

## 扩展示例

如需添加新的示例，请按照以下规范：

1. 在对应的功能目录下创建子目录
2. 使用 `main.go` 作为示例文件名
3. 包含详细的中文注释
4. 提供完整的错误处理
5. 在 README.md 中更新说明

## 贡献

欢迎提交新的示例代码或改进现有示例。请确保：

- 代码符合项目的编码规范
- 包含适当的中文注释
- 提供完整的错误处理
- 测试通过后再提交 