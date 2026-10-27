package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/magedq1/BinaryTableData/tinygo/binarytable"
)

func TestTableWorkflow(t *testing.T) {
	app := NewApp()
	defer app.CloseAll()
	dir := t.TempDir()
	path := filepath.Join(dir, "users.db")

	info, err := app.CreateTable(CreateTableRequest{
		Path: path,
		Rows: 4,
		Columns: []ColumnSpec{
			{Name: "_id", Type: "int"},
			{Name: "name", Type: "string", Length: 16},
			{Name: "payload", Type: "raw", Length: 8},
			{Name: "score", Type: "double"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if info.Rows != 4 || len(info.Columns) != 4 || info.ReadOnly {
		t.Fatalf("info = %#v", info)
	}
	if _, err := app.CreateTable(CreateTableRequest{Path: path, Rows: 1, Columns: []ColumnSpec{{Name: "x", Type: "int"}}}); err == nil {
		t.Fatal("creating over an existing table without overwrite must fail")
	}
	if _, err := app.CreateTable(CreateTableRequest{
		Path: path, Rows: 4, Overwrite: true,
		Columns: []ColumnSpec{{Name: "_id", Type: "int"}, {Name: "name", Type: "string", Length: 16}, {Name: "payload", Type: "raw", Length: 8}, {Name: "score", Type: "double"}},
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := app.SetTableCell(path, 0, 0, "7", "typed"); err != nil {
		t.Fatal(err)
	}
	if _, err := app.SetTableCell(path, 0, 1, "Alice", "typed"); err != nil {
		t.Fatal(err)
	}
	cell, err := app.SetTableCell(path, 0, 2, "010203", "hex")
	if err != nil {
		t.Fatal(err)
	}
	if cell.RawHex != "0102030000000000" {
		t.Fatalf("raw = %s", cell.RawHex)
	}
	if _, err := app.SetTableCell(path, 0, 3, "12.5", "typed"); err != nil {
		t.Fatal(err)
	}
	if _, err := app.SetTableCell(path, 0, 0, "not-an-int", "typed"); err == nil {
		t.Fatal("invalid INT must fail")
	}

	page, err := app.GetTableRows(path, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 4 || len(page.Rows) != 4 {
		t.Fatalf("page = %#v", page)
	}
	if got := page.Rows[0].Cells[0].Value; got != "7" {
		t.Fatalf("id = %q", got)
	}
	if got := page.Rows[0].Cells[1].Value; got != "Alice" {
		t.Fatalf("name = %q", got)
	}
	if got := page.Rows[0].Cells[3].Value; got != "12.5" {
		t.Fatalf("score = %q", got)
	}

	hits, err := app.SearchTableRows(path, 1, "Alice", "typed", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0] != 0 {
		t.Fatalf("hits = %v", hits)
	}
	hits, err = app.SearchTableRows(path, 2, "0102030000000000", "hex", 10)
	if err != nil || len(hits) != 1 {
		t.Fatalf("raw hits = %v, %v", hits, err)
	}

	source := filepath.Join(dir, "blob.bin")
	if err := os.WriteFile(source, []byte("too-large"), 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := app.SetTableCellFromFile(path, 0, 2, source); err == nil {
		t.Fatal("embedding a file larger than the column must fail")
	}

	if err := app.ClearTableRow(path, 0); err != nil {
		t.Fatal(err)
	}
	page, err = app.GetTableRows(path, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, cell := range page.Rows[0].Cells {
		if !cell.Empty {
			t.Fatalf("cell %s is not empty: %#v", cell.Name, cell)
		}
	}

	if err := app.CloseTable(path); err != nil {
		t.Fatal(err)
	}
	readOnly, err := app.OpenTableReadOnly(path)
	if err != nil {
		t.Fatal(err)
	}
	if !readOnly.ReadOnly {
		t.Fatal("expected read-only info")
	}
	if _, err := app.SetTableCell(path, 0, 0, "1", "typed"); !errors.Is(err, binarytable.ErrReadOnly) {
		t.Fatalf("read-only write = %v", err)
	}
	if _, err := app.GetTableRows(path, 0, 2); err != nil {
		t.Fatal(err)
	}
}

func TestListWorkflow(t *testing.T) {
	app := NewApp()
	defer app.CloseAll()
	dir := t.TempDir()
	listPath := filepath.Join(dir, "blobs.bin")

	info, err := app.OpenList(listPath, 4)
	if err != nil {
		t.Fatal(err)
	}
	if info.Entries != 0 || info.MaxEntries != 4 {
		t.Fatalf("info = %#v", info)
	}

	first := filepath.Join(dir, "first.txt")
	second := filepath.Join(dir, "second.bin")
	if err := os.WriteFile(first, []byte("first blob"), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte{0, 1, 0, 255, 0}, 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := app.AddBlobFromPath(listPath, first); err != nil {
		t.Fatal(err)
	}
	if _, err := app.AddBlobFromPath(listPath, second); err != nil {
		t.Fatal(err)
	}
	if _, err := app.AddBlobFromBase64(listPath, "third.txt", "aGVsbG8="); err != nil {
		t.Fatal(err)
	}

	info, err = app.OpenList(listPath, 0)
	if err != nil {
		t.Fatal(err)
	}
	if info.Entries != 3 {
		t.Fatalf("entries = %d", info.Entries)
	}
	if info.MaxEntries != 4 {
		t.Fatalf("inferred max = %d, want 4", info.MaxEntries)
	}

	preview, err := app.GetBlob(listPath, 0, 1024)
	if err != nil {
		t.Fatal(err)
	}
	if !preview.IsText || preview.Text != "first blob" {
		t.Fatalf("preview = %#v", preview)
	}
	if preview.MIME == "" || preview.Base64 == "" {
		t.Fatalf("preview metadata = %#v", preview)
	}

	out := filepath.Join(dir, "exported.txt")
	if err := app.ExportBlob(listPath, 0, out); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil || string(data) != "first blob" {
		t.Fatalf("exported = %q, %v", data, err)
	}

	replacement := filepath.Join(dir, "replacement.txt")
	if err := os.WriteFile(replacement, []byte("replaced"), 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := app.ReplaceBlobFromPath(listPath, 1, replacement); err != nil {
		t.Fatal(err)
	}
	preview, err = app.GetBlob(listPath, 1, 1024)
	if err != nil || preview.Text != "replaced" {
		t.Fatalf("replacement = %#v, %v", preview, err)
	}

	info, err = app.DeleteBlob(listPath, 0)
	if err != nil {
		t.Fatal(err)
	}
	if info.Entries != 2 {
		t.Fatalf("entries after delete = %d", info.Entries)
	}
	preview, err = app.GetBlob(listPath, 0, 1024)
	if err != nil || preview.Text != "replaced" {
		t.Fatalf("first after delete = %#v, %v", preview, err)
	}
}

func TestCompressedViewsAndExplicitListCapacity(t *testing.T) {
	app := NewApp()
	defer app.CloseAll()
	dir := t.TempDir()

	listPath := filepath.Join(dir, "blobs.bin")
	info, err := app.CreateList(CreateListRequest{Path: listPath, MaxEntries: 3})
	if err != nil {
		t.Fatal(err)
	}
	if info.MaxEntries != 3 || info.Entries != 0 {
		t.Fatalf("created list = %#v", info)
	}
	if err := app.CloseList(listPath); err != nil {
		t.Fatal(err)
	}
	reopened, err := app.OpenList(listPath, 0)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.MaxEntries != 3 {
		t.Fatalf("recovered max = %d", reopened.MaxEntries)
	}
	source := filepath.Join(dir, "blob.bin")
	if err := os.WriteFile(source, []byte("compressed blob"), 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := app.AddBlobFromPath(listPath, source); err != nil {
		t.Fatal(err)
	}
	listDeflate := filepath.Join(dir, "blobs.deflate")
	if _, err := app.PackDeflate(listPath, listDeflate, 512, 0, false); err != nil {
		t.Fatal(err)
	}
	compressedList, err := app.OpenCompressedList(listDeflate, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !compressedList.ReadOnly || !compressedList.Compressed || compressedList.Entries != 1 || compressedList.MaxEntries != 3 {
		t.Fatalf("compressed list = %#v", compressedList)
	}
	preview, err := app.GetBlob(listDeflate, 0, 1024)
	if err != nil || preview.Text != "compressed blob" {
		t.Fatalf("compressed preview = %#v, %v", preview, err)
	}
	if _, err := app.AddBlobFromPath(listDeflate, source); !errors.Is(err, binarytable.ErrReadOnly) {
		t.Fatalf("compressed add = %v", err)
	}
	if _, err := app.DeleteBlob(listDeflate, 0); !errors.Is(err, binarytable.ErrReadOnly) {
		t.Fatalf("compressed delete = %v", err)
	}

	tablePath := filepath.Join(dir, "users.db")
	if _, err := app.CreateTable(CreateTableRequest{
		Path: tablePath, Rows: 4,
		Columns: []ColumnSpec{{Name: "_id", Type: "int"}, {Name: "name", Type: "string", Length: 16}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.SetTableCell(tablePath, 0, 0, "9", "typed"); err != nil {
		t.Fatal(err)
	}
	if _, err := app.SetTableCell(tablePath, 0, 1, "Compressed", "typed"); err != nil {
		t.Fatal(err)
	}
	tableDeflate := filepath.Join(dir, "users.db.deflate")
	if _, err := app.PackDeflate(tablePath, tableDeflate, 512, 0, false); err != nil {
		t.Fatal(err)
	}
	compressedTable, err := app.OpenCompressedTable(tableDeflate)
	if err != nil {
		t.Fatal(err)
	}
	if !compressedTable.ReadOnly || !compressedTable.Compressed || compressedTable.Rows != 4 {
		t.Fatalf("compressed table = %#v", compressedTable)
	}
	row, err := app.GetTableRows(tableDeflate, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got := row.Rows[0].Cells[1].Value; got != "Compressed" {
		t.Fatalf("compressed name = %q", got)
	}
	if _, err := app.SetTableCell(tableDeflate, 0, 0, "1", "typed"); !errors.Is(err, binarytable.ErrReadOnly) {
		t.Fatalf("compressed table write = %v", err)
	}
}

func TestDeflateWorkflow(t *testing.T) {
	app := NewApp()
	dir := t.TempDir()
	input := filepath.Join(dir, "data.bin")
	packed := filepath.Join(dir, "data.deflate")
	unpacked := filepath.Join(dir, "data.out")

	raw := make([]byte, 20000)
	for i := range raw {
		raw[i] = byte(i % 64)
	}
	if err := os.WriteFile(input, raw, 0o666); err != nil {
		t.Fatal(err)
	}

	info, err := app.PackDeflate(input, packed, 1024, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if info.BlockSize != 1024 || info.Blocks == 0 || info.OriginalLength != int64(len(raw)) {
		t.Fatalf("info = %#v", info)
	}
	if _, err := app.PackDeflate(input, packed, 1024, 0, false); err == nil {
		t.Fatal("packing over an existing file without overwrite must fail")
	}

	inspected, err := app.InspectDeflate(packed)
	if err != nil {
		t.Fatal(err)
	}
	if inspected.OriginalLength != int64(len(raw)) || inspected.Blocks != info.Blocks {
		t.Fatalf("inspected = %#v", inspected)
	}

	result, err := app.UnpackDeflate(packed, unpacked, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.OutputSize != int64(len(raw)) {
		t.Fatalf("output size = %d", result.OutputSize)
	}
	got, err := os.ReadFile(unpacked)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, raw) {
		t.Fatal("unpacked data differs")
	}
	if _, err := app.UnpackDeflate(packed, unpacked, 2, false); err == nil {
		t.Fatal("unpacking over an existing file without overwrite must fail")
	}
}
