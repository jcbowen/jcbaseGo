# Attachment 组件

附件管理组件提供统一的文件上传、存储和管理功能，支持本地存储和多种云存储服务。

## 概述

Attachment 组件是一个功能强大的文件附件管理工具，支持多种文件类型（图片、音频、视频、办公文件、压缩文件等）的上传、存储和访问。组件提供统一的 API 接口，支持本地存储、腾讯云 COS、阿里云 OSS、FTP 和 SFTP 等多种存储方式。

## 功能特性

- **多存储支持**：支持本地存储、腾讯云 COS、阿里云 OSS、FTP、SFTP
- **多种文件类型**：支持图片、音频、视频、办公文件、压缩文件等
- **灵活配置**：可自定义文件大小限制、允许的文件扩展名
- **自动目录管理**：按年月自动创建目录结构
- **Base64 支持**：支持 Base64 编码的文件数据上传
- **图片处理**：自动获取图片尺寸信息
- **错误处理**：完善的错误处理和验证机制
- **回调函数**：支持保存前的自定义回调函数

## 快速开始

### 安装

```go
go get github.com/jcbowen/jcbaseGo/component/attachment
```

### 基本使用示例

```go
package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/attachment"
)

func main() {
    r := gin.Default()

    r.POST("/upload", func(c *gin.Context) {
        // 获取上传的文件
        file, err := c.FormFile("file")
        if err != nil {
            c.JSON(400, gin.H{"error": "文件上传失败"})
            return
        }

        // 配置附件存储
        config := &jcbaseGo.AttachmentStruct{
            StorageType: "local",
            LocalDir:    "uploads",
        }

        // 创建附件实例
        att := attachment.New(c, config)

        // 配置上传选项
        opt := &attachment.Options{
            FileData: file,
            FileType: "image",
            MaxSize:  5 * 1024 * 1024, // 5MB
            AllowExt: []string{".jpg", ".jpeg", ".png", ".gif"},
        }

        // 上传并保存文件
        result := att.Upload(opt).Save()

        if result.HasError() {
            c.JSON(400, gin.H{"error": result.Error().Error()})
            return
        }

        // 返回上传结果
        c.JSON(200, gin.H{
            "message": "文件上传成功",
            "data": gin.H{
                "filename": result.FileName,
                "filepath": result.FileAttachment,
                "filesize": result.FileSize,
                "width":    result.Width,
                "height":   result.Height,
                "md5":      result.FileMD5,
            },
        })
    })

    r.Run(":8080")
}
```

### Base64 图片上传示例

```go
package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/attachment"
)

func main() {
    r := gin.Default()

    r.POST("/upload-base64", func(c *gin.Context) {
        var req struct {
            ImageData string `json:"image_data"`
        }
        
        if err := c.BindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": "请求参数错误"})
            return
        }

        config := &jcbaseGo.AttachmentStruct{
            StorageType: "local",
            LocalDir:    "uploads",
        }

        att := attachment.New(c, config)

        opt := &attachment.Options{
            FileData: req.ImageData, // Base64 编码的图片数据
            FileType: "image",
            MaxSize:  2 * 1024 * 1024, // 2MB
            AllowExt: []string{".jpg", ".jpeg", ".png"},
        }

        result := att.Upload(opt).Save()

        if result.HasError() {
            c.JSON(400, gin.H{"error": result.Error().Error()})
            return
        }

        c.JSON(200, gin.H{
            "message": "Base64 图片上传成功",
            "data": gin.H{
                "filename": result.FileName,
                "filepath": result.FileAttachment,
                "full_url": att.ToMedia(result.FileAttachment),
            },
        })
    })

    r.Run(":8080")
}
```

### 使用回调函数示例

```go
package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/attachment"
)

func main() {
    r := gin.Default()

    r.POST("/upload-with-callback", func(c *gin.Context) {
        file, err := c.FormFile("file")
        if err != nil {
            c.JSON(400, gin.H{"error": "文件上传失败"})
            return
        }

        config := &jcbaseGo.AttachmentStruct{
            StorageType: "local",
            LocalDir:    "uploads",
        }

        att := attachment.New(c, config)

        opt := &attachment.Options{
            FileData: file,
            FileType: "image",
            MaxSize:  5 * 1024 * 1024,
        }

        // 设置保存前的回调函数
        att.SetBeforeSave(func(a *attachment.Attachment) bool {
            // 检查文件大小是否超过 1MB
            if a.FileSize > 1*1024*1024 {
                a.addError(fmt.Errorf("文件大小不能超过 1MB"))
                return false
            }
            
            // 可以在这里添加其他自定义验证逻辑
            fmt.Printf("文件信息: %s, 大小: %d bytes\n", a.FileName, a.FileSize)
            return true
        })

        result := att.Upload(opt).Save()

        if result.HasError() {
            c.JSON(400, gin.H{"error": result.Error().Error()})
            return
        }

        c.JSON(200, gin.H{
            "message": "文件上传成功",
            "data": gin.H{
                "filename": result.FileName,
                "filepath": result.FileAttachment,
            },
        })
    })

    r.Run(":8080")
}
```

## 详细配置

### AttachmentStruct 配置

AttachmentStruct 是附件存储的基础配置结构体：

```go
type AttachmentStruct struct {
    StorageType      string `json:"storage_type" default:"local"` // 存储类型: local/cos/oss/ftp/sftp
    LocalDir         string `json:"local_dir" default:"uploads"`  // 本地存储目录
    VisitDomain      string `json:"visit_domain" default:"/"`     // 访问域名
    LocalVisitDomain string `json:"local_visit_domain"`           // 本地访问域名
}
```

### Options 配置

Options 是文件上传的选项配置：

```go
type Options struct {
    Group    string      // 附件分组
    FileData interface{} // 文件数据: *multipart.FileHeader, string(base64), []byte
    FileType string      // 文件类型: image/voice/video/office/zip
    MaxSize  int64       // 最大文件大小（字节）
    AllowExt []string    // 允许的文件扩展名
}
```

### 预定义文件类型

组件预定义了多种文件类型的配置：

```go
// 图片类型
Types["image"] = &typeInfo{
    TypeName: "图片",
    AllowExt: []string{".gif", ".jpg", ".jpeg", ".bmp", ".png", ".ico"},
    MaxSize:  5 * 1024 * 1024, // 5MB
}

// 音频类型  
Types["voice"] = &typeInfo{
    TypeName: "音频",
    AllowExt: []string{".mp3", ".wma", ".wav", ".amr"},
    MaxSize:  50 * 1024 * 1024, // 50MB
}

// 视频类型
Types["video"] = &typeInfo{
    TypeName: "视频",
    AllowExt: []string{".rm", ".rmvb", ".wmv", ".avi", ".mpg", ".mpeg", ".mp4"},
    MaxSize:  300 * 1024 * 1024, // 300MB
}

// 办公文件类型
Types["office"] = &typeInfo{
    TypeName: "办公文件",
    AllowExt: []string{".wps", ".wpt", ".doc", ".dot", ".docx", ".docm", ".dotm", 
                      ".et", ".ett", ".xls", ".xlt", ".xlsx", ".xlsm", ".xltx", ".xltm", ".xlsb",
                      ".dps", ".dpt", ".ppt", ".pps", ".pot", ".pptx", ".ppsx", ".potx",
                      ".txt", ".csv", ".prn", ".pdf", ".xml"},
    MaxSize:  50 * 1024 * 1024, // 50MB
}

// 压缩文件类型
Types["zip"] = &typeInfo{
    TypeName: "压缩文件",
    AllowExt: []string{".zip", ".rar"},
    MaxSize:  500 * 1024 * 1024, // 500MB
}
```

## 高级功能

### 使用云存储（腾讯云 COS）

```go
package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/attachment"
)

func main() {
    r := gin.Default()

    r.POST("/upload-cos", func(c *gin.Context) {
        file, err := c.FormFile("file")
        if err != nil {
            c.JSON(400, gin.H{"error": "文件上传失败"})
            return
        }

        // 基础配置
        config := &jcbaseGo.AttachmentStruct{
            StorageType: "cos",
            LocalDir:    "uploads",
        }

        // 腾讯云 COS 配置
        cosConfig := jcbaseGo.COSStruct{
            SecretId:  "your-secret-id",
            SecretKey: "your-secret-key",
            Region:    "ap-beijing",
            BucketName: "your-bucket-name",
            CustomizeVisitDomain: "https://your-cdn-domain.com",
        }

        // 创建附件实例（传入 COS 配置）
        att := attachment.New(c, config, cosConfig)

        opt := &attachment.Options{
            FileData: file,
            FileType: "image",
            MaxSize:  10 * 1024 * 1024, // 10MB
        }

        result := att.Upload(opt).Save()

        if result.HasError() {
            c.JSON(400, gin.H{"error": result.Error().Error()})
            return
        }

        c.JSON(200, gin.H{
            "message": "文件上传到 COS 成功",
            "data": gin.H{
                "filename": result.FileName,
                "filepath": result.FileAttachment,
                "full_url": att.ToMedia(result.FileAttachment),
            },
        })
    })

    r.Run(":8080")
}
```

### 使用分组管理

```go
package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/jcbowen/jcbaseGo"
    "github.com/jcbowen/jcbaseGo/component/attachment"
)

func main() {
    r := gin.Default()

    r.POST("/upload-group", func(c *gin.Context) {
        file, err := c.FormFile("file")
        if err != nil {
            c.JSON(400, gin.H{"error": "文件上传失败"})
            return
        }

        config := &jcbaseGo.AttachmentStruct{
            StorageType: "local",
            LocalDir:    "uploads",
        }

        att := attachment.New(c, config)

        // 使用分组管理
        opt := &attachment.Options{
            Group:    "user_avatar", // 分组名称
            FileData: file,
            FileType: "image",
            MaxSize:  2 * 1024 * 1024,
        }

        result := att.Upload(opt).Save()

        if result.HasError() {
            c.JSON(400, gin.H{"error": result.Error().Error()})
            return
        }

        c.JSON(200, gin.H{
            "message": "分组文件上传成功",
            "data": gin.H{
                "filename": result.FileName,
                "filepath": result.FileAttachment,
                "group":   "user_avatar",
            },
        })
    })

    r.Run(":8080")
}
```

### 生成访问 URL

```go
package main

import (
    "fmt"
    "github.com/jcbowen/jcbaseGo/component/attachment"
)

func main() {
    // 创建附件实例（不需要 gin 上下文）
    att := attachment.New(nil, nil)

    // 生成完整的访问 URL
    filePath := "images/2023/10/0112345678901234567890.jpg"
    
    // 本地访问 URL
    localURL := att.ToMedia(filePath, true, false)
    fmt.Printf("本地访问 URL: %s\n", localURL)
    
    // 远程访问 URL（带缓存时间戳）
    remoteURL := att.ToMedia(filePath, false, true)
    fmt.Printf("远程访问 URL: %s\n", remoteURL)
    
    // 不带时间戳的远程访问 URL
    remoteURLNoCache := att.ToMedia(filePath, false, false)
    fmt.Printf("不带缓存的远程访问 URL: %s\n", remoteURLNoCache)
}
```

## 错误处理

### 错误检查方法

```go
result := att.Upload(opt).Save()

// 检查是否有错误
if result.HasError() {
    // 获取第一个错误
    fmt.Printf("错误: %v\n", result.Error())
    
    // 获取所有错误
    errors := result.Errors()
    for _, err := range errors {
        fmt.Printf("详细错误: %v\n", err)
    }
}
```

### 常见错误类型

- **文件类型不支持**：文件扩展名不在允许的列表中
- **文件大小超限**：文件大小超过配置的最大限制
- **目录创建失败**：无法创建文件存储目录
- **文件保存失败**：无法将文件保存到目标位置
- **Base64 解析失败**：Base64 数据格式不正确

## 性能优化建议

1. **文件大小限制**：根据实际需求设置合理的文件大小限制
2. **存储类型选择**：根据访问频率选择合适的存储类型
3. **CDN 加速**：对于图片等静态资源，建议使用 CDN 加速
4. **异步上传**：对于大文件上传，可以考虑使用异步处理

## 安全考虑

- 严格验证文件类型，防止恶意文件上传
- 设置合理的文件大小限制
- 对上传的文件进行病毒扫描（如果适用）
- 使用安全的存储服务，配置适当的访问权限

## API 参考

### 主要结构体

#### Attachment
附件管理的主要结构体

**字段：**
- `Opt *Options` - 上传选项配置
- `GinContext *gin.Context` - Gin 上下文
- `BaseConfig *jcbaseGo.AttachmentStruct` - 基础配置
- `RemoteConfig interface{}` - 远程存储配置
- `FileType string` - 文件类型
- `FileName string` - 文件名
- `FileSize int64` - 文件大小
- `FileAttachment string` - 文件相对路径
- `FileMD5 string` - 文件 MD5 值
- `FileExt string` - 文件扩展名
- `Width int` - 图片宽度
- `Height int` - 图片高度

**方法：**
- `New(args ...interface{}) *Attachment` - 创建附件实例
- `Upload(opt *Options) *Attachment` - 配置上传选项
- `SetBeforeSave(fn func(a *Attachment) bool) *Attachment` - 设置保存前回调
- `Save() *Attachment` - 保存文件
- `ToMedia(src string, args ...interface{}) string` - 生成访问 URL
- `HasError() bool` - 检查是否有错误
- `Error() error` - 获取第一个错误
- `Errors() []error` - 获取所有错误

#### Options
上传选项配置结构体

**字段：**
- `Group string` - 附件分组
- `FileData interface{}` - 文件数据
- `FileType string` - 文件类型
- `MaxSize int64` - 最大文件大小
- `AllowExt []string` - 允许的文件扩展名

## 版本历史

- v1.0.0：初始版本，包含基本的文件上传和管理功能
- 支持多种存储类型和文件类型

## 贡献指南

欢迎提交 Issue 和 Pull Request 来改进 Attachment 组件。

## 许可证

Attachment 组件遵循 MIT 许可证。