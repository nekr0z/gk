package crypt

import (
	"crypto/rand"
	"crypto/sha256"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncrypt_Success(t *testing.T) {
	data := mockUnencryptedData("test")
	passPhrase := "password"

	encrypted, err := Encrypt(data, passPhrase)
	assert.NoError(t, err)

	assert.NotNil(t, encrypted.Data)
	assert.NotZero(t, len(encrypted.Data))

	assert.NotNil(t, encrypted.Hash)

	expectedHash := sha256.Sum256(encrypted.Data)
	assert.Equal(t, expectedHash, encrypted.Hash)
}

func TestEncrypt_SaltError(t *testing.T) {
	origRandReader := rand.Reader
	defer func() { rand.Reader = origRandReader }()

	rand.Reader = &faultyReader{failOnRead: 1, realReader: origRandReader}
	data := mockUnencryptedData("test")

	_, err := Encrypt(data, "password")
	assert.Error(t, err)
}

func TestEncrypt_NonceError(t *testing.T) {
	origRandReader := rand.Reader
	defer func() { rand.Reader = origRandReader }()

	rand.Reader = &faultyReader{failOnRead: 2, realReader: origRandReader}
	data := mockUnencryptedData("test")

	_, err := Encrypt(data, "password")
	assert.Error(t, err)
}

func TestDecrypt_Success(t *testing.T) {
	data := mockUnencryptedData("test")
	passPhrase := "password"
	encrypted, err := Encrypt(data, passPhrase)
	require.NoError(t, err)

	decrypted, err := Decrypt(encrypted, passPhrase)
	assert.NoError(t, err)

	assert.Equal(t, data.Marshal(), decrypted)
}

func TestDecrypt_WrongPassphrase(t *testing.T) {
	data := mockUnencryptedData("test")
	encrypted, err := Encrypt(data, "password")
	require.NoError(t, err)

	_, err = Decrypt(encrypted, "wrongpassword")
	assert.Error(t, err)
}

func TestDecrypt_TamperedData(t *testing.T) {
	data := mockUnencryptedData("test")
	passPhrase := "password"
	encrypted, err := Encrypt(data, passPhrase)
	require.NoError(t, err)

	encrypted.Data[0] ^= 0xFF

	_, err = Decrypt(encrypted, passPhrase)
	assert.Error(t, err)
}

func TestDecrypt_TamperedHash(t *testing.T) {
	data := mockUnencryptedData("test")
	passPhrase := "password"
	encrypted, err := Encrypt(data, passPhrase)
	require.NoError(t, err)

	encrypted.Hash[0] ^= 0xFF

	_, err = Decrypt(encrypted, passPhrase)
	assert.Error(t, err)
}

type mockUnencryptedData string

func (m mockUnencryptedData) Marshal() []byte {
	return []byte(m)
}

// faultyReader is an io.Reader that fails after specified reads
type faultyReader struct {
	readCount  int
	failOnRead int
	realReader io.Reader
}

func (f *faultyReader) Read(p []byte) (n int, err error) {
	f.readCount++
	if f.readCount == f.failOnRead {
		return 0, assert.AnError
	}
	return f.realReader.Read(p)
}
