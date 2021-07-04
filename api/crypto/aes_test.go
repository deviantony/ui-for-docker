package crypto

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_encryptAndDecrypt_withTheSamePassword(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("", "encrypt")
	defer os.RemoveAll(tmpdir)

	var (
		originFilePath    = filepath.Join(tmpdir, "origin")
		encryptedFilePath = filepath.Join(tmpdir, "encrypted")
		decryptedFilePath = filepath.Join(tmpdir, "decrypted")
	)

	content := []byte("content")
	ioutil.WriteFile(originFilePath, content, 0600)

	originFile, _ := os.Open(originFilePath)
	defer originFile.Close()

	encryptedFileWriter, _ := os.Create(encryptedFilePath)
	defer encryptedFileWriter.Close()

	err := AesEncrypt(originFile, encryptedFileWriter, []byte("passphrase"))
	assert.Nil(t, err, "Failed to encrypt a file")
	encryptedContent, err := ioutil.ReadFile(encryptedFilePath)
	assert.Nil(t, err, "Couldn't read encrypted file")
	assert.NotEqual(t, encryptedContent, content, "Content wasn't encrypted")

	encryptedFileReader, _ := os.Open(encryptedFilePath)
	defer encryptedFileReader.Close()

	decryptedFileWriter, _ := os.Create(decryptedFilePath)
	defer decryptedFileWriter.Close()

	decryptedReader, err := AesDecrypt(encryptedFileReader, []byte("passphrase"))
	assert.Nil(t, err, "Failed to decrypt file")

	io.Copy(decryptedFileWriter, decryptedReader)

	decryptedContent, _ := ioutil.ReadFile(decryptedFilePath)
	assert.Equal(t, content, decryptedContent, "Original and decrypted content should match")
}

func Test_encryptAndDecrypt_withEmptyPassword(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("", "encrypt")
	defer os.RemoveAll(tmpdir)

	var (
		originFilePath    = filepath.Join(tmpdir, "origin")
		encryptedFilePath = filepath.Join(tmpdir, "encrypted")
		decryptedFilePath = filepath.Join(tmpdir, "decrypted")
	)

	content := []byte("content")
	ioutil.WriteFile(originFilePath, content, 0600)

	originFile, _ := os.Open(originFilePath)
	defer originFile.Close()

	encryptedFileWriter, _ := os.Create(encryptedFilePath)
	defer encryptedFileWriter.Close()

	err := AesEncrypt(originFile, encryptedFileWriter, []byte(""))
	assert.Nil(t, err, "Failed to encrypt a file")
	encryptedContent, err := ioutil.ReadFile(encryptedFilePath)
	assert.Nil(t, err, "Couldn't read encrypted file")
	assert.NotEqual(t, encryptedContent, content, "Content wasn't encrypted")

	encryptedFileReader, _ := os.Open(encryptedFilePath)
	defer encryptedFileReader.Close()

	decryptedFileWriter, _ := os.Create(decryptedFilePath)
	defer decryptedFileWriter.Close()

	decryptedReader, err := AesDecrypt(encryptedFileReader, []byte(""))
	assert.Nil(t, err, "Failed to decrypt file")

	io.Copy(decryptedFileWriter, decryptedReader)

	decryptedContent, _ := ioutil.ReadFile(decryptedFilePath)
	assert.Equal(t, content, decryptedContent, "Original and decrypted content should match")
}

func Test_decryptWithDifferentPassphrase_shouldProduceWrongResult(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("", "encrypt")
	defer os.RemoveAll(tmpdir)

	var (
		originFilePath    = filepath.Join(tmpdir, "origin")
		encryptedFilePath = filepath.Join(tmpdir, "encrypted")
		decryptedFilePath = filepath.Join(tmpdir, "decrypted")
	)

	content := []byte("content")
	ioutil.WriteFile(originFilePath, content, 0600)

	originFile, _ := os.Open(originFilePath)
	defer originFile.Close()

	encryptedFileWriter, _ := os.Create(encryptedFilePath)
	defer encryptedFileWriter.Close()

	err := AesEncrypt(originFile, encryptedFileWriter, []byte("passphrase"))
	assert.Nil(t, err, "Failed to encrypt a file")
	encryptedContent, err := ioutil.ReadFile(encryptedFilePath)
	assert.Nil(t, err, "Couldn't read encrypted file")
	assert.NotEqual(t, encryptedContent, content, "Content wasn't encrypted")

	encryptedFileReader, _ := os.Open(encryptedFilePath)
	defer encryptedFileReader.Close()

	decryptedFileWriter, _ := os.Create(decryptedFilePath)
	defer decryptedFileWriter.Close()

	decryptedReader, err := AesDecrypt(encryptedFileReader, []byte("garbage"))
	assert.Nil(t, err, "Should allow to decrypt with wrong passphrase")

	io.Copy(decryptedFileWriter, decryptedReader)

	decryptedContent, _ := ioutil.ReadFile(decryptedFilePath)
	assert.NotEqual(t, content, decryptedContent, "Original and decrypted content should NOT match")
}
