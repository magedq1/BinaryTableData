# Deflate compression

**English** | [العربية](compression.ar.md)

Package: `binarytable`

There are two levels of compression:

1. **Direct zlib compression** of a byte slice.
2. **Block Deflate streams** compatible with Kotlin's
   `DeflateBinaryStreamReader` and `DeflateBinaryStreamWriter`.

---

## 1. Direct compression

```go
compressed, err := binarytable.DeflateData(data)
restored, err := binarytable.InflateData(compressed)
```

For a range:

```go
compressed, err := binarytable.DeflateDataRange(data, start, length)
```

- Uses `compress/zlib`.
- Compatible with `java.util.zip.Deflater` (zlib, not raw deflate).
- `InflateData` returns `ErrCorrupt` for invalid input.

---

## 2. Block Deflate streams

### Writing

```go
source := binarytable.NewByteArrayReader(data)
output := binarytable.NewAutoSizeStream(4096)

err := binarytable.Deflate(
	source,
	output,
	binarytable.DefaultDeflateBlockSize, // 8192
	binarytable.DefaultDeflateHeaderSize, // 0
)
```

`Deflate` reads from a `Reader`, writes to a `Writer`, and leaves `headerSize`
bytes at the start for caller data.

### Reading

```go
reader, err := binarytable.NewDeflateReader(output, 8) // cache 8 blocks
if err != nil {
	return err
}
data, err := binarytable.ReadAll(reader)
```

Or:

```go
data, err := binarytable.InflateStream(output, 8)
```

- `reader.Size()` returns the original uncompressed length.
- `reader.ReadAt` works across block boundaries.
- Decompressed blocks are kept in an LRU cache; `cacheMax <= 0` disables it.

The header can be read without opening a reader:

```go
header, err := binarytable.ReadDeflateHeader(source)
fmt.Println(header.BlockSize, header.Blocks, header.OriginalLength)
```

`ReadDeflateHeader` is the direct Go equivalent of the metadata exposed by
`DeflateBinaryStreamReader`. The CLI exposes it through `deflate info` and
`deflate read`.

Because `DeflateReader` implements `binarytable.Reader`, it can be passed
directly to `FromReader` or wrapped with `AsStream` for `BinaryList`. The GUI
uses this to browse compressed tables and lists read-only without extracting the
file first.

---

## 3. Format

```text
100  int32  blockSize
104  int32  headerSize
108  int32  maxBlocksCount
112  int32  blocksCount
116  int64  originalLength
headerSize + 512                        block index
headerSize + 512 + maxBlocksCount*16    compressed blocks
```

Each index entry is `int64 start` and `int64 end`, relative to the block area.
Each block is zlib-compressed.

See [Data format](format.md) for the complete layout.

---

## 4. Choosing a block size

- The default is 8192 bytes, matching Kotlin.
- Smaller blocks: faster random access, larger index, slightly worse
  compression.
- Larger blocks: better compression, more memory per decompressed block.
- On a memory-constrained microcontroller doing random access, use small blocks
  and a small cache (for example 2–4).

---

## 5. Example

```go
raw := make([]byte, 20000)
for i := range raw {
	raw[i] = byte(i % 64)
}

compressed := binarytable.NewAutoSizeStream(1024)
if err := binarytable.Deflate(
	binarytable.NewByteArrayReader(raw),
	compressed,
	1024,
	0,
); err != nil {
	return err
}

restored, err := binarytable.InflateStream(compressed, 4)
if err != nil {
	return err
}
if !bytes.Equal(raw, restored) {
	return errors.New("compression round trip failed")
}
```

---

## 6. Compatibility

- `deflate.bin` in `testdata/` was produced by Kotlin and is read by
  `TestKotlinDeflateFixture`.
- Kotlin reading a Go-written stream was verified during development.
- The exact zlib bytes may differ slightly between Java and Go, but the stream
  format, original length, and block index are compatible.
