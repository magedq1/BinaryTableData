package binarytable

import (
	"fmt"
	"math"
	"strings"
)

// ColumnInfo describes the position of a column inside a row.
type ColumnInfo struct {
	Index    int
	Column   Column
	StartPos int
}

// Header is an immutable table schema.
//
// The schema is stored at the beginning of every table file:
//
//	offset 0: int32  max rows
//	offset 4: int32  schema byte length
//	offset 8: UTF-8  "name/length/TYPE\n" for every column
//
// Header values must be created with NewHeader. Always use the accessor methods
// rather than building the derived fields yourself.
type Header struct {
	maxRows      int
	columns      []Column
	schema       []byte
	schemaSize   int
	maxRowLength int
	byIndex      []ColumnInfo
	byName       map[string]ColumnInfo
}

// NewHeader validates columns and computes the row layout.
func NewHeader(maxRows int, columns []Column) (*Header, error) {
	if maxRows < 0 || int64(maxRows) > math.MaxInt32 {
		return nil, fmt.Errorf("%w: max rows must fit in int32", ErrOutOfBounds)
	}
	cols := make([]Column, len(columns))
	copy(cols, columns)

	byIndex := make([]ColumnInfo, len(cols))
	byName := make(map[string]ColumnInfo, len(cols))
	pos := 0
	for i, c := range cols {
		if c.Name == "" {
			return nil, fmt.Errorf("binarytable: column %d has an empty name", i)
		}
		if strings.ContainsAny(c.Name, "/\n\r") {
			return nil, fmt.Errorf("binarytable: column name %q contains a reserved character", c.Name)
		}
		if c.Length <= 0 {
			return nil, fmt.Errorf("binarytable: column %q must have a positive length", c.Name)
		}
		if c.Type == TypeInt && c.Length < 4 {
			return nil, fmt.Errorf("binarytable: INT column %q needs at least 4 bytes", c.Name)
		}
		if (c.Type == TypeLong || c.Type == TypeDouble) && c.Length < 8 {
			return nil, fmt.Errorf("binarytable: column %q needs at least 8 bytes", c.Name)
		}
		if _, exists := byName[c.Name]; exists {
			return nil, fmt.Errorf("binarytable: duplicate column name %q", c.Name)
		}
		if pos > maxInt()-c.Length {
			return nil, fmt.Errorf("%w: row length overflows int", ErrOutOfBounds)
		}
		info := ColumnInfo{Index: i, Column: c, StartPos: pos}
		byIndex[i] = info
		byName[c.Name] = info
		pos += c.Length
	}

	schema := encodeColumns(cols)
	if len(schema) > math.MaxInt32-8 {
		return nil, fmt.Errorf("%w: schema is too large", ErrOutOfBounds)
	}
	schemaSize := len(schema) + 8
	if pos > 0 && maxRows > 0 && int64(pos)*int64(maxRows) > math.MaxInt64-int64(schemaSize) {
		return nil, fmt.Errorf("%w: table capacity overflows int64", ErrOutOfBounds)
	}

	return &Header{
		maxRows:      maxRows,
		columns:      cols,
		schema:       schema,
		schemaSize:   schemaSize,
		maxRowLength: pos,
		byIndex:      byIndex,
		byName:       byName,
	}, nil
}

// MaxRows returns the fixed row capacity.
func (h *Header) MaxRows() int { return h.maxRows }

// ColumnCount returns the number of columns.
func (h *Header) ColumnCount() int { return len(h.columns) }

// Columns returns a copy of the schema columns.
func (h *Header) Columns() []Column {
	out := make([]Column, len(h.columns))
	copy(out, h.columns)
	return out
}

// SchemaSize returns the number of bytes occupied by the header (including the
// two leading int32 fields).
func (h *Header) SchemaSize() int { return h.schemaSize }

// MaxRowLength returns the sum of all column lengths.
func (h *Header) MaxRowLength() int { return h.maxRowLength }

// SchemaBytes returns a copy of the encoded schema text.
func (h *Header) SchemaBytes() []byte {
	out := make([]byte, len(h.schema))
	copy(out, h.schema)
	return out
}

// ColumnInfo returns layout information for a zero-based column index.
func (h *Header) ColumnInfo(index int) (ColumnInfo, bool) {
	if index < 0 || index >= len(h.byIndex) {
		return ColumnInfo{}, false
	}
	return h.byIndex[index], true
}

// ColumnInfoByName returns layout information for a column name.
func (h *Header) ColumnInfoByName(name string) (ColumnInfo, bool) {
	info, ok := h.byName[name]
	return info, ok
}

// ColumnIndex returns the zero-based index of a column name.
func (h *Header) ColumnIndex(name string) (int, bool) {
	info, ok := h.byName[name]
	if !ok {
		return 0, false
	}
	return info.Index, true
}

// Capacity returns the total table size in bytes: schema size plus all rows.
func (h *Header) Capacity() int64 {
	return int64(h.schemaSize) + int64(h.maxRows)*int64(h.maxRowLength)
}
