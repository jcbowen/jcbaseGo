package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/jcbowen/jcbaseGo/helper"
)

type AES struct {
	Text string `json:"text" default:""`
	Key  string `json:"key" default:"jcbase.aes_key__"`
	Iv   string `json:"iv" default:"jcbase.aes_iv___"`
}

// Encrypt encrypts the given text using the provided key and iv
func (a AES) Encrypt(cipherText *string) error {
	_ = helper.CheckAndSetDefault(&a)

	if err := validateKey(a.Key); err != nil {
		return err
	}
	if err := validateIv(a.Iv); err != nil {
		return err
	}

	block, err := aes.NewCipher([]byte(a.Key))
	if err != nil {
		return err
	}

	plainText := pkcs7Pad([]byte(a.Text), block.BlockSize())
	cipherByteArr := make([]byte, len(plainText))
	ivBytes := []byte(a.Iv)

	mode := cipher.NewCBCEncrypter(block, ivBytes)
	mode.CryptBlocks(cipherByteArr, plainText)

	*cipherText = base64.StdEncoding.EncodeToString(cipherByteArr)
	return nil
}

// Decrypt decrypts the given cipherText using the provided key and iv
func (a AES) Decrypt(plaintext *string) error {
	_ = helper.CheckAndSetDefault(&a)

	if err := validateKey(a.Key); err != nil {
		return err
	}
	if err := validateIv(a.Iv); err != nil {
		return err
	}

	block, err := aes.NewCipher([]byte(a.Key))
	if err != nil {
		return err
	}

	cipherBytes, err := base64.StdEncoding.DecodeString(a.Text)
	if err != nil {
		return err
	}

	if len(cipherBytes) < aes.BlockSize {
		return errors.New("ciphertext too short")
	}

	ivBytes := []byte(a.Iv)
	mode := cipher.NewCBCDecrypter(block, ivBytes)
	mode.CryptBlocks(cipherBytes, cipherBytes)

	plainByteArr, err := pkcs7Unpad(cipherBytes, block.BlockSize())
	if err != nil {
		return err
	}

	*plaintext = string(plainByteArr)

	return nil
}

// validateKey checks if the provided key is valid
func validateKey(key string) error {
	keyLen := len(key)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return fmt.Errorf("key length must be 16, 24, or 32 bytes; got key len (%d)", keyLen)
	}
	return nil
}

// validateIv checks if the provided iv is valid
func validateIv(iv string) error {
	if len(iv) != aes.BlockSize {
		return errors.New("IV length must be 16 bytes")
	}
	return nil
}

// pkcs7Pad pads the plaintext to a multiple of the block size using PKCS7 padding
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7Unpad removes the PKCS7 padding from the plaintext
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("input data is empty")
	}
	if len(data)%blockSize != 0 {
		return nil, errors.New("input data is not a multiple of the block size")
	}

	padding := data[len(data)-1]
	padLen := int(padding)
	if padLen > blockSize || padLen == 0 {
		return nil, errors.New("invalid padding length")
	}

	for i := len(data) - padLen; i < len(data); i++ {
		if data[i] != padding {
			return nil, errors.New("invalid padding")
		}
	}

	return data[:len(data)-padLen], nil
}
