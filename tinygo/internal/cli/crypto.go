package cli

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/magedq1/BinaryTableData/tinygo/cryptoutil"
)

func runAES(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(stdout, `Usage: binarytable aes <command> [flags]

Commands:
  iv       Generate a random 16-byte IV and print it as hexadecimal
  encrypt  Encrypt with AES-256-CBC/PKCS#5
  decrypt  Decrypt AES-256-CBC/PKCS#5 data

The password is hashed with SHA-256, matching the Kotlin AESEncryptor.
Prefer -password-file over -password to avoid exposing secrets in process
arguments. -in and -out accept - for stdin/stdout.

Examples:
  binarytable aes iv
  binarytable aes encrypt -in plain.bin -out secret.bin -password-file key.txt -iv 000102030405060708090a0b0c0d0e0f
  binarytable aes decrypt -in secret.bin -out plain.bin -password-file key.txt -iv 000102030405060708090a0b0c0d0e0f
`)
		return 0
	}
	switch args[0] {
	case "iv":
		return aesIV(args[1:], stdout, stderr)
	case "encrypt":
		return aesCrypt(args[1:], stdin, stdout, stderr, true)
	case "decrypt":
		return aesCrypt(args[1:], stdin, stdout, stderr, false)
	default:
		return usageError(stderr, "unknown aes command %q", args[0])
	}
}

func aesIV(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("aes iv", stderr)
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	iv, err := cryptoutil.GenerateIV()
	if err != nil {
		return commandError(stderr, err)
	}
	fmt.Fprintln(stdout, hex.EncodeToString(iv))
	return 0
}

func aesCrypt(args []string, stdin io.Reader, stdout, stderr io.Writer, encrypt bool) int {
	name := "aes decrypt"
	if encrypt {
		name = "aes encrypt"
	}
	fs := newFlagSet(name, stderr)
	in := fs.String("in", "", "input file, or - for stdin")
	out := fs.String("out", "", "output file, or - for stdout")
	password := fs.String("password", "", "password (visible in process arguments)")
	passwordFile := fs.String("password-file", "", "file containing the password")
	ivHex := fs.String("iv", "", "16-byte IV as hexadecimal")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *in == "" || *out == "" || *ivHex == "" {
		return usageError(stderr, "%s needs -in, -out, and -iv", name)
	}
	secret, err := resolvePassword(fs, *password, *passwordFile)
	if err != nil {
		return usageError(stderr, "%s: %v", name, err)
	}
	iv, err := hex.DecodeString(strings.TrimPrefix(strings.TrimSpace(*ivHex), "0x"))
	if err != nil || len(iv) != cryptoutil.IVSize {
		return usageError(stderr, "%s: -iv must be %d hexadecimal bytes", name, cryptoutil.IVSize)
	}
	data, err := readInput(*in, stdin)
	if err != nil {
		return commandError(stderr, err)
	}
	var result []byte
	if encrypt {
		result, err = cryptoutil.Encrypt(secret, iv, data)
	} else {
		result, err = cryptoutil.Decrypt(secret, iv, data)
	}
	if err != nil {
		return commandError(stderr, err)
	}
	if err := writeOutput(*out, stdout, result); err != nil {
		return commandError(stderr, err)
	}
	status := stdout
	if *out == "-" {
		status = stderr
	}
	fmt.Fprintf(status, "%s: %d bytes -> %d bytes\n", name, len(data), len(result))
	return 0
}

func resolvePassword(fs *flag.FlagSet, password, passwordFile string) (string, error) {
	visited := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) { visited[f.Name] = true })
	if visited["password"] && visited["password-file"] {
		return "", fmt.Errorf("use only one of -password and -password-file")
	}
	if visited["password-file"] {
		data, err := os.ReadFile(passwordFile)
		if err != nil {
			return "", err
		}
		return strings.TrimRight(string(data), "\r\n"), nil
	}
	if visited["password"] {
		return password, nil
	}
	return "", fmt.Errorf("specify -password or -password-file")
}

func runRSA(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(stdout, `Usage: binarytable rsa <command> [flags]

Commands:
  generate  Create a PKCS#8 private key and a PKIX public key
  sign      Sign a file with SHA256withRSA
  verify    Verify a SHA256withRSA signature
  encrypt   Encrypt with RSA PKCS#1 v1.5
  decrypt   Decrypt RSA PKCS#1 v1.5 data

Examples:
  binarytable rsa generate -bits 2048 -private private.pem -public public.pem
  binarytable rsa sign -key private.pem -in message.bin -out message.sig
  binarytable rsa verify -key public.pem -in message.bin -sig message.sig

RSA is intended for hosted targets. It is usually too heavy for small
microcontrollers.
`)
		return 0
	}
	switch args[0] {
	case "generate":
		return rsaGenerate(args[1:], stdout, stderr)
	case "sign":
		return rsaSign(args[1:], stdin, stdout, stderr)
	case "verify":
		return rsaVerify(args[1:], stdin, stdout, stderr)
	case "encrypt":
		return rsaEncrypt(args[1:], stdin, stdout, stderr)
	case "decrypt":
		return rsaDecrypt(args[1:], stdin, stdout, stderr)
	default:
		return usageError(stderr, "unknown rsa command %q", args[0])
	}
}

func rsaGenerate(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("rsa generate", stderr)
	bits := fs.Int("bits", 2048, "key size in bits")
	publicPath := fs.String("public", "", "public key output path")
	privatePath := fs.String("private", "", "private key output path")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *publicPath == "" || *privatePath == "" {
		return usageError(stderr, "rsa generate needs -public and -private")
	}
	privateKey, err := cryptoutil.GenerateRSAKeyPair(*bits)
	if err != nil {
		return commandError(stderr, err)
	}
	publicPEM, err := cryptoutil.EncodeRSAPublicKeyPEM(&privateKey.PublicKey)
	if err != nil {
		return commandError(stderr, err)
	}
	privatePEM, err := cryptoutil.EncodeRSAPrivateKeyPEM(privateKey)
	if err != nil {
		return commandError(stderr, err)
	}
	if err := os.WriteFile(*publicPath, publicPEM, 0o644); err != nil {
		return commandError(stderr, err)
	}
	if err := os.WriteFile(*privatePath, privatePEM, 0o600); err != nil {
		return commandError(stderr, err)
	}
	fmt.Fprintf(stdout, "wrote public key to %s and private key to %s (%d bits)\n", *publicPath, *privatePath, *bits)
	return 0
}

func rsaSign(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("rsa sign", stderr)
	keyPath := fs.String("key", "", "private key PEM file")
	in := fs.String("in", "", "input file, or - for stdin")
	out := fs.String("out", "", "signature output file, or - for stdout")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *keyPath == "" || *in == "" || *out == "" {
		return usageError(stderr, "rsa sign needs -key, -in, and -out")
	}
	keyData, err := os.ReadFile(*keyPath)
	if err != nil {
		return commandError(stderr, err)
	}
	privateKey, err := cryptoutil.ParseRSAPrivateKeyPEM(keyData)
	if err != nil {
		return commandError(stderr, err)
	}
	data, err := readInput(*in, stdin)
	if err != nil {
		return commandError(stderr, err)
	}
	signature, err := cryptoutil.SignSHA256(privateKey, data)
	if err != nil {
		return commandError(stderr, err)
	}
	if err := writeOutput(*out, stdout, signature); err != nil {
		return commandError(stderr, err)
	}
	status := stdout
	if *out == "-" {
		status = stderr
	}
	fmt.Fprintf(status, "signed %d bytes\n", len(data))
	return 0
}

func rsaVerify(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("rsa verify", stderr)
	keyPath := fs.String("key", "", "public key PEM file")
	in := fs.String("in", "", "input file, or - for stdin")
	sigPath := fs.String("sig", "", "signature file")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *keyPath == "" || *in == "" || *sigPath == "" {
		return usageError(stderr, "rsa verify needs -key, -in, and -sig")
	}
	keyData, err := os.ReadFile(*keyPath)
	if err != nil {
		return commandError(stderr, err)
	}
	publicKey, err := cryptoutil.ParseRSAPublicKeyPEM(keyData)
	if err != nil {
		return commandError(stderr, err)
	}
	data, err := readInput(*in, stdin)
	if err != nil {
		return commandError(stderr, err)
	}
	signature, err := os.ReadFile(*sigPath)
	if err != nil {
		return commandError(stderr, err)
	}
	if err := cryptoutil.VerifySHA256(publicKey, data, signature); err != nil {
		fmt.Fprintln(stderr, "signature invalid")
		return 1
	}
	fmt.Fprintln(stdout, "signature valid")
	return 0
}

func rsaEncrypt(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("rsa encrypt", stderr)
	keyPath := fs.String("key", "", "public key PEM file")
	in := fs.String("in", "", "input file, or - for stdin")
	out := fs.String("out", "", "output file, or - for stdout")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *keyPath == "" || *in == "" || *out == "" {
		return usageError(stderr, "rsa encrypt needs -key, -in, and -out")
	}
	keyData, err := os.ReadFile(*keyPath)
	if err != nil {
		return commandError(stderr, err)
	}
	publicKey, err := cryptoutil.ParseRSAPublicKeyPEM(keyData)
	if err != nil {
		return commandError(stderr, err)
	}
	data, err := readInput(*in, stdin)
	if err != nil {
		return commandError(stderr, err)
	}
	encrypted, err := cryptoutil.EncryptRSA(publicKey, data)
	if err != nil {
		return commandError(stderr, err)
	}
	if err := writeOutput(*out, stdout, encrypted); err != nil {
		return commandError(stderr, err)
	}
	status := stdout
	if *out == "-" {
		status = stderr
	}
	fmt.Fprintf(status, "encrypted %d bytes into %d bytes\n", len(data), len(encrypted))
	return 0
}

func rsaDecrypt(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("rsa decrypt", stderr)
	keyPath := fs.String("key", "", "private key PEM file")
	in := fs.String("in", "", "input file, or - for stdin")
	out := fs.String("out", "", "output file, or - for stdout")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *keyPath == "" || *in == "" || *out == "" {
		return usageError(stderr, "rsa decrypt needs -key, -in, and -out")
	}
	keyData, err := os.ReadFile(*keyPath)
	if err != nil {
		return commandError(stderr, err)
	}
	privateKey, err := cryptoutil.ParseRSAPrivateKeyPEM(keyData)
	if err != nil {
		return commandError(stderr, err)
	}
	data, err := readInput(*in, stdin)
	if err != nil {
		return commandError(stderr, err)
	}
	decrypted, err := cryptoutil.DecryptRSA(privateKey, data)
	if err != nil {
		return commandError(stderr, err)
	}
	if err := writeOutput(*out, stdout, decrypted); err != nil {
		return commandError(stderr, err)
	}
	status := stdout
	if *out == "-" {
		status = stderr
	}
	fmt.Fprintf(status, "decrypted %d bytes into %d bytes\n", len(data), len(decrypted))
	return 0
}
