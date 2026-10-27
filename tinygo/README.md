# BinaryTableData — Go / TinyGo port

**English** | [العربية](README.ar.md)

This is a documented Go port of the Kotlin **BinaryTableData** library. It keeps
the **exact same big-endian on-disk format**, adds an idiomatic Go API, and is
designed to stay usable with TinyGo. Compatibility is verified in both
directions: Go reads files produced by Kotlin, and Kotlin reads files produced by
Go.

- Core package: `github.com/magedq1/BinaryTableData/tinygo/binarytable`
- Optional crypto package: `github.com/magedq1/BinaryTableData/tinygo/cryptoutil`
- CLI: `github.com/magedq1/BinaryTableData/tinygo/cmd/binarytable`
- The core library and CLI use only the Go standard library. The optional
  Wails GUI has its own Wails/React frontend dependencies.

---

## Contents

1. [Compatibility](#compatibility)
2. [Requirements and installation](#requirements-and-installation)
3. [Quick start](#quick-start)
4. [CLI](#cli)
5. [Desktop GUI (Wails)](#desktop-gui-wails)
6. [Project layout](#project-layout)
7. [Documentation](#documentation)
8. [TinyGo notes](#tinygo-notes)
9. [Tests](#tests)
10. [Intentional differences from Kotlin](#intentional-differences-from-kotlin)

---

## Compatibility

Compatibility was checked with the original compiled Kotlin library:

| Item | Kotlin writes → Go reads | Go writes → Kotlin reads |
| --- | --- | --- |
| `BinaryTableData` in memory | ✅ | ✅ |
| `BinaryTableData` file | ✅ | ✅ |
| `BinaryList` | ✅ | ✅ |
| Block Deflate stream | ✅ | ✅ |
| AES/CBC/PKCS#5 | ✅ (byte-identical) | ✅ (byte-identical) |

The binary fixtures in `testdata/` were produced by Kotlin and are read by the
`*_compat_test.go` files. The Java scripts in `testdata/kotlin/` document how to
regenerate them and how to run the reverse check.

---

## Requirements and installation

- **Go 1.22 or newer** for the library and CLI.
- A recent **TinyGo** when targeting microcontrollers or WASI.
- **Go 1.23+, Wails v2.17+, Node.js 20+, and WebView2** for the desktop GUI.
- The library has no external dependencies and needs no CGO; only the GUI module
  pulls Wails and frontend packages.

The module lives inside the Kotlin repository. Use it with a `replace` directive:

```bash
go mod init example.com/myapp
go mod edit -replace github.com/magedq1/BinaryTableData/tinygo=../BinaryTableData/tinygo
go get github.com/magedq1/BinaryTableData/tinygo
```

Or work inside `tinygo/`:

```bash
go test ./...
go run ./examples/demo
go run ./cmd/binarytable help
```

Install the CLI from the local checkout:

```bash
go install ./cmd/binarytable
```

---

## Quick start

```go
package main

import (
	"fmt"
	"log"

	"github.com/magedq1/BinaryTableData/tinygo/binarytable"
)

func main() {
	header, err := binarytable.NewHeader(1000, []binarytable.Column{
		binarytable.IntColumn("_id"),
		binarytable.StringColumn("first_name", 64),
		binarytable.StringColumn("last_name", 64),
	})
	if err != nil {
		log.Fatal(err)
	}

	// In-memory table. Use CreateFile for a fixed-capacity file.
	stream := binarytable.NewAutoSizeStream(1024)
	table, err := binarytable.New(header, stream)
	if err != nil {
		log.Fatal(err)
	}

	if err := table.SetInt(0, 0, 1); err != nil {
		log.Fatal(err)
	}
	if err := table.SetString(0, 1, "Alice"); err != nil {
		log.Fatal(err)
	}
	if err := table.SetStringByName(0, "last_name", "Saleh"); err != nil {
		log.Fatal(err)
	}

	name, err := table.GetStringByName(0, "first_name")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(name) // Alice

	// Reopen from raw bytes (read-only).
	reopened, err := binarytable.FromBytes(stream.ToBytes())
	if err != nil {
		log.Fatal(err)
	}
	_ = reopened
}
```

A runnable example is in [`examples/demo/main.go`](examples/demo/main.go).

---

## CLI

The `binarytable` command covers tables, binary lists, compression, and crypto.

```bash
binarytable help
binarytable table help
```

Create and use a table:

```bash
binarytable table create -file users.db -rows 1000 \
  -col _id:int -col name:string:64 -col score:double

binarytable table set -file users.db -row 0 -col _id -value 1
binarytable table set -file users.db -row 0 -col name -value Alice
binarytable table get -file users.db -row 0 -col name
binarytable table dump -file users.db -row 0
binarytable table search -file users.db -col name -value Alice
binarytable table info -file users.db
```

Binary lists:

```bash
binarytable list create -file blobs.bin -max 5000
binarytable list add -file blobs.bin -in avatar.png
binarytable list get -file blobs.bin -index 0 -out avatar-copy.png
binarytable list dump -file blobs.bin
```

`list info` and `list add` recover the capacity automatically for lists created
by this port; use `-max N` for legacy or ambiguous files.

Compression:

```bash
binarytable zlib compress -in data.bin -out data.zlib
binarytable zlib decompress -in data.zlib -out data.bin

binarytable deflate pack -in data.bin -out data.deflate -block 8192
binarytable deflate unpack -in data.deflate -out data.bin
```

Crypto:

```bash
binarytable aes iv
binarytable aes encrypt -in plain.bin -out secret.bin \
  -password-file password.txt -iv 000102030405060708090a0b0c0d0e0f

binarytable rsa generate -bits 2048 -public public.pem -private private.pem
binarytable rsa sign   -key private.pem -in message.bin -out message.sig
binarytable rsa verify -key public.pem  -in message.bin -sig message.sig
```

`-in -` and `-out -` mean stdin/stdout for the commands that support them.
Status messages are written to stderr when binary output goes to stdout.
`deflate info` reads the block header, and `deflate read` is an explicit alias
for the `DeflateBinaryStreamReader`-style unpack operation.

See [CLI documentation](docs/cli.md) for all commands and flags.

---

## Desktop GUI (Wails)

The `gui/` module is a full desktop editor built with Wails v2, React 19, and
Tailwind CSS 4:

- Table schema builder, paginated data grid, type-aware cell inspector, search,
  row clearing, and file embedding into RAW cells.
- Binary list editor with explicit capacity, multi-file streaming embed,
  image/text/hex previews, export, replace, and atomic delete/rebuild.
- Deflate workbench with pack, unpack, header inspection, cache tuning, and
  **read-only browsing of compressed tables/lists through `DeflateReader`**.
- Global drag-and-drop, native file dialogs, command palette (`Ctrl+K`), and a
  dark OLED design system.

```bash
cd tinygo/gui
go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.0
wails dev       # hot reload
wails build -clean
```

The production executable is written to `build/bin/BinaryTableStudio.exe`.
See the [GUI documentation](gui/README.md) for features, shortcuts, design
tokens, and architecture.

---

## Project layout

```
tinygo/
├── go.mod
├── README.md / README.ar.md
├── binarytable/          core library
│   ├── stream.go         Reader / Writer / Stream / Grower
│   ├── bytearray.go      in-memory streams
│   ├── file.go           file-backed stream (ReadAt/WriteAt)
│   ├── subreader.go      SubReader and JoinedReader
│   ├── column.go         columns and schema encoding
│   ├── header.go         table header
│   ├── table.go          table operations
│   ├── search.go         indexOf helpers
│   ├── binarylist.go     variable-length blobs
│   ├── compression.go    direct zlib
│   ├── deflate.go        block deflate format
│   └── cache.go          LRU cache
├── cryptoutil/           AES, RSA, random
├── cmd/binarytable/      CLI entry point
├── internal/cli/         CLI implementation and tests
├── gui/                  Wails desktop editor (Go + React)
├── examples/demo/        runnable example
├── docs/                 documentation (English default, Arabic .ar.md)
└── testdata/             Kotlin fixtures and compatibility tools
```

---

## Documentation

- [Data format and compatibility](docs/format.md)
- [Tables, headers, and columns](docs/core.md)
- [Streams and random access](docs/streams.md)
- [Binary lists](docs/binarylist.md)
- [Deflate compression](docs/compression.md)
- [Helpers: search, encoding, cache](docs/helpers.md)
- [AES and RSA](docs/security.md)
- [CLI reference](docs/cli.md)
- [Desktop GUI (Wails)](gui/README.md)
- [TinyGo: constraints and targets](docs/tinygo.md)
- [Migrating from Kotlin](docs/migration.md)

Arabic versions are available next to each file with the `.ar.md` suffix, for
example [`docs/core.ar.md`](docs/core.ar.md).

---

## TinyGo notes

- **No reflection-based serialization.** Kotlin uses `kotlinx.serialization`;
  TinyGo only partially supports `reflect`. This port is **schema-first**:
  declare columns, then use typed helpers.
- **Memory streams work on microcontrollers.** `NewAutoSizeStream`,
  `NewByteArrayStream`, and `ByteArrayReader` do not need a filesystem.
- **File streams need a hosted target.** `OpenFile` and `CreateFile` work on
  Linux, Windows, WASI, and similar targets, not on bare-metal.
- **Compression** uses `compress/zlib`, which is supported by TinyGo.
- **AES** uses `crypto/aes`, `crypto/cipher`, and `crypto/sha256`.
  `cryptoutil` also provides RSA, but it is optional and heavy on small devices.
- **The Wails GUI is a desktop target**, not a TinyGo target. The core library
  remains TinyGo-friendly; the GUI is an additional host application.
- Build the example for WASI:

  ```bash
  tinygo build -target=wasi -o demo.wasm ./examples/demo
  ```

- Run tests with TinyGo (depending on the available target):

  ```bash
  tinygo test ./...
  ```

---

## Tests

```bash
cd tinygo
gofmt -l .
go vet ./...
go test ./...
go test -race ./...   # optional, requires CGO
```

The test suite covers streams (memory, files, sub/joined), tables, search,
`BinaryList`, Deflate across block boundaries, AES/RSA, CLI commands, GUI
services, and Kotlin fixture compatibility.

For the GUI module:

```bash
cd gui
go test ./...
cd frontend && npm run build
```

---

## Intentional differences from Kotlin

1. **Errors instead of exceptions.** Functions return `error`; use `errors.Is`
   with `ErrOverSize`, `ErrReadOnly`, `ErrOutOfBounds`, `ErrCorrupt`, or
   `ErrNoReader`.
2. **`SetString` zero-fills the rest of the column**, preventing stale bytes
   when a shorter string replaces a longer one. Raw `SetValue` keeps the Kotlin
   behaviour and does not zero-fill.
3. **No mmap.** File streams use `ReadAt`/`WriteAt` instead of
   `MappedByteBuffer`.
4. **Short reads.** Reading a cell beyond the physical end of a growable stream
   zero-fills the missing bytes, matching Kotlin in-memory streams.
5. **`BinaryList` keeps the 32-bit offsets** of the original format, so the
   practical limit is about 2 GiB.
6. **No reflection serialization.** See
   [Migrating from Kotlin](docs/migration.md).
