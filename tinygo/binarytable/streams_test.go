package binarytable

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestByteArrayReader(t *testing.T) {
	data := make([]byte, 1024)
	rng := rand.New(rand.NewSource(1))
	if _, err := rng.Read(data); err != nil {
		t.Fatal(err)
	}
	reader := NewByteArrayReader(data)
	if reader.Size() != int64(len(data)) {
		t.Fatalf("size = %d", reader.Size())
	}
	for i := 0; i < 100; i++ {
		pos := rng.Intn(len(data))
		n := rng.Intn(len(data) - pos)
		got, err := ReadBytesAt(reader, int64(pos), n)
		if err != nil {
			t.Fatalf("read at %d: %v", pos, err)
		}
		if !bytes.Equal(got, data[pos:pos+n]) {
			t.Fatalf("mismatch at %d/%d", pos, n)
		}
	}
	if _, err := ReadBytesAt(reader, int64(len(data)+1), 1); err == nil {
		t.Fatal("read past end must fail")
	}
}

func TestByteArrayStream(t *testing.T) {
	stream := NewByteArrayStream(make([]byte, 16))
	if stream.Size() != 0 {
		t.Fatalf("initial size = %d", stream.Size())
	}
	payload := []byte("hello")
	n, err := stream.WriteAt(payload, 4)
	if err != nil || n != len(payload) {
		t.Fatalf("write = %d, %v", n, err)
	}
	if stream.Size() != 9 {
		t.Fatalf("size = %d", stream.Size())
	}
	got := stream.ToBytes()
	if want := append(make([]byte, 4), payload...); !bytes.Equal(got, want) {
		t.Fatalf("bytes = %v, want %v", got, want)
	}

	// A write that crosses the end is clipped.
	n, err = stream.WriteAt([]byte{1, 2, 3, 4}, 14)
	if !errors.Is(err, io.ErrShortWrite) || n != 2 {
		t.Fatalf("clipped write = %d, %v", n, err)
	}
	if _, err := stream.WriteAt([]byte{1}, 17); !errors.Is(err, ErrOutOfBounds) {
		t.Fatalf("past-end write = %v", err)
	}

	fixed := NewByteArrayStreamFixed(make([]byte, 4))
	if fixed.Size() != 4 {
		t.Fatalf("fixed size = %d", fixed.Size())
	}
}

func TestAutoSizeStream(t *testing.T) {
	stream := NewAutoSizeStream(7)
	rng := rand.New(rand.NewSource(2))
	want := make([]byte, 0, 4096)
	for i := 0; i < 200; i++ {
		off := rng.Intn(4096)
		length := rng.Intn(40)
		data := make([]byte, length)
		if _, err := rng.Read(data); err != nil {
			t.Fatal(err)
		}
		if _, err := stream.WriteAt(data, int64(off)); err != nil {
			t.Fatalf("write at %d: %v", off, err)
		}
		needed := off + length
		if needed > len(want) {
			want = append(want, make([]byte, needed-len(want))...)
		}
		copy(want[off:], data)
	}
	if stream.Size() != int64(len(want)) {
		t.Fatalf("size = %d, want %d", stream.Size(), len(want))
	}
	got := stream.ToBytes()
	if !bytes.Equal(got, want) {
		t.Fatal("stream content mismatch")
	}
	for i := 0; i < 100; i++ {
		off := rng.Intn(len(want))
		n := rng.Intn(len(want) - off)
		buf, err := ReadBytesAt(stream, int64(off), n)
		if err != nil {
			t.Fatalf("read at %d: %v", off, err)
		}
		if !bytes.Equal(buf, want[off:off+n]) {
			t.Fatalf("read mismatch at %d/%d", off, n)
		}
	}
}

func TestSubReader(t *testing.T) {
	reader := NewByteArrayReader([]byte("0123456789"))
	sub, err := NewSubReader(reader, 2, 7)
	if err != nil {
		t.Fatal(err)
	}
	if sub.Size() != 5 || sub.Start() != 2 || sub.End() != 7 {
		t.Fatalf("sub = %d %d %d", sub.Size(), sub.Start(), sub.End())
	}
	got, err := ReadAll(sub)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "23456" {
		t.Fatalf("got %q", got)
	}
	if _, err := NewSubReader(reader, 5, 2); !errors.Is(err, ErrOutOfBounds) {
		t.Fatalf("invalid sub = %v", err)
	}
	if _, err := NewSubReader(reader, -1, 2); !errors.Is(err, ErrOutOfBounds) {
		t.Fatalf("negative sub = %v", err)
	}
	nested, err := sub.SubReader(1, 4)
	if err != nil {
		t.Fatal(err)
	}
	nestedBytes, err := ReadAll(nested)
	if err != nil {
		t.Fatal(err)
	}
	if string(nestedBytes) != "345" {
		t.Fatalf("nested = %q", nestedBytes)
	}
}

func TestJoinedReader(t *testing.T) {
	parts := []string{"abc", "defg", "hijkl"}
	readers := make([]Reader, len(parts))
	var want bytes.Buffer
	for i, p := range parts {
		readers[i] = NewByteArrayReader([]byte(p))
		want.WriteString(p)
	}
	joined := NewJoinedReader(readers...)
	if joined.Size() != int64(want.Len()) {
		t.Fatalf("size = %d", joined.Size())
	}
	got, err := ReadAll(joined)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("joined = %q, want %q", got, want.String())
	}
	rng := rand.New(rand.NewSource(3))
	for i := 0; i < 100; i++ {
		off := rng.Intn(want.Len())
		n := rng.Intn(want.Len() - off)
		buf, err := ReadBytesAt(joined, int64(off), n)
		if err != nil {
			t.Fatalf("read at %d: %v", off, err)
		}
		if !bytes.Equal(buf, want.Bytes()[off:off+n]) {
			t.Fatalf("mismatch at %d/%d", off, n)
		}
	}
}

func TestFileStream(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stream.bin")
	stream, err := CreateFileStream(path, 32, ReadWrite)
	if err != nil {
		t.Fatal(err)
	}
	if stream.Size() != 32 {
		t.Fatalf("size = %d", stream.Size())
	}
	if err := WriteAtFull(stream, 4, []byte("test")); err != nil {
		t.Fatal(err)
	}
	got, err := ReadBytesAt(stream, 4, 4)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "test" {
		t.Fatalf("got %q", got)
	}
	if err := stream.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := stream.Close(); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 32 {
		t.Fatalf("file size = %d", info.Size())
	}

	ro, err := OpenFileStream(path, ReadOnly)
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	if _, err := ro.WriteAt([]byte{1}, 0); !errors.Is(err, ErrReadOnly) {
		t.Fatalf("read-only write = %v", err)
	}
}

func TestOpenOrCreateFileStream(t *testing.T) {
	path := filepath.Join(t.TempDir(), "created.bin")
	stream, err := OpenOrCreateFileStream(path)
	if err != nil {
		t.Fatal(err)
	}
	if stream.Size() != 0 {
		t.Fatalf("size = %d", stream.Size())
	}
	if err := WriteAtFull(stream, 2, []byte("ok")); err != nil {
		t.Fatal(err)
	}
	if err := stream.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := OpenOrCreateFileStream(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if reopened.Size() != 4 {
		t.Fatalf("reopened size = %d", reopened.Size())
	}
	got, err := ReadBytesAt(reopened, 2, 2)
	if err != nil || string(got) != "ok" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestCopyAt(t *testing.T) {
	src := NewByteArrayReader([]byte("abcdefghij"))
	dst := NewAutoSizeStream(4)
	if err := CopyAt(dst, 3, src, 2, 5); err != nil {
		t.Fatal(err)
	}
	got, err := ReadBytesAt(dst, 3, 5)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "cdefg" {
		t.Fatalf("got %q", got)
	}
}

func TestIOAdapters(t *testing.T) {
	reader := binarytableReaderForIO()
	got, err := io.ReadAll(NewIOReader(reader))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "io adapter" {
		t.Fatalf("reader = %q", got)
	}

	stream := NewAutoSizeStream(4)
	writer := NewIOWriter(stream, 3)
	if _, err := writer.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	written, err := ReadBytesAt(stream, 3, 5)
	if err != nil {
		t.Fatal(err)
	}
	if string(written) != "hello" {
		t.Fatalf("writer = %q", written)
	}
}

func binarytableReaderForIO() Reader {
	return NewByteArrayReader([]byte("io adapter"))
}

func TestCache(t *testing.T) {
	cache := NewCache[string, int](2)
	cache.Put("a", 1)
	cache.Put("b", 2)
	if v, ok := cache.Get("a"); !ok || v != 1 {
		t.Fatalf("get a = %d, %v", v, ok)
	}
	cache.Put("c", 3)
	if cache.Contains("b") {
		t.Fatal("b should have been evicted")
	}
	if !cache.Contains("a") || !cache.Contains("c") {
		t.Fatal("a and c should be present")
	}
	cache.Put("a", 10)
	if v, _ := cache.Get("a"); v != 10 {
		t.Fatalf("updated a = %d", v)
	}
	cache.Remove("a")
	if cache.Contains("a") {
		t.Fatal("a should be removed")
	}
	cache.Clear()
	if cache.Len() != 0 {
		t.Fatalf("len = %d", cache.Len())
	}
}
