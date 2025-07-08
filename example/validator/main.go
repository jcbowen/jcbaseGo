package main

import (
	"fmt"

	"github.com/jcbowen/jcbaseGo/component/validator"
)

func main() {
	fmt.Println("=== 数据验证器使用示例 ===\n")

	// 1. 基本验证示例
	fmt.Println("1. 基本验证示例:")
	basicValidation()

	// 2. 网络相关验证示例
	fmt.Println("\n2. 网络相关验证示例:")
	networkValidation()

	// 3. 身份证验证示例
	fmt.Println("\n3. 身份证验证示例:")
	idCardValidation()

	// 4. 批量验证示例
	fmt.Println("\n4. 批量验证示例:")
	batchValidation()
}

// basicValidation 基本验证示例
func basicValidation() {
	// 验证邮箱
	emails := []string{
		"test@example.com",
		"invalid-email",
		"user.name@domain.org",
		"test@",
		"@domain.com",
	}

	fmt.Println("邮箱验证:")
	for _, email := range emails {
		if validator.IsEmail(email) {
			fmt.Printf("  ✅ %s: 有效邮箱\n", email)
		} else {
			fmt.Printf("  ❌ %s: 无效邮箱\n", email)
		}
	}

	// 验证手机号
	phones := []string{
		"13800138000",
		"123",
		"12345678901",
		"1380013800",
		"15812345678",
	}

	fmt.Println("\n手机号验证:")
	for _, phone := range phones {
		if validator.IsMobile(phone) {
			fmt.Printf("  ✅ %s: 有效手机号\n", phone)
		} else {
			fmt.Printf("  ❌ %s: 无效手机号\n", phone)
		}
	}
}

// networkValidation 网络相关验证示例
func networkValidation() {
	// 验证 URL
	urls := []string{
		"https://www.example.com",
		"http://localhost:8080",
		"ftp://files.example.com",
		"not-a-url",
		"https://",
	}

	fmt.Println("URL 验证:")
	for _, url := range urls {
		if validator.IsURL(url) {
			fmt.Printf("  ✅ %s: 有效URL\n", url)
		} else {
			fmt.Printf("  ❌ %s: 无效URL\n", url)
		}
	}

	// 验证 IP 地址
	ips := []string{
		"192.168.1.1",
		"10.0.0.1",
		"2001:db8::1",
		"256.256.256.256",
		"invalid-ip",
	}

	fmt.Println("\nIP 地址验证:")
	for _, ip := range ips {
		isValid, ipType := validator.IsIP(ip)
		if isValid {
			switch ipType {
			case validator.IPv4:
				fmt.Printf("  ✅ %s: 有效IPv4地址\n", ip)
			case validator.IPv6:
				fmt.Printf("  ✅ %s: 有效IPv6地址\n", ip)
			}
		} else {
			fmt.Printf("  ❌ %s: 无效IP地址\n", ip)
		}
	}

	// 验证端口
	ports := []string{
		"80",
		"443",
		"8080",
		"65535",
		"65536",
		"0",
		"-1",
		"abc",
	}

	fmt.Println("\n端口验证:")
	for _, port := range ports {
		if validator.IsPort(port) {
			fmt.Printf("  ✅ %s: 有效端口\n", port)
		} else {
			fmt.Printf("  ❌ %s: 无效端口\n", port)
		}
	}
}

// idCardValidation 身份证验证示例
func idCardValidation() {
	// 验证身份证号
	idCards := []string{
		"110101199001011234", // 18位身份证
		"110101900101123",    // 15位身份证
		"123456789012345678", // 无效18位
		"12345678901234",     // 无效15位
		"11010119900101123X", // 18位带X
		"invalid-id-card",
	}

	fmt.Println("身份证验证:")
	for _, idCard := range idCards {
		if validator.IsChineseIDCard(idCard) {
			fmt.Printf("  ✅ %s: 有效身份证号\n", idCard)
		} else {
			fmt.Printf("  ❌ %s: 无效身份证号\n", idCard)
		}
	}
}

// batchValidation 批量验证示例
func batchValidation() {
	// 批量验证邮箱
	emails := []string{
		"admin@company.com",
		"user.name@domain.org",
		"test+tag@example.net",
		"invalid.email",
		"@domain.com",
		"user@",
		"user@domain",
	}

	fmt.Println("批量邮箱验证:")
	validCount := 0
	for _, email := range emails {
		if validator.IsEmail(email) {
			validCount++
			fmt.Printf("  ✅ %s\n", email)
		} else {
			fmt.Printf("  ❌ %s\n", email)
		}
	}
	fmt.Printf("有效邮箱数量: %d/%d\n", validCount, len(emails))

	// 批量验证手机号
	phones := []string{
		"13800138000",
		"13912345678",
		"15098765432",
		"12345678901",
		"1380013800",
		"138001380001",
		"abc12345678",
	}

	fmt.Println("\n批量手机号验证:")
	validPhoneCount := 0
	for _, phone := range phones {
		if validator.IsMobile(phone) {
			validPhoneCount++
			fmt.Printf("  ✅ %s\n", phone)
		} else {
			fmt.Printf("  ❌ %s\n", phone)
		}
	}
	fmt.Printf("有效手机号数量: %d/%d\n", validPhoneCount, len(phones))

	// 验证结果统计
	fmt.Printf("\n验证统计:\n")
	fmt.Printf("邮箱验证通过率: %.1f%%\n", float64(validCount)/float64(len(emails))*100)
	fmt.Printf("手机号验证通过率: %.1f%%\n", float64(validPhoneCount)/float64(len(phones))*100)
}
