package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/attachment"
	"github.com/jcbowen/jcbaseGo/component/attachment/remote"
)

// main 演示如何生成腾讯云 COS 预签名上传 URL，供客户端直传文件。
func main() {
	r := gin.Default()

	r.POST("/presign-cos", func(c *gin.Context) {
		// 基础配置，存储类型必须为 cos
		baseConfig := &jcbaseGo.AttachmentStruct{
			StorageType: "cos",
			LocalDir:    "uploads",
		}

		// COS 远程配置，请替换为真实配置
		cosConfig := jcbaseGo.COSStruct{
			SecretId:  "your-secret-id",
			SecretKey: "your-secret-key",
			Region:    "ap-guangzhou",
			Bucket:    "your-bucket-1250000000",
			Url:       "https://your-bucket-1250000000.cos.ap-guangzhou.myqcloud.com",
		}

		// 创建附件实例
		att := attachment.New(c, baseConfig, cosConfig)

		// 生成预签名上传 URL，有效期 10 分钟
		opts := &remote.PresignOptions{
			Expires: 10 * time.Minute,
			// ContentType: "image/jpeg", // 如需限制文件类型可取消注释
			// Metadata:    map[string]string{"x-cos-meta-uid": "12345"}, // 自定义元数据需使用 x-cos-meta- 前缀
		}
		url, headers, err := att.GetPresignURL(c.Request.Context(), "images/2024/01/example.jpg", opts)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"message": "预签名 URL 生成成功",
			"data": gin.H{
				"url":     url,
				"method":  "PUT",
				"headers": headers,
			},
		})
	})

	if err := r.Run(":8080"); err != nil {
		fmt.Printf("服务启动失败: %v\n", err)
	}
}
