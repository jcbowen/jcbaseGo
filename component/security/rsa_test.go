package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// genTestKeyPair 生成测试用 RSA 2048 密钥对，返回裸 base64 私钥串与公钥串
func genTestKeyPair(t *testing.T) (priRaw, pubRaw string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}
	priDER := x509.MarshalPKCS1PrivateKey(key)
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("导出公钥失败: %v", err)
	}
	return base64.StdEncoding.EncodeToString(priDER), base64.StdEncoding.EncodeToString(pubDER)
}

// TestRSASignAndVerify 验证签名后可正常验签通过
func TestRSASignAndVerify(t *testing.T) {
	priRaw, pubRaw := genTestKeyPair(t)

	data := "app_id=2021000100000000,nonce=abc123,timestamp=1700000000000\nGET\n/v3/alipay/trade/query\n"
	sign, err := RSASign(data, priRaw, false)
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}

	ok, err := RSAVerify(data, sign, pubRaw, false)
	if err != nil {
		t.Fatalf("验签异常: %v", err)
	}
	if !ok {
		t.Fatal("验签应通过")
	}
}

// TestRSAVerifyTampered 验证内容被篡改后验签失败
func TestRSAVerifyTampered(t *testing.T) {
	priRaw, pubRaw := genTestKeyPair(t)

	sign, err := RSASign("original content", priRaw, false)
	if err != nil {
		t.Fatalf("签名失败: %v", err)
	}

	ok, err := RSAVerify("tampered content", sign, pubRaw, false)
	if err != nil {
		t.Fatalf("验签异常: %v", err)
	}
	if ok {
		t.Fatal("篡改后验签应失败")
	}
}

// TestRSASignWithPEMAndFile 验证 PEM 文本与文件路径两种来源等价
func TestRSASignWithPEMAndFile(t *testing.T) {
	priRaw, pubRaw := genTestKeyPair(t)

	// 文件方式
	dir := t.TempDir()
	priFile := filepath.Join(dir, "pri.pem")
	pubFile := filepath.Join(dir, "pub.pem")
	if err := os.WriteFile(priFile, []byte(FormatPrivateKeyPEM(priRaw)), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pubFile, []byte(FormatPublicKeyPEM(pubRaw)), 0644); err != nil {
		t.Fatal(err)
	}

	data := "hello alipay"
	sign1, err := RSASign(data, priRaw, false)
	if err != nil {
		t.Fatalf("裸串签名失败: %v", err)
	}
	sign2, err := RSASign(data, priFile, true)
	if err != nil {
		t.Fatalf("文件签名失败: %v", err)
	}
	if sign1 != sign2 {
		t.Fatal("裸串与文件方式签名结果应一致")
	}

	ok, err := RSAVerify(data, sign1, pubFile, true)
	if err != nil || !ok {
		t.Fatalf("文件公钥验签失败: ok=%v err=%v", ok, err)
	}

	// 标准带换行的 PEM 文本签名
	sign3, err := RSASign(data, FormatPrivateKeyPEM(priRaw), false)
	if err != nil {
		t.Fatalf("PEM 文本签名失败: %v", err)
	}
	if sign1 != sign3 {
		t.Fatal("裸串与 PEM 文本签名结果应一致")
	}
}

// TestRSASignWithPKCS8 验证 PKCS8 私钥的裸串与 PEM 文本两种形式均可正常签名
func TestRSASignWithPKCS8(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("导出 PKCS8 失败: %v", err)
	}
	priRaw := base64.StdEncoding.EncodeToString(der)

	data := "pkcs8 test"
	// 裸 PKCS8 base64 串
	sign1, err := RSASign(data, priRaw, false)
	if err != nil {
		t.Fatalf("PKCS8 裸串签名失败: %v", err)
	}

	// 标准 PKCS8 PEM 文本（"PRIVATE KEY" 头）
	var sb strings.Builder
	sb.WriteString("-----BEGIN PRIVATE KEY-----\n")
	for i := 0; i < len(priRaw); i += 64 {
		end := i + 64
		if end > len(priRaw) {
			end = len(priRaw)
		}
		sb.WriteString(priRaw[i:end])
		sb.WriteString("\n")
	}
	sb.WriteString("-----END PRIVATE KEY-----")
	sign2, err := RSASign(data, sb.String(), false)
	if err != nil {
		t.Fatalf("PKCS8 PEM 签名失败: %v", err)
	}
	if sign1 != sign2 {
		t.Fatal("PKCS8 裸串与 PEM 文本签名结果应一致")
	}

	// 用对应公钥验签
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("导出公钥失败: %v", err)
	}
	ok, err := RSAVerify(data, sign1, base64.StdEncoding.EncodeToString(pubDER), false)
	if err != nil || !ok {
		t.Fatalf("PKCS8 签名验签失败: ok=%v err=%v", ok, err)
	}
}

// TestRSASignInvalidKey 验证非法私钥返回错误
func TestRSASignInvalidKey(t *testing.T) {
	if _, err := RSASign("data", "not-a-valid-key!!", false); err == nil {
		t.Fatal("非法私钥应返回错误")
	}
	if _, err := RSASign("data", "", false); err == nil {
		t.Fatal("空私钥应返回错误")
	}
	if _, err := RSAVerify("data", "sign", "", false); err == nil {
		t.Fatal("空公钥应返回错误")
	}
}

// TestFormatKeyPEM 验证 PEM 包装输出格式
func TestFormatKeyPEM(t *testing.T) {
	out := FormatPrivateKeyPEM("AAAA  BBBB\nCCCC")
	if !strings.HasPrefix(out, "-----BEGIN RSA PRIVATE KEY-----") {
		t.Fatal("私钥 PEM 应含正确头")
	}
	if strings.Contains(strings.Split(out, "\n")[1], " ") {
		t.Fatal("PEM 内容行不应含空格")
	}

	pubOut := FormatPublicKeyPEM("AAAA\nBBBB")
	if !strings.HasPrefix(pubOut, "-----BEGIN PUBLIC KEY-----") {
		t.Fatal("公钥 PEM 应含正确头")
	}

	// 包装后可被 pem.Decode
	block, _ := pem.Decode([]byte(pubOut))
	if block == nil {
		t.Fatal("包装后的公钥应可被 pem.Decode 解析")
	}
}
