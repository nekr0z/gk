// Package crypt handles data encryption.
package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltLen = 8
	keyLen  = 32
	iter    = 1024 * 1024
)

// Data is a piece of encrypted data.
type Data struct {
	Data []byte   // Encrypted data.
	Hash [32]byte // Hash of the data.
}

// UnencryptedData is a piece of unencrypted data.
type UnencryptedData interface {
	Payload() []byte
}

// Encrypt encrypts the data.
func Encrypt(in UnencryptedData, passPhrase string) (Data, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return Data{}, err
	}

	gcm, err := getGCM(passPhrase, salt)
	if err != nil {
		return Data{}, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return Data{}, err
	}

	salt = append(salt, nonce...)

	encryptedData := append(salt, gcm.Seal(nil, nonce, in.Payload(), nil)...)

	hash := sha256.Sum256(encryptedData)

	return Data{
		Data: encryptedData,
		Hash: hash,
	}, nil
}

// Decrypt decrypts the data to JSON.
func Decrypt(in Data, passPhrase string) ([]byte, error) {
	// Verify data integrity hash
	if actualHash := sha256.Sum256(in.Data); actualHash != in.Hash {
		return nil, errors.New("data corruption detected")
	}

	if len(in.Data) < saltLen {
		return nil, errors.New("invalid ciphertext")
	}

	salt, ciphertext := in.Data[:saltLen], in.Data[saltLen:]

	gcm, err := getGCM(passPhrase, salt)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("invalid ciphertext")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func getGCM(passPhrase string, salt []byte) (cipher.AEAD, error) {
	key := pbkdf2.Key([]byte(passPhrase), salt, iter, keyLen, sha256.New)

	c, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	return gcm, nil
}
