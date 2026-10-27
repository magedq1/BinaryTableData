package cryptoutil

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// The AES fixture is produced by the original Kotlin AESEncryptor, using
// AES/CBC/PKCS5Padding and a SHA-256-derived key.
func TestKotlinAESFixture(t *testing.T) {
	dir := filepath.Join("..", "testdata")
	ciphertext, err := os.ReadFile(filepath.Join(dir, "aes.bin"))
	if err != nil {
		t.Fatal(err)
	}
	plaintext, err := os.ReadFile(filepath.Join(dir, "aes-plain.bin"))
	if err != nil {
		t.Fatal(err)
	}
	iv, err := os.ReadFile(filepath.Join(dir, "aes-iv.bin"))
	if err != nil {
		t.Fatal(err)
	}

	got, err := Decrypt("secret-key", iv, ciphertext)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("decrypted = %q, want %q", got, plaintext)
	}

	// CBC encryption is deterministic for a fixed key/IV/padding, so the Go
	// output must be byte-for-byte identical to Java's output.
	encrypted, err := Encrypt("secret-key", iv, plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encrypted, ciphertext) {
		t.Fatal("Go AES output differs from the Kotlin fixture")
	}
}
