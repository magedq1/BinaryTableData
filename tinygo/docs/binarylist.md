# Binary lists

**English** | [العربية](binarylist.ar.md)

Package: `binarytable`

`BinaryList` stores variable-length blobs (images, documents, messages) in one
stream, with a separate index. Unlike `BinaryTableData`, it does not require
fixed-size rows.

---

## 1. Format

```text
offset 0                       int32  realSize
offset 100                     int32  start offset of item 0
...
offset 100 + maxListSize*4     start of data
```

- `schemaSize = 100` bytes.
- `indexSize = maxListSize * 4` bytes.
- Data starts at `max(streamLength, 100 + indexSize)`.
- The end of an item is the start of the next item, or the stream end for the
  last item.
- Offsets are 32-bit, so the practical limit is about 2 GiB.

This format is identical to Kotlin's `BinaryList.kt`; lists are readable in both
directions.

### Capacity is not stored as a field

The Kotlin format has no explicit `maxListSize` field. A newly initialized list
in this port writes `realSize = 0` and reserves `100 + maxEntries × 4` bytes.
That lets `GuessListCapacity` recover the capacity from the prefix even when the
list is empty. For a populated legacy list it can also use the first data
offset.

```go
list, _ := binarytable.NewBinaryList(stream, 5000)
err := list.Initialize() // reserve the prefix for a new file

capacity, _ := binarytable.GuessListCapacity(reader) // 5000
```

---

## 2. Creating a list

```go
stream := binarytable.NewAutoSizeStream(1024)
list, err := binarytable.NewBinaryList(stream, 1000)
```

The second argument is `maxListSize`. A non-positive value selects 1000. The
entry count is read from the first 4 bytes, so an existing list can be reopened:

```go
list, err := binarytable.NewBinaryList(
	binarytable.NewByteArrayStreamFixed(existingBytes),
	1000,
)
```

---

## 3. Adding and reading

```go
index, err := list.Add([]byte("maged"))
index, err = list.AddReader(reader) // stream without buffering the blob
```

`Add` returns the new index, or an error if the list is full. `AddReader`
streams directly into the list, so large files do not have to fit in memory.

```go
index, err := list.AddFunc(func(w io.Writer) error {
	_, err := w.Write([]byte("ahmed"))
	return err
})
```

`AddFunc` builds the blob in memory and then appends it, like the Kotlin
`add { outputStream -> ... }` overload.

```go
data, err := list.Bytes(index)
reader, err := list.Reader(index) // SubReader
ioReader, err := list.Open(index) // sequential io.Reader
```

- `Bytes` returns a full copy.
- `Reader` returns a `SubReader` without copying.
- `Open` returns a sequential `io.Reader`.
- `ToBytes` returns a copy of the whole stream.
- `Size` returns the number of entries.
- `SchemaSize`, `IndexSize`, and `MaxSize` are available for inspection.

---

## 4. Empty blobs

Adding an empty blob writes no payload, so it does not extend the stream
physically. To keep the index valid, `BinaryList` extends the stream to the data
start through the `Grower` interface. The built-in streams (`AutoSizeStream`,
`ByteArrayStream`, `FileStream`) support it.

If you use an external stream that does not implement `Grower`, `Add(nil)`
returns a clear error instead of writing a corrupt index.

---

## 5. Errors

| Condition | Error |
| --- | --- |
| Index out of range | `ErrOutOfBounds` |
| List is full | `ErrOutOfBounds` |
| Invalid start offset | `ErrCorrupt` |
| End smaller than start | `ErrCorrupt` |
| `nil` stream | ordinary error |
| Stream cannot store an empty blob | ordinary error |

---

## 6. Example with a table

Store an image in a `BinaryList` and link it to a table row:

```go
avatarID, err := list.Add(imageBytes)
if err != nil {
	return err
}

if err := table.SetIntByName(row, "avatar_id", avatarID); err != nil {
	return err
}

// Later
id, err := table.GetIntByName(row, "avatar_id")
image, err := list.Bytes(int(id))
```

This keeps table rows small and fixed-size while large payloads live in a
separate list, matching the original design.
