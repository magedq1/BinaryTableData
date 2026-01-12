# Core Components: BinaryTableData

The core functionality revolves around `BinaryTableData`, which manages the persistent storage and retrieval of structured data.

## BinaryTableData
The main class representing the table.
- **Path**: `com.qapsoft.io.BinaryTableData`
- **Purpose**: Provides random access to rows based on a fixed header schema.

### Key Capabilities
- **Direct Value Access**: `getInt`, `getString`, `getDouble`, `getValue` (raw bytes).
- **Direct Value Writing**: `setValue` for all supported types.
- **Search**: `indexOf(value, column)` allows linear search functionality.

```kotlin
val table = BinaryTableData(header, file)
val id = table.getInt(rowIndex = 5, columnName = "_id")
```

## BinaryTableHeader
Defines the schema of the table. Once created, the header is written to the beginning of the file.
- **MaxRowsCount**: Fixed capacity of the table.
- **Columns**: List of `BinaryColumn` definitions.
- **RowSize**: Automatically calculated sum of all column lengths.

## BinaryColumn
Defines a single column in the table.
- **Type**: `Int` (4 bytes), `Long` (8 bytes), `Double` (8 bytes), `String` (variable fixed length), `Raw` (variable fixed length).
- **Length**: Must be defined for `String` and `Raw`.

```kotlin
val col = BinaryColumn.String("username", length = 32)
```
