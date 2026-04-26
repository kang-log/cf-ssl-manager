package app

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"runtime"

	"golang.org/x/crypto/pbkdf2"
)

func deriveKey() ([]byte, error) {
	hostname, _ := os.Hostname()
	user := os.Getenv("USERNAME")
	if user == "" {
		user = os.Getenv("USER")
	}
	salt := fmt.Sprintf("%s-%s-%s", hostname, user, runtime.GOOS)
	// Use a fixed application secret combined with machine info
	secret := "cf-ssl-manager-2026-secret-key"
	return pbkdf2.Key([]byte(secret), []byte(salt), 10000, 32, sha256.New), nil
}

func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		fmt.Printf("[Encrypt] input empty, returning empty\n")
		return "", nil
	}
	fmt.Printf("[Encrypt] input len=%d\n", len(plaintext))
	key, err := deriveKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	result := base64.StdEncoding.EncodeToString(ciphertext)
	fmt.Printf("[Encrypt] output len=%d\n", len(result))
	return result, nil
}

func Decrypt(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	key, err := deriveKey()
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
