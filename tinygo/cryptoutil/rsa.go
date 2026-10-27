package cryptoutil

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// GenerateRSAKeyPair generates an RSA key pair with the requested bit size.
//
// RSA support is optional and intended for hosted targets (Linux, Windows,
// WASI). On very small microcontrollers, prefer AES or a pre-shared public key
// and avoid generating keys on-device.
func GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, error) {
	if bits < 1024 {
		return nil, fmt.Errorf("cryptoutil: RSA key size must be at least 1024 bits")
	}
	return rsa.GenerateKey(rand.Reader, bits)
}

// EncodeRSAPublicKeyPEM encodes a public key as a PKIX "PUBLIC KEY" PEM block,
// the format accepted by RSAEncryptHelper.readPublicKey.
func EncodeRSAPublicKeyPEM(publicKey *rsa.PublicKey) ([]byte, error) {
	der, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}), nil
}

// EncodeRSAPrivateKeyPEM encodes a private key as an unencrypted PKCS#8
// "PRIVATE KEY" PEM block, the format accepted by
// RSAEncryptHelper.readPrivateKey.
func EncodeRSAPrivateKeyPEM(privateKey *rsa.PrivateKey) ([]byte, error) {
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), nil
}

// ParseRSAPublicKeyPEM parses a PKIX "PUBLIC KEY" PEM block.
func ParseRSAPublicKeyPEM(data []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("cryptoutil: no PEM block found")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	publicKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("cryptoutil: PEM block does not contain an RSA public key")
	}
	return publicKey, nil
}

// ParseRSAPrivateKeyPEM parses a PKCS#8 "PRIVATE KEY" PEM block. PKCS#1
// "RSA PRIVATE KEY" blocks are accepted as well.
func ParseRSAPrivateKeyPEM(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("cryptoutil: no PEM block found")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		privateKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("cryptoutil: PEM block does not contain an RSA private key")
		}
		return privateKey, nil
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// EncryptRSA encrypts data with RSA PKCS#1 v1.5 padding, matching Java's
// default "RSA" cipher transformation used by RSAEncryptHelper.processData.
//
// data must be no longer than publicKey.Size()-11 bytes. Use hybrid encryption
// for larger payloads.
func EncryptRSA(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
}

// DecryptRSA decrypts RSA PKCS#1 v1.5 data.
func DecryptRSA(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)
}

// SignSHA256 signs data with SHA-256 and RSASSA-PKCS1-v1_5.
func SignSHA256(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	digest := sha256.Sum256(data)
	return rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
}

// VerifySHA256 verifies an RSASSA-PKCS1-v1_5 SHA-256 signature.
func VerifySHA256(publicKey *rsa.PublicKey, data, signature []byte) error {
	digest := sha256.Sum256(data)
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest[:], signature)
}
