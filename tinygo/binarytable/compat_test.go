package binarytable

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// These fixtures are produced by the original Kotlin implementation. They prove
// that the Go port reads the exact same on-disk format. See testdata/README.md
// for regeneration instructions.
func TestKotlinTableFixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "testdata", "table.bin"))
	if err != nil {
		t.Fatal(err)
	}
	table, err := FromBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := table.GetIntByName(0, "_id"); err != nil || got != 7 {
		t.Fatalf("id = %d, %v", got, err)
	}
	if got, err := table.GetStringByName(0, "first_name"); err != nil || got != "Alice" {
		t.Fatalf("first_name = %q, %v", got, err)
	}
	if got, err := table.GetStringByName(0, "last_name"); err != nil || got != "Saleh" {
		t.Fatalf("last_name = %q, %v", got, err)
	}
	if got, err := table.GetLongByName(0, "created"); err != nil || got != 1_700_000_000_000 {
		t.Fatalf("created = %d, %v", got, err)
	}
	if got, err := table.GetDoubleByName(0, "score"); err != nil || got != 12.5 {
		t.Fatalf("score = %v, %v", got, err)
	}
	payload, err := table.GetValueByName(0, "payload")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(payload, []byte{1, 2, 3, 0, 0, 0, 0, 0}) {
		t.Fatalf("payload = %v", payload)
	}
	if got, err := table.GetInt(3, 0); err != nil || got != 99 {
		t.Fatalf("row 3 id = %d, %v", got, err)
	}
	if got, err := table.GetString(3, 1); err != nil || got != "Maged" {
		t.Fatalf("row 3 name = %q, %v", got, err)
	}
}

func TestKotlinTableFileFixture(t *testing.T) {
	path := filepath.Join("..", "testdata", "table-file.bin")
	table, err := OpenFileReadOnly(path)
	if err != nil {
		t.Fatal(err)
	}
	defer table.Close()
	if got, err := table.GetInt(3, 0); err != nil || got != 99 {
		t.Fatalf("id = %d, %v", got, err)
	}
	if got, err := table.GetString(3, 1); err != nil || got != "Maged" {
		t.Fatalf("name = %q, %v", got, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != table.Header.Capacity() {
		t.Fatalf("file size = %d, capacity = %d", info.Size(), table.Header.Capacity())
	}
}

func TestKotlinBinaryListFixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "testdata", "list.bin"))
	if err != nil {
		t.Fatal(err)
	}
	list, err := NewBinaryList(NewByteArrayStreamFixed(data), 8)
	if err != nil {
		t.Fatal(err)
	}
	if list.Size() != 3 {
		t.Fatalf("size = %d", list.Size())
	}
	for i, want := range [][]byte{
		[]byte("maged"),
		[]byte("ahmed"),
		{0, 1, 0, 255, 0},
	} {
		got, err := list.Bytes(i)
		if err != nil {
			t.Fatalf("blob %d: %v", i, err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("blob %d = %v, want %v", i, got, want)
		}
	}
}

func TestKotlinDeflateFixture(t *testing.T) {
	compressed, err := os.ReadFile(filepath.Join("..", "testdata", "deflate.bin"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join("..", "testdata", "deflate-raw.bin"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := InflateStream(NewByteArrayReader(compressed), 4)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("inflated Kotlin fixture differs from the original data")
	}
}
