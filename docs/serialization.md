# Serialization and Schema Generation

The library integrates with `kotlinx.serialization` to seamlessly map Kotlin objects to Binary Rows.

## Automatic Schema Creation
Instead of manually defining `BinaryTableHeader`, let the library inspect your class.

```kotlin
@Serializable
data class Product(
    val id: Int,
    @BinaryColumnLength(50) val name: String
)

val table = BinaryTableData.create<Product>(file, 1000)
```
- **@BinaryColumnLength**: Mandatory for `String` and `Raw` fields to define the fixed byte width in the row.

## Encoder / Decoder
- **BinaryTableEncoder**: Implements `kotlinx.serialization.Encoder`. Writes object properties to columns matching by name.
- **BinaryTableDecoder**: Implements `kotlinx.serialization.Decoder`. Reads column values into object properties.

### Convenience Extensions
```kotlin
// Read
val product: Product = table.decodeRow(0)
val all: Sequence<Product> = table.decodeAll()

// Write
table.encodeRow(0, product)
```

### Raw Data Support
You can map fields to `ByteArray`, `BinaryStream` (read-only), or `BinaryStreamReader` (read-only) for efficient access to RAW columns.

```kotlin
@Serializable
data class ImageRecord(
    val id: Int,
    @BinaryColumnLength(1024) val data: ByteArray
)
```
