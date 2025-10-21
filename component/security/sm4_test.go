package security

import (
	"strings"
	"testing"
)

// TestSM4_EncryptDecryptCBC 测试SM4 CBC模式加解密
func TestSM4_EncryptDecryptCBC(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		key       string
		iv        string
		wantErr   bool
	}{
		{
			name:      "正常加解密测试",
			plaintext: "Hello, SM4 World!",
			key:       "1234567890123456", // 16字节密钥
			iv:        "abcdefghijklmnop", // 16字节IV
			wantErr:   false,
		},
		{
			name:      "中文文本加解密测试",
			plaintext: "这是一个SM4加密测试",
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
			plaintext: "这是一个很长的文本用于测试SM4加密算法的性能和正确性，包含中英文混合内容：Hello World! 123456789",
			key:       "1234567890123456",
			iv:        "abcdefghijklmnop",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm4Instance := SM4{
				Text: tt.plaintext,
				Key:  tt.key,
				Iv:   tt.iv,
				Mode: "CBC",
			}

			// 测试加密
			var cipherText string
			err := sm4Instance.EncryptCBC(&cipherText)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptCBC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cipherText == "" {
				t.Error("EncryptCBC() 加密结果为空")
				return
			}

			// 测试解密
			sm4Instance.Text = cipherText
			var decryptedText string
			err = sm4Instance.DecryptCBC(&decryptedText)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptCBC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && decryptedText != tt.plaintext {
				t.Errorf("DecryptCBC() 解密结果不匹配，期望: %s, 实际: %s", tt.plaintext, decryptedText)
			}
		})
	}
}

// TestSM4_EncryptDecryptGCM 测试SM4 GCM模式加解密
func TestSM4_EncryptDecryptGCM(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		key       string
		wantErr   bool
	}{
		{
			name:      "正常GCM加解密测试",
			plaintext: "Hello, SM4 GCM!",
			key:       "1234567890123456",
			wantErr:   false,
		},
		{
			name:      "中文GCM加解密测试",
			plaintext: "这是SM4 GCM模式测试",
			key:       "1234567890123456",
			wantErr:   false,
		},
		{
			name:      "空文本GCM测试",
			plaintext: "",
			key:       "1234567890123456",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm4Instance := SM4{
				Text: tt.plaintext,
				Key:  tt.key,
				Mode: "GCM",
			}

			// 测试加密
			var cipherText string
			err := sm4Instance.EncryptGCM(&cipherText)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptGCM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cipherText == "" {
				t.Error("EncryptGCM() 加密结果为空")
				return
			}

			// 测试解密
			sm4Instance.Text = cipherText
			var decryptedText string
			err = sm4Instance.DecryptGCM(&decryptedText)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptGCM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && decryptedText != tt.plaintext {
				t.Errorf("DecryptGCM() 解密结果不匹配，期望: %s, 实际: %s", tt.plaintext, decryptedText)
			}
		})
	}
}

// TestSM4_UniversalEncryptDecrypt 测试通用加解密方法
func TestSM4_UniversalEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		key       string
		iv        string
		mode      string
		wantErr   bool
	}{
		{
			name:      "CBC模式通用方法测试",
			plaintext: "Universal CBC Test",
			key:       "1234567890123456",
			iv:        "abcdefghijklmnop",
			mode:      "CBC",
			wantErr:   false,
		},
		{
			name:      "GCM模式通用方法测试",
			plaintext: "Universal GCM Test",
			key:       "1234567890123456",
			iv:        "", // GCM模式不需要IV
			mode:      "GCM",
			wantErr:   false,
		},
		{
			name:      "不支持的模式测试",
			plaintext: "Test",
			key:       "1234567890123456",
			iv:        "abcdefghijklmnop",
			mode:      "UNSUPPORTED",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm4Instance := SM4{
				Text: tt.plaintext,
				Key:  tt.key,
				Iv:   tt.iv,
				Mode: tt.mode,
			}

			// 测试加密
			var cipherText string
			err := sm4Instance.Encrypt(&cipherText)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return // 期望错误的情况，不继续测试解密
			}

			if cipherText == "" {
				t.Error("Encrypt() 加密结果为空")
				return
			}

			// 测试解密
			sm4Instance.Text = cipherText
			var decryptedText string
			err = sm4Instance.Decrypt(&decryptedText)
			if err != nil {
				t.Errorf("Decrypt() error = %v", err)
				return
			}

			if decryptedText != tt.plaintext {
				t.Errorf("Decrypt() 解密结果不匹配，期望: %s, 实际: %s", tt.plaintext, decryptedText)
			}
		})
	}
}

// TestSM4_KeyValidation 测试密钥验证
func TestSM4_KeyValidation(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "正确长度密钥",
			key:     "1234567890123456", // 16字节
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
			err := validateSM4Key(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSM4Key() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestSM4_IvValidation 测试IV验证
func TestSM4_IvValidation(t *testing.T) {
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
			err := validateSM4Iv(tt.iv)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSM4Iv() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSM4_EncodingOptionsCBC(t *testing.T) {
	isHexLower := func(s string) bool {
		if len(s)%2 != 0 { // hex should be even length
			return false
		}
		for _, c := range s {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				return false
			}
		}
		return true
	}

	cases := []struct {
		name     string
		encoding string
	}{
		{name: "Std Base64", encoding: "Std"},
		{name: "Raw Base64 (no padding)", encoding: "Raw"},
		{name: "RawURL Base64 (url-safe no padding)", encoding: "RawURL"},
		{name: "Hex", encoding: "Hex"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sm4Instance := SM4{
				Text:     "Hello, SM4 CBC with Encoding!",
				Key:      "1234567890123456",
				Iv:       "abcdefghijklmnop",
				Mode:     "CBC",
				Encoding: tc.encoding,
			}

			var cipherText string
			if err := sm4Instance.EncryptCBC(&cipherText); err != nil {
				t.Fatalf("EncryptCBC error: %v", err)
			}
			if cipherText == "" {
				t.Fatalf("EncryptCBC produced empty cipher text")
			}

			switch tc.encoding {
			case "Raw":
				if strings.Contains(cipherText, "=") {
					t.Errorf("Raw encoding should not contain '=', got: %s", cipherText)
				}
			case "RawURL":
				if strings.ContainsAny(cipherText, "+/=") {
					t.Errorf("RawURL encoding should not contain '+', '/', '='; got: %s", cipherText)
				}
			case "Hex":
				if !isHexLower(cipherText) {
					t.Errorf("Hex encoding should be lowercase hex without symbols, got: %s", cipherText)
				}
			}

			// Decrypt using same encoding
			sm4Instance.Text = cipherText
			var plain string
			if err := sm4Instance.DecryptCBC(&plain); err != nil {
				t.Fatalf("DecryptCBC error: %v", err)
			}
			if plain != "Hello, SM4 CBC with Encoding!" {
				t.Errorf("DecryptCBC mismatch, want %q, got %q", "Hello, SM4 CBC with Encoding!", plain)
			}
		})
	}
}

func TestSM4_EncodingOptionsGCM(t *testing.T) {
	isHexLower := func(s string) bool {
		if len(s)%2 != 0 {
			return false
		}
		for _, c := range s {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				return false
			}
		}
		return true
	}

	cases := []struct {
		name     string
		encoding string
	}{
		{name: "Std Base64", encoding: "Std"},
		{name: "Raw Base64 (no padding)", encoding: "Raw"},
		{name: "RawURL Base64 (url-safe no padding)", encoding: "RawURL"},
		{name: "Hex", encoding: "Hex"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sm4Instance := SM4{
				Text:     "Hello, SM4 GCM with Encoding!",
				Key:      "1234567890123456",
				Mode:     "GCM",
				Encoding: tc.encoding,
			}

			var cipherText string
			if err := sm4Instance.EncryptGCM(&cipherText); err != nil {
				t.Fatalf("EncryptGCM error: %v", err)
			}
			if cipherText == "" {
				t.Fatalf("EncryptGCM produced empty cipher text")
			}

			switch tc.encoding {
			case "Raw":
				if strings.Contains(cipherText, "=") {
					t.Errorf("Raw encoding should not contain '=', got: %s", cipherText)
				}
			case "RawURL":
				if strings.ContainsAny(cipherText, "+/=") {
					t.Errorf("RawURL encoding should not contain '+', '/', '='; got: %s", cipherText)
				}
			case "Hex":
				if !isHexLower(cipherText) {
					t.Errorf("Hex encoding should be lowercase hex without symbols, got: %s", cipherText)
				}
			}

			// Decrypt using same encoding
			sm4Instance.Text = cipherText
			var plain string
			if err := sm4Instance.DecryptGCM(&plain); err != nil {
				t.Fatalf("DecryptGCM error: %v", err)
			}
			if plain != "Hello, SM4 GCM with Encoding!" {
				t.Errorf("DecryptGCM mismatch, want %q, got %q", "Hello, SM4 GCM with Encoding!", plain)
			}
		})
	}
}
