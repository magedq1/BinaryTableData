# Helper Utilities

## ContentHelper (Search)
Provides linear search capabilities over columns.
- `indexOf(value, columnIndex)`
- `indexOf(value, columnName)`

Supported for `Int`, `Long`, `Double`, and `String`.

## IoHelper
General IO extension functions.
- `asInputStream()`: Convert BinaryStreamReader to InputStream.
- `asOutputStream()`: Convert BinaryStreamWriter to OutputStream.
- Type conversions: `ByteArray.asInt()`, `Long.toByteArray()`, etc.

## CacheManager
Simple LRU Cache implementation used internally by `DeflateBinaryStreamReader` but available for general use.
