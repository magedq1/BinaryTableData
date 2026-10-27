# Data format and compatibility

**English** | [العربية](format.ar.md)

Every number in this library is **big-endian**, exactly like the original Kotlin
implementation. This is what makes files portable between both languages.

---

## 1. Table header

A table file starts with:

| Offset | Size | Type | Meaning |
| --- | --- | --- | --- |
| `0` | 4 | `int32` | `maxRowsCount` |
| `4` | 4 | `int32` | schema text length in bytes |
| `8` | variable | UTF-8 | schema text |

The schema contains one line per column:

```
name/length/TYPE\n
```

Example:

```
_id/4/INT
first_name/16/STRING
last_name/16/STRING
created/8/LONG
score/8/DOUBLE
payload/8/RAW
```

Valid types are `INT`, `LONG`, `DOUBLE`, `STRING`, and `RAW`. These are the same
names produced by `BinaryColumn.ColumnType.name` in Kotlin.

### Offset calculation

```text
schemaSize  = 8 + len(schema)
rowLength   = sum of all column lengths
cellOffset  = schemaSize + rowIndex*rowLength + columnStartPosition
capacity    = schemaSize + maxRowsCount*rowLength
```

`columnStartPosition` is the sum of the lengths of all preceding columns.

---

## 2. Column types

| Type | Default length | Encoding |
| --- | --- | --- |
| `INT` | 4 bytes | `int32`, big-endian |
| `LONG` | 8 bytes | `int64`, big-endian |
| `DOUBLE` | 8 bytes | IEEE-754, big-endian |
| `STRING` | caller-defined | UTF-8 in a fixed-size field |
| `RAW` | caller-defined | raw bytes |

### Strings

- `SetString` writes the string and zero-fills the rest of the column.
- `GetString` removes all `0x00` bytes, like Kotlin's `getString`.
- Length is measured in bytes, not runes. Arabic text often needs roughly twice
  as many bytes as characters, so choose a generous length.

### Unwritten cells

A new file or stream is zero-filled, so any cell that was never written reads as
zero. This matches Kotlin.

---

## 3. Growable and fixed-capacity tables

There are two common ways to create a table:

1. **A growable memory stream** such as `AutoSizeStream`. Its physical length
   ends at the last written byte; reads beyond that return zeros.
2. **A fixed-capacity file** via `CreateFile`. The file is exactly
   `header.Capacity()` bytes, matching
   `BinaryStreamImplMapped(file, schemaSize + maxRows*rowLength)` in Kotlin.

`GetAllRows()` returns bytes from the end of the schema to the physical end of
the stream. For a fixed file this is every row slot; for a growable stream it is
only the portion written so far.

---

## 4. Binary list

```text
offset 0                         int32  realSize
offset 100                       int32  start offset of item 0
offset 100 + 4                   int32  start offset of item 1
...
offset 100 + maxListSize*4       start of the data area
```

- `schemaSize = 100`.
- `indexSize = maxListSize * 4`.
- Data starts at `max(streamLength, 100 + indexSize)`.
- The end of an item is the start of the next item, or the stream end for the
  last item.
- Offsets are 32-bit, limiting a list to roughly 2 GiB.
- Adding an empty blob requires extending the stream to the data start. The Go
  port does this automatically through the `Grower` interface on built-in
  streams.
- `maxListSize` is not stored as a field. A new Go-created list reserves
  `100 + maxListSize*4` bytes, so `GuessListCapacity` can recover it later; for
  populated legacy lists the first data offset is used instead.

---

## 5. Block Deflate format

The header reserves 512 bytes, followed by the block index and compressed
blocks:

| Offset | Size | Meaning |
| --- | --- | --- |
| `100` | 4 | `blockSize` |
| `104` | 4 | `headerSize` (caller data) |
| `108` | 4 | `maxBlocksCount` |
| `112` | 4 | `blocksCount` |
| `116` | 8 | original length before compression |
| `headerSize + 512` | `maxBlocksCount*16` | block index |
| `headerSize + 512 + maxBlocksCount*16` | variable | blocks |

Each index entry contains:

- `int64 start` relative to the start of the block area.
- `int64 end` relative to the start of the block area.

Each block is compressed with **zlib** (not raw deflate), because Kotlin's
`java.util.zip.Deflater` defaults to zlib. The Go port uses `compress/zlib`, and
outputs are readable in both directions.

Bytes `[0, headerSize)` are left untouched, allowing callers to reserve their
own header.

---

## 6. Compatibility checks

The binary files in `testdata/` were produced by the original Kotlin code:

- `table.bin` — in-memory table.
- `table-file.bin` — fixed-capacity file table.
- `list.bin` — `BinaryList`.
- `deflate.bin` + `deflate-raw.bin` — Deflate stream and original data.
- `aes.bin` + `aes-iv.bin` + `aes-plain.bin` — AES/CBC/PKCS#5.

The `compat_test.go` files read these fixtures. The reverse direction was also
verified: Kotlin read files written by Go. Reproducible scripts live in
`testdata/kotlin/`.
