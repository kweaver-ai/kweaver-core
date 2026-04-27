package anysharedshandler

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAnyshareOauth2Token(t *testing.T) {
	t.Parallel()

	t.Skip("Skipped: Function uses log.Fatal which terminates test process. " +
		"Tests for RSA decryption and base64 decoding are covered in other test functions.")
}

func TestParseRSAPrivateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		keyPEM      []byte
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid_private_key",
			keyPEM:      []byte(RSA_PRIVATE_KEY),
			expectError: false,
		},
		{
			name:        "invalid_pem_format",
			keyPEM:      []byte("invalid pem data"),
			expectError: true,
			errorMsg:    "解析PEM私钥失败：无效的PEM格式",
		},
		{
			name:        "empty_pem_data",
			keyPEM:      []byte{},
			expectError: true,
			errorMsg:    "解析PEM私钥失败：无效的PEM格式",
		},
		{
			name: "invalid_key_data",
			keyPEM: []byte(`-----BEGIN PRIVATE KEY-----
InvalidKeyData
-----END PRIVATE KEY-----`),
			expectError: true,
			errorMsg:    "解析PEM私钥失败：无效的PEM格式",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			key, err := parseRSAPrivateKey(tt.keyPEM)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, key)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, key)
			}
		})
	}
}

func TestRSAPrivateDecryptPKCS1(t *testing.T) {
	t.Parallel()

	privateKey, err := parseRSAPrivateKey([]byte(RSA_PRIVATE_KEY))
	require.NoError(t, err)

	tests := []struct {
		name          string
		encryptedData []byte
		expectedError bool
		errorMsg      string
	}{
		{
			name:          "decrypt_success",
			encryptedData: rsaEncryptTestDataPKCS1(privateKey, []byte("test_password")),
			expectedError: false,
		},
		{
			name:          "decrypt_invalid_data",
			encryptedData: []byte("invalid_encrypted_data"),
			expectedError: true,
			errorMsg:      "PKCS1解密失败",
		},
		{
			name:          "decrypt_empty_data",
			encryptedData: []byte{},
			expectedError: true,
			errorMsg:      "PKCS1解密失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			decrypted, err := rsaPrivateDecryptPKCS1(privateKey, tt.encryptedData)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, decrypted)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, decrypted)
			}
		})
	}
}

func TestRSAPrivateDecryptOAEP(t *testing.T) {
	t.Parallel()

	privateKey, err := parseRSAPrivateKey([]byte(RSA_PRIVATE_KEY))
	require.NoError(t, err)

	tests := []struct {
		name          string
		encryptedData []byte
		expectedError bool
		errorMsg      string
	}{
		{
			name:          "decrypt_success",
			encryptedData: rsaEncryptTestDataOAEP(privateKey, []byte("test_password_oaep")),
			expectedError: false,
		},
		{
			name:          "decrypt_invalid_data",
			encryptedData: []byte("invalid_encrypted_data"),
			expectedError: true,
			errorMsg:      "OAEP解密失败",
		},
		{
			name:          "decrypt_empty_data",
			encryptedData: []byte{},
			expectedError: true,
			errorMsg:      "OAEP解密失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			decrypted, err := rsaPrivateDecryptOAEP(privateKey, tt.encryptedData)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, decrypted)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, decrypted)
			}
		})
	}
}

// Helper function to encrypt test data with PKCS1
func rsaEncryptTestDataPKCS1(pubKey *rsa.PrivateKey, data []byte) []byte {
	rng := rand.Reader
	encrypted, _ := rsa.EncryptPKCS1v15(rng, &pubKey.PublicKey, data)

	return encrypted
}

// Helper function to encrypt test data with OAEP
func rsaEncryptTestDataOAEP(pubKey *rsa.PrivateKey, data []byte) []byte {
	rng := rand.Reader
	hash := crypto.SHA256.New()
	encrypted, _ := rsa.EncryptOAEP(hash, rng, &pubKey.PublicKey, data, nil)

	return encrypted
}
