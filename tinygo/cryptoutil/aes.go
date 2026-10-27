// Package cryptoutil contains the Go port of the Kotlin util package.
//
// AES helpers are compatible with the Kotlin AESEncryptor:
//   - the password is hashed with SHA-256 to obtain a 256-bit key;
//   - data is encrypted with AES-256 in CBC mode;
//   - padding is PKCS#5/PKCS#7, as used by Java's "AES/CBC/PKCS5Padding".
//
// RSA helpers are optional and intended for hosted targets. They are not
// recommended for small microcontrollers because key generation and X.509
// parsing pull in a large part of the standard library.
package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
)

const (
	// IVSize is the AES block size and IV size in bytes.
	IVSize = aes.BlockSize
	// KeySize is the AES-256 key size in bytes.
	KeySize = 32
)

// PasswordToKey derives a 32-byte AES key from a password using SHA-256. This
// matches AESEncryptor.convertPasswordTo32Bytes in the Kotlin library.
func PasswordToKey(password string) []byte {
	sum := sha256.Sum256([]byte(password))
	return sum[:]
}

// GenerateIV returns a cryptographically random 16-byte IV.
func GenerateIV() ([]byte, error) {
	iv := make([]byte, IVSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	return iv, nil
}

// Encrypt encrypts plaintext with AES-256-CBC and PKCS#7 padding.
func Encrypt(password string, iv, plaintext []byte) ([]byte, error) {
	if len(iv) != IVSize {
		return nil, fmt.Errorf("cryptoutil: IV must be %d bytes, got %d", IVSize, len(iv))
	}
	block, err := aes.NewCipher(PasswordToKey(password))
	if err != nil {
		return nil, err
	}
	padded := pkcs7Pad(plaintext, aes.BlockSize)
	out := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, padded)
	return out, nil
}

// Decrypt decrypts AES-256-CBC data and removes PKCS#7 padding.
func Decrypt(password string, iv, ciphertext []byte) ([]byte, error) {
	if len(iv) != IVSize {
		return nil, fmt.Errorf("cryptoutil: IV must be %d bytes, got %d", IVSize, len(iv))
	}
	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("cryptoutil: ciphertext length %d is not a multiple of %d", len(ciphertext), aes.BlockSize)
	}
	block, err := aes.NewCipher(PasswordToKey(password))
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(out, ciphertext)
	return pkcs7Unpad(out, aes.BlockSize)
}

// EncryptStream reads all of src, encrypts it, and writes the result to dst.
//
// The Kotlin API streams through CipherOutputStream. The Go port reads the
// plaintext into memory because CBC encryption needs complete blocks and TinyGo
// targets are usually memory constrained anyway; use Encrypt directly when the
// input is already in memory.
func EncryptStream(password string, iv []byte, src io.Reader, dst io.Writer) error {
	plaintext, err := io.ReadAll(src)
	if err != nil {
		return err
	}
	ciphertext, err := Encrypt(password, iv, plaintext)
	if err != nil {
		return err
	}
	return writeAll(dst, ciphertext)
}

// DecryptStream reads all encrypted bytes from src, decrypts them, and writes
// the plaintext to dst.
func DecryptStream(password string, iv []byte, src io.Reader, dst io.Writer) error {
	ciphertext, err := io.ReadAll(src)
	if err != nil {
		return err
	}
	plaintext, err := Decrypt(password, iv, ciphertext)
	if err != nil {
		return err
	}
	return writeAll(dst, plaintext)
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	out := make([]byte, len(data)+padding)
	copy(out, data)
	for i := len(data); i < len(out); i++ {
		out[i] = byte(padding)
	}
	return out
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, fmt.Errorf("cryptoutil: invalid padded data length")
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize || padding > len(data) {
		return nil, fmt.Errorf("cryptoutil: invalid PKCS#7 padding")
	}
	for _, b := range data[len(data)-padding:] {
		if int(b) != padding {
			return nil, fmt.Errorf("cryptoutil: invalid PKCS#7 padding")
		}
	}
	return data[:len(data)-padding], nil
}

func writeAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if n > 0 {
			data = data[n:]
		}
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
	}
	return nil
}
