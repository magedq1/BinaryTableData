package cryptoutil

import (
	"bytes"
	"testing"
)

func TestRSA(t *testing.T) {
	privateKey, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatal(err)
	}
	publicKey := &privateKey.PublicKey

	encrypted, err := EncryptRSA(publicKey, []byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := DecryptRSA(privateKey, encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decrypted, []byte("secret")) {
		t.Fatal("RSA round trip mismatch")
	}

	signature, err := SignSHA256(privateKey, []byte("payload"))
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifySHA256(publicKey, []byte("payload"), signature); err != nil {
		t.Fatal(err)
	}
	if err := VerifySHA256(publicKey, []byte("other"), signature); err == nil {
		t.Fatal("signature must not verify")
	}

	publicPEM, err := EncodeRSAPublicKeyPEM(publicKey)
	if err != nil {
		t.Fatal(err)
	}
	parsedPublic, err := ParseRSAPublicKeyPEM(publicPEM)
	if err != nil {
		t.Fatal(err)
	}
	if parsedPublic.N.Cmp(publicKey.N) != 0 {
		t.Fatal("public key mismatch")
	}

	privatePEM, err := EncodeRSAPrivateKeyPEM(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	parsedPrivate, err := ParseRSAPrivateKeyPEM(privatePEM)
	if err != nil {
		t.Fatal(err)
	}
	if parsedPrivate.N.Cmp(privateKey.N) != 0 {
		t.Fatal("private key mismatch")
	}
}
