package security

import (
	"testing"
)

// TestAES_EncryptDecrypt 测试AES加解密功能
func TestAES_EncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		key       string
		iv        string
		wantErr   bool
	}{
		{
			name:      "正常加解密测试",
			plaintext: "Hello, AES World!",
			key:       "1234567890123456", // 16字节密钥
			iv:        "abcdefghijklmnop", // 16字节IV
			wantErr:   false,
		},
		{
			name:      "中文文本加解密测试",
			plaintext: "这是一个AES加密测试",
			key:       "1234567890123456",
			iv:        "abcdefghijklmnop",
			wantErr:   false,
		},
		{
			name:      "空文本加解密测试",
			plaintext: "",
			key:       "1234567890123456",
			iv:        "abcdefghijklmnop",
			wantErr:   false,
		},
		{
			name:      "长文本加解密测试",
			plaintext: "这是一个很长的文本用于测试AES加密算法的性能和正确性，包含中英文混合内容：Hello World! 123456789",
			key:       "1234567890123456",
			iv:        "abcdefghijklmnop",
			wantErr:   false,
		},
		{
			name:      "24字节密钥测试",
			plaintext: "24字节密钥测试",
			key:       "123456789012345678901234", // 24字节密钥
			iv:        "abcdefghijklmnop",
			wantErr:   false,
		},
		{
			name:      "32字节密钥测试",
			plaintext: "32字节密钥测试",
			key:       "12345678901234567890123456789012", // 32字节密钥
			iv:        "abcdefghijklmnop",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aesInstance := AES{
				Text: tt.plaintext,
				Key:  tt.key,
				Iv:   tt.iv,
			}

			// 测试加密
			var cipherText string
			err := aesInstance.Encrypt(&cipherText)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cipherText == "" {
				t.Error("Encrypt() 加密结果为空")
				return
			}

			// 测试解密
			aesInstance.Text = cipherText
			var decryptedText string
			err = aesInstance.Decrypt(&decryptedText)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && decryptedText != tt.plaintext {
				t.Errorf("Decrypt() 解密结果不匹配，期望: %s, 实际: %s", tt.plaintext, decryptedText)
			}
		})
	}
}

// TestAES_DefaultValues 测试AES默认值功能
func TestAES_DefaultValues(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		wantErr   bool
	}{
		{
			name:      "使用默认值测试",
			plaintext: "测试默认值",
			wantErr:   false,
		},
		{
			name:      "空文本默认值测试",
			plaintext: "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aesInstance := AES{
				Text: tt.plaintext,
				// 不设置 Key 和 Iv，使用默认值
			}

			// 测试加密
			var cipherText string
			err := aesInstance.Encrypt(&cipherText)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cipherText == "" {
				t.Error("Encrypt() 加密结果为空")
				return
			}

			// 测试解密
			aesInstance.Text = cipherText
			var decryptedText string
			err = aesInstance.Decrypt(&decryptedText)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && decryptedText != tt.plaintext {
				t.Errorf("Decrypt() 解密结果不匹配，期望: %s, 实际: %s", tt.plaintext, decryptedText)
			}
		})
	}
}

// TestAES_KeyValidation 测试密钥验证
func TestAES_KeyValidation(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "16字节密钥",
			key:     "1234567890123456",
			wantErr: false,
		},
		{
			name:    "24字节密钥",
			key:     "123456789012345678901234",
			wantErr: false,
		},
		{
			name:    "32字节密钥",
			key:     "12345678901234567890123456789012",
			wantErr: false,
		},
		{
			name:    "密钥过短",
			key:     "123456789012345", // 15字节
			wantErr: true,
		},
		{
			name:    "密钥过长",
			key:     "12345678901234567", // 17字节
			wantErr: true,
		},
		{
			name:    "空密钥",
			key:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAES_IvValidation 测试IV验证
func TestAES_IvValidation(t *testing.T) {
	tests := []struct {
		name    string
		iv      string
		wantErr bool
	}{
		{
			name:    "正确长度IV",
			iv:      "abcdefghijklmnop", // 16字节
			wantErr: false,
		},
		{
			name:    "IV过短",
			iv:      "abcdefghijklmno", // 15字节
			wantErr: true,
		},
		{
			name:    "IV过长",
			iv:      "abcdefghijklmnopq", // 17字节
			wantErr: true,
		},
		{
			name:    "空IV",
			iv:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIv(tt.iv)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateIv() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAES_PKCS7Padding 测试PKCS7填充功能
func TestAES_PKCS7Padding(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		blockSize int
		wantLen   int
	}{
		{
			name:      "需要填充的数据",
			data:      []byte("Hello"),
			blockSize: 16,
			wantLen:   16,
		},
		{
			name:      "刚好一个块大小的数据",
			data:      []byte("1234567890123456"),
			blockSize: 16,
			wantLen:   32, // 需要填充一个完整的块
		},
		{
			name:      "空数据",
			data:      []byte{},
			blockSize: 16,
			wantLen:   16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			padded := pkcs7Pad(tt.data, tt.blockSize)

			if len(padded) != tt.wantLen {
				t.Errorf("pkcs7Pad() 长度不匹配，期望: %d, 实际: %d", tt.wantLen, len(padded))
			}

			if len(padded)%tt.blockSize != 0 {
				t.Errorf("pkcs7Pad() 结果不是块大小的倍数")
			}

			// 测试去填充
			unpadded, err := pkcs7Unpad(padded, tt.blockSize)
			if err != nil {
				t.Errorf("pkcs7Unpad() error = %v", err)
				return
			}

			if string(unpadded) != string(tt.data) {
				t.Errorf("pkcs7Unpad() 结果不匹配，期望: %s, 实际: %s", string(tt.data), string(unpadded))
			}
		})
	}
}

// TestAES_PKCS7UnpadErrors 测试PKCS7去填充错误情况
func TestAES_PKCS7UnpadErrors(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		blockSize int
		wantErr   bool
	}{
		{
			name:      "空数据",
			data:      []byte{},
			blockSize: 16,
			wantErr:   true,
		},
		{
			name:      "不是块大小倍数的数据",
			data:      []byte("123456789012345"), // 15字节
			blockSize: 16,
			wantErr:   true,
		},
		{
			name:      "无效的填充长度",
			data:      []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 0}, // 填充长度为0
			blockSize: 16,
			wantErr:   true,
		},
		{
			name:      "填充长度超过块大小",
			data:      []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 17}, // 填充长度为17
			blockSize: 16,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pkcs7Unpad(tt.data, tt.blockSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("pkcs7Unpad() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAES_ErrorCases 测试AES错误情况
func TestAES_ErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		key       string
		iv        string
		wantErr   bool
	}{
		{
			name:      "无效密钥长度",
			plaintext: "测试文本",
			key:       "123", // 无效长度
			iv:        "abcdefghijklmnop",
			wantErr:   true,
		},
		{
			name:      "无效IV长度",
			plaintext: "测试文本",
			key:       "1234567890123456",
			iv:        "123", // 无效长度
			wantErr:   true,
		},
		{
			name:      "无效密钥长度",
			plaintext: "测试文本",
			key:       "123", // 无效长度
			iv:        "abcdefghijklmnop",
			wantErr:   true,
		},
		{
			name:      "无效IV长度",
			plaintext: "测试文本",
			key:       "1234567890123456",
			iv:        "123", // 无效长度
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aesInstance := AES{
				Text: tt.plaintext,
				Key:  tt.key,
				Iv:   tt.iv,
			}

			// 测试加密
			var cipherText string
			err := aesInstance.Encrypt(&cipherText)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAES_DecryptErrorCases 测试AES解密错误情况
func TestAES_DecryptErrorCases(t *testing.T) {
	tests := []struct {
		name       string
		cipherText string
		key        string
		iv         string
		wantErr    bool
	}{
		{
			name:       "无效的Base64密文",
			cipherText: "invalid-base64!@#",
			key:        "1234567890123456",
			iv:         "abcdefghijklmnop",
			wantErr:    true,
		},
		{
			name:       "密文长度过短",
			cipherText: "aGVsbG8=", // "hello" 的base64，但长度小于16字节
			key:        "1234567890123456",
			iv:         "abcdefghijklmnop",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aesInstance := AES{
				Text: tt.cipherText,
				Key:  tt.key,
				Iv:   tt.iv,
			}

			var decryptedText string
			err := aesInstance.Decrypt(&decryptedText)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
