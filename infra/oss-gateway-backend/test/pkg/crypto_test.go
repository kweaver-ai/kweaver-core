package pkg_test

import (
	"oss-gateway/pkg/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAESCrypto_ValidKey(t *testing.T) {
	key := "" // 32 bytes
	aes, err := crypto.NewAESCrypto(key)

	assert.NoError(t, err)
	assert.NotNil(t, aes)
}

func TestNewAESCrypto_InvalidKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{
			name: "Key too short",
			key:  "short",
		},
		{
			name: "Key too long",
			key:  "123456789012345678901234567890123",
		},
		{
			name: "Empty key",
			key:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aes, err := crypto.NewAESCrypto(tt.key)
			assert.Error(t, err)
			assert.Nil(t, aes)
		})
	}
}

func TestAESCrypto_Encrypt_Success(t *testing.T) {
	key := ""
	aes, _ := crypto.NewAESCrypto(key)

	plaintext := "Hello, World!"
	ciphertext, err := aes.Encrypt(plaintext)

	assert.NoError(t, err)
	assert.NotEmpty(t, ciphertext)
	assert.NotEqual(t, plaintext, ciphertext)
}

func TestAESCrypto_Decrypt_Success(t *testing.T) {
	key := ""
	aes, _ := crypto.NewAESCrypto(key)

	plaintext := "Hello, World!"
	ciphertext, _ := aes.Encrypt(plaintext)

	decrypted, err := aes.Decrypt(ciphertext)

	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestAESCrypto_EncryptDecrypt_MultipleValues(t *testing.T) {
	key := ""
	aes, _ := crypto.NewAESCrypto(key)

	testCases := []string{
		"",
		"a",
		"short text",
		"This is a longer text with special characters: !@#$%^&*()",
		"中文测试",
		"12345678901234567890123456789012345678901234567890",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			encrypted, err := aes.Encrypt(tc)
			assert.NoError(t, err)

			decrypted, err := aes.Decrypt(encrypted)
			assert.NoError(t, err)
			assert.Equal(t, tc, decrypted)
		})
	}
}

func TestAESCrypto_Decrypt_InvalidCiphertext(t *testing.T) {
	key := ""
	aes, _ := crypto.NewAESCrypto(key)

	tests := []struct {
		name       string
		ciphertext string
	}{
		{
			name:       "Invalid base64",
			ciphertext: "not-valid-base64!@#",
		},
		{
			name:       "Too short",
			ciphertext: "YWJj", // "abc" in base64, too short for AES
		},
		{
			name:       "Empty string",
			ciphertext: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := aes.Decrypt(tt.ciphertext)
			assert.Error(t, err)
		})
	}
}

func TestAESCrypto_EncryptProducesDifferentOutputs(t *testing.T) {
	key := ""
	aes, _ := crypto.NewAESCrypto(key)

	plaintext := "same plaintext"

	encrypted1, _ := aes.Encrypt(plaintext)
	encrypted2, _ := aes.Encrypt(plaintext)

	// 由于使用随机IV，两次加密结果应该不同
	assert.NotEqual(t, encrypted1, encrypted2)

	// 但都应该能正确解密
	decrypted1, _ := aes.Decrypt(encrypted1)
	decrypted2, _ := aes.Decrypt(encrypted2)

	assert.Equal(t, plaintext, decrypted1)
	assert.Equal(t, plaintext, decrypted2)
}

func TestAESCrypto_DifferentKeysProduceDifferentOutputs(t *testing.T) {
	key1 := ""
	key2 := "abcdefghijklmnopqrstuvwxyz123456"

	aes1, _ := crypto.NewAESCrypto(key1)
	aes2, _ := crypto.NewAESCrypto(key2)

	plaintext := "test message"

	encrypted1, _ := aes1.Encrypt(plaintext)
	encrypted2, _ := aes2.Encrypt(plaintext)

	// 不同密钥加密结果应该不同
	assert.NotEqual(t, encrypted1, encrypted2)

	// 用错误的密钥解密应该失败或得到错误结果
	_, err := aes1.Decrypt(encrypted2)
	// 可能会失败，也可能得到错误的结果
	if err == nil {
		decrypted, _ := aes1.Decrypt(encrypted2)
		assert.NotEqual(t, plaintext, decrypted)
	}
}

func TestAESCrypto_LargeData(t *testing.T) {
	key := ""
	aes, _ := crypto.NewAESCrypto(key)

	// 生成一个大字符串
	largeText := ""
	for i := 0; i < 10000; i++ {
		largeText += "a"
	}

	encrypted, err := aes.Encrypt(largeText)
	assert.NoError(t, err)

	decrypted, err := aes.Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, largeText, decrypted)
}
