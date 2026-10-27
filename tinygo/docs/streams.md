# Streams and random access

**English** | [العربية](streams.ar.md)

Package: `binarytable`

---

## 1. Interfaces

```go
type Reader interface {
	ReadAt(p []byte, off int64) (n int, err error)
	Size() int64
}

type Writer interface {
	WriteAt(p []byte, off int64) (n int, err error)
}

type Stream interface {
	Reader
	Writer
}

type Grower interface {
	Grow(size int64) error
}
```

- `ReadAt` follows `io.ReaderAt`: it returns `io.EOF` when fewer bytes than
  requested are available at the end.
- `Size` returns the current logical length.
- `Grower` is used by `BinaryList` to reserve the data start for empty blobs.

### Adapting a reader to a stream

```go
stream := binarytable.AsStream(reader)
```

Writes on the result return `ErrReadOnly`.

---

## 2. Memory streams

### ByteArrayReader

Read-only over an existing slice, without copying:

```go
reader := binarytable.NewByteArrayReader(data)
```

### ByteArrayStream

A stream over a fixed-size slice:

```go
stream := binarytable.NewByteArrayStream(make([]byte, 4096))
```

- `Size` is the highest written position.
- Length cannot exceed the backing slice.
- A write crossing the end is clipped and returns `io.ErrShortWrite`.
- A write starting past the absolute end returns `ErrOutOfBounds`.
- `NewByteArrayStreamFixed(data)` makes `Size` always equal `len(data)`.
- `NewByteArrayStreamCapacity(n)` allocates a zero-filled slice.
- `ToBytes()` returns a copy of the written data.

### AutoSizeStream

Grows in fixed-size blocks:

```go
stream := binarytable.NewAutoSizeStream(1024)
```

- A non-positive block size selects 1024.
- `ToBytes()` returns a copy of the logical content.
- `Grow(n)` extends the stream with zeros without writing payload bytes.

This is the best choice for in-memory tables and microcontrollers.

---

## 3. File streams

```go
stream, err := binarytable.OpenFileStream("data.bin", binarytable.ReadWrite)
defer stream.Close()

stream, err = binarytable.OpenFileStream("data.bin", binarytable.ReadOnly)

stream, err = binarytable.CreateFileStream("data.bin", size, binarytable.ReadWrite)

stream, err = binarytable.OpenOrCreateFileStream("data.bin")
```

- `CreateFileStream` creates or truncates the file to the requested size.
- `binarytable.CreateFile(header, path)` creates a table file with the right
  capacity.
- `Grow` extends a file with `Truncate`.
- `Sync` flushes data to stable storage.
- `Name` returns the path.

Instead of Kotlin's `MappedByteBuffer`, the Go port uses `ReadAt/WriteAt`. This
keeps random access without mmap and works on hosted TinyGo targets.

---

## 4. Views and concatenation

### SubReader

```go
sub, err := binarytable.NewSubReader(reader, start, end)
sub, err = binarytable.NewSubReader(reader, start) // to the end
```

- Read-only window.
- `Start`, `End`, and `Source` are available.
- `SubReader(start, end)` creates a nested window.

### JoinedReader

```go
joined := binarytable.NewJoinedReader(part1, part2, part3)
```

Concatenates several readers into one logical stream. Reads cross part
boundaries automatically.

---

## 5. Helper functions

```go
b, err := binarytable.ReadBytesAt(reader, offset, size)
b, err := binarytable.ReadAll(reader)
err = binarytable.ReadAtFull(reader, offset, buffer)
err = binarytable.WriteAtFull(writer, offset, buffer)
err = binarytable.CopyAt(dst, dstOffset, src, srcOffset, size)

// asInputStream / asOutputStream equivalents
ioReader := binarytable.NewIOReader(reader)
ioWriter := binarytable.NewIOWriter(writer, startOffset)
```

- `ReadAtFull` and `WriteAtFull` return an error if the full length cannot be
  transferred.
- `CopyAt` copies between streams in chunks, without loading everything into
  memory.

---

## 6. Concurrency

- `ByteArrayStream`, `AutoSizeStream`, and `FileStream` are safe for concurrent
  use.
- `ByteArrayReader` is read-only and needs no lock.
- `SubReader` and `JoinedReader` depend on the safety of their source.
- `Table` has no internal lock; use an external lock for concurrent writes to
  the same rows.

---

## 7. Complete example

```go
src := binarytable.NewAutoSizeStream(256)
_, _ = src.WriteAt([]byte("hello world"), 0)

sub, _ := binarytable.NewSubReader(src, 6, 11)
b, _ := binarytable.ReadAll(sub)
fmt.Println(string(b)) // world

joined := binarytable.NewJoinedReader(
	binarytable.NewByteArrayReader([]byte("abc")),
	binarytable.NewByteArrayReader([]byte("def")),
)
all, _ := binarytable.ReadAll(joined)
fmt.Println(string(all)) // abcdef
```
