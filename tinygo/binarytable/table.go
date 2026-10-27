package binarytable

import (
	"fmt"
	"io"
	"math"
)

// Table provides random access to fixed-size rows.
//
// A Table can be read-only (StreamWriter == nil), write-only, or read/write.
// Rows are addressed by a zero-based index. Values must fit in their column.
type Table struct {
	// Header is the immutable schema of the table.
	Header *Header

	// StreamReader is the underlying reader, or nil for a write-only table.
	StreamReader Reader

	// StreamWriter is the underlying writer, or nil for a read-only table.
	StreamWriter Writer

	closer io.Closer
}

// New creates a table on stream, writes the header at offset zero, and returns
// a read/write table.
func New(header *Header, stream Stream) (*Table, error) {
	return NewWithStreams(header, stream, stream)
}

// NewWithStreams creates a table with separate reader and writer. When writer is
// non-nil the header is written immediately, matching the Kotlin initializer.
func NewWithStreams(header *Header, reader Reader, writer Writer) (*Table, error) {
	if header == nil {
		return nil, fmt.Errorf("binarytable: nil header")
	}
	t := &Table{Header: header, StreamReader: reader, StreamWriter: writer}
	if writer != nil {
		if err := t.writeHeader(); err != nil {
			return nil, err
		}
	}
	return t, nil
}

// OpenReader creates a read-only table from an existing reader without
// rewriting the header.
func OpenReader(header *Header, reader Reader) (*Table, error) {
	if header == nil {
		return nil, fmt.Errorf("binarytable: nil header")
	}
	if reader == nil {
		return nil, ErrNoReader
	}
	return &Table{Header: header, StreamReader: reader}, nil
}

// FromReader reads the schema from the beginning of reader and returns a
// read-only table.
func FromReader(reader Reader) (*Table, error) {
	if reader == nil {
		return nil, ErrNoReader
	}
	if reader.Size() < 8 {
		return nil, fmt.Errorf("%w: stream is shorter than the 8-byte table prefix", ErrCorrupt)
	}
	prefix, err := ReadBytesAt(reader, 0, 8)
	if err != nil {
		return nil, err
	}
	maxRows, err := DecodeInt32(prefix[:4])
	if err != nil {
		return nil, err
	}
	schemaSize, err := DecodeInt32(prefix[4:])
	if err != nil {
		return nil, err
	}
	if maxRows < 0 || schemaSize < 0 {
		return nil, fmt.Errorf("%w: negative max rows or schema size", ErrCorrupt)
	}
	if int64(schemaSize) > reader.Size()-8 {
		return nil, fmt.Errorf("%w: schema size %d exceeds stream size %d", ErrCorrupt, schemaSize, reader.Size())
	}
	if int64(schemaSize) > int64(maxInt()) {
		return nil, fmt.Errorf("%w: schema does not fit in memory", ErrOutOfBounds)
	}
	schema, err := ReadBytesAt(reader, 8, int(schemaSize))
	if err != nil {
		return nil, err
	}
	columns, err := parseColumns(schema)
	if err != nil {
		return nil, err
	}
	header, err := NewHeader(int(maxRows), columns)
	if err != nil {
		return nil, err
	}
	return &Table{Header: header, StreamReader: reader}, nil
}

// FromBytes reads a table from an in-memory byte slice. The resulting table is
// read-only and shares the supplied slice.
func FromBytes(data []byte) (*Table, error) {
	return FromReader(NewByteArrayReader(data))
}

// OpenFile opens an existing table file for reading and writing.
func OpenFile(path string) (*Table, error) {
	stream, err := OpenFileStream(path, ReadWrite)
	if err != nil {
		return nil, err
	}
	table, err := FromReader(stream)
	if err != nil {
		_ = stream.Close()
		return nil, err
	}
	table.StreamWriter = stream
	table.closer = stream
	return table, nil
}

// OpenFileReadOnly opens an existing table file for reading only.
func OpenFileReadOnly(path string) (*Table, error) {
	stream, err := OpenFileStream(path, ReadOnly)
	if err != nil {
		return nil, err
	}
	table, err := FromReader(stream)
	if err != nil {
		_ = stream.Close()
		return nil, err
	}
	table.closer = stream
	return table, nil
}

// CreateFile creates a new fixed-capacity table file and writes its header.
//
// The file is truncated to header.Capacity() bytes. This reproduces the size
// used by the Kotlin BinaryStreamImplMapped constructor.
func CreateFile(header *Header, path string) (*Table, error) {
	if header == nil {
		return nil, fmt.Errorf("binarytable: nil header")
	}
	stream, err := CreateFileStream(path, header.Capacity(), ReadWrite)
	if err != nil {
		return nil, err
	}
	table, err := NewWithStreams(header, stream, stream)
	if err != nil {
		_ = stream.Close()
		return nil, err
	}
	table.closer = stream
	return table, nil
}

// Close closes the underlying file, when the table was opened with OpenFile,
// OpenFileReadOnly, or CreateFile. It is safe to call Close on a table created
// over an in-memory or user-owned stream; in that case it is a no-op.
func (t *Table) Close() error {
	if t.closer == nil {
		return nil
	}
	err := t.closer.Close()
	t.closer = nil
	return err
}

// Position returns the absolute byte offset of a cell.
func (t *Table) Position(row, column int) (int64, error) {
	info, err := t.column(column)
	if err != nil {
		return 0, err
	}
	if err := t.checkRow(row); err != nil {
		return 0, err
	}
	pos := int64(t.Header.SchemaSize()) + int64(row)*int64(t.Header.MaxRowLength()) + int64(info.StartPos)
	if pos < 0 {
		return 0, ErrOutOfBounds
	}
	return pos, nil
}

// SetValue writes raw bytes into a cell. The bytes must not be longer than the
// column. Bytes shorter than the column leave the remaining bytes untouched,
// matching the Kotlin setValue(ByteArray) behaviour.
func (t *Table) SetValue(row, column int, value []byte) error {
	return t.setValue(row, column, value)
}

// SetValueByName writes raw bytes into a named cell.
func (t *Table) SetValueByName(row int, columnName string, value []byte) error {
	info, err := t.columnByName(columnName)
	if err != nil {
		return err
	}
	return t.setValue(row, info.Index, value)
}

func (t *Table) setValue(row, column int, value []byte) error {
	info, err := t.column(column)
	if err != nil {
		return err
	}
	if t.StreamWriter == nil {
		return ErrReadOnly
	}
	if len(value) > info.Column.Length {
		return fmt.Errorf("%w: column %q holds %d bytes, got %d", ErrOverSize, info.Column.Name, info.Column.Length, len(value))
	}
	pos, err := t.Position(row, column)
	if err != nil {
		return err
	}
	return WriteAtFull(t.StreamWriter, pos, value)
}

// GetValue returns a copy of a complete cell. The returned slice always has the
// configured column length.
func (t *Table) GetValue(row, column int) ([]byte, error) {
	info, err := t.column(column)
	if err != nil {
		return nil, err
	}
	return t.getValue(row, info)
}

// GetValueByName returns a copy of a named cell.
func (t *Table) GetValueByName(row int, columnName string) ([]byte, error) {
	info, err := t.columnByName(columnName)
	if err != nil {
		return nil, err
	}
	return t.getValue(row, info)
}

func (t *Table) getValue(row int, info ColumnInfo) ([]byte, error) {
	if t.StreamReader == nil {
		return nil, ErrNoReader
	}
	pos, err := t.Position(row, info.Index)
	if err != nil {
		return nil, err
	}
	// The Kotlin implementation allocates a zero-filled result and ignores a
	// short read. This matters for growable in-memory streams, whose physical
	// length may end before a partially written last column.
	value := make([]byte, info.Column.Length)
	n, err := t.StreamReader.ReadAt(value, pos)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if n < 0 {
		return nil, ErrCorrupt
	}
	return value, nil
}

// GetValueAsStream returns a read-only stream over one cell.
func (t *Table) GetValueAsStream(row, column int) (Stream, error) {
	info, err := t.column(column)
	if err != nil {
		return nil, err
	}
	return t.getValueAsStream(row, info)
}

// GetValueAsStreamByName returns a read-only stream over one named cell.
func (t *Table) GetValueAsStreamByName(row int, columnName string) (Stream, error) {
	info, err := t.columnByName(columnName)
	if err != nil {
		return nil, err
	}
	return t.getValueAsStream(row, info)
}

func (t *Table) getValueAsStream(row int, info ColumnInfo) (Stream, error) {
	if t.StreamReader == nil {
		return nil, ErrNoReader
	}
	pos, err := t.Position(row, info.Index)
	if err != nil {
		return nil, err
	}
	sub, err := NewSubReader(t.StreamReader, pos, pos+int64(info.Column.Length))
	if err != nil {
		return nil, err
	}
	return AsStream(sub), nil
}

// GetAllRows returns every row byte, starting immediately after the schema.
func (t *Table) GetAllRows() ([]byte, error) {
	if t.StreamReader == nil {
		return nil, ErrNoReader
	}
	start := int64(t.Header.SchemaSize())
	end := t.StreamReader.Size()
	if end <= start {
		return []byte{}, nil
	}
	size := end - start
	if size > int64(maxInt()) {
		return nil, ErrOutOfBounds
	}
	return ReadBytesAt(t.StreamReader, start, int(size))
}

// GetInt reads a 4-byte INT cell.
func (t *Table) GetInt(row, column int) (int32, error) {
	b, err := t.GetValue(row, column)
	if err != nil {
		return 0, err
	}
	return DecodeInt32(b)
}

// GetIntByName reads a named INT cell.
func (t *Table) GetIntByName(row int, columnName string) (int32, error) {
	b, err := t.GetValueByName(row, columnName)
	if err != nil {
		return 0, err
	}
	return DecodeInt32(b)
}

// SetInt writes a 4-byte INT cell.
func (t *Table) SetInt(row, column int, value int32) error {
	return t.SetValue(row, column, EncodeInt32(value))
}

// SetIntByName writes a named INT cell.
func (t *Table) SetIntByName(row int, columnName string, value int32) error {
	return t.SetValueByName(row, columnName, EncodeInt32(value))
}

// GetLong reads an 8-byte LONG cell.
func (t *Table) GetLong(row, column int) (int64, error) {
	b, err := t.GetValue(row, column)
	if err != nil {
		return 0, err
	}
	return DecodeInt64(b)
}

// GetLongByName reads a named LONG cell.
func (t *Table) GetLongByName(row int, columnName string) (int64, error) {
	b, err := t.GetValueByName(row, columnName)
	if err != nil {
		return 0, err
	}
	return DecodeInt64(b)
}

// SetLong writes an 8-byte LONG cell.
func (t *Table) SetLong(row, column int, value int64) error {
	return t.SetValue(row, column, EncodeInt64(value))
}

// SetLongByName writes a named LONG cell.
func (t *Table) SetLongByName(row int, columnName string, value int64) error {
	return t.SetValueByName(row, columnName, EncodeInt64(value))
}

// GetDouble reads an 8-byte DOUBLE cell.
func (t *Table) GetDouble(row, column int) (float64, error) {
	b, err := t.GetValue(row, column)
	if err != nil {
		return 0, err
	}
	return DecodeFloat64(b)
}

// GetDoubleByName reads a named DOUBLE cell.
func (t *Table) GetDoubleByName(row int, columnName string) (float64, error) {
	b, err := t.GetValueByName(row, columnName)
	if err != nil {
		return 0, err
	}
	return DecodeFloat64(b)
}

// SetDouble writes an 8-byte DOUBLE cell.
func (t *Table) SetDouble(row, column int, value float64) error {
	return t.SetValue(row, column, EncodeFloat64(value))
}

// SetDoubleByName writes a named DOUBLE cell.
func (t *Table) SetDoubleByName(row int, columnName string, value float64) error {
	return t.SetValueByName(row, columnName, EncodeFloat64(value))
}

// GetString reads a STRING cell. Zero bytes are removed, matching the Kotlin
// getString extension.
func (t *Table) GetString(row, column int) (string, error) {
	b, err := t.GetValue(row, column)
	if err != nil {
		return "", err
	}
	return string(stripZeroBytes(b)), nil
}

// GetStringByName reads a named STRING cell.
func (t *Table) GetStringByName(row int, columnName string) (string, error) {
	b, err := t.GetValueByName(row, columnName)
	if err != nil {
		return "", err
	}
	return string(stripZeroBytes(b)), nil
}

// SetString writes a UTF-8 STRING cell and zero-fills the rest of the column.
// Zero-filling avoids stale bytes when a shorter value replaces a longer one.
func (t *Table) SetString(row, column int, value string) error {
	info, err := t.column(column)
	if err != nil {
		return err
	}
	if t.StreamWriter == nil {
		return ErrReadOnly
	}
	if len(value) > info.Column.Length {
		return fmt.Errorf("%w: column %q holds %d bytes, string needs %d", ErrOverSize, info.Column.Name, info.Column.Length, len(value))
	}
	buf := make([]byte, info.Column.Length)
	copy(buf, value)
	pos, err := t.Position(row, column)
	if err != nil {
		return err
	}
	return WriteAtFull(t.StreamWriter, pos, buf)
}

// SetStringByName writes a named STRING cell.
func (t *Table) SetStringByName(row int, columnName, value string) error {
	info, err := t.columnByName(columnName)
	if err != nil {
		return err
	}
	return t.SetString(row, info.Index, value)
}

func (t *Table) writeHeader() error {
	header := t.Header
	if header.MaxRows() > math.MaxInt32 {
		return fmt.Errorf("%w: max rows do not fit in int32", ErrOutOfBounds)
	}
	schema := header.SchemaBytes()
	if len(schema)+8 > math.MaxInt32 {
		return fmt.Errorf("%w: schema does not fit in int32", ErrOutOfBounds)
	}
	var prefix [8]byte
	putInt32(prefix[:4], int32(header.MaxRows()))
	putInt32(prefix[4:], int32(len(schema)))
	if err := WriteAtFull(t.StreamWriter, 0, prefix[:]); err != nil {
		return err
	}
	if len(schema) == 0 {
		return nil
	}
	return WriteAtFull(t.StreamWriter, 8, schema)
}

func (t *Table) checkRow(row int) error {
	if row < 0 || row >= t.Header.MaxRows() {
		return fmt.Errorf("%w: row %d, max rows %d", ErrOutOfBounds, row, t.Header.MaxRows())
	}
	return nil
}

func (t *Table) column(column int) (ColumnInfo, error) {
	info, ok := t.Header.ColumnInfo(column)
	if !ok {
		return ColumnInfo{}, fmt.Errorf("%w: column index %d", ErrOutOfBounds, column)
	}
	return info, nil
}

func (t *Table) columnByName(name string) (ColumnInfo, error) {
	info, ok := t.Header.ColumnInfoByName(name)
	if !ok {
		return ColumnInfo{}, fmt.Errorf("binarytable: unknown column %q", name)
	}
	return info, nil
}

func stripZeroBytes(b []byte) []byte {
	first := -1
	for i, v := range b {
		if v != 0 {
			first = i
			break
		}
	}
	if first < 0 {
		return b[:0]
	}
	hasZero := false
	for _, v := range b[first:] {
		if v == 0 {
			hasZero = true
			break
		}
	}
	if !hasZero {
		return b[first:]
	}
	out := make([]byte, 0, len(b)-first)
	for _, v := range b[first:] {
		if v != 0 {
			out = append(out, v)
		}
	}
	return out
}
