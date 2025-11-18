package helper

import (
	"net"
	"strings"
)

// IP 提供IP地址处理相关的工具函数
type IP struct {
	IP string
}

// NewIP 创建一个新的IP实例
// 参数:
//   - ip: IP地址字符串
//
// 返回值:
//   - *IP: IP实例指针
func NewIP(ip string) *IP {
	return &IP{
		IP: ip,
	}
}

// IsValid 验证IP地址是否有效
// 参数:
//   - 无
//
// 返回值:
//   - bool: 如果IP有效返回true，否则返回false
func (ip *IP) IsValid() bool {
	if ip.IP == "" {
		return false
	}

	// 检查是否为本地回环地址
	if ip.IsLoopback() {
		return false
	}

	// 验证IP格式
	parsedIP := net.ParseIP(ip.IP)
	if parsedIP == nil {
		return false
	}

	// 检查是否为私有地址（可选，根据需求决定是否过滤）
	if ip.IsPrivate() {
		// 可以根据实际需求决定是否允许私有地址
		// 这里暂时允许私有地址通过验证
	}

	return true
}

// IsLoopback 检查是否为本地回环地址
// 参数:
//   - 无
//
// 返回值:
//   - bool: 如果是回环地址返回true，否则返回false
func (ip *IP) IsLoopback() bool {
	// 常见的本地回环地址
	loopbackIPs := []string{
		"127.0.0.1", "::1", "localhost",
		"127.0.0.0", "127.255.255.255", // 127.0.0.0/8 网段
		"0.0.0.0", "0:0:0:0:0:0:0:0", "0:0:0:0:0:0:0:1",
		"fe80::1%lo0", // IPv6 本地回环
	}

	for _, loopback := range loopbackIPs {
		if ip.IP == loopback {
			return true
		}
	}

	// 检查是否为 127.0.0.0/8 网段
	if strings.HasPrefix(ip.IP, "127.") {
		return true
	}

	// 检查是否为 IPv6 回环地址 ::1
	if ip.IP == "::1" {
		return true
	}

	return false
}

// IsPrivate 检查是否为私有IP地址
// 参数:
//   - 无
//
// 返回值:
//   - bool: 如果是私有地址返回true，否则返回false
func (ip *IP) IsPrivate() bool {
	parsedIP := net.ParseIP(ip.IP)
	if parsedIP == nil {
		return false
	}

	// 检查是否为私有地址范围
	// 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16
	if parsedIP.IsPrivate() {
		return true
	}

	// 检查是否为链路本地地址
	if parsedIP.IsLinkLocalUnicast() || parsedIP.IsLinkLocalMulticast() {
		return true
	}

	return false
}

// IsAllowed 检查IP是否被允许（基于白名单和黑名单）
// 参数:
//   - whitelist: IP白名单列表，如果提供则只允许白名单内的IP
//   - blacklist: IP黑名单列表，如果提供则拒绝黑名单内的IP
//
// 返回值:
//   - bool: 如果IP被允许返回true，否则返回false
func (ip *IP) IsAllowed(whitelist, blacklist []string) bool {
	// 如果提供了白名单，检查IP是否在白名单中
	if len(whitelist) > 0 {
		for _, allowedIP := range whitelist {
			if ip.IP == allowedIP || ip.IsInCIDR(allowedIP) {
				return true
			}
		}
		return false // 不在白名单中，拒绝
	}

	// 如果提供了黑名单，检查IP是否在黑名单中
	if len(blacklist) > 0 {
		for _, deniedIP := range blacklist {
			if ip.IP == deniedIP || ip.IsInCIDR(deniedIP) {
				return false // 在黑名单中，拒绝
			}
		}
	}

	// 如果没有提供过滤列表，或者IP不在黑名单中，允许访问
	return true
}

// IsInCIDR 检查IP是否在CIDR范围内
// 参数:
//   - cidr: CIDR表示法（如 "192.168.1.0/24" 或 "2001:db8::/32"）
//
// 返回值:
//   - bool: 如果IP在CIDR范围内返回true，否则返回false
func (ip *IP) IsInCIDR(cidr string) bool {
	// 解析IP地址
	parsedIP := net.ParseIP(ip.IP)
	if parsedIP == nil {
		return false
	}

	// 解析CIDR
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		// 如果不是CIDR格式，尝试作为普通IP处理
		if cidrIP := net.ParseIP(cidr); cidrIP != nil {
			return parsedIP.Equal(cidrIP)
		}
		return false
	}

	// 检查IP是否在CIDR范围内
	return ipNet.Contains(parsedIP)
}

// GetValidClientIP 从IP列表中获取有效的客户端IP
// 该方法会优先选择公网IP，如果没有公网IP则选择私有IP，最后选择第一个有效IP
//
// 参数:
//   - ips: IP地址列表
//
// 返回值:
//   - string: 有效的客户端IP地址，如果找不到则返回空字符串
func GetValidClientIP(ips []string) string {
	// 从前往后遍历，找到第一个有效的公网IP
	// 跳过可能的本地地址和私有地址（除非没有其他选择）
	var candidateIP string

	for _, ip := range ips {
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			continue
		}

		// 优先选择公网IP
		if !parsedIP.IsPrivate() && !parsedIP.IsLoopback() {
			return ip
		}

		// 如果没有公网IP，记录第一个有效的私有IP作为候选
		if candidateIP == "" && !parsedIP.IsLoopback() {
			candidateIP = ip
		}
	}

	// 如果没有找到公网IP，返回第一个有效的私有IP
	if candidateIP != "" {
		return candidateIP
	}

	// 如果连私有IP都没有，返回第一个有效的IP
	if len(ips) > 0 {
		return ips[0]
	}

	return ""
}

// GetRealIPFromHeaders 从HTTP头信息中获取真实客户端IP
// 函数名：GetRealIPFromHeaders
// 参数：
// - headers map[string]string — 经过提取与归一化的请求头映射
// 返回值：
// - string — 解析出的真实客户端IP；当不可判定时返回空字符串
// 异常：不触发 panic
// 使用示例：
//
//	ip := GetRealIPFromHeaders(map[string]string{
//	    "Ali-Cdn-Real-Ip":  "203.0.113.1",
//	    "X-Forwarded-For":  "203.0.113.1, 198.51.100.2",
//	    "X-Real-IP":        "198.51.100.2",
//	})
//	// ip == "203.0.113.1"
func GetRealIPFromHeaders(headers map[string]string) string {
	// 1) 先读各厂商/通用的“客户端真实IP”专用头（单值）
	for _, k := range []string{"Ali-Cdn-Real-Ip", "CF-Connecting-IP", "True-Client-IP", "X-Client-IP"} {
		v := strings.TrimSpace(headers[k])
		if v != "" && NewIP(v).IsValid() {
			return v
		}
	}

	// 2) 再读标准代理链：X-Forwarded-For（取第一个有效客户端IP，优先公网）
	if xff := strings.TrimSpace(headers["X-Forwarded-For"]); xff != "" {
		ips := SplitAndValidateIPs(xff)
		if len(ips) > 0 {
			if ip := GetValidClientIP(ips); ip != "" && NewIP(ip).IsValid() {
				return ip
			}
		}
	}

	// 3) 兜底使用 X-Real-IP（常见反向代理传入）
	if xr := strings.TrimSpace(headers["X-Real-IP"]); xr != "" && NewIP(xr).IsValid() {
		return xr
	}

	// 4) 进一步兜底：历史/部分代理头（需校验）
	for _, k := range []string{"X-Cluster-Client-IP", "X-Originating-IP"} {
		v := strings.TrimSpace(headers[k])
		if v != "" && NewIP(v).IsValid() {
			return v
		}
	}

	// 不返回非IP头（如 X-Forwarded、X-Forwarded-Host）以避免误判域名/主机名
	return ""
}

// SplitIPs 分割多个IP地址字符串
// 该方法会将逗号分隔的IP字符串分割成数组
//
// 参数:
//   - s: 包含多个IP地址的字符串，用逗号分隔
//
// 返回值:
//   - []string: 分割后的IP地址列表
func SplitIPs(s string) []string {
	var result []string
	for _, item := range strings.Split(s, ",") {
		result = append(result, strings.TrimSpace(item))
	}
	return result
}

// SplitAndValidateIPs 分割并验证多个IP地址
// 该方法会将逗号分隔的IP字符串分割成数组，并验证每个IP的有效性
//
// 参数:
//   - s: 包含多个IP地址的字符串，用逗号分隔
//
// 返回值:
//   - []string: 经过验证的有效IP地址列表
func SplitAndValidateIPs(s string) []string {
	var result []string
	ips := SplitIPs(s)

	for _, ip := range ips {
		if NewIP(ip).IsValid() {
			result = append(result, ip)
		}
	}

	return result
}
