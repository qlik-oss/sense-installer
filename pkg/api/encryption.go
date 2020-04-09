package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
)

const (
	key_file_name = "user_secret_key"
)

// GenerateAndStoreSecretKey generates and stores key
func GenerateAndStoreSecretKey(secretsDir string) (string, error) {
	// creating contexts/qlik-default/secrets/user_secret_key
	keyFile := filepath.Join(secretsDir, key_file_name)
	key, err := GenerateKey()
	if err != nil {
		return "", err
	}
	if err := writeContentToFile([]byte(key), keyFile); err != nil {
		return "", err
	}
	return key, nil
}
func LoadSecretKey(secretsDir string) (string, error) {
	keyFile := filepath.Join(secretsDir, key_file_name)
	by, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return "", err
	}
	return string(by), nil
}

// writeContentToFile writes keys to a file
func writeContentToFile(keyData []byte, fileName string) error {
	err := ioutil.WriteFile(fileName, keyData, 0600)
	if err != nil {
		log.Printf("error writing to file (%s): %v", fileName, err)
		return err
	}
	return nil
}

func GenerateKey() (string, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	s := fmt.Sprintf("%x", salt)
	return s, nil
}

func EncryptData(plaintext []byte, userKey string) ([]byte, error) {
	key, _ := hex.DecodeString(userKey)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return aesgcm.Seal(nonce, nonce, plaintext, nil), nil
}

func DecryptData(ciphertext []byte, userKey string) ([]byte, error) {
	key, _ := hex.DecodeString(userKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
