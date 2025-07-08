package main

import (
	"fmt"

	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/security"
)

func main() {
	fmt.Println("=== AES 加密组件使用示例 ===\n")

	// 1. 使用默认值进行加解密
	fmt.Println("1. 使用默认值进行加解密:")
	defaultExample()

	// 2. 不同密钥长度测试
	fmt.Println("\n2. 不同密钥长度测试:")
	keyLengthExample()

	// 3. 自定义密钥和IV进行加解密
	fmt.Println("\n3. 自定义密钥和IV进行加解密:")
	customExample()

	// 4. 错误处理示例
	fmt.Println("\n4. 错误处理示例:")
	errorExample()
}

// defaultExample 使用默认值进行加解密
func defaultExample() {
	// 创建 AES 实例，使用默认值
	aes := security.AES{
		Text: "这是使用默认值的AES测试文本",
	}

	// 应用默认值
	helper.CheckAndSetDefault(&aes)
	fmt.Printf("默认密钥: %s (长度: %d)\n", aes.Key, len(aes.Key))
	fmt.Printf("默认IV: %s (长度: %d)\n", aes.Iv, len(aes.Iv))

	// 加密
	var cipherText string
	err := aes.Encrypt(&cipherText)
	if err != nil {
		fmt.Printf("加密失败: %v\n", err)
		return
	}
	fmt.Printf("加密结果: %s\n", cipherText)

	// 解密
	aes.Text = cipherText
	var decryptedText string
	err = aes.Decrypt(&decryptedText)
	if err != nil {
		fmt.Printf("解密失败: %v\n", err)
		return
	}
	fmt.Printf("解密结果: %s\n", decryptedText)
	fmt.Printf("加解密测试: %t\n", decryptedText == "这是使用默认值的AES测试文本")
}

// keyLengthExample 不同密钥长度测试
func keyLengthExample() {
	// 16字节密钥测试
	fmt.Println("16字节密钥测试:")
	aes16 := security.AES{
		Text: "16字节密钥测试",
		Key:  "1234567890123456", // 16字节密钥
		Iv:   "abcdefghijklmnop", // 16字节IV
	}

	var cipherText string
	err := aes16.Encrypt(&cipherText)
	if err != nil {
		fmt.Printf("16字节密钥加密失败: %v\n", err)
	} else {
		fmt.Printf("16字节密钥加密结果: %s\n", cipherText)
	}

	// 24字节密钥测试
	fmt.Println("24字节密钥测试:")
	aes24 := security.AES{
		Text: "24字节密钥测试",
		Key:  "123456789012345678901234", // 24字节密钥
		Iv:   "abcdefghijklmnop",         // 16字节IV
	}

	var cipherText24 string
	err = aes24.Encrypt(&cipherText24)
	if err != nil {
		fmt.Printf("24字节密钥加密失败: %v\n", err)
	} else {
		fmt.Printf("24字节密钥加密结果: %s\n", cipherText24)
	}

	// 32字节密钥测试
	fmt.Println("32字节密钥测试:")
	aes32 := security.AES{
		Text: "32字节密钥测试",
		Key:  "12345678901234567890123456789012", // 32字节密钥
		Iv:   "abcdefghijklmnop",                 // 16字节IV
	}

	var cipherText32 string
	err = aes32.Encrypt(&cipherText32)
	if err != nil {
		fmt.Printf("32字节密钥加密失败: %v\n", err)
	} else {
		fmt.Printf("32字节密钥加密结果: %s\n", cipherText32)
	}
}

// customExample 自定义密钥和IV进行加解密
func customExample() {
	// 自定义密钥和IV
	aes := security.AES{
		Text: "自定义AES密钥测试",
		Key:  "my-aes-key-16bytes", // 16字节密钥
		Iv:   "my-aes-iv-16bytes",  // 16字节IV
	}

	// 加密
	var cipherText string
	err := aes.Encrypt(&cipherText)
	if err != nil {
		fmt.Printf("加密失败: %v\n", err)
		return
	}
	fmt.Printf("加密结果: %s\n", cipherText)

	// 解密
	aes.Text = cipherText
	var decryptedText string
	err = aes.Decrypt(&decryptedText)
	if err != nil {
		fmt.Printf("解密失败: %v\n", err)
		return
	}
	fmt.Printf("解密结果: %s\n", decryptedText)
}

// errorExample 错误处理示例
func errorExample() {
	// 测试无效密钥长度
	fmt.Println("测试无效密钥长度:")
	aes := security.AES{
		Text: "测试文本",
		Key:  "123", // 无效长度
		Iv:   "abcdefghijklmnop",
	}

	var cipherText string
	err := aes.Encrypt(&cipherText)
	if err != nil {
		fmt.Printf("预期的错误: %v\n", err)
	} else {
		fmt.Println("意外成功")
	}

	// 测试无效IV长度
	fmt.Println("测试无效IV长度:")
	aes.Key = "1234567890123456"
	aes.Iv = "123" // 无效长度

	err = aes.Encrypt(&cipherText)
	if err != nil {
		fmt.Printf("预期的错误: %v\n", err)
	} else {
		fmt.Println("意外成功")
	}

	// 测试无效的Base64密文
	fmt.Println("测试无效的Base64密文:")
	aes.Iv = "abcdefghijklmnop"
	aes.Text = "invalid-base64!@#"

	var decryptedText string
	err = aes.Decrypt(&decryptedText)
	if err != nil {
		fmt.Printf("预期的错误: %v\n", err)
	} else {
		fmt.Println("意外成功")
	}
}
