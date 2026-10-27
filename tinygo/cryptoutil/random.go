package cryptoutil

import (
	"crypto/rand"
	"fmt"
	"io"
)

// RandomStringCharacters is the alphabet used by EncryptionHelper.randomString
// in the Kotlin library.
const RandomStringCharacters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()_+"

// RandomString returns a cryptographically random string of the requested
// length, drawn from RandomStringCharacters.
func RandomString(length int) (string, error) {
	if length < 0 {
		return "", fmt.Errorf("cryptoutil: negative length %d", length)
	}
	if length == 0 {
		return "", nil
	}
	// Rejection sampling avoids the modulo bias introduced by a naive
	// randomByte % len(alphabet) implementation.
	limit := 256 - (256 % len(RandomStringCharacters))
	out := make([]byte, length)
	written := 0
	var buf [64]byte
	for written < length {
		if _, err := io.ReadFull(rand.Reader, buf[:]); err != nil {
			return "", err
		}
		for _, b := range buf {
			if int(b) >= limit {
				continue
			}
			out[written] = RandomStringCharacters[int(b)%len(RandomStringCharacters)]
			written++
			if written == length {
				break
			}
		}
	}
	return string(out), nil
}
