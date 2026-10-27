# CLI reference

**English** | [العربية](cli.ar.md)

The `binarytable` command is built from `cmd/binarytable` and implemented in
`internal/cli`. It uses only the Go standard library.

```bash
go run ./cmd/binarytable help
go install ./cmd/binarytable
binarytable version
```

Exit codes:

| Code | Meaning |
| --- | --- |
| `0` | success |
| `1` | runtime error, invalid data, or "not found" |
| `2` | usage error |

`-in -` and `-out -` mean stdin and stdout for the commands that support them.
When binary data is written to stdout, status messages go to stderr.

---

## 1. Table commands

```text
binarytable table <create|info|schema|get|set|dump|search> [flags]
```

### Column specifications

`-col` accepts `name:type[:length]` and can be repeated.

| Type | Meaning | Length |
| --- | --- | --- |
| `int`, `int32` | 4-byte integer | 4 or omitted |
| `long`, `int64` | 8-byte integer | 8 or omitted |
| `double`, `float64` | 8-byte float | 8 or omitted |
| `string`, `str`, `text` | UTF-8 string | default 64 |
| `raw`, `bytes`, `blob` | raw bytes | required |

Examples:

```bash
binarytable table create -file users.db -rows 1000 \
  -col _id:int -col name:string:64 -col score:double -col payload:raw:1024
```

### `table create`

| Flag | Required | Description |
| --- | --- | --- |
| `-file` | yes | output file |
| `-rows` | yes | row capacity |
| `-col` | yes | column spec, repeatable |
| `-force` | no | overwrite an existing file |

`CreateFile` truncates the target, so the CLI refuses to overwrite without
`-force`.

### `table info`

```bash
binarytable table info -file users.db
```

Prints file size, row count, row size, schema size, capacity, and columns.

### `table schema`

```bash
binarytable table schema -file users.db
```

Prints the raw schema text (`name/length/TYPE\n`).

### `table get`

| Flag | Description |
| --- | --- |
| `-file` | table file |
| `-row` | row index |
| `-col` | column name or zero-based index |
| `-hex` | print raw bytes as hex |
| `-out` | write raw bytes to a file, or `-` for stdout |

```bash
binarytable table get -file users.db -row 0 -col name
binarytable table get -file users.db -row 0 -col payload -out payload.bin
binarytable table get -file users.db -row 0 -col payload -hex
```

RAW columns print as hex by default. `-hex` and `-out` cannot be combined.

### `table set`

| Flag | Description |
| --- | --- |
| `-file` | table file |
| `-row` | row index |
| `-col` | column name or zero-based index |
| `-value` | typed value: string, int, long, double, or raw text |
| `-value-hex` | raw bytes as hex |
| `-value-file` | read raw bytes from a file, or `-` for stdin |

Exactly one value flag is required.

```bash
binarytable table set -file users.db -row 0 -col _id -value 1
binarytable table set -file users.db -row 0 -col name -value Alice
binarytable table set -file users.db -row 0 -col payload -value-hex 010203
binarytable table set -file users.db -row 0 -col payload -value-file avatar.png
```

`-value-hex` and `-value-file` write raw bytes. For `STRING`, prefer `-value`
so the rest of the column is zero-filled.

### `table dump`

```bash
binarytable table dump -file users.db -row 0
```

Output is tab-separated: `index`, `name`, `TYPE`, `value`. Strings are quoted;
raw values are hex.

### `table search`

| Flag | Description |
| --- | --- |
| `-file` | table file |
| `-col` | column name or index |
| `-value` | value to find |
| `-start` | first row to inspect (default 0) |

```bash
binarytable table search -file users.db -col name -value Alice
```

Prints the row index, or `-1` and exits with code 1 when not found. RAW search is
not supported.

---

## 2. Binary list commands

```text
binarytable list <create|info|add|get|dump> [flags]
```

### `list create`

```bash
binarytable list create -file blobs.bin -max 5000
```

- `-max` is required to be positive and is stored by reserving
  `100 + maxEntries × 4` bytes.
- `-force` overwrites an existing file.
- The reserved prefix lets a later `list info` / `list add` recover the capacity
  automatically even when the list is empty.

### `list info`

```bash
binarytable list info -file blobs.bin
```

`-max 0` (the default) recovers the capacity from the reserved prefix or from
the first data offset. Pass `-max N` explicitly for legacy files where detection
is ambiguous.

### `list add`

| Flag | Description |
| --- | --- |
| `-file` | list file (created if missing) |
| `-in` | blob file, or `-` for stdin |
| `-max` | capacity; `0` recovers it, new files default to 1000 |

```bash
binarytable list add -file blobs.bin -in avatar.png
printf 'hello' | binarytable list add -file blobs.bin -in -
```

### `list get`

| Flag | Description |
| --- | --- |
| `-file` | list file |
| `-index` | blob index |
| `-max` | list capacity |
| `-hex` | print hex |
| `-out` | write to a file, or `-` for stdout |

```bash
binarytable list get -file blobs.bin -index 0 -out avatar-copy.png
binarytable list get -file blobs.bin -index 0 -hex
```

### `list dump`

```bash
binarytable list dump -file blobs.bin
```

Prints `index<TAB>size` for every blob.

---

## 3. Compression commands

### `deflate`

```text
binarytable deflate <pack|write|unpack|read|info> [flags]
```

`pack` / `write`:

| Flag | Default | Description |
| --- | --- | --- |
| `-in` | required | input file, or `-` for stdin |
| `-out` | required | output file, or `-` for stdout |
| `-block` | 8192 | block size |
| `-header` | 0 | reserved caller header bytes |
| `-force` | false | overwrite the output file |

`unpack` / `read`:

| Flag | Default | Description |
| --- | --- | --- |
| `-in` | required | input file, or `-` for stdin |
| `-out` | required | output file, or `-` for stdout |
| `-cache` | 8 | decompressed blocks to cache |
| `-force` | false | overwrite the output file |

`info` reads the block header and prints the original size, compressed size,
ratio, block size, caller header, and block counts:

```bash
binarytable deflate pack -in data.bin -out data.deflate -block 8192
binarytable deflate unpack -in data.deflate -out data.bin -cache 4
binarytable deflate info -in data.deflate

# Explicit DeflateBinaryStreamReader-style read, piped to another tool:
binarytable deflate read -in data.deflate -out - | sha256sum
```

`read` is an alias for `unpack`; `write` is an alias for `pack`.

### `zlib`

```text
binarytable zlib <compress|decompress> -in PATH -out PATH
```

```bash
binarytable zlib compress -in data.bin -out data.zlib
binarytable zlib decompress -in data.zlib -out data.bin
```

Both commands accept `-` for stdin/stdout.

---

## 4. AES commands

```text
binarytable aes <iv|encrypt|decrypt> [flags]
```

`aes iv` prints a random 16-byte IV in hex.

`aes encrypt` / `aes decrypt`:

| Flag | Description |
| --- | --- |
| `-in` | input file, or `-` for stdin |
| `-out` | output file, or `-` for stdout |
| `-password` | password (visible in process arguments) |
| `-password-file` | file containing the password |
| `-iv` | 16-byte IV in hex, required |

Exactly one of `-password` and `-password-file` is required. Prefer
`-password-file` because command-line arguments can be visible to other
processes.

```bash
binarytable aes iv
binarytable aes encrypt -in plain.bin -out secret.bin \
  -password-file password.txt -iv 000102030405060708090a0b0c0d0e0f
binarytable aes decrypt -in secret.bin -out plain.bin \
  -password-file password.txt -iv 000102030405060708090a0b0c0d0e0f
```

---

## 5. RSA commands

```text
binarytable rsa <generate|sign|verify|encrypt|decrypt> [flags]
```

| Command | Flags |
| --- | --- |
| `generate` | `-bits` (default 2048), `-public`, `-private` |
| `sign` | `-key`, `-in`, `-out` |
| `verify` | `-key`, `-in`, `-sig` |
| `encrypt` | `-key`, `-in`, `-out` |
| `decrypt` | `-key`, `-in`, `-out` |

```bash
binarytable rsa generate -bits 2048 -public public.pem -private private.pem
binarytable rsa sign   -key private.pem -in message.bin -out message.sig
binarytable rsa verify -key public.pem  -in message.bin -sig message.sig
```

`verify` prints `signature valid` and exits 0, or exits 1 with
`signature invalid`.

RSA is intended for hosted targets; it is usually too heavy for small
microcontrollers.

---

## 6. Shell examples

Create, populate, and inspect a table:

```bash
set -e

binarytable table create -file users.db -rows 100 -force \
  -col _id:int -col name:string:32 -col score:double

binarytable table set -file users.db -row 0 -col _id -value 1
binarytable table set -file users.db -row 0 -col name -value Alice
binarytable table set -file users.db -row 0 -col score -value 9.5

binarytable table info -file users.db
binarytable table dump -file users.db -row 0
```

Compress a table file:

```bash
binarytable deflate pack -in users.db -out users.db.deflate
binarytable deflate unpack -in users.db.deflate -out users.db.copy
```
