package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jcbowen/jcbaseGo/component/attachment/remote"
)

func main() {
	ctx := context.Background()

	// 使用FTP
	ftpConfig := remote.FTPConfig{
		Address:  "ftp.example.com:21",
		Username: "username",
		Password: "password",
		Timeout:  10 * time.Second, // 可选
	}
	client, err := remote.NewClient(remote.TypeFTP, ftpConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	err = client.Upload(ctx, "/remote/path/file.txt", []byte("Hello FTP"))
	if err != nil {
		log.Fatal(err)
	}

	data, err := client.Download(ctx, "/remote/path/file.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Downloaded data:", string(data))

	// 分页列举文件
	listOptions := remote.ListOptions{
		Prefix:  "/remote/path/",
		MaxKeys: 10,
	}
	for {
		result, err := client.List(ctx, listOptions)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("FTP Files:", result.Files)

		if !result.IsTruncated {
			break
		}
		listOptions.Marker = result.NextMarker
	}

	// 其他存储类型的使用方法类似，注意根据需要设置ListOptions和上下文
}
