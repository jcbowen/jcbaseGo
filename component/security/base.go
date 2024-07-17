package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/pbkdf2"
	"io"
)

// CipherConfig stores the cipher configuration
type CipherConfig struct {
	Cipher               string
	AllowedCiphers       map[string][]int
	KdfHash              string
	MacHash              string
	AuthKeyInfo          string
	DerivationIterations int
}

// DefaultCipherConfig is the default cipher configuration
var DefaultCipherConfig = CipherConfig{
	Cipher: "AES-128-CBC",
	AllowedCiphers: map[string][]int{
		"AES-128-CBC": {16, 16},
		"AES-192-CBC": {16, 24},
		"AES-256-CBC": {16, 32},
	},
	KdfHash:              "sha256",
	MacHash:              "sha256",
	AuthKeyInfo:          "AuthorizationKey",
	DerivationIterations: 100000,
}

// GenerateRandomBytes generates random bytes of specified length
func GenerateRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// Encrypt encrypts data using a password or a key
func Encrypt(data, secret string, passwordBased bool, config CipherConfig) (string, error) {
	cipherConfig, exists := config.AllowedCiphers[config.Cipher]
	if !exists {
		return "", errors.New("cipher not allowed")
	}
	keySize := cipherConfig[1]

	keySalt, err := GenerateRandomBytes(keySize)
	if err != nil {
		return "", err
	}

	var derivedKey []byte
	if passwordBased {
		derivedKey = pbkdf2.Key([]byte(secret), keySalt, config.DerivationIterations, keySize, sha256.New)
	} else {
		hkdfReader := hkdf.New(sha256.New, []byte(secret), keySalt, []byte(config.AuthKeyInfo))
		derivedKey = make([]byte, keySize)
		if _, err := io.ReadFull(hkdfReader, derivedKey); err != nil {
			return "", err
		}
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return "", err
	}

	aesGcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGcm.Seal(nil, nonce, []byte(data), nil)
	authKey := hmac.New(sha256.New, derivedKey)
	authKey.Write(nonce)
	authKey.Write(ciphertext)
	mac := authKey.Sum(nil)

	return base64.URLEncoding.EncodeToString(append(append(append(keySalt, mac...), nonce...), ciphertext...)), nil
}

// Decrypt decrypts data using a password or a key
func Decrypt(data, secret string, passwordBased bool, config CipherConfig) (string, error) {
	decodedData, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	cipherConfig, exists := config.AllowedCiphers[config.Cipher]
	if !exists {
		return "", errors.New("cipher not allowed")
	}
	keySize := cipherConfig[1]

	keySalt := decodedData[:keySize]
	mac := decodedData[keySize : keySize+32]
	nonce := decodedData[keySize+32 : keySize+32+12]
	ciphertext := decodedData[keySize+32+12:]

	var derivedKey []byte
	if passwordBased {
		derivedKey = pbkdf2.Key([]byte(secret), keySalt, config.DerivationIterations, keySize, sha256.New)
	} else {
		hkdfReader := hkdf.New(sha256.New, []byte(secret), keySalt, []byte(config.AuthKeyInfo))
		derivedKey = make([]byte, keySize)
		if _, err := io.ReadFull(hkdfReader, derivedKey); err != nil {
			return "", err
		}
	}

	authKey := hmac.New(sha256.New, derivedKey)
	authKey.Write(nonce)
	authKey.Write(ciphertext)
	expectedMac := authKey.Sum(nil)

	if !hmac.Equal(mac, expectedMac) {
		return "", errors.New("authentication failed")
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return "", err
	}

	aesGcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := aesGcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
