package model

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var encryptionKey []byte

func SetEncryptionKey(key string) {
	k := []byte(key)
	if len(k) < 32 {
		padded := make([]byte, 32)
		copy(padded, k)
		k = padded
	}
	encryptionKey = k[:32]
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.PasswordHash != "" && len(u.PasswordHash) < 60 {
		hashed, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.PasswordHash = string(hashed)
	}
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("PasswordHash") && len(u.PasswordHash) < 60 {
		hashed, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.PasswordHash = string(hashed)
	}
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func encrypt(plaintext string) (string, error) {
	if len(encryptionKey) == 0 {
		return plaintext, nil
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(ciphertext string) (string, error) {
	if len(encryptionKey) == 0 {
		return ciphertext, nil
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func (p *Provider) BeforeCreate(tx *gorm.DB) error {
	if p.APIKey != "" {
		encrypted, err := encrypt(p.APIKey)
		if err != nil {
			return err
		}
		p.APIKey = encrypted
	}
	return nil
}

func (p *Provider) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("APIKey") && p.APIKey != "" {
		encrypted, err := encrypt(p.APIKey)
		if err != nil {
			return err
		}
		p.APIKey = encrypted
	}
	return nil
}

func (p *Provider) GetDecryptedAPIKey() string {
	if p.APIKey == "" {
		return ""
	}
	decrypted, err := decrypt(p.APIKey)
	if err != nil {
		return p.APIKey
	}
	return decrypted
}
