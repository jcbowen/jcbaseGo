package security

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
)

// RSASign 使用 RSA 私钥对待签名内容进行 SHA256withRSA（RSA2）签名。
//
// 兼容三种私钥来源：
//  1. 裸 base64 私钥串（无 PEM 头尾、无换行），自动按 PKCS1 包装；
//  2. 标准 PEM 文本（PKCS1 "RSA PRIVATE KEY" 或 PKCS8 "PRIVATE KEY"）；
//  3. 私钥文件路径（keyFromFile=true 时按文件读取，内容同样支持上述两种格式）。
//
// 参数:
//   - data: 待签名的原始字符串（如支付平台签名串）
//   - privateKey: 私钥字符串或私钥文件路径
//   - keyFromFile: true 表示 privateKey 为文件路径
//
// 返回值:
//   - string: base64 编码后的签名值
//   - error: 私钥格式错误或签名失败时返回错误
//
// 示例:
//
//	sign, err := security.RSASign("待签名内容", "MIIEpAIBAAK...", false)
func RSASign(data, privateKey string, keyFromFile bool) (string, error) {
	if strings.TrimSpace(privateKey) == "" {
		return "", errors.New("私钥不能为空")
	}

	keyBytes := []byte(privateKey)
	if keyFromFile {
		content, err := os.ReadFile(privateKey)
		if err != nil {
			return "", fmt.Errorf("读取私钥文件失败: %w", err)
		}
		keyBytes = content
	}

	priKey, err := ParseRSAPrivateKey(keyBytes)
	if err != nil {
		return "", err
	}

	digest := sha256.Sum256([]byte(data))
	signature, err := rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", fmt.Errorf("RSA签名失败: %w", err)
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// RSAVerify 使用公钥对签名进行 SHA256withRSA（RSA2）验签。
//
// 兼容三种公钥来源：
//  1. 裸 base64 公钥串（无 PEM 头尾、无换行），自动按 X.509 SubjectPublicKeyInfo 包装；
//  2. 标准 PEM 公钥文本（"PUBLIC KEY" 或 "RSA PUBLIC KEY"）；
//  3. 公钥文件路径（keyFromFile=true 时按文件读取）。
//
// 参数:
//   - data: 被签名的原始内容
//   - sign: base64 编码的签名值
//   - publicKey: 公钥字符串或公钥文件路径
//   - keyFromFile: true 表示 publicKey 为文件路径
//
// 返回值:
//   - bool: true 验签通过；false 验签不通过
//   - error: 公钥格式错误等异常情况（验签不通过不算异常，返回 false,nil）
//
// 示例:
//
//	ok, err := security.RSAVerify("待验签内容", sign, platformPublicKey, false)
func RSAVerify(data, sign, publicKey string, keyFromFile bool) (bool, error) {
	if strings.TrimSpace(publicKey) == "" {
		return false, errors.New("公钥不能为空")
	}
	if strings.TrimSpace(sign) == "" {
		return false, errors.New("签名值不能为空")
	}

	keyBytes := []byte(publicKey)
	if keyFromFile {
		content, err := os.ReadFile(publicKey)
		if err != nil {
			return false, fmt.Errorf("读取公钥文件失败: %w", err)
		}
		keyBytes = content
	}

	pubKey, err := ParseRSAPublicKey(keyBytes)
	if err != nil {
		return false, err
	}

	signBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false, fmt.Errorf("签名值 base64 解码失败: %w", err)
	}

	digest := sha256.Sum256([]byte(data))
	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, digest[:], signBytes)
	if err != nil {
		// 验签不通过属于正常的业务结果，由返回的 false 表达，不向上抛错误
		return false, nil //nolint:nilerr // 有意吞掉 VerifyPKCS1v15 的错误，语义是「签名无效」而非「校验失败」
	}
	return true, nil
}

// ParseRSAPrivateKey 解析私钥字节为 rsa.PrivateKey。
//
// 自动兼容：PKCS1 PEM、PKCS8 PEM、裸 base64 串（自动包装为 PKCS1 PEM）。
//
// 参数:
//   - keyBytes: 私钥内容字节
//
// 返回值:
//   - *rsa.PrivateKey: 解析后的私钥对象
//   - error: 无法识别的格式时返回错误
func ParseRSAPrivateKey(keyBytes []byte) (*rsa.PrivateKey, error) {
	trimmed := strings.TrimSpace(string(keyBytes))
	// 裸私钥串：不含 PEM 头时包装为标准 PKCS1 PEM（每行 64 字符）
	if !strings.Contains(trimmed, "-----BEGIN") {
		trimmed = FormatPrivateKeyPEM(trimmed)
		keyBytes = []byte(trimmed)
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("私钥格式错误：无法解析 PEM 块，请检查是否为合法的 PKCS1/PKCS8 私钥")
	}

	// 优先尝试 PKCS1，其次 PKCS8
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("私钥解析失败（已尝试 PKCS1 与 PKCS8）: %w", err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("私钥类型不是 RSA: %T", key)
	}
	return rsaKey, nil
}

// ParseRSAPublicKey 解析公钥字节为 rsa.PublicKey。
//
// 自动兼容：X.509 SubjectPublicKeyInfo PEM（"PUBLIC KEY"）、PKCS1 PEM（"RSA PUBLIC KEY"）、
// 裸 base64 串（自动包装为 "PUBLIC KEY" PEM）。
//
// 参数:
//   - keyBytes: 公钥内容字节
//
// 返回值:
//   - *rsa.PublicKey: 解析后的公钥对象
//   - error: 无法识别的格式时返回错误
func ParseRSAPublicKey(keyBytes []byte) (*rsa.PublicKey, error) {
	trimmed := strings.TrimSpace(string(keyBytes))
	if !strings.Contains(trimmed, "-----BEGIN") {
		trimmed = FormatPublicKeyPEM(trimmed)
		keyBytes = []byte(trimmed)
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("公钥格式错误：无法解析 PEM 块")
	}

	// 优先尝试 X.509 SubjectPublicKeyInfo，其次 PKCS1
	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("公钥类型不是 RSA: %T", key)
		}
		return rsaKey, nil
	}
	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("公钥解析失败（已尝试 PKIX 与 PKCS1）: %w", err)
	}
	return key, nil
}

// FormatPrivateKeyPEM 将裸 base64 私钥串包装为标准 PKCS1 PEM 文本。
//
// 参数:
//   - rawKey: 不含 PEM 头尾的 base64 私钥串（可含空格，内部会清理）
//
// 返回值:
//   - string: 标准 PEM 格式私钥文本（每行 64 字符）
//
// 示例:
//
//	pemText := security.FormatPrivateKeyPEM("MIIEpAIBAAK...")
func FormatPrivateKeyPEM(rawKey string) string {
	body := strings.NewReplacer(" ", "", "\n", "", "\r", "").Replace(rawKey)
	var sb strings.Builder
	sb.WriteString("-----BEGIN RSA PRIVATE KEY-----\n")
	for i := 0; i < len(body); i += 64 {
		end := i + 64
		if end > len(body) {
			end = len(body)
		}
		sb.WriteString(body[i:end])
		sb.WriteString("\n")
	}
	sb.WriteString("-----END RSA PRIVATE KEY-----")
	return sb.String()
}

// FormatPublicKeyPEM 将裸 base64 公钥串包装为标准 X.509 SubjectPublicKeyInfo PEM 文本。
//
// 参数:
//   - rawKey: 不含 PEM 头尾的 base64 公钥串（可含空格，内部会清理）
//
// 返回值:
//   - string: 标准 PEM 格式公钥文本（每行 64 字符）
//
// 示例:
//
//	pemText := security.FormatPublicKeyPEM("MIIBIjANBgkq...")
func FormatPublicKeyPEM(rawKey string) string {
	body := strings.NewReplacer(" ", "", "\n", "", "\r", "").Replace(rawKey)
	var sb strings.Builder
	sb.WriteString("-----BEGIN PUBLIC KEY-----\n")
	for i := 0; i < len(body); i += 64 {
		end := i + 64
		if end > len(body) {
			end = len(body)
		}
		sb.WriteString(body[i:end])
		sb.WriteString("\n")
	}
	sb.WriteString("-----END PUBLIC KEY-----")
	return sb.String()
}
