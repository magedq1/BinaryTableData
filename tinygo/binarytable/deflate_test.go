package binarytable

import (
	"bytes"
	"errors"
	"math/rand"
	"testing"
)

func randomBytes(t *testing.T, size int, max byte) []byte {
	t.Helper()
	rng := rand.New(rand.NewSource(42))
	data := make([]byte, size)
	if _, err := rng.Read(data); err != nil {
		t.Fatal(err)
	}
	for i := range data {
		data[i] %= max
	}
	return data
}

func TestDataCompression(t *testing.T) {
	data := randomBytes(t, 100*1024, 255)
	compressed, err := DeflateData(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(compressed) == 0 {
		t.Fatal("compressed data is empty")
	}
	inflated, err := InflateData(compressed)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(inflated, data) {
		t.Fatal("inflate mismatch")
	}
	if _, err := InflateData([]byte("not zlib")); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("invalid zlib error = %v", err)
	}
}

func TestDeflateStream(t *testing.T) {
	for _, size := range []int{0, 1, 8192, 8193, 10*1024*1024 + 5} {
		data := randomBytes(t, size, 64)
		source := NewByteArrayReader(data)
		output := NewAutoSizeStream(4096)
		if err := Deflate(source, output, DefaultDeflateBlockSize, DefaultDeflateHeaderSize); err != nil {
			t.Fatalf("size %d: deflate: %v", size, err)
		}
		reader, err := NewDeflateReader(output, 4)
		if err != nil {
			t.Fatalf("size %d: reader: %v", size, err)
		}
		if reader.Size() != int64(size) {
			t.Fatalf("size %d: reported %d", size, reader.Size())
		}
		got, err := ReadAll(reader)
		if err != nil {
			t.Fatalf("size %d: read: %v", size, err)
		}
		if !bytes.Equal(got, data) {
			t.Fatalf("size %d: content mismatch", size)
		}

		// Reading across block boundaries must also work.
		if size >= 8192+10 {
			off := int64(8192 - 10)
			buf, err := ReadBytesAt(reader, off, 20)
			if err != nil {
				t.Fatalf("size %d: cross-block read: %v", size, err)
			}
			if !bytes.Equal(buf, data[off:off+20]) {
				t.Fatalf("size %d: cross-block mismatch", size)
			}
		}
	}
}

func TestDeflateWithHeaderSize(t *testing.T) {
	data := randomBytes(t, 5000, 255)
	output := NewAutoSizeStream(256)
	if err := Deflate(NewByteArrayReader(data), output, 1024, 64); err != nil {
		t.Fatal(err)
	}
	header, err := ReadBytesAt(output, DeflatePosHeaderSize, 4)
	if err != nil {
		t.Fatal(err)
	}
	gotHeader, _ := DecodeInt32(header)
	if gotHeader != 64 {
		t.Fatalf("header size = %d", gotHeader)
	}
	parsed, err := ReadDeflateHeader(output)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.HeaderSize != 64 || parsed.BlockSize != 1024 || parsed.OriginalLength != int64(len(data)) {
		t.Fatalf("parsed header = %#v", parsed)
	}
	reader, err := NewDeflateReader(output, 2)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("content mismatch")
	}
}

func TestDeflateCorrupt(t *testing.T) {
	data := randomBytes(t, 10000, 255)
	output := NewAutoSizeStream(256)
	if err := Deflate(NewByteArrayReader(data), output, 1024, 0); err != nil {
		t.Fatal(err)
	}
	raw := output.ToBytes()

	// Break the first block index entry so that the block points outside the
	// stream.
	corrupt := append([]byte(nil), raw...)
	corrupt[DeflateIndexReserve] = 0x7f
	def, err := NewDeflateReader(NewByteArrayReader(corrupt), 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ReadAll(def); !errors.Is(err, ErrCorrupt) && !errors.Is(err, ErrOutOfBounds) {
		t.Fatalf("corrupt read error = %v", err)
	}

	if _, err := NewDeflateReader(NewByteArrayReader(raw[:100]), 1); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("short header error = %v", err)
	}
}

func TestInflateStream(t *testing.T) {
	data := randomBytes(t, 12345, 100)
	output := NewAutoSizeStream(256)
	if err := Deflate(NewByteArrayReader(data), output, 2048, 0); err != nil {
		t.Fatal(err)
	}
	got, err := InflateStream(output, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("content mismatch")
	}
}
