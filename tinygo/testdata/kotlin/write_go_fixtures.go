// This helper is intentionally inside testdata, so it is not part of the
// package build. It writes fixtures with the Go port for the reverse
// compatibility check performed by VerifyGoFixtures.java.
//
// Run from the tinygo directory:
//
//	go run ./testdata/kotlin/write_go_fixtures.go /path/to/output
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/magedq1/BinaryTableData/tinygo/binarytable"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./testdata/kotlin/write_go_fixtures.go <output-dir>")
		os.Exit(2)
	}
	dir := os.Args[1]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		panic(err)
	}

	header, err := binarytable.NewHeader(16, []binarytable.Column{
		binarytable.IntColumn("_id"),
		binarytable.StringColumn("first_name", 16),
		binarytable.StringColumn("last_name", 16),
		binarytable.LongColumn("created"),
		binarytable.DoubleColumn("score"),
		binarytable.RawColumn("payload", 8),
	})
	check(err)

	memory := binarytable.NewAutoSizeStream(64)
	table, err := binarytable.New(header, memory)
	check(err)
	check(table.SetInt(0, 0, 7))
	check(table.SetString(0, 1, "Alice"))
	check(table.SetString(0, 2, "Saleh"))
	check(table.SetLong(0, 3, 1_700_000_000_000))
	check(table.SetDouble(0, 4, 12.5))
	check(table.SetValue(0, 5, []byte{1, 2, 3}))
	check(table.SetInt(3, 0, 99))
	check(table.SetString(3, 1, "Maged"))
	check(os.WriteFile(filepath.Join(dir, "go-table.bin"), memory.ToBytes(), 0o666))

	fileTable, err := binarytable.CreateFile(header, filepath.Join(dir, "go-table-file.bin"))
	check(err)
	check(fileTable.SetInt(3, 0, 99))
	check(fileTable.SetString(3, 1, "Maged"))
	check(fileTable.Close())

	listStream := binarytable.NewAutoSizeStream(16)
	list, err := binarytable.NewBinaryList(listStream, 8)
	check(err)
	_, err = list.Add([]byte("maged"))
	check(err)
	_, err = list.Add([]byte("ahmed"))
	check(err)
	_, err = list.Add([]byte{0, 1, 0, 255, 0})
	check(err)
	check(os.WriteFile(filepath.Join(dir, "go-list.bin"), listStream.ToBytes(), 0o666))

	raw := make([]byte, 20000)
	for i := range raw {
		raw[i] = byte(i % 64)
	}
	deflated := binarytable.NewAutoSizeStream(4096)
	check(binarytable.Deflate(binarytable.NewByteArrayReader(raw), deflated, 1024, 0))
	check(os.WriteFile(filepath.Join(dir, "go-deflate.bin"), deflated.ToBytes(), 0o666))
	check(os.WriteFile(filepath.Join(dir, "go-deflate-raw.bin"), raw, 0o666))
	fmt.Println("wrote Go fixtures to", dir)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
