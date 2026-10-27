# Migrating from Kotlin

**English** | [العربية](migration.ar.md)

Quick mapping between the original Kotlin API and the Go port.

---

## 1. Columns and headers

| Kotlin | Go |
| --- | --- |
| `BinaryColumn.Int("id")` | `binarytable.IntColumn("id")` |
| `BinaryColumn.Long("n")` | `binarytable.LongColumn("n")` |
| `BinaryColumn.Double("d")` | `binarytable.DoubleColumn("d")` |
| `BinaryColumn.String("s", 64)` | `binarytable.StringColumn("s", 64)` |
| `BinaryColumn.Raw("r", 1024)` | `binarytable.RawColumn("r", 1024)` |
| `BinaryTableHeader(maxRows, columns)` | `binarytable.NewHeader(maxRows, columns)` |
| `header.maxRowsCount` | `header.MaxRows()` |
| `header.schemaSize` | `header.SchemaSize()` |
| `header.maxRowLength` | `header.MaxRowLength()` |
| `header.columnsInfoByName` | `header.ColumnInfoByName(name)` |

---

## 2. Creating tables

| Kotlin | Go |
| --- | --- |
| `BinaryTableData(header, file)` | `binarytable.CreateFile(header, path)` |
| `BinaryTableData(header, stream)` | `binarytable.New(header, stream)` |
| `BinaryTableData(header, reader, writer)` | `binarytable.NewWithStreams(header, reader, writer)` |
| `BinaryTableData(header, streamWriter = w)` | `binarytable.NewWithStreams(header, nil, w)` |
| `BinaryTableData.from(file)` | `binarytable.OpenFile(path)` |
| `BinaryTableData.from(bytes)` | `binarytable.FromBytes(data)` |
| `BinaryTableData.from(reader)` | `binarytable.FromReader(reader)` |

`FromBytes` and `FromReader` return read-only tables. Use `OpenFile` or
`NewWithStreams` for writing.

---

## 3. Accessing values

| Kotlin | Go |
| --- | --- |
| `getValue(row, name)` | `GetValueByName(row, name)` |
| `getValue(row, index)` | `GetValue(row, index)` |
| `setValue(row, name, bytes)` | `SetValueByName(row, name, bytes)` |
| `setValue(row, index, bytes)` | `SetValue(row, index, bytes)` |
| `getInt(row, name)` | `GetIntByName(row, name)` |
| `getLong(row, name)` | `GetLongByName(row, name)` |
| `getDouble(row, name)` | `GetDoubleByName(row, name)` |
| `getString(row, name)` | `GetStringByName(row, name)` |
| `setValue(row, name, int)` | `SetIntByName(row, name, value)` |
| `setValue(row, name, long)` | `SetLongByName(row, name, value)` |
| `setValue(row, name, double)` | `SetDoubleByName(row, name, value)` |
| `setValue(row, name, string)` | `SetStringByName(row, name, value)` |
| `getValueAsStream` | `GetValueAsStream` / `GetValueAsStreamByName` |
| `getAllRows()` | `GetAllRows()` |
| `getPos(row, column)` | `Position(row, column)` |

Kotlin throws exceptions. In Go, every operation that can fail returns `error`.

---

## 4. Searching

| Kotlin | Go |
| --- | --- |
| `indexOf(int, columnIndex)` | `IndexOfInt(value, column, start)` |
| `indexOf(int, columnName)` | `IndexOfIntByName(value, name, start)` |
| `indexOf(long, ...)` | `IndexOfLong` / `IndexOfLongByName` |
| `indexOf(double, ...)` | `IndexOfDouble` / `IndexOfDoubleByName` |
| `indexOf(string, ...)` | `IndexOfString` / `IndexOfStringByName` |

---

## 5. Streams

| Kotlin | Go |
| --- | --- |
| `ByteArrayStreamReader(bytes)` | `binarytable.NewByteArrayReader(data)` |
| `ByteArrayStream(bytes)` | `binarytable.NewByteArrayStream(data)` |
| `ByteArrayStreamAutoSize()` | `binarytable.NewAutoSizeStream(1024)` |
| `BinaryStreamImpRandomAccess(file)` | `binarytable.OpenFileStream(path, binarytable.ReadWrite)` |
| `BinaryStreamImplMapped(file, size)` | `binarytable.CreateFileStream(path, size, binarytable.ReadWrite)` |
| `stream.subReader(a, b)` | `binarytable.NewSubReader(reader, a, b)` |
| `JoinedBinaryStreamReader(list)` | `binarytable.NewJoinedReader(readers...)` |
| `reader.toByteArray()` | `binarytable.ReadAll(reader)` |
| `reader.getBytesAt(pos, size)` | `binarytable.ReadBytesAt(reader, pos, size)` |
| `reader.asInputStream()` | `binarytable.NewIOReader(reader)` |
| `writer.asOutputStream(start)` | `binarytable.NewIOWriter(writer, start)` |

`ByteArrayStreamAutoSize` corresponds to `AutoSizeStream`.

---

## 6. BinaryList

| Kotlin | Go |
| --- | --- |
| `BinaryList(stream, max)` | `binarytable.NewBinaryList(stream, max)` |
| `list.add(bytes)` | `list.Add(bytes)` or `list.AddReader(reader)` |
| `list.add { os -> ... }` | `list.AddFunc(func(w io.Writer) error { ... })` |
| `list.getBytes(i)` | `list.Bytes(i)` |
| `list.getAsReader(i)` | `list.Reader(i)` |
| `list.get(i)` | `list.Open(i)` (`io.Reader`) |
| `list.toByteArray()` | `list.ToBytes()` |
| `list.size` | `list.Size()` |

---

## 7. Compression

| Kotlin | Go |
| --- | --- |
| `DataCompression.deflateData(bytes)` | `binarytable.DeflateData(bytes)` |
| `DataCompression.inflateData(bytes)` | `binarytable.InflateData(bytes)` |
| `reader.deflate(writer, blockSize, headerSize)` | `binarytable.Deflate(reader, writer, blockSize, headerSize)` |
| `DeflateBinaryStreamReader(source, cache)` | `binarytable.NewDeflateReader(source, cache)` |

---

## 8. Crypto

| Kotlin | Go |
| --- | --- |
| `AESEncryptor.encrypt(key, iv, in, out)` | `cryptoutil.EncryptStream(key, iv, in, out)` |
| `AESEncryptor.decrypt(key, iv, in, out)` | `cryptoutil.DecryptStream(key, iv, in, out)` |
| `AESEncryptor.convertPasswordTo32Bytes(key)` | `cryptoutil.PasswordToKey(key)` |
| `AESEncryptor.generateIV()` | `cryptoutil.GenerateIV()` |
| `EncryptionHelper.randomString(n)` | `cryptoutil.RandomString(n)` |
| `RSAEncryptHelper.*` | `cryptoutil.*` (see [Crypto](security.md)) |

---

## 9. Automatic serialization

There is no equivalent of `@Serializable` and `@BinaryColumnLength` in this
port. Replace:

```kotlin
@Serializable
data class User(val id: Int, @BinaryColumnLength(64) val name: String)
```

with:

```go
type User struct {
	ID   int32
	Name string
}
```

and `readUser`/`writeUser` functions that use `GetIntByName` and
`SetStringByName`. See [Tables](core.md#8-manual-mapping-to-go-structs) for a
complete example.

---

## 10. Migration checklist

- [ ] Convert `@Serializable` classes to explicit column definitions.
- [ ] Make sure `STRING` lengths are large enough for UTF-8 (especially Arabic).
- [ ] Replace exceptions with `error` checks.
- [ ] Use `errors.Is` with the library sentinel errors.
- [ ] Test reading existing Kotlin files before writing to them.
- [ ] Keep a backup: `CreateFile` truncates an existing file.
- [ ] Use `OpenFileReadOnly` for safe inspection.
