package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitEncryption(t *testing.T) {
	// 测试初始化加密
	err := InitEncryption()
	assert.NoError(t, err)
	assert.NotEmpty(t, encryptionKey)
	assert.Equal(t, 32, len(encryptionKey)) // 确保密钥长度为32字节
}

func TestEncryptDecrypt(t *testing.T) {
	// 测试加密和解密功能
	testPlaintext := "test plaintext for encryption"

	// 加密
	ciphertext, err := Encrypt(testPlaintext)
	assert.NoError(t, err)
	assert.NotEmpty(t, ciphertext)

	// 解密
	decryptedText, err := Decrypt(ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, testPlaintext, decryptedText)
}

func TestEncryptDecryptMultiple(t *testing.T) {
	// 测试多次加密解密
	testCases := []string{
		"",
		"short",
		"this is a longer test string for encryption",
		"string with special characters: !@#$%^&*()",
	}

	for _, testCase := range testCases {
		// 加密
		ciphertext, err := Encrypt(testCase)
		assert.NoError(t, err)
		assert.NotEmpty(t, ciphertext)

		// 解密
		decryptedText, err := Decrypt(ciphertext)
		assert.NoError(t, err)
		assert.Equal(t, testCase, decryptedText)
	}
}
