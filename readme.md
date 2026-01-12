# BinaryTableData Project Documentation

**BinaryTableData** is a comprehensive Kotlin library for structured binary data storage, providing random access, serialization, and stream management.

## Table of Contents
1.  [**Core Components**](docs/core.md)
    *   `BinaryTableData`, `BinaryTableHeader`, Columns, and basic manipulation.
2.  [**Streams & BinaryList**](docs/streams.md)
    *   `BinaryStream` ecosystem (RandomAccess, Mapped, Memory).
    *   `BinaryList` for variable-length item storage.
    *   Stream interop (`InputStream`/`OutputStream`).
3.  [**Serialization**](docs/serialization.md)
    *   Mapping Kotlin Objects to Binary Rows.
    *   Automatic Schema Creation.
    *   Annotations.
4.  [**Data Compression**](docs/compression.md)
    *   Deflate Streams for compressed storage with read access.
5.  [**Helpers**](docs/helpers.md)
    *   Search (`indexOf`), Cache, and IO utilities.

---

## Quick Start

### Installation
Add the serialization plugin and dependency (see [Serialization](docs/serialization.md)).

### Example: Storing Users
```kotlin
@Serializable
data class User(
    val id: Int,
    @BinaryColumnLength(64) val name: String
)

fun main() {
    // Create new table
    val table = BinaryTableData.create<User>(File("users.db"), 1000)
    
    // Write
    table.encodeRow(0, User(1, "Alice"))
    
    // Read
    val user = table.decodeRow<User>(0)
    println(user.name) // Alice
}
```

### Example: Using Streams
```kotlin
val fileStream = BinaryStreamImpRandomAccess(File("large_blob.bin"), "rw")
val outputStream = fileStream.asOutputStream()
// Use standard java IO
outputStream.write(byteArrayOf(1, 2, 3))
```
