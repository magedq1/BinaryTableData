# AES and RSA

**English** | [العربية](security.ar.md)

Package: `cryptoutil`

This package is separate from `binarytable`, so programs that do not need crypto
do not pay for it.

---

## 1. AES compatible with Kotlin

The original implementation uses:

- a key derived from the password with SHA-256 (32 bytes);
- `AES/CBC/PKCS5Padding`;
- a 16-byte IV.

The Go port performs the same steps, so the output is **byte-for-byte identical**
for the same key, IV, and plaintext. This is verified by
`TestKotlinAESFixture`.

### Encrypt

```go
import "github.com/magedq1/BinaryTableData/tinygo/cryptoutil"

iv, err := cryptoutil.GenerateIV()
if err != nil {
	return err
}

ciphertext, err := cryptoutil.Encrypt("my password", iv, plaintext)
```

### Decrypt

```go
plaintext, err := cryptoutil.Decrypt("my password", iv, ciphertext)
```

### Streams

```go
err := cryptoutil.EncryptStream(password, iv, srcReader, dstWriter)
err = cryptoutil.DecryptStream(password, iv, srcReader, dstWriter)
```

`EncryptStream` and `DecryptStream` read the whole input into memory because CBC
needs complete blocks. For very large data on a hosted target, process it in
chunks with `Encrypt`/`Decrypt` or write a custom loop.

### IV storage

The IV is not secret, but it must be random and unique for each encryption
operation. Store it with the ciphertext (for example, the first 16 bytes of the
file) and never reuse it with the same key.

### Key derivation

```go
key := cryptoutil.PasswordToKey("my password") // 32 bytes
```

This is a plain SHA-256, not a slow KDF such as PBKDF2 or Argon2. For real
systems protecting human passwords, use a modern KDF. This function exists for
compatibility with Kotlin.

---

## 2. RSA

`cryptoutil` provides:

```go
privateKey, err := cryptoutil.GenerateRSAKeyPair(2048)
publicKey := &privateKey.PublicKey

ciphertext, err := cryptoutil.EncryptRSA(publicKey, data)
data, err = cryptoutil.DecryptRSA(privateKey, ciphertext)

signature, err := cryptoutil.SignSHA256(privateKey, data)
err = cryptoutil.VerifySHA256(publicKey, data, signature)
```

PEM encoding and parsing:

```go
publicPEM, err := cryptoutil.EncodeRSAPublicKeyPEM(publicKey)
privatePEM, err := cryptoutil.EncodeRSAPrivateKeyPEM(privateKey)

publicKey, err = cryptoutil.ParseRSAPublicKeyPEM(publicPEM)
privateKey, err = cryptoutil.ParseRSAPrivateKeyPEM(privatePEM)
```

- Encryption uses PKCS#1 v1.5, matching Java's `Cipher.getInstance("RSA")`.
- Signing uses `SHA256withRSA`.
- The plaintext must be at most `publicKey.Size()-11` bytes. For larger data,
  use hybrid encryption: a random AES key plus RSA for the key.
- Key generation and X.509 parsing are expensive on microcontrollers. Prefer a
  pre-shared public key and avoid generating keys on-device.

---

## 3. Random strings

```go
s, err := cryptoutil.RandomString(32)
```

- Uses `crypto/rand`.
- Uses the same alphabet as Kotlin's `EncryptionHelper.randomString`.
- Uses rejection sampling to avoid modulo bias.

---

## 4. TinyGo constraints

- `crypto/aes`, `crypto/cipher`, `crypto/sha256`, and `crypto/rand` are
  supported.
- `crypto/rsa` and `crypto/x509` are importable but heavy, and may depend on
  reflection features that are incomplete on some targets. Test them on your
  target before relying on them.
- On a small microcontroller, prefer AES or verify a signature generated
  off-device instead of running RSA on the device.
