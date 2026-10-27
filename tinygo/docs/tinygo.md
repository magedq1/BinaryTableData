# TinyGo: constraints and targets

**English** | [العربية](tinygo.ar.md)

This page explains how the port was designed for TinyGo and what to keep in mind
when moving from standard Go to TinyGo.

---

## 1. What works?

| Feature | TinyGo |
| --- | --- |
| Memory streams | ✅ |
| `ByteArrayReader` / `AutoSizeStream` | ✅ |
| Fixed-schema tables | ✅ |
| `BinaryList` | ✅ |
| zlib / block Deflate | ✅ |
| Files through `os.File` | ✅ on hosted targets |
| AES / SHA-256 / randomness | ✅ |
| RSA / X.509 | ⚠️ heavy; test on your target |
| Reflection serialization | ❌ |
| mmap | ❌ (replaced by ReadAt/WriteAt) |

The core library does not depend on `reflect` or `unsafe`. The standard-library
packages used are on TinyGo's supported list:

`bytes`, `compress/zlib`, `container/list`, `crypto/aes`, `crypto/cipher`,
`crypto/rand`, `crypto/rsa`, `crypto/sha256`, `crypto/x509`, `encoding/pem`,
`errors`, `fmt`, `io`, `math`, `os`, `strconv`, `strings`, `sync`, `testing`.

The CLI additionally uses `flag` and `encoding/hex`, both supported by TinyGo.

The **Wails GUI is a desktop application**, not a TinyGo target. It depends on
WebView2 and the Wails runtime; the core library remains TinyGo-compatible.

---

## 2. No automatic serialization

In Kotlin:

```kotlin
@Serializable
data class User(val id: Int, @BinaryColumnLength(64) val name: String)

val table = BinaryTableData.create<User>(file, 1000)
table.encodeRow(0, User(1, "Alice"))
```

There is no safe reflection equivalent in TinyGo. The replacement:

```go
header, _ := binarytable.NewHeader(1000, []binarytable.Column{
	binarytable.IntColumn("id"),
	binarytable.StringColumn("name", 64),
})

func writeUser(t *binarytable.Table, row int, u User) error {
	if err := t.SetIntByName(row, "id", u.ID); err != nil {
		return err
	}
	return t.SetStringByName(row, "name", u.Name)
}
```

This means writing a small mapping function, but it:

- works on TinyGo;
- is faster than reflection;
- makes schema changes easy to review;
- avoids a `kotlinx.serialization` dependency.

For large projects, you can generate the mapping functions with your own
codegen.

---

## 3. Memory and blocks

- `AutoSizeStream` allocates blocks of `blockSize` bytes (default 1024).
- On a small microcontroller, use smaller blocks and size the table precisely.
- `ByteArrayReader` does not copy data; use it for read-only data.
- `GetValueAsStream` avoids copying large raw columns.
- `DeflateReader` keeps a limited number of decompressed blocks. Use a small
  `cacheMax` (2–4) on constrained devices.
- `EncryptStream` reads all input into memory. For large data, use `Encrypt` on
  chunks.

---

## 4. Files and targets

`OpenFile`, `CreateFile`, `OpenFileStream`, and `CreateFileStream` rely on
`os.File`. They work on:

- Linux
- Windows
- macOS
- WASI

They are usually not used on bare-metal targets. Use memory streams there, or
implement `Reader`/`Writer` over your own storage (Flash, SD, EEPROM, ...). The
interfaces are small:

```go
type Reader interface {
	ReadAt(p []byte, off int64) (int, error)
	Size() int64
}

type Writer interface {
	WriteAt(p []byte, off int64) (int, error)
}
```

Any storage implementing these two methods can be used with tables, `BinaryList`,
and compression.

---

## 5. Build commands

From `tinygo/`:

```bash
# Standard Go
go build ./...
go test ./...

# TinyGo for WASI
tinygo build -target=wasi -o demo.wasm ./examples/demo

# TinyGo for Linux (when available)
tinygo build -target=linux -o demo ./examples/demo

# TinyGo tests, depending on the target
tinygo test ./...
```

### Reducing size

- Do not import `cryptoutil` unless you need crypto.
- Use `-opt=z` or `-opt=s` as appropriate.
- Avoid `fmt` on hot paths on microcontrollers.
- Use `-print-allocs=.` to inspect allocations (depends on the TinyGo version).

---

## 6. What may differ?

- `fmt` and `reflect` in TinyGo may not support every advanced case. The library
  does not rely on reflection.
- Some targets do not fully support `recover`. The library does not use it.
- `os` may be limited on some targets.
- `crypto/rsa` may only work on hosted targets. Test it.
- Map performance is relatively slower; column names are resolved once while
  building the header, then column info is accessed by index.

---

## 7. Practical recommendation

For a memory-constrained microcontroller:

1. Use a small fixed schema.
2. Use `AutoSizeStream` or a custom Flash stream.
3. Store large payloads in `BinaryList` instead of huge RAW columns.
4. Use Deflate only if the data actually compresses.
5. Do not run RSA on-device; use AES or verify a precomputed signature.
