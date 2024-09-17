package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jcbowen/jcbaseGo/component/attachment/remote"
)

func main() {
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
	defer func(client remote.Client) {
		_ = client.Close()
	}(client)

	err = client.Upload("/remote/path/file.txt", []byte("Hello FTP"))
	if err != nil {
		log.Fatal(err)
	}

	files, err := client.List("/remote/path")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("FTP Files:", files)

	// 使用SFTP
	sftpConfig := remote.SFTPConfig{
		Address:  "sftp.example.com:22",
		Username: "username",
		Password: "password",
		// PrivateKey: []byte("..."), // 如果使用密钥认证
		// HostKeyCallback: ssh.FixedHostKey(hostKey), // 如果需要验证主机密钥
		Timeout: 10 * time.Second, // 可选
	}
	client, err = remote.NewClient(remote.TypeSFTP, sftpConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer func(client remote.Client) {
		_ = client.Close()
	}(client)

	err = client.Upload("/remote/path/file.txt", []byte("Hello SFTP"))
	if err != nil {
		log.Fatal(err)
	}

	files, err = client.List("/remote/path")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("SFTP Files:", files)

	// 使用腾讯云COS
	cosConfig := remote.COSConfig{
		BucketURL: "https://your-bucket.cos.region.myqcloud.com",
		SecretID:  "your-secret-id",
		SecretKey: "your-secret-key",
	}
	client, err = remote.NewClient(remote.TypeCOS, cosConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer func(client remote.Client) {
		_ = client.Close()
	}(client)

	err = client.Upload("remote/path/file.txt", []byte("Hello COS"))
	if err != nil {
		log.Fatal(err)
	}

	files, err = client.List("remote/path")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("COS Files:", files)

	// 使用阿里云OSS
	ossConfig := remote.OSSConfig{
		Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
		AccessKeyID:     "your-access-key-id",
		AccessKeySecret: "your-access-key-secret",
		BucketName:      "your-bucket-name",
	}
	client, err = remote.NewClient(remote.TypeOSS, ossConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer func(client remote.Client) {
		_ = client.Close()
	}(client)

	err = client.Upload("remote/path/file.txt", []byte("Hello OSS"))
	if err != nil {
		log.Fatal(err)
	}

	files, err = client.List("remote/path")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("OSS Files:", files)
}
