# Data Compression

Built-in support for Deflate compression allows you to store compressed data while retaining the ability to read it back efficiently.

## DeflateBinaryStreamReader
Wraps another `BinaryStreamReader` (the source of compressed data) and presents uncompressed data to the caller.
- **Mechanism**: Reads compressed blocks, inflates them, and caches them.
- **Usage**:
  ```kotlin
  val reader = DeflateBinaryStreamReader(compressedSource)
  val data = reader.readAt(0, buffer) // Reads UNCOMPRESSED data
  ```

## DeflateBinaryStreamWriter
Compresses data into blocks and writes them to a target `BinaryStreamWriter`.
- **Mechanism**: Buffers input -> Deflates -> Writes block + index.
- **Usage**:
  ```kotlin
  sourceReader.deflate(targetWriter)
  ```

## DataCompression Helper
Static utility methods for direct byte array compression.
- `deflateData(bytes)`
- `inflateData(bytes)`
