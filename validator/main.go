package validator

import (
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const (
	InvalidIP = iota
	IPv4
	IPv6
)

// IsMobile 检查字符串是否为有效的手机号
func IsMobile(mobile string) bool {
	reg := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return reg.MatchString(mobile)
}

// IsEmail 检查字符串是否为有效的电子邮件地址
func IsEmail(email string) bool {
	// 基本格式校验
	if len(email) < 3 || len(email) > 254 {
		return false
	}

	// 正则表达式验证电子邮件格式
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	if !match {
		return false
	}

	// 进一步验证域名部分的合法性
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	domain := parts[1]
	if len(domain) < 3 {
		return false
	}

	return true
}

// IsURL 检查字符串是否为有效的 URL
func IsURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return strings.HasPrefix(u.Scheme, "http") || strings.HasPrefix(u.Scheme, "https")
}

// IsIP 检查字符串是否为有效的 IP 地址，并返回 IP 类型
func IsIP(ip string) (bool, int) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false, InvalidIP
	}
	if parsedIP.To4() != nil {
		return true, IPv4
	}
	if parsedIP.To16() != nil {
		return true, IPv6
	}
	return false, InvalidIP
}

// IsPort 检查字符串是否为有效的端口
func IsPort(portStr string) bool {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return false
	}
	if port < 0 || port > 65535 {
		return false
	}
	return true
}

// IsChineseIDCard 检查字符串是否为有效的中国大陆身份证号码（支持15位和18位）
func IsChineseIDCard(idCard string) bool {
	// 长度验证
	if len(idCard) != 15 && len(idCard) != 18 {
		return false
	}

	// 正则表达式验证格式
	var pattern string
	if len(idCard) == 15 {
		pattern = `^\d{15}$`
	} else {
		pattern = `^\d{17}[\dxX]$`
	}
	match, _ := regexp.MatchString(pattern, idCard)
	if !match {
		return false
	}

	// 验证区域码（省份代码）
	if !isValidRegionCode(idCard) {
		return false
	}

	// 验证生日
	if !isValidBirthday(idCard) {
		return false
	}

	// 验证18位身份证的校验码
	if len(idCard) == 18 && !isValidChecksum(idCard) {
		return false
	}

	return true
}

// isValidRegionCode 验证身份证的区域码（省份代码）
func isValidRegionCode(idCard string) bool {
	regionCode := idCard[:2]
	return regionCode >= "11" && regionCode <= "91"
}

// isValidBirthday 验证身份证的生日部分
func isValidBirthday(idCard string) bool {
	var birthday string
	if len(idCard) == 15 {
		birthday = "19" + idCard[6:12]
	} else {
		birthday = idCard[6:14]
	}
	_, err := strconv.ParseInt(birthday, 10, 64)
	return err == nil
}

// isValidChecksum 验证18位身份证的校验码
func isValidChecksum(idCard string) bool {
	idCard = strings.ToUpper(idCard) // 统一转为大写
	var (
		factor   = []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
		checksum = []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}
		sum      int
	)

	for i := 0; i < 17; i++ {
		num, _ := strconv.Atoi(string(idCard[i]))
		sum += num * factor[i]
	}
	mod := sum % 11
	return idCard[17] == checksum[mod]
}
