// Command demo shows the main parts of the TinyGo port: a fixed-schema table,
// a binary list of variable-length blobs, and block compression.
//
// Run it with:
//
//	go run ./examples/demo
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"github.com/magedq1/BinaryTableData/tinygo/binarytable"
)

func main() {
	header, err := binarytable.NewHeader(1000, []binarytable.Column{
		binarytable.IntColumn("_id"),
		binarytable.StringColumn("first_name", 64),
		binarytable.StringColumn("last_name", 64),
		binarytable.LongColumn("created"),
		binarytable.DoubleColumn("score"),
		binarytable.RawColumn("avatar", 128),
	})
	check(err)

	// In-memory table. Use binarytable.CreateFile for a fixed-capacity file.
	stream := binarytable.NewAutoSizeStream(1024)
	table, err := binarytable.New(header, stream)
	check(err)

	for row := 0; row < 3; row++ {
		check(table.SetInt(row, 0, int32(100+row)))
		check(table.SetString(row, 1, fmt.Sprintf("User%d", row)))
		check(table.SetString(row, 2, "Example"))
		check(table.SetLong(row, 3, int64(1_700_000_000_000+row)))
		check(table.SetDouble(row, 4, float64(row)*1.25))
		check(table.SetValueByName(row, "avatar", []byte{0x89, 'P', 'N', 'G'}))
	}

	id, err := table.GetIntByName(1, "_id")
	check(err)
	name, err := table.GetStringByName(1, "first_name")
	check(err)
	fmt.Printf("row 1: id=%d name=%q\n", id, name)

	found, err := table.IndexOfStringByName("User2", "first_name", 0)
	check(err)
	fmt.Printf("IndexOfString(User2) = %d\n", found)

	// Variable-length blobs.
	blobs := binarytable.NewAutoSizeStream(256)
	list, err := binarytable.NewBinaryList(blobs, 1000)
	check(err)
	_, err = list.Add([]byte("first blob"))
	check(err)
	_, err = list.AddFunc(func(w io.Writer) error {
		_, err := w.Write([]byte("second blob"))
		return err
	})
	check(err)
	second, err := list.Bytes(1)
	check(err)
	fmt.Printf("blob 1 = %q\n", second)

	// Block compression.
	payload := bytes.Repeat([]byte("tinygo binary table data "), 1000)
	compressed := binarytable.NewAutoSizeStream(4096)
	check(binarytable.Deflate(binarytable.NewByteArrayReader(payload), compressed, binarytable.DefaultDeflateBlockSize, 0))
	restored, err := binarytable.InflateStream(compressed, 8)
	check(err)
	if !bytes.Equal(payload, restored) {
		log.Fatal("compression round trip failed")
	}
	fmt.Printf("compressed %d bytes to %d bytes, then restored them\n", len(payload), compressed.Size())

	// A table can be opened again from its raw bytes (read-only).
	reopened, err := binarytable.FromBytes(stream.ToBytes())
	check(err)
	name, err = reopened.GetStringByName(2, "first_name")
	check(err)
	fmt.Printf("reopened row 2 name = %q\n", name)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
