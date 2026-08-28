package security

import (
	"testing"

	"github.com/jcbowen/jcbaseGo/component/helper"
)

func newTestIDCipher(t *testing.T) *IDCipher {
	t.Helper()
	sm4 := &SM4{
		Mode:     "CBC",
		Encoding: "Std",
	}
	_ = helper.CheckAndSetDefault(sm4)
	return NewIDCipher(sm4)
}

func TestIDCipherRoundTrip(t *testing.T) {
	c := newTestIDCipher(t)

	ids := []uint64{0, 1, 12345, 1 << 32, 1<<63 - 1}
	for _, id := range ids {
		cipher, err := c.EncryptUint64(id)
		if err != nil {
			t.Fatalf("EncryptUint64(%d) returned error: %v", id, err)
		}
		got, err := c.DecryptUint64(cipher)
		if err != nil {
			t.Fatalf("DecryptUint64(EncryptUint64(%d)) returned error: %v", id, err)
		}
		if got != id {
			t.Errorf("DecryptUint64(EncryptUint64(%d)) = %d, want %d", id, got, id)
		}
	}
}

// TestIDCipherCiphertextCompat 验证密文为保留填充的 URL-safe base64，
// 且与历史（标准 base64）密文互相兼容。
func TestIDCipherCiphertextCompat(t *testing.T) {
	c := newTestIDCipher(t)

	cipher, err := c.EncryptUint64(12345)
	if err != nil {
		t.Fatalf("EncryptUint64 returned error: %v", err)
	}

	// URL-safe 替换后不应含标准 base64 的 + 和 /
	for _, ch := range []byte{'+', '/'} {
		if i := indexByte(cipher, ch); i >= 0 {
			t.Fatalf("cipher contains illegal char %q: %s", ch, cipher)
		}
	}
	// 保留 = 填充（长度应为 4 的倍数，结尾含 =）
	if len(cipher)%4 != 0 {
		t.Fatalf("cipher length %d not multiple of 4: %s", len(cipher), cipher)
	}

	// 模拟旧格式标准 base64 密文（含 + 或 / 时），应仍可解密
	// 通过把 URL-safe 字符还原为标准 base64 再解密
	std := make([]byte, len(cipher))
	for i := 0; i < len(cipher); i++ {
		switch cipher[i] {
		case '-':
			std[i] = '+'
		case '_':
			std[i] = '/'
		default:
			std[i] = cipher[i]
		}
	}
	got, err := c.DecryptUint64(string(std))
	if err != nil {
		t.Fatalf("DecryptUint64(standard base64) returned error: %v", err)
	}
	if got != 12345 {
		t.Errorf("DecryptUint64(standard base64) = %d, want 12345", got)
	}
}

func TestIDCipherNil(t *testing.T) {
	var c *IDCipher
	if _, err := c.EncryptUint64(1); err == nil {
		t.Error("nil IDCipher EncryptUint64 expected error, got nil")
	}
	if _, err := c.DecryptUint64("x"); err == nil {
		t.Error("nil IDCipher DecryptUint64 expected error, got nil")
	}
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
