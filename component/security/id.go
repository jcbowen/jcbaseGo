package security

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// IDCipher uint64 ID 加解密器。
//
// 底层复用 SM4（CBC + 标准 Base64），并在标准 Base64 基础上做 URL-safe 字符替换
// （+ → -、/ → _）且保留 = 填充。注意：SM4 自带的 Encoding:"RawURL" 会去掉填充，
// 若业务需要保留填充，应使用 "Std" 编码并依赖本实现的字符替换。
type IDCipher struct {
	sm4               *SM4
	urlSafeReplacer   *strings.Replacer
	stdBase64Replacer *strings.Replacer
}

// NewIDCipher 创建 ID 加解密器。
// sm4 需预先配置 Mode/Encoding（Key/Iv 可留空，加解密时会应用默认值）。
// 若需要保留 base64 填充，sm4.Encoding 应使用 "Std"。
func NewIDCipher(sm4 *SM4) *IDCipher {
	return &IDCipher{
		sm4:               sm4,
		urlSafeReplacer:   strings.NewReplacer("+", "-", "/", "_"),
		stdBase64Replacer: strings.NewReplacer("-", "+", "_", "/"),
	}
}

// EncryptUint64 将 uint64 ID 加密为 URL-safe base64 字符串（保留填充）。
//
// 参数:
//   - id: 待加密的 ID
//
// 返回值:
//   - string: 加密后的 URL-safe base64 字符串
//   - error: SM4 未配置或加密失败时返回错误
func (c *IDCipher) EncryptUint64(id uint64) (string, error) {
	if c == nil || c.sm4 == nil {
		return "", errors.New("IDCipher 未初始化 SM4")
	}

	s := *c.sm4
	s.Text = strconv.FormatUint(id, 10)
	var cipherText string
	if err := s.Encrypt(&cipherText); err != nil {
		return "", err
	}
	return c.urlSafeReplacer.Replace(cipherText), nil
}

// DecryptUint64 解密 URL-safe base64 字符串为 uint64（兼容 URL-safe 与标准 base64 两种字符集）。
// 解密前将 URL-safe 字符归一化为标准 base64；标准 base64 密文不含 - 和 _，不受影响。
//
// 参数:
//   - cipherText: 加密后的 ID 字符串
//
// 返回值:
//   - uint64: 解密后的 ID
//   - error: SM4 未配置、解密失败或结果无法解析为 uint64 时返回错误
func (c *IDCipher) DecryptUint64(cipherText string) (uint64, error) {
	if c == nil || c.sm4 == nil {
		return 0, errors.New("IDCipher 未初始化 SM4")
	}

	s := *c.sm4
	s.Text = c.stdBase64Replacer.Replace(cipherText)
	var plain string
	if err := s.Decrypt(&plain); err != nil {
		return 0, fmt.Errorf("decrypt id failed: %w", err)
	}
	v, err := strconv.ParseUint(plain, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse decrypted id failed: %w", err)
	}
	return v, nil
}
