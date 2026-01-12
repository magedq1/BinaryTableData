# Streams and BinaryList

The library abstracts storage access using a powerful `BinaryStream` interface ecosystem, allowing data to live in memory, on disk, or be split across multiple sources.

## BinaryStream Interface
`com.qapsoft.io.BinaryStream`
Combines `BinaryStreamReader` and `BinaryStreamWriter`.
- `readAt(pos, buffer)`: Random access read.
- `writeAt(pos, buffer)`: Random access write.
- `length()`: Total size.

### Implementations
1.  **BinaryStreamImpRandomAccess**:
    - Wraps `java.io.RandomAccessFile`.
    - Best for large files on disk.
    - Thread-safe via `FileChannel`.
2.  **BinaryStreamImplMapped**:
    - Wraps `MappedByteBuffer`.
    - Incredible performance for frequent random access.
    - **Warning**: Map size limited by OS (2GB on some 32-bit systems, though less of an issue on 64-bit).
3.  **ByteArrayStream**:
    - Simple wrapper around a fixed-size `ByteArray`.
4.  **ByteArrayStreamAutoSize**:
    - In-memory stream that grows dynamically.
    - Uses a list of blocks meant to minimize resizing overhead.

### Helpers
- **JoinedBinaryStreamReader**: Combines multiple readers into one. Great for split files!
- **SubBinaryStreamReader**: Limits access to a specific range of an existing stream.
- **Interoperability**:
    - `stream.asInputStream()`: Converts to `java.io.InputStream`.
    - `writer.asOutputStream()`: Converts to `java.io.OutputStream`.

---

## BinaryList
`com.qapsoft.io.BinaryList`

A specialized structure to manage a **list of variable-length blobs** using a `BinaryStream`. Unlike `BinaryTableData` which requires fixed rows, `BinaryList` stores an index and data separately.

### Usage
```kotlin
val list = BinaryList(stream)
val index = list.add(myByteArray)
val data = list.getBytes(index)
```
- **Schema**:
    - Header: 4 bytes (size).
    - Index Area: `maxListSize * 4` bytes.
    - Data Area: Appended after index.

This is perfect for storing things like images or distinct large documents linked by ID from the main `BinaryTableData`.
