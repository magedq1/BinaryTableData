package binarytable

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func testHeader(t *testing.T, rows int) *Header {
	t.Helper()
	header, err := NewHeader(rows, []Column{
		IntColumn("_id"),
		StringColumn("first_name", 16),
		StringColumn("last_name", 16),
		LongColumn("created"),
		DoubleColumn("score"),
		RawColumn("payload", 8),
	})
	if err != nil {
		t.Fatalf("NewHeader: %v", err)
	}
	return header
}

func TestHeaderLayout(t *testing.T) {
	header := testHeader(t, 10)
	wantSchema := "_id/4/INT\nfirst_name/16/STRING\nlast_name/16/STRING\ncreated/8/LONG\nscore/8/DOUBLE\npayload/8/RAW\n"
	if got := string(header.SchemaBytes()); got != wantSchema {
		t.Fatalf("schema = %q, want %q", got, wantSchema)
	}
	if got, want := header.MaxRowLength(), 4+16+16+8+8+8; got != want {
		t.Fatalf("row length = %d, want %d", got, want)
	}
	if got, want := header.SchemaSize(), 8+len(wantSchema); got != want {
		t.Fatalf("schema size = %d, want %d", got, want)
	}
	if got, want := header.Capacity(), int64(header.SchemaSize()+10*header.MaxRowLength()); got != want {
		t.Fatalf("capacity = %d, want %d", got, want)
	}
	info, ok := header.ColumnInfoByName("score")
	if !ok {
		t.Fatal("score column not found")
	}
	if info.Index != 4 || info.StartPos != 4+16+16+8 {
		t.Fatalf("score info = %#v", info)
	}
}

func TestHeaderValidation(t *testing.T) {
	if _, err := NewHeader(1, []Column{IntColumn("a"), IntColumn("a")}); err == nil {
		t.Fatal("duplicate names must fail")
	}
	if _, err := NewHeader(1, []Column{StringColumn("bad/name", 4)}); err == nil {
		t.Fatal("reserved name character must fail")
	}
	if _, err := NewHeader(1, []Column{IntColumn("a"), RawColumn("zero", 0)}); err == nil {
		t.Fatal("zero-length column must fail")
	}
	if _, err := NewHeader(-1, nil); err == nil {
		t.Fatal("negative row count must fail")
	}
}

func TestTableRoundTrip(t *testing.T) {
	header := testHeader(t, 32)
	stream := NewAutoSizeStream(64)
	table, err := New(header, stream)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := table.SetInt(0, 0, 7); err != nil {
		t.Fatal(err)
	}
	if err := table.SetString(0, 1, "Alice"); err != nil {
		t.Fatal(err)
	}
	if err := table.SetString(0, 2, "Saleh"); err != nil {
		t.Fatal(err)
	}
	if err := table.SetLong(0, 3, 1_700_000_000_000); err != nil {
		t.Fatal(err)
	}
	if err := table.SetDouble(0, 4, 12.5); err != nil {
		t.Fatal(err)
	}
	if err := table.SetValueByName(0, "payload", []byte{1, 2, 3}); err != nil {
		t.Fatal(err)
	}

	if got, err := table.GetIntByName(0, "_id"); err != nil || got != 7 {
		t.Fatalf("GetInt = %d, %v", got, err)
	}
	if got, err := table.GetString(0, 1); err != nil || got != "Alice" {
		t.Fatalf("GetString = %q, %v", got, err)
	}
	if got, err := table.GetStringByName(0, "last_name"); err != nil || got != "Saleh" {
		t.Fatalf("GetString = %q, %v", got, err)
	}
	if got, err := table.GetLong(0, 3); err != nil || got != 1_700_000_000_000 {
		t.Fatalf("GetLong = %d, %v", got, err)
	}
	if got, err := table.GetDouble(0, 4); err != nil || got != 12.5 {
		t.Fatalf("GetDouble = %v, %v", got, err)
	}
	raw, err := table.GetValueByName(0, "payload")
	if err != nil {
		t.Fatal(err)
	}
	if want := []byte{1, 2, 3, 0, 0, 0, 0, 0}; !bytes.Equal(raw, want) {
		t.Fatalf("raw = %v, want %v", raw, want)
	}

	cell, err := table.GetValueAsStreamByName(0, "first_name")
	if err != nil {
		t.Fatal(err)
	}
	cellBytes, err := ReadAll(cell)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(stripZeroBytes(cellBytes)); got != "Alice" {
		t.Fatalf("cell stream = %q", got)
	}
}

func TestTableHeaderPrefix(t *testing.T) {
	header := testHeader(t, 3)
	stream := NewAutoSizeStream(64)
	if _, err := New(header, stream); err != nil {
		t.Fatal(err)
	}
	got := stream.ToBytes()
	if len(got) < 8 {
		t.Fatalf("stream too short: %d", len(got))
	}
	rows, _ := DecodeInt32(got[:4])
	schemaSize, _ := DecodeInt32(got[4:8])
	if rows != 3 {
		t.Fatalf("rows = %d", rows)
	}
	if int(schemaSize) != header.SchemaSize()-8 {
		t.Fatalf("schema size = %d, want %d", schemaSize, header.SchemaSize()-8)
	}
	if !bytes.Equal(got[8:8+schemaSize], header.SchemaBytes()) {
		t.Fatal("schema bytes do not match")
	}
}

func TestTablePositions(t *testing.T) {
	header := testHeader(t, 8)
	stream := NewAutoSizeStream(64)
	table, err := New(header, stream)
	if err != nil {
		t.Fatal(err)
	}
	pos, err := table.Position(2, 2)
	if err != nil {
		t.Fatal(err)
	}
	want := int64(header.SchemaSize()) + 2*int64(header.MaxRowLength()) + int64(4+16)
	if pos != want {
		t.Fatalf("position = %d, want %d", pos, want)
	}
	if _, err := table.Position(-1, 0); !errors.Is(err, ErrOutOfBounds) {
		t.Fatalf("negative row error = %v", err)
	}
	if _, err := table.Position(0, 99); !errors.Is(err, ErrOutOfBounds) {
		t.Fatalf("bad column error = %v", err)
	}
}

func TestTableStringOverwriteClearsTail(t *testing.T) {
	header, err := NewHeader(2, []Column{StringColumn("name", 16)})
	if err != nil {
		t.Fatal(err)
	}
	table, err := New(header, NewAutoSizeStream(32))
	if err != nil {
		t.Fatal(err)
	}
	if err := table.SetString(0, 0, "a-very-long-name"); err != nil {
		t.Fatal(err)
	}
	if err := table.SetString(0, 0, "ok"); err != nil {
		t.Fatal(err)
	}
	if got, err := table.GetString(0, 0); err != nil || got != "ok" {
		t.Fatalf("string = %q, %v", got, err)
	}
}

func TestTableErrors(t *testing.T) {
	header, err := NewHeader(2, []Column{StringColumn("name", 4), IntColumn("id")})
	if err != nil {
		t.Fatal(err)
	}
	table, err := New(header, NewAutoSizeStream(32))
	if err != nil {
		t.Fatal(err)
	}
	if err := table.SetString(0, 0, "too long"); !errors.Is(err, ErrOverSize) {
		t.Fatalf("oversize error = %v", err)
	}
	if err := table.SetValue(0, 0, make([]byte, 5)); !errors.Is(err, ErrOverSize) {
		t.Fatalf("raw oversize error = %v", err)
	}
	if err := table.SetInt(5, 1, 1); !errors.Is(err, ErrOutOfBounds) {
		t.Fatalf("row error = %v", err)
	}
	if err := table.SetInt(0, 9, 1); !errors.Is(err, ErrOutOfBounds) {
		t.Fatalf("column error = %v", err)
	}
}

func TestTableFromBytesIsReadOnly(t *testing.T) {
	header := testHeader(t, 4)
	stream := NewAutoSizeStream(32)
	table, err := New(header, stream)
	if err != nil {
		t.Fatal(err)
	}
	if err := table.SetInt(1, 0, 42); err != nil {
		t.Fatal(err)
	}
	data := stream.ToBytes()
	got, err := FromBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	if id, err := got.GetInt(1, 0); err != nil || id != 42 {
		t.Fatalf("id = %d, %v", id, err)
	}
	if err := got.SetInt(1, 0, 9); !errors.Is(err, ErrReadOnly) {
		t.Fatalf("write error = %v", err)
	}
	if _, err := FromBytes([]byte{1, 2, 3}); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("short data error = %v", err)
	}
}

func TestTableSearch(t *testing.T) {
	header, err := NewHeader(10, []Column{IntColumn("id"), LongColumn("n"), DoubleColumn("d"), StringColumn("s", 8)})
	if err != nil {
		t.Fatal(err)
	}
	table, err := New(header, NewAutoSizeStream(64))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		if err := table.SetInt(i, 0, int32(i*10)); err != nil {
			t.Fatal(err)
		}
		if err := table.SetLong(i, 1, int64(i)*100); err != nil {
			t.Fatal(err)
		}
		if err := table.SetDouble(i, 2, float64(i)+0.5); err != nil {
			t.Fatal(err)
		}
		if err := table.SetString(i, 3, "row"+string(rune('0'+i))); err != nil {
			t.Fatal(err)
		}
	}
	if row, err := table.IndexOfIntByName(30, "id", 0); err != nil || row != 3 {
		t.Fatalf("int index = %d, %v", row, err)
	}
	if row, err := table.IndexOfLong(200, 1, 0); err != nil || row != 2 {
		t.Fatalf("long index = %d, %v", row, err)
	}
	if row, err := table.IndexOfDouble(4.5, 2, 0); err != nil || row != 4 {
		t.Fatalf("double index = %d, %v", row, err)
	}
	if row, err := table.IndexOfString("row2", 3, 0); err != nil || row != 2 {
		t.Fatalf("string index = %d, %v", row, err)
	}
	if row, err := table.IndexOfInt(999, 0, 0); err != nil || row != -1 {
		t.Fatalf("missing index = %d, %v", row, err)
	}
}

func TestTableFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.db")
	header := testHeader(t, 16)
	table, err := CreateFile(header, path)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != header.Capacity() {
		t.Fatalf("file size = %d, want %d", info.Size(), header.Capacity())
	}
	if err := table.SetInt(3, 0, 99); err != nil {
		t.Fatal(err)
	}
	if err := table.SetStringByName(3, "first_name", "Maged"); err != nil {
		t.Fatal(err)
	}
	if err := table.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := OpenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if id, err := reopened.GetInt(3, 0); err != nil || id != 99 {
		t.Fatalf("id = %d, %v", id, err)
	}
	if name, err := reopened.GetString(3, 1); err != nil || name != "Maged" {
		t.Fatalf("name = %q, %v", name, err)
	}
	all, err := reopened.GetAllRows()
	if err != nil {
		t.Fatal(err)
	}
	if want := header.MaxRows() * header.MaxRowLength(); len(all) != want {
		t.Fatalf("all rows length = %d, want %d", len(all), want)
	}

	readOnly, err := OpenFileReadOnly(path)
	if err != nil {
		t.Fatal(err)
	}
	defer readOnly.Close()
	if id, err := readOnly.GetInt(3, 0); err != nil || id != 99 {
		t.Fatalf("read-only id = %d, %v", id, err)
	}
	if err := readOnly.SetInt(3, 0, 1); !errors.Is(err, ErrReadOnly) {
		t.Fatalf("read-only write = %v", err)
	}
}

func TestColumnParsing(t *testing.T) {
	columns, err := parseColumns([]byte("_id/4/INT\r\nname/8/STRING\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(columns) != 2 || columns[0].Type != TypeInt || columns[1].Name != "name" {
		t.Fatalf("columns = %#v", columns)
	}
	if _, err := parseColumns([]byte("bad")); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("bad schema error = %v", err)
	}
	if _, err := parseColumns([]byte("x/4/NOPE")); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("bad type error = %v", err)
	}
}
