package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runCLI(t *testing.T, stdin []byte, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run(args, bytes.NewReader(stdin), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func mustRun(t *testing.T, stdin []byte, args ...string) string {
	t.Helper()
	code, stdout, stderr := runCLI(t, stdin, args...)
	if code != 0 {
		t.Fatalf("%v exited %d\nstdout: %s\nstderr: %s", args, code, stdout, stderr)
	}
	return stdout
}

func TestHelpVersionAndErrors(t *testing.T) {
	if code, out, _ := runCLI(t, nil); code != 0 || !strings.Contains(out, "Usage:") {
		t.Fatalf("help = %d, %q", code, out)
	}
	if code, out, _ := runCLI(t, nil, "version"); code != 0 || !strings.Contains(out, Version) {
		t.Fatalf("version = %d, %q", code, out)
	}
	if code, _, errOut := runCLI(t, nil, "nope"); code != 2 || !strings.Contains(errOut, "unknown command") {
		t.Fatalf("unknown = %d, %q", code, errOut)
	}
	if code, _, _ := runCLI(t, nil, "table", "create", "-h"); code != 0 {
		t.Fatalf("command help = %d", code)
	}
}

func TestTableCLI(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "users.db")
	create := []string{
		"table", "create", "-file", file, "-rows", "10",
		"-col", "_id:int", "-col", "name:string:16",
		"-col", "score:double", "-col", "payload:raw:8",
	}
	out := mustRun(t, nil, create...)
	if !strings.Contains(out, "rows=10") {
		t.Fatalf("create output: %q", out)
	}

	mustRun(t, nil, "table", "set", "-file", file, "-row", "0", "-col", "_id", "-value", "7")
	mustRun(t, nil, "table", "set", "-file", file, "-row", "0", "-col", "name", "-value", "Alice")
	mustRun(t, nil, "table", "set", "-file", file, "-row", "0", "-col", "score", "-value", "12.5")
	mustRun(t, nil, "table", "set", "-file", file, "-row", "0", "-col", "payload", "-value-hex", "010203")

	if got := strings.TrimSpace(mustRun(t, nil, "table", "get", "-file", file, "-row", "0", "-col", "name")); got != "Alice" {
		t.Fatalf("get name = %q", got)
	}
	if got := strings.TrimSpace(mustRun(t, nil, "table", "get", "-file", file, "-row", "0", "-col", "score")); got != "12.5" {
		t.Fatalf("get score = %q", got)
	}
	if got := strings.TrimSpace(mustRun(t, nil, "table", "get", "-file", file, "-row", "0", "-col", "payload")); got != "0102030000000000" {
		t.Fatalf("get payload = %q", got)
	}
	if got := strings.TrimSpace(mustRun(t, nil, "table", "get", "-file", file, "-row", "0", "-col", "payload", "-hex")); got != "0102030000000000" {
		t.Fatalf("get hex = %q", got)
	}

	outFile := filepath.Join(dir, "payload.bin")
	mustRun(t, nil, "table", "get", "-file", file, "-row", "0", "-col", "payload", "-out", outFile)
	raw, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, []byte{1, 2, 3, 0, 0, 0, 0, 0}) {
		t.Fatalf("payload file = %v", raw)
	}

	dump := mustRun(t, nil, "table", "dump", "-file", file, "-row", "0")
	if !strings.Contains(dump, "name\tSTRING\t\"Alice\"") {
		t.Fatalf("dump = %q", dump)
	}
	if got := strings.TrimSpace(mustRun(t, nil, "table", "info", "-file", file)); !strings.Contains(got, "Rows: 10") {
		t.Fatalf("info = %q", got)
	}
	if got := mustRun(t, nil, "table", "schema", "-file", file); !strings.Contains(got, "_id/4/INT") {
		t.Fatalf("schema = %q", got)
	}
	if got := strings.TrimSpace(mustRun(t, nil, "table", "search", "-file", file, "-col", "name", "-value", "Alice")); got != "0" {
		t.Fatalf("search = %q", got)
	}
	if code, out, _ := runCLI(t, nil, "table", "search", "-file", file, "-col", "name", "-value", "Bob"); code != 1 || strings.TrimSpace(out) != "-1" {
		t.Fatalf("missing search = %d, %q", code, out)
	}

	if code, _, _ := runCLI(t, nil, create...); code != 1 {
		t.Fatalf("create without force = %d", code)
	}
	if code, _, _ := runCLI(t, nil, append(create, "-force")...); code != 0 {
		t.Fatalf("create with force = %d", code)
	}
	if code, _, _ := runCLI(t, nil, "table", "set", "-file", file, "-row", "0", "-col", "name", "-value", strings.Repeat("x", 20)); code != 1 {
		t.Fatalf("oversize set = %d", code)
	}
}

func TestListCLI(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "blobs.bin")
	if got := mustRun(t, nil, "list", "create", "-file", file, "-max", "4"); !strings.Contains(got, "max_entries=4") {
		t.Fatalf("create = %q", got)
	}
	if code, _, _ := runCLI(t, nil, "list", "create", "-file", file, "-max", "4"); code != 1 {
		t.Fatalf("create duplicate = %d", code)
	}
	if got := mustRun(t, nil, "list", "info", "-file", file); !strings.Contains(got, "Max entries: 4") {
		t.Fatalf("created info = %q", got)
	}
	if got := strings.TrimSpace(mustRun(t, []byte("maged"), "list", "add", "-file", file, "-in", "-")); got != "index=0 size=5" {
		t.Fatalf("add stdin = %q", got)
	}
	input := filepath.Join(dir, "blob.bin")
	if err := os.WriteFile(input, []byte{0, 1, 0, 255, 0}, 0o666); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(mustRun(t, nil, "list", "add", "-file", file, "-in", input)); got != "index=1 size=5" {
		t.Fatalf("add file = %q", got)
	}
	if got := strings.TrimSpace(mustRun(t, nil, "list", "get", "-file", file, "-index", "0")); got != "maged" {
		t.Fatalf("get = %q", got)
	}
	if got := strings.TrimSpace(mustRun(t, nil, "list", "get", "-file", file, "-index", "1", "-hex")); got != "000100ff00" {
		t.Fatalf("get hex = %q", got)
	}
	outFile := filepath.Join(dir, "out.bin")
	mustRun(t, nil, "list", "get", "-file", file, "-index", "1", "-out", outFile)
	out, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, []byte{0, 1, 0, 255, 0}) {
		t.Fatalf("out = %v", out)
	}
	if got := mustRun(t, nil, "list", "info", "-file", file); !strings.Contains(got, "Entries: 2") || !strings.Contains(got, "Max entries: 4") {
		t.Fatalf("info = %q", got)
	}
	if got := mustRun(t, nil, "list", "dump", "-file", file); !strings.Contains(got, "1\t5") {
		t.Fatalf("dump = %q", got)
	}
	if code, _, _ := runCLI(t, nil, "list", "get", "-file", file, "-index", "2"); code != 1 {
		t.Fatalf("out-of-range get = %d", code)
	}
}

func TestCompressionCLI(t *testing.T) {
	dir := t.TempDir()
	data := bytes.Repeat([]byte("compress me "), 500)
	input := filepath.Join(dir, "raw.bin")
	if err := os.WriteFile(input, data, 0o666); err != nil {
		t.Fatal(err)
	}
	zlibFile := filepath.Join(dir, "data.zlib")
	mustRun(t, nil, "zlib", "compress", "-in", input, "-out", zlibFile)
	zlibOut := filepath.Join(dir, "data.out")
	mustRun(t, nil, "zlib", "decompress", "-in", zlibFile, "-out", zlibOut)
	got, err := os.ReadFile(zlibOut)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("zlib round trip mismatch")
	}

	deflateFile := filepath.Join(dir, "data.deflate")
	mustRun(t, nil, "deflate", "pack", "-in", input, "-out", deflateFile, "-block", "256")
	if info := mustRun(t, nil, "deflate", "info", "-in", deflateFile); !strings.Contains(info, "Block size: 256") || !strings.Contains(info, "Blocks:") {
		t.Fatalf("deflate info = %q", info)
	}
	deflateOut := filepath.Join(dir, "data.raw")
	mustRun(t, nil, "deflate", "unpack", "-in", deflateFile, "-out", deflateOut, "-cache", "2")
	got, err = os.ReadFile(deflateOut)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("deflate round trip mismatch")
	}
	readOut := filepath.Join(dir, "data.read")
	mustRun(t, nil, "deflate", "read", "-in", deflateFile, "-out", readOut)
	got, err = os.ReadFile(readOut)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("deflate read alias mismatch")
	}

	// stdin -> stdout; status must go to stderr so stdout stays binary.
	code, stdout, stderr := runCLI(t, data, "zlib", "compress", "-in", "-", "-out", "-")
	if code != 0 {
		t.Fatalf("stream compress = %d, %s", code, stderr)
	}
	plain, err := decompressBytes([]byte(stdout))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plain, data) {
		t.Fatal("stream zlib mismatch")
	}
	if !strings.Contains(stderr, "compressed") {
		t.Fatalf("stream status = %q", stderr)
	}
}

func decompressBytes(data []byte) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	code := Run(
		[]string{"zlib", "decompress", "-in", "-", "-out", "-"},
		bytes.NewReader(data),
		&stdout,
		&stderr,
	)
	if code != 0 {
		return nil, os.ErrInvalid
	}
	return stdout.Bytes(), nil
}

func TestAESRSA_CLI(t *testing.T) {
	dir := t.TempDir()
	plain := []byte("aes and rsa command line")
	input := filepath.Join(dir, "plain.bin")
	if err := os.WriteFile(input, plain, 0o666); err != nil {
		t.Fatal(err)
	}
	passwordFile := filepath.Join(dir, "password.txt")
	if err := os.WriteFile(passwordFile, []byte("secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	iv := "000102030405060708090a0b0c0d0e0f"
	encrypted := filepath.Join(dir, "secret.bin")
	mustRun(t, nil, "aes", "encrypt", "-in", input, "-out", encrypted, "-password-file", passwordFile, "-iv", iv)
	decrypted := filepath.Join(dir, "plain.out")
	mustRun(t, nil, "aes", "decrypt", "-in", encrypted, "-out", decrypted, "-password-file", passwordFile, "-iv", iv)
	got, err := os.ReadFile(decrypted)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatal("AES round trip mismatch")
	}
	if code, _, _ := runCLI(t, nil, "aes", "encrypt", "-in", input, "-out", encrypted, "-password", "x", "-password-file", passwordFile, "-iv", iv); code != 2 {
		t.Fatalf("both password flags = %d", code)
	}

	publicKey := filepath.Join(dir, "public.pem")
	privateKey := filepath.Join(dir, "private.pem")
	mustRun(t, nil, "rsa", "generate", "-bits", "1024", "-public", publicKey, "-private", privateKey)
	signature := filepath.Join(dir, "data.sig")
	mustRun(t, nil, "rsa", "sign", "-key", privateKey, "-in", input, "-out", signature)
	if got := strings.TrimSpace(mustRun(t, nil, "rsa", "verify", "-key", publicKey, "-in", input, "-sig", signature)); got != "signature valid" {
		t.Fatalf("verify = %q", got)
	}
	if code, _, _ := runCLI(t, nil, "rsa", "verify", "-key", publicKey, "-in", passwordFile, "-sig", signature); code != 1 {
		t.Fatalf("invalid signature = %d", code)
	}
	rsaFile := filepath.Join(dir, "rsa.enc")
	rsaOut := filepath.Join(dir, "rsa.out")
	mustRun(t, nil, "rsa", "encrypt", "-key", publicKey, "-in", input, "-out", rsaFile)
	mustRun(t, nil, "rsa", "decrypt", "-key", privateKey, "-in", rsaFile, "-out", rsaOut)
	got, err = os.ReadFile(rsaOut)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatal("RSA round trip mismatch")
	}
}
