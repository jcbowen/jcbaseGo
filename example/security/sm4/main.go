package main

import (
	"fmt"

	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/security"
)

func main() {
	fmt.Println("=== SM4 加密组件使用示例 ===\n")

	// 1. 使用默认值进行加解密
	fmt.Println("1. 使用默认值进行加解密:")
	defaultExample()

	// 2. 自定义密钥和IV进行加解密
	fmt.Println("\n2. 自定义密钥和IV进行加解密:")
	customExample()

	// 3. 不同加密模式测试
	fmt.Println("\n3. 不同加密模式测试:")
	modeExample()

	// 4. 不同编码格式测试
	fmt.Println("\n4. 不同编码格式测试:")
	encodingExample()

	// 5. 直接方法调用示例
	fmt.Println("\n5. 直接方法调用示例:")
	directMethodExample()

	// 6. 错误处理示例
	fmt.Println("\n6. 错误处理示例:")
	errorExample()

	// 7. 自定义参数测试
	fmt.Println("\n7. 自定义参数测试:")
	customExample2()
}

// defaultExample 使用默认值进行加解密
func defaultExample() {
	// 创建 SM4 实例，使用默认值
	sm4 := security.SM4{
		Text: "这是使用默认值的测试文本",
	}

	// 应用默认值
	helper.CheckAndSetDefault(&sm4)
	fmt.Printf("默认密钥: %s (长度: %d)\n", sm4.Key, len(sm4.Key))
	fmt.Printf("默认IV: %s (长度: %d)\n", sm4.Iv, len(sm4.Iv))
	fmt.Printf("默认模式: %s\n", sm4.Mode)

	// 加密
	var cipherText string
	err := sm4.Encrypt(&cipherText)
	if err != nil {
		fmt.Printf("加密失败: %v\n", err)
		return
	}
	fmt.Printf("加密结果: %s\n", cipherText)

	// 解密
	sm4.Text = cipherText
	var decryptedText string
	err = sm4.Decrypt(&decryptedText)
	if err != nil {
		fmt.Printf("解密失败: %v\n", err)
		return
	}
	fmt.Printf("解密结果: %s\n", decryptedText)
	fmt.Printf("加解密测试: %t\n", decryptedText == "这是使用默认值的测试文本")
}

// customExample 自定义密钥和IV进行加解密
func customExample() {
	// 自定义密钥和IV
	sm4 := security.SM4{
		Text: "自定义密钥测试",
		Key:  "my-custom-key-16", // 16字节密钥
		Iv:   "my-custom-iv-16b", // 16字节IV
		Mode: "CBC",
	}

	// 加密
	var cipherText string
	err := sm4.Encrypt(&cipherText)
	if err != nil {
		fmt.Printf("加密失败: %v\n", err)
		return
	}
	fmt.Printf("加密结果: %s\n", cipherText)

	// 解密
	sm4.Text = cipherText
	var decryptedText string
	err = sm4.Decrypt(&decryptedText)
	if err != nil {
		fmt.Printf("解密失败: %v\n", err)
		return
	}
	fmt.Printf("解密结果: %s\n", decryptedText)
}

// modeExample 不同加密模式测试
func modeExample() {
	// CBC 模式测试
	fmt.Println("CBC 模式:")
	sm4CBC := security.SM4{
		Text: "CBC模式测试",
		Key:  "1234567890123456",
		Iv:   "abcdefghijklmnop",
		Mode: "CBC",
	}

	var cipherText string
	err := sm4CBC.Encrypt(&cipherText)
	if err != nil {
		fmt.Printf("CBC加密失败: %v\n", err)
	} else {
		fmt.Printf("CBC加密结果: %s\n", cipherText)
	}

	// GCM 模式测试
	fmt.Println("GCM 模式:")
	sm4GCM := security.SM4{
		Text: "GCM模式测试",
		Key:  "1234567890123456",
		Mode: "GCM",
	}

	var gcmCipherText string
	err = sm4GCM.Encrypt(&gcmCipherText)
	if err != nil {
		fmt.Printf("GCM加密失败: %v\n", err)
	} else {
		fmt.Printf("GCM加密结果: %s\n", gcmCipherText)
	}
}

// encodingExample 不同编码格式测试
func encodingExample() {
	text := "编码格式测试文本"
	key := "1234567890123456"
	iv := "abcdefghijklmnop"

	encodings := []struct {
		name     string
		encoding string
		desc     string
	}{
		{"Std", "Std", "标准Base64编码（带填充）"},
		{"Raw", "Raw", "无填充Base64编码"},
		{"RawURL", "RawURL", "URL安全无填充Base64编码"},
		{"Hex", "Hex", "十六进制编码"},
	}

	for _, enc := range encodings {
		fmt.Printf("%s (%s):\n", enc.name, enc.desc)

		sm4 := security.SM4{
			Text:     text,
			Key:      key,
			Iv:       iv,
			Mode:     "CBC",
			Encoding: enc.encoding,
		}

		// 加密
		var cipherText string
		err := sm4.Encrypt(&cipherText)
		if err != nil {
			fmt.Printf("  加密失败: %v\n", err)
			continue
		}
		fmt.Printf("  加密结果: %s\n", cipherText)

		// 解密
		sm4.Text = cipherText
		var decryptedText string
		err = sm4.Decrypt(&decryptedText)
		if err != nil {
			fmt.Printf("  解密失败: %v\n", err)
			continue
		}
		fmt.Printf("  解密结果: %s\n", decryptedText)
		fmt.Printf("  验证结果: %t\n\n", decryptedText == text)
	}
}

// directMethodExample 直接方法调用示例
func directMethodExample() {
	text := "直接方法调用测试"
	key := "1234567890123456"
	iv := "abcdefghijklmnop"

	fmt.Println("CBC模式直接方法调用:")
	sm4CBC := security.SM4{
		Text: text,
		Key:  key,
		Iv:   iv,
		Mode: "CBC",
	}

	// 使用EncryptCBC和DecryptCBC直接方法
	var cipherText string
	err := sm4CBC.EncryptCBC(&cipherText)
	if err != nil {
		fmt.Printf("  CBC加密失败: %v\n", err)
	} else {
		fmt.Printf("  CBC加密结果: %s\n", cipherText)

		sm4CBC.Text = cipherText
		var decryptedText string
		err = sm4CBC.DecryptCBC(&decryptedText)
		if err != nil {
			fmt.Printf("  CBC解密失败: %v\n", err)
		} else {
			fmt.Printf("  CBC解密结果: %s\n", decryptedText)
			fmt.Printf("  CBC验证结果: %t\n", decryptedText == text)
		}
	}

	fmt.Println("\nGCM模式直接方法调用:")
	sm4GCM := security.SM4{
		Text: text,
		Key:  key,
		Mode: "GCM",
	}

	var gcmCipherText string
	err = sm4GCM.EncryptGCM(&gcmCipherText)
	if err != nil {
		fmt.Printf("  GCM加密失败: %v\n", err)
	} else {
		fmt.Printf("  GCM加密结果: %s\n", gcmCipherText)

		sm4GCM.Text = gcmCipherText
		var gcmDecryptedText string
		err = sm4GCM.DecryptGCM(&gcmDecryptedText)
		if err != nil {
			fmt.Printf("  GCM解密失败: %v\n", err)
		} else {
			fmt.Printf("  GCM解密结果: %s\n", gcmDecryptedText)
			fmt.Printf("  GCM验证结果: %t\n", gcmDecryptedText == text)
		}
	}
}

// errorExample 错误处理示例
func errorExample() {
	// 测试无效密钥长度
	fmt.Println("测试无效密钥长度:")
	sm4 := security.SM4{
		Text: "测试文本",
		Key:  "123", // 无效长度
		Iv:   "abcdefghijklmnop",
		Mode: "CBC",
	}

	var cipherText string
	err := sm4.Encrypt(&cipherText)
	if err != nil {
		fmt.Printf("  预期的错误: %v\n", err)
	} else {
		fmt.Println("  意外成功")
	}

	// 测试无效IV长度
	fmt.Println("\n测试无效IV长度:")
	sm4.Key = "1234567890123456"
	sm4.Iv = "123" // 无效长度

	err = sm4.Encrypt(&cipherText)
	if err != nil {
		fmt.Printf("  预期的错误: %v\n", err)
	} else {
		fmt.Println("  意外成功")
	}

	// 测试不支持的模式
	fmt.Println("\n测试不支持的模式:")
	sm4.Iv = "abcdefghijklmnop"
	sm4.Mode = "UNSUPPORTED"

	err = sm4.Encrypt(&cipherText)
	if err != nil {
		fmt.Printf("  预期的错误: %v\n", err)
	} else {
		fmt.Println("  意外成功")
	}

	// 测试GCM模式下的IV验证
	fmt.Println("\n测试GCM模式下的IV处理:")
	sm4GCM := security.SM4{
		Text: "GCM测试",
		Key:  "1234567890123456",
		Iv:   "GCM不需要IV", // GCM模式会忽略IV
		Mode: "GCM",
	}

	var gcmCipherText string
	err = sm4GCM.EncryptGCM(&gcmCipherText)
	if err != nil {
		fmt.Printf("  GCM加密失败: %v\n", err)
	} else {
		fmt.Println("  GCM加密成功（IV被忽略）")
	}
}

// customExample 自定义参数测试
func customExample2() {
	text := "86-527924-424955-7"

	fmt.Println("自定义参数测试:")
	sm4 := security.SM4{
		Text:     text,
		Encoding: "RawURL",
	}

	var cipherText string
	err := sm4.Encrypt(&cipherText)
	if err != nil {
		fmt.Printf("  加密失败: %v\n", err)
	} else {
		fmt.Printf("  加密结果: %s\n", cipherText)
	}

	sm4.Text = cipherText
	var plainText string
	err = sm4.Decrypt(&plainText)
	if err != nil {
		fmt.Printf("  解密失败: %v\n", err)
	} else {
		fmt.Printf("  解密结果: %s\n", plainText)
	}
}
