# Tables, headers, and columns

**English** | [العربية](core.ar.md)

Package: `binarytable`

---

## 1. Defining columns

```go
col := binarytable.IntColumn("_id")          // 4 bytes
col = binarytable.LongColumn("created")      // 8 bytes
col = binarytable.DoubleColumn("score")      // 8 bytes
col = binarytable.StringColumn("name", 64)   // 64-byte UTF-8 field
col = binarytable.RawColumn("avatar", 8192)  // 8192 raw bytes
```

You can also construct a column directly:

```go
col := binarytable.Column{
	Name:   "custom",
	Length: 32,
	Type:   binarytable.TypeRaw,
}
```

Available types: `TypeInt`, `TypeLong`, `TypeDouble`, `TypeString`, `TypeRaw`.

### Constraints

- Names cannot be empty or contain `/`, `\n`, or `\r`.
- Length must be greater than zero.
- `INT` needs at least 4 bytes; `LONG` and `DOUBLE` need at least 8 bytes.
- Duplicate names are rejected.

---

## 2. Creating a header

```go
header, err := binarytable.NewHeader(1000, []binarytable.Column{
	binarytable.IntColumn("_id"),
	binarytable.StringColumn("first_name", 64),
})
if err != nil {
	// invalid schema
}
```

### Header methods

| Method | Result |
| --- | --- |
| `MaxRows()` | row capacity |
| `ColumnCount()` | number of columns |
| `Columns()` | copy of the column list |
| `SchemaSize()` | header size in bytes, including the first 8 bytes |
| `MaxRowLength()` | sum of column lengths |
| `Capacity()` | total table size in bytes |
| `SchemaBytes()` | copy of the encoded schema text |
| `ColumnInfo(index)` | column layout by index |
| `ColumnInfoByName(name)` | column layout by name |
| `ColumnIndex(name)` | column index |

A header is immutable after creation. Do not modify the slice returned by
`Columns()`.

---

## 3. Creating and opening tables

### In-memory table

```go
stream := binarytable.NewAutoSizeStream(1024)
table, err := binarytable.New(header, stream)
```

`New` writes the header and uses the stream for both reading and writing.

### Separate reader and writer

```go
table, err := binarytable.NewWithStreams(header, reader, writer)
```

### Read-only table over an existing reader

```go
table, err := binarytable.OpenReader(header, reader)
```

### Reading the header from a stream

```go
table, err := binarytable.FromReader(reader)
```

### From bytes

```go
table, err := binarytable.FromBytes(data) // read-only
```

### From files

```go
table, err := binarytable.CreateFile(header, "users.db") // create/truncate
defer table.Close()

table, err = binarytable.OpenFile("users.db")           // read/write
defer table.Close()

table, err = binarytable.OpenFileReadOnly("users.db")   // read-only
defer table.Close()
```

`CreateFile` creates a file of exactly `header.Capacity()` bytes. Tables opened
with `OpenFile` or `CreateFile` are closed with `Close`; streams that you own are
not closed automatically.

---

## 4. Accessing values

### Raw

```go
err := table.SetValue(row, column, rawBytes)
err = table.SetValueByName(row, "payload", rawBytes)

raw, err := table.GetValue(row, column)
raw, err = table.GetValueByName(row, "payload")

stream, err := table.GetValueAsStreamByName(row, "payload")
```

`GetValue` returns a full-length copy of the column. `SetValue` rejects values
longer than the column with `ErrOverSize`.

### Typed access

| Type | Read | Write |
| --- | --- | --- |
| `int32` | `GetInt`, `GetIntByName` | `SetInt`, `SetIntByName` |
| `int64` | `GetLong`, `GetLongByName` | `SetLong`, `SetLongByName` |
| `float64` | `GetDouble`, `GetDoubleByName` | `SetDouble`, `SetDoubleByName` |
| `string` | `GetString`, `GetStringByName` | `SetString`, `SetStringByName` |

Example:

```go
if err := table.SetInt(0, 0, 42); err != nil { ... }
id, err := table.GetInt(0, 0)

if err := table.SetStringByName(0, "name", "Alice"); err != nil { ... }
name, err := table.GetStringByName(0, "name")
```

`SetString` zero-fills the rest of the column, avoiding stale bytes when a
shorter value replaces a longer one.

---

## 5. Positions and size

```go
pos, err := table.Position(3, 1)
```

Returns the absolute byte offset of a cell. Rows are indexed from 0 to
`MaxRows()-1`; columns from 0 to `ColumnCount()-1`.

```go
rows, err := table.GetAllRows()
```

Returns all bytes after the schema.

---

## 6. Searching

```go
row, err := table.IndexOfInt(value, column, start)
row, err = table.IndexOfIntByName(value, "id", 0)
row, err = table.IndexOfLong(value, column, start)
row, err = table.IndexOfDouble(value, column, start)
row, err = table.IndexOfString(value, column, start)
```

Returns `-1` when no row matches. `start` is the first row to inspect.

```go
row, err := table.IndexOfStringByName("Alice", "first_name", 0)
if err != nil {
	return err
}
if row >= 0 {
	fmt.Println("found at row", row)
}
```

---

## 7. Errors

```go
var (
	binarytable.ErrReadOnly
	binarytable.ErrOutOfBounds
	binarytable.ErrOverSize
	binarytable.ErrCorrupt
	binarytable.ErrNoReader
)
```

Use `errors.Is`:

```go
if err := table.SetString(row, col, value); errors.Is(err, binarytable.ErrOverSize) {
	// value is too long for the column
}
```

---

## 8. Manual mapping to Go structs

There is no reflection-based serialization in TinyGo. Write two small mapping
functions instead:

```go
type User struct {
	ID    int32
	Name  string
	Score float64
}

func writeUser(table *binarytable.Table, row int, user User) error {
	if err := table.SetIntByName(row, "_id", user.ID); err != nil {
		return err
	}
	if err := table.SetStringByName(row, "name", user.Name); err != nil {
		return err
	}
	return table.SetDoubleByName(row, "score", user.Score)
}

func readUser(table *binarytable.Table, row int) (User, error) {
	id, err := table.GetIntByName(row, "_id")
	if err != nil {
		return User{}, err
	}
	name, err := table.GetStringByName(row, "name")
	if err != nil {
		return User{}, err
	}
	score, err := table.GetDoubleByName(row, "score")
	if err != nil {
		return User{}, err
	}
	return User{ID: id, Name: name, Score: score}, nil
}
```

This pattern is TinyGo-friendly, faster than reflection, and makes schema changes
easy to review.
