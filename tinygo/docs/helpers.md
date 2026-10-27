# Helpers

**English** | [العربية](helpers.ar.md)

Package: `binarytable`

---

## 1. Number encoding

All numbers are big-endian, like Kotlin.

```go
b := binarytable.EncodeInt32(42)
v, err := binarytable.DecodeInt32(b)

b = binarytable.EncodeInt64(1_700_000_000_000)
v64, err := binarytable.DecodeInt64(b)

b = binarytable.EncodeFloat64(12.5)
f, err := binarytable.DecodeFloat64(b)
```

Append without an extra allocation:

```go
dst = binarytable.AppendInt32(dst, value)
dst = binarytable.AppendInt64(dst, value)
dst = binarytable.AppendFloat64(dst, value)
```

---

## 2. Searching

```go
row, err := table.IndexOfInt(value, column, start)
row, err = table.IndexOfLong(value, column, start)
row, err = table.IndexOfDouble(value, column, start)
row, err = table.IndexOfString(value, column, start)

row, err = table.IndexOfIntByName(value, "id", start)
row, err = table.IndexOfLongByName(value, "created", start)
row, err = table.IndexOfDoubleByName(value, "score", start)
row, err = table.IndexOfStringByName(value, "name", start)
```

- `start` is the first row to inspect.
- `-1` means no match.
- The search is linear, like `ContentHelper.kt` in Kotlin.

---

## 3. LRU cache

```go
cache := binarytable.NewCache[string, int](50)
cache.Put("a", 1)

if value, ok := cache.Get("a"); ok {
	fmt.Println(value)
}

cache.Remove("a")
cache.Clear()
```

- Equivalent to Kotlin's `CacheManager`.
- Thread-safe.
- Capacity `<= 0` disables caching.
- Used internally by `DeflateReader`.

---

## 4. Stream helpers

```go
data, err := binarytable.ReadAll(reader)
data, err = binarytable.ReadBytesAt(reader, offset, size)
err = binarytable.ReadAtFull(reader, offset, buffer)
err = binarytable.WriteAtFull(writer, offset, buffer)
err = binarytable.CopyAt(dst, dstOffset, src, srcOffset, size)
```

- `ReadAtFull` guarantees a full buffer or an error.
- `WriteAtFull` guarantees every byte is written.
- `CopyAt` copies in chunks without loading everything into memory.

---

## 5. Common errors

```go
binarytable.ErrReadOnly
binarytable.ErrOutOfBounds
binarytable.ErrOverSize
binarytable.ErrCorrupt
binarytable.ErrNoReader
```

```go
if errors.Is(err, binarytable.ErrOverSize) {
	// value is larger than the column
}
```

---

## 6. Concurrency

- `Cache` is safe for concurrent use.
- `ByteArrayStream`, `AutoSizeStream`, and `FileStream` are safe.
- `Table` has no internal lock.
- `DeflateReader` uses the thread-safe cache, but concurrent reads are not
  guaranteed to be consistent if the source is modified meanwhile.

---

## 7. Performance tips

- Use `GetValueAsStream` for large raw data instead of copying it.
- Use an appropriate `AutoSizeStream` block size (1024–8192) to reduce
  reallocations.
- Use `JoinedReader` for split files without merging them on disk.
- Use `BinaryList.Reader` to pass a blob to another parser without copying.
- In TinyGo, avoid allocations in hot loops when possible.
