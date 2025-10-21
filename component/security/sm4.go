package security

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/emmansun/gmsm/sm4"
	"github.com/jcbowen/jcbaseGo/component/helper"
)

// SM4 SM4加解密结构体
// 提供SM4对称加密算法的加解密功能
type SM4 struct {
	Text     string `json:"text" default:""`                // 待加密/解密的文本内容，实际加密或解密操作将基于该文本进行
	Key      string `json:"key" default:"jcbase.sm4_key__"` // 加密密钥，SM4算法要求密钥长度必须为16字节，用于加密和解密数据
	Iv       string `json:"iv" default:"jcbase.sm4_iv___"`  // 初始化向量（Initialization Vector），SM4算法要求IV长度必须为16字节，用于增强加密安全性
	Mode     string `json:"mode" default:"CBC"`             // 加密模式，可选值包括：CBC、GCM、CFB、OFB、CTR，指定加密操作使用的具体模式
	Encoding string `json:"encoding" default:"Std"`         // 输出/输入编码格式，可选值包括：Std（标准Base64）、Raw（无填充Base64）、RawURL（URL安全无填充Base64）、Hex（十六进制）
}

// EncryptCBC 使用CBC模式加密数据
// 参数：
//   - cipherText: 加密后的密文输出指针
//
// 返回：
//   - error: 加密失败时的错误信息
func (s SM4) EncryptCBC(cipherText *string) error {
	_ = helper.CheckAndSetDefault(&s)

	if err := validateSM4Key(s.Key); err != nil {
		return err
	}
	if err := validateSM4Iv(s.Iv); err != nil {
		return err
	}

	block, err := sm4.NewCipher([]byte(s.Key))
	if err != nil {
		return fmt.Errorf("创建SM4密码器失败: %w", err)
	}

	// 使用PKCS7填充
	plainText := pkcs7Pad([]byte(s.Text), sm4.BlockSize)
	cipherByteArr := make([]byte, len(plainText))
	ivBytes := []byte(s.Iv)

	mode := cipher.NewCBCEncrypter(block, ivBytes)
	mode.CryptBlocks(cipherByteArr, plainText)

	*cipherText = s.encodeBytes(cipherByteArr)
	return nil
}

// DecryptCBC 使用CBC模式解密数据
// 参数：
//   - plaintext: 解密后的明文输出指针
//
// 返回：
//   - error: 解密失败时的错误信息
func (s SM4) DecryptCBC(plaintext *string) error {
	_ = helper.CheckAndSetDefault(&s)

	if err := validateSM4Key(s.Key); err != nil {
		return err
	}
	if err := validateSM4Iv(s.Iv); err != nil {
		return err
	}

	block, err := sm4.NewCipher([]byte(s.Key))
	if err != nil {
		return fmt.Errorf("创建SM4密码器失败: %w", err)
	}

	cipherBytes, err := s.decodeString(s.Text)
	if err != nil {
		return fmt.Errorf("密文解码失败: %w", err)
	}

	if len(cipherBytes) < sm4.BlockSize {
		return errors.New("密文长度过短")
	}

	ivBytes := []byte(s.Iv)
	mode := cipher.NewCBCDecrypter(block, ivBytes)
	mode.CryptBlocks(cipherBytes, cipherBytes)

	plainByteArr, err := pkcs7Unpad(cipherBytes, sm4.BlockSize)
	if err != nil {
		return fmt.Errorf("去除填充失败: %w", err)
	}

	*plaintext = string(plainByteArr)
	return nil
}

// EncryptGCM 使用GCM模式加密数据（推荐使用，提供认证加密）
// 参数：
//   - cipherText: 加密后的密文输出指针
//
// 返回：
//   - error: 加密失败时的错误信息
func (s SM4) EncryptGCM(cipherText *string) error {
	_ = helper.CheckAndSetDefault(&s)

	if err := validateSM4Key(s.Key); err != nil {
		return err
	}

	block, err := sm4.NewCipher([]byte(s.Key))
	if err != nil {
		return fmt.Errorf("创建SM4密码器失败: %w", err)
	}

	sm4gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("创建GCM模式失败: %w", err)
	}

	// 生成随机nonce
	nonce := make([]byte, sm4gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("生成随机nonce失败: %w", err)
	}

	// 加密数据
	cipherBytes := sm4gcm.Seal(nil, nonce, []byte(s.Text), nil)

	// 将nonce和密文组合后进行base64编码
	result := append(nonce, cipherBytes...)
	*cipherText = s.encodeBytes(result)
	return nil
}

// DecryptGCM 使用GCM模式解密数据
// 参数：
//   - plaintext: 解密后的明文输出指针
//
// 返回：
//   - error: 解密失败时的错误信息
func (s SM4) DecryptGCM(plaintext *string) error {
	_ = helper.CheckAndSetDefault(&s)

	if err := validateSM4Key(s.Key); err != nil {
		return err
	}

	block, err := sm4.NewCipher([]byte(s.Key))
	if err != nil {
		return fmt.Errorf("创建SM4密码器失败: %w", err)
	}

	sm4gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("创建GCM模式失败: %w", err)
	}

	// 解码base64数据
	data, err := s.decodeString(s.Text)
	if err != nil {
		return fmt.Errorf("密文解码失败: %w", err)
	}

	nonceSize := sm4gcm.NonceSize()
	if len(data) < nonceSize {
		return errors.New("密文长度过短")
	}

	// 分离nonce和密文
	nonce := data[:nonceSize]
	cipherBytes := data[nonceSize:]

	// 解密数据
	plainBytes, err := sm4gcm.Open(nil, nonce, cipherBytes, nil)
	if err != nil {
		return fmt.Errorf("GCM解密失败: %w", err)
	}

	*plaintext = string(plainBytes)
	return nil
}

// Encrypt 通用加密方法，根据Mode字段选择加密模式
// 参数：
//   - cipherText: 加密后的密文输出指针
//
// 返回：
//   - error: 加密失败时的错误信息
func (s SM4) Encrypt(cipherText *string) error {
	_ = helper.CheckAndSetDefault(&s)

	switch s.Mode {
	case "CBC":
		return s.EncryptCBC(cipherText)
	case "GCM":
		return s.EncryptGCM(cipherText)
	default:
		return fmt.Errorf("不支持的加密模式: %s", s.Mode)
	}
}

// Decrypt 通用解密方法，根据Mode字段选择解密模式
// 参数：
//   - plaintext: 解密后的明文输出指针
//
// 返回：
//   - error: 解密失败时的错误信息
func (s SM4) Decrypt(plaintext *string) error {
	_ = helper.CheckAndSetDefault(&s)

	switch s.Mode {
	case "CBC":
		return s.DecryptCBC(plaintext)
	case "GCM":
		return s.DecryptGCM(plaintext)
	default:
		return fmt.Errorf("不支持的解密模式: %s", s.Mode)
	}
}

// validateSM4Key 验证SM4密钥是否有效
// SM4密钥必须为16字节长度
func validateSM4Key(key string) error {
	if len(key) != 16 {
		return fmt.Errorf("SM4密钥长度必须为16字节，当前长度: %d", len(key))
	}
	return nil
}

// validateSM4Iv 验证SM4初始化向量是否有效
// SM4的IV必须为16字节长度
func validateSM4Iv(iv string) error {
	if len(iv) != sm4.BlockSize {
		return fmt.Errorf("SM4初始化向量长度必须为%d字节，当前长度: %d", sm4.BlockSize, len(iv))
	}
	return nil
}

func (s SM4) encodeBytes(b []byte) string {
	switch s.Encoding {
	case "Raw":
		return base64.RawStdEncoding.EncodeToString(b)
	case "RawURL":
		return base64.RawURLEncoding.EncodeToString(b)
	case "Hex":
		return hex.EncodeToString(b)
	default:
		return base64.StdEncoding.EncodeToString(b)
	}
}

func (s SM4) decodeString(str string) ([]byte, error) {
	switch s.Encoding {
	case "Raw":
		return base64.RawStdEncoding.DecodeString(str)
	case "RawURL":
		return base64.RawURLEncoding.DecodeString(str)
	case "Hex":
		return hex.DecodeString(str)
	default:
		return base64.StdEncoding.DecodeString(str)
	}
}
