package cryptoutil

import (
	"bytes"
	"strings"
	"testing"
)

func TestAESRoundTrip(t *testing.T) {
	plaintext := []byte(strings.Repeat("binary-table-data-", 512))
	iv, err := GenerateIV()
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := Encrypt("correct horse battery staple", iv, plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext equals plaintext")
	}
	got, err := Decrypt("correct horse battery staple", iv, ciphertext)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatal("plaintext mismatch")
	}
}

func TestAESErrors(t *testing.T) {
	if _, err := Encrypt("k", make([]byte, 8), []byte("x")); err == nil {
		t.Fatal("short IV must fail")
	}
	if _, err := Decrypt("k", make([]byte, 16), []byte{1, 2, 3}); err == nil {
		t.Fatal("invalid ciphertext length must fail")
	}
	iv, _ := GenerateIV()
	ciphertext, err := Encrypt("k", iv, []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	ciphertext[len(ciphertext)-1] ^= 0xff
	if _, err := Decrypt("k", iv, ciphertext); err == nil {
		t.Fatal("tampered padding must fail")
	}
}

func TestStreams(t *testing.T) {
	plaintext := []byte("streaming through TinyGo")
	iv, _ := GenerateIV()
	var encrypted bytes.Buffer
	if err := EncryptStream("key", iv, bytes.NewReader(plaintext), &encrypted); err != nil {
		t.Fatal(err)
	}
	var decrypted bytes.Buffer
	if err := DecryptStream("key", iv, bytes.NewReader(encrypted.Bytes()), &decrypted); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decrypted.Bytes(), plaintext) {
		t.Fatal("stream mismatch")
	}
}

func TestPasswordToKey(t *testing.T) {
	key := PasswordToKey("password")
	if len(key) != KeySize {
		t.Fatalf("key length = %d", len(key))
	}
	again := PasswordToKey("password")
	if !bytes.Equal(key, again) {
		t.Fatal("key derivation is not deterministic")
	}
}

func TestRandomString(t *testing.T) {
	for _, length := range []int{0, 1, 16, 256} {
		got, err := RandomString(length)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != length {
			t.Fatalf("length = %d, want %d", len(got), length)
		}
		for _, r := range got {
			if !strings.ContainsRune(RandomStringCharacters, r) {
				t.Fatalf("unexpected character %q", r)
			}
		}
	}
	if _, err := RandomString(-1); err == nil {
		t.Fatal("negative length must fail")
	}
}
