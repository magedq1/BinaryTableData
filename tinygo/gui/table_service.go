package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/magedq1/BinaryTableData/tinygo/binarytable"
)

// ColumnDTO describes one column of an open table.
type ColumnDTO struct {
	Index    int    `json:"index"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Length   int    `json:"length"`
	StartPos int    `json:"startPos"`
}

// TableInfo describes an open table.
type TableInfo struct {
	Path       string      `json:"path"`
	ReadOnly   bool        `json:"readOnly"`
	Compressed bool        `json:"compressed"`
	Rows       int         `json:"rows"`
	RowSize    int         `json:"rowSize"`
	SchemaSize int         `json:"schemaSize"`
	Capacity   int64       `json:"capacity"`
	FileSize   int64       `json:"fileSize"`
	Schema     string      `json:"schema"`
	Columns    []ColumnDTO `json:"columns"`
}

// CellDTO is one editable cell.
type CellDTO struct {
	Column int    `json:"column"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Length int    `json:"length"`
	Value  string `json:"value"`
	RawHex string `json:"rawHex"`
	Empty  bool   `json:"empty"`
}

// RowDTO is one table row.
type RowDTO struct {
	Index int       `json:"index"`
	Cells []CellDTO `json:"cells"`
}

// TablePage is a page of rows.
type TablePage struct {
	Start int      `json:"start"`
	Total int      `json:"total"`
	Limit int      `json:"limit"`
	Rows  []RowDTO `json:"rows"`
}

// ColumnSpec is used when creating a table.
type ColumnSpec struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Length int    `json:"length"`
}

// CreateTableRequest contains all parameters for table creation.
type CreateTableRequest struct {
	Path      string       `json:"path"`
	Rows      int          `json:"rows"`
	Columns   []ColumnSpec `json:"columns"`
	Overwrite bool         `json:"overwrite"`
}

type tableSession struct {
	mu         sync.Mutex
	path       string
	readOnly   bool
	compressed bool
	table      *binarytable.Table
	source     io.Closer
}

func (s *tableSession) close() error {
	var first error
	if s.table != nil {
		first = s.table.Close()
	}
	if s.source != nil {
		if err := s.source.Close(); first == nil {
			first = err
		}
		s.source = nil
	}
	return first
}

// OpenTable opens an existing table read/write.
func (a *App) OpenTable(path string) (*TableInfo, error) {
	return a.openTable(path, false)
}

// OpenTableReadOnly opens an existing table in read-only mode.
func (a *App) OpenTableReadOnly(path string) (*TableInfo, error) {
	return a.openTable(path, true)
}

func (a *App) openTable(path string, readOnly bool) (*TableInfo, error) {
	key, err := normalizePath(path)
	if err != nil {
		return nil, err
	}

	a.mu.Lock()
	existing := a.tables[key]
	if existing != nil && !existing.compressed && (existing.readOnly || !readOnly) {
		a.mu.Unlock()
		return existing.info()
	}
	if existing != nil {
		delete(a.tables, key)
	}
	a.mu.Unlock()

	if existing != nil {
		existing.mu.Lock()
		_ = existing.close()
		existing.mu.Unlock()
	}

	var table *binarytable.Table
	if readOnly {
		table, err = binarytable.OpenFileReadOnly(key)
	} else {
		table, err = binarytable.OpenFile(key)
	}
	if err != nil {
		return nil, err
	}
	session := &tableSession{path: key, readOnly: readOnly, table: table}
	a.mu.Lock()
	a.tables[key] = session
	a.mu.Unlock()
	return session.info()
}

// OpenCompressedTable reads a table through a DeflateBinaryStreamReader
// equivalent. The resulting table is read-only.
func (a *App) OpenCompressedTable(path string) (*TableInfo, error) {
	key, err := normalizePath(path)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	existing := a.tables[key]
	if existing != nil && existing.compressed {
		a.mu.Unlock()
		return existing.info()
	}
	if existing != nil {
		delete(a.tables, key)
	}
	a.mu.Unlock()
	if existing != nil {
		existing.mu.Lock()
		_ = existing.close()
		existing.mu.Unlock()
	}

	source, err := binarytable.OpenFileStream(key, binarytable.ReadOnly)
	if err != nil {
		return nil, err
	}
	reader, err := binarytable.NewDeflateReader(source, 8)
	if err != nil {
		_ = source.Close()
		return nil, err
	}
	table, err := binarytable.FromReader(reader)
	if err != nil {
		_ = source.Close()
		return nil, err
	}
	session := &tableSession{path: key, readOnly: true, compressed: true, table: table, source: source}
	a.mu.Lock()
	a.tables[key] = session
	a.mu.Unlock()
	return session.info()
}

// CloseTable closes one open table.
func (a *App) CloseTable(path string) error {
	key, err := normalizePath(path)
	if err != nil {
		return err
	}
	a.mu.Lock()
	session := a.tables[key]
	delete(a.tables, key)
	a.mu.Unlock()
	if session == nil {
		return nil
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	return session.close()
}

// CreateTable creates a new fixed-capacity table file and opens it.
func (a *App) CreateTable(request CreateTableRequest) (*TableInfo, error) {
	key, err := normalizePath(request.Path)
	if err != nil {
		return nil, err
	}
	if request.Rows < 0 {
		return nil, fmt.Errorf("row capacity must be >= 0")
	}
	if len(request.Columns) == 0 {
		return nil, fmt.Errorf("add at least one column")
	}
	columns := make([]binarytable.Column, 0, len(request.Columns))
	for _, spec := range request.Columns {
		column, err := columnFromSpec(spec)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	header, err := binarytable.NewHeader(request.Rows, columns)
	if err != nil {
		return nil, err
	}
	if !request.Overwrite {
		if _, err := os.Stat(key); err == nil {
			return nil, fmt.Errorf("%s already exists", key)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(key), 0o755); err != nil {
		return nil, err
	}

	a.mu.Lock()
	if existing := a.tables[key]; existing != nil {
		delete(a.tables, key)
		existing.mu.Lock()
		_ = existing.close()
		existing.mu.Unlock()
	}
	if existing := a.lists[key]; existing != nil {
		delete(a.lists, key)
		existing.mu.Lock()
		_ = existing.close()
		existing.mu.Unlock()
	}
	a.mu.Unlock()

	table, err := binarytable.CreateFile(header, key)
	if err != nil {
		return nil, err
	}
	session := &tableSession{path: key, table: table}
	a.mu.Lock()
	a.tables[key] = session
	a.mu.Unlock()
	return session.info()
}

func columnFromSpec(spec ColumnSpec) (binarytable.Column, error) {
	name := strings.TrimSpace(spec.Name)
	if name == "" {
		return binarytable.Column{}, fmt.Errorf("column name is required")
	}
	kind := strings.ToLower(strings.TrimSpace(spec.Type))
	length := spec.Length
	switch kind {
	case "int", "int32", "i32":
		return binarytable.IntColumn(name), nil
	case "long", "int64", "i64":
		return binarytable.LongColumn(name), nil
	case "double", "float64", "f64":
		return binarytable.DoubleColumn(name), nil
	case "string", "str", "text":
		if length <= 0 {
			length = 64
		}
		return binarytable.StringColumn(name, length), nil
	case "raw", "bytes", "blob":
		if length <= 0 {
			return binarytable.Column{}, fmt.Errorf("RAW column %q needs a length", name)
		}
		return binarytable.RawColumn(name, length), nil
	default:
		return binarytable.Column{}, fmt.Errorf("unknown column type %q", spec.Type)
	}
}

func (s *tableSession) info() (*TableInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stat, err := os.Stat(s.path)
	if err != nil {
		return nil, err
	}
	header := s.table.Header
	columns := make([]ColumnDTO, 0, header.ColumnCount())
	for _, column := range header.Columns() {
		info, _ := header.ColumnInfoByName(column.Name)
		columns = append(columns, ColumnDTO{
			Index:    info.Index,
			Name:     column.Name,
			Type:     column.Type.String(),
			Length:   column.Length,
			StartPos: info.StartPos,
		})
	}
	return &TableInfo{
		Path:       s.path,
		ReadOnly:   s.readOnly,
		Compressed: s.compressed,
		Rows:       header.MaxRows(),
		RowSize:    header.MaxRowLength(),
		SchemaSize: header.SchemaSize(),
		Capacity:   header.Capacity(),
		FileSize:   stat.Size(),
		Schema:     string(header.SchemaBytes()),
		Columns:    columns,
	}, nil
}

func (a *App) getTable(path string) (*tableSession, error) {
	key, err := normalizePath(path)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	session := a.tables[key]
	a.mu.Unlock()
	if session == nil {
		return nil, fmt.Errorf("table is not open: %s", key)
	}
	return session, nil
}

// GetTableRows returns a page of rows.
func (a *App) GetTableRows(path string, start, limit int) (*TablePage, error) {
	session, err := a.getTable(path)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	if start < 0 {
		start = 0
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	total := session.table.Header.MaxRows()
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	rows := make([]RowDTO, 0, end-start)
	for row := start; row < end; row++ {
		rowValue, err := rowFor(session.table, row)
		if err != nil {
			return nil, err
		}
		rows = append(rows, rowValue)
	}
	return &TablePage{Start: start, Total: total, Limit: limit, Rows: rows}, nil
}

func rowFor(table *binarytable.Table, row int) (RowDTO, error) {
	header := table.Header
	cells := make([]CellDTO, 0, header.ColumnCount())
	for _, column := range header.Columns() {
		info, _ := header.ColumnInfoByName(column.Name)
		cell, err := cellFor(table, row, info)
		if err != nil {
			return RowDTO{}, err
		}
		cells = append(cells, cell)
	}
	return RowDTO{Index: row, Cells: cells}, nil
}

func cellFor(table *binarytable.Table, row int, info binarytable.ColumnInfo) (CellDTO, error) {
	raw, err := table.GetValue(row, info.Index)
	if err != nil {
		return CellDTO{}, err
	}
	empty := true
	for _, b := range raw {
		if b != 0 {
			empty = false
			break
		}
	}
	cell := CellDTO{
		Column: info.Index,
		Name:   info.Column.Name,
		Type:   info.Column.Type.String(),
		Length: info.Column.Length,
		RawHex: hex.EncodeToString(raw),
		Empty:  empty,
	}
	switch info.Column.Type {
	case binarytable.TypeInt:
		value, err := binarytable.DecodeInt32(raw)
		if err != nil {
			return CellDTO{}, err
		}
		cell.Value = strconv.FormatInt(int64(value), 10)
	case binarytable.TypeLong:
		value, err := binarytable.DecodeInt64(raw)
		if err != nil {
			return CellDTO{}, err
		}
		cell.Value = strconv.FormatInt(value, 10)
	case binarytable.TypeDouble:
		value, err := binarytable.DecodeFloat64(raw)
		if err != nil {
			return CellDTO{}, err
		}
		cell.Value = strconv.FormatFloat(value, 'g', -1, 64)
	case binarytable.TypeString:
		cell.Value = string(stripZeroBytes(raw))
	case binarytable.TypeRaw:
		cell.Value = cell.RawHex
	}
	return cell, nil
}

func stripZeroBytes(data []byte) []byte {
	out := make([]byte, 0, len(data))
	for _, b := range data {
		if b != 0 {
			out = append(out, b)
		}
	}
	return out
}

// SetTableCell writes one cell. format is "typed", "hex", or "text".
func (a *App) SetTableCell(path string, row, column int, value, format string) (*CellDTO, error) {
	session, err := a.getTable(path)
	if err != nil {
		return nil, err
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.readOnly {
		return nil, binarytable.ErrReadOnly
	}
	info, ok := session.table.Header.ColumnInfo(column)
	if !ok {
		return nil, binarytable.ErrOutOfBounds
	}
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "typed"
	}
	switch format {
	case "hex":
		text := strings.ReplaceAll(strings.TrimSpace(value), " ", "")
		text = strings.TrimPrefix(text, "0x")
		data, err := hex.DecodeString(text)
		if err != nil {
			return nil, fmt.Errorf("invalid hexadecimal value: %w", err)
		}
		if err := writePadded(session.table, row, info, data); err != nil {
			return nil, err
		}
	case "text":
		if err := writePadded(session.table, row, info, []byte(value)); err != nil {
			return nil, err
		}
	default:
		switch info.Column.Type {
		case binarytable.TypeInt:
			parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid INT value %q", value)
			}
			if err := writePadded(session.table, row, info, binarytable.EncodeInt32(int32(parsed))); err != nil {
				return nil, err
			}
		case binarytable.TypeLong:
			parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid LONG value %q", value)
			}
			if err := writePadded(session.table, row, info, binarytable.EncodeInt64(parsed)); err != nil {
				return nil, err
			}
		case binarytable.TypeDouble:
			parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid DOUBLE value %q", value)
			}
			if err := writePadded(session.table, row, info, binarytable.EncodeFloat64(parsed)); err != nil {
				return nil, err
			}
		case binarytable.TypeString:
			if err := session.table.SetString(row, column, value); err != nil {
				return nil, err
			}
		case binarytable.TypeRaw:
			if err := writePadded(session.table, row, info, []byte(value)); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported column type %s", info.Column.Type)
		}
	}
	cell, err := cellFor(session.table, row, info)
	if err != nil {
		return nil, err
	}
	return &cell, nil
}

func writePadded(table *binarytable.Table, row int, info binarytable.ColumnInfo, data []byte) error {
	if len(data) > info.Column.Length {
		return fmt.Errorf("%w: column %q holds %d bytes, got %d", binarytable.ErrOverSize, info.Column.Name, info.Column.Length, len(data))
	}
	buf := make([]byte, info.Column.Length)
	copy(buf, data)
	return table.SetValue(row, info.Index, buf)
}

// SetTableCellFromFile embeds a file into one cell.
func (a *App) SetTableCellFromFile(path string, row, column int, sourcePath string) (*CellDTO, error) {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, err
	}
	session, err := a.getTable(path)
	if err != nil {
		return nil, err
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.readOnly {
		return nil, binarytable.ErrReadOnly
	}
	info, ok := session.table.Header.ColumnInfo(column)
	if !ok {
		return nil, binarytable.ErrOutOfBounds
	}
	if err := writePadded(session.table, row, info, data); err != nil {
		return nil, err
	}
	cell, err := cellFor(session.table, row, info)
	if err != nil {
		return nil, err
	}
	return &cell, nil
}

// ClearTableRow zero-fills every cell of a row.
func (a *App) ClearTableRow(path string, row int) error {
	session, err := a.getTable(path)
	if err != nil {
		return err
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.readOnly {
		return binarytable.ErrReadOnly
	}
	for _, column := range session.table.Header.Columns() {
		info, _ := session.table.Header.ColumnInfoByName(column.Name)
		if err := session.table.SetValue(row, info.Index, make([]byte, column.Length)); err != nil {
			return err
		}
	}
	return nil
}

// SearchTableRows returns matching row indices.
func (a *App) SearchTableRows(path string, column int, value, format string, limit int) ([]int, error) {
	session, err := a.getTable(path)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	info, ok := session.table.Header.ColumnInfo(column)
	if !ok {
		return nil, binarytable.ErrOutOfBounds
	}
	format = strings.ToLower(strings.TrimSpace(format))
	var rawTarget []byte
	if format == "hex" || info.Column.Type == binarytable.TypeRaw {
		if format == "hex" {
			text := strings.ReplaceAll(strings.TrimSpace(value), " ", "")
			text = strings.TrimPrefix(text, "0x")
			rawTarget, err = hex.DecodeString(text)
			if err != nil {
				return nil, fmt.Errorf("invalid hexadecimal value: %w", err)
			}
		} else {
			rawTarget = []byte(value)
		}
	}
	hits := make([]int, 0, min(limit, 32))
	if rawTarget != nil {
		total := session.table.Header.MaxRows()
		for row := 0; row < total && len(hits) < limit; row++ {
			raw, err := session.table.GetValue(row, info.Index)
			if err != nil {
				return nil, err
			}
			if bytes.Equal(raw, rawTarget) {
				hits = append(hits, row)
			}
		}
		return hits, nil
	}
	start := 0
	for len(hits) < limit {
		var row int
		switch info.Column.Type {
		case binarytable.TypeInt:
			parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid INT value %q", value)
			}
			row, err = session.table.IndexOfInt(int32(parsed), info.Index, start)
			if err != nil {
				return nil, err
			}
		case binarytable.TypeLong:
			parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid LONG value %q", value)
			}
			row, err = session.table.IndexOfLong(parsed, info.Index, start)
			if err != nil {
				return nil, err
			}
		case binarytable.TypeDouble:
			parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid DOUBLE value %q", value)
			}
			row, err = session.table.IndexOfDouble(parsed, info.Index, start)
			if err != nil {
				return nil, err
			}
		case binarytable.TypeString:
			row, err = session.table.IndexOfString(value, info.Index, start)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported column type %s", info.Column.Type)
		}
		if row < 0 {
			break
		}
		hits = append(hits, row)
		start = row + 1
	}
	return hits, nil
}
