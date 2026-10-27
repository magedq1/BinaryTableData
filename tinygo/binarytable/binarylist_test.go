package binarytable

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestBinaryList(t *testing.T) {
	stream := NewAutoSizeStream(16)
	list, err := NewBinaryList(stream, 8)
	if err != nil {
		t.Fatal(err)
	}
	if list.Size() != 0 {
		t.Fatalf("initial size = %d", list.Size())
	}
	index, err := list.Add([]byte("maged"))
	if err != nil || index != 0 {
		t.Fatalf("add = %d, %v", index, err)
	}
	index, err = list.AddFunc(func(w io.Writer) error {
		_, err := w.Write([]byte("ahmed"))
		return err
	})
	if err != nil || index != 1 {
		t.Fatalf("add func = %d, %v", index, err)
	}
	if list.Size() != 2 {
		t.Fatalf("size = %d", list.Size())
	}
	got, err := list.Bytes(0)
	if err != nil || string(got) != "maged" {
		t.Fatalf("bytes 0 = %q, %v", got, err)
	}
	got, err = list.Bytes(1)
	if err != nil || string(got) != "ahmed" {
		t.Fatalf("bytes 1 = %q, %v", got, err)
	}
	reader, err := list.Reader(1)
	if err != nil {
		t.Fatal(err)
	}
	got, err = ReadAll(reader)
	if err != nil || string(got) != "ahmed" {
		t.Fatalf("reader = %q, %v", got, err)
	}
	if _, err := list.Bytes(2); !errors.Is(err, ErrOutOfBounds) {
		t.Fatalf("out of bounds = %v", err)
	}
	if _, err := list.Bytes(-1); !errors.Is(err, ErrOutOfBounds) {
		t.Fatalf("negative = %v", err)
	}

	// Data must start after the 100-byte prefix and the reserved index.
	dataStart := int64(list.SchemaSize() + list.IndexSize())
	if int64(len(stream.ToBytes())) <= dataStart {
		t.Fatal("data was not written after the index area")
	}
	first, err := ReadBytesAt(stream, dataStart, 5)
	if err != nil || string(first) != "maged" {
		t.Fatalf("first blob at data start = %q, %v", first, err)
	}

	all, err := list.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := NewBinaryList(NewByteArrayStreamFixed(all), 8)
	if err != nil {
		t.Fatal(err)
	}
	got, err = reopened.Bytes(1)
	if err != nil || string(got) != "ahmed" {
		t.Fatalf("reopened = %q, %v", got, err)
	}
}

func TestBinaryListOnlyEmptyBlob(t *testing.T) {
	stream := NewAutoSizeStream(16)
	list, err := NewBinaryList(stream, 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := list.Add(nil); err != nil {
		t.Fatal(err)
	}
	if stream.Size() < int64(list.SchemaSize()+list.IndexSize()) {
		t.Fatalf("stream did not reserve the data area: %d", stream.Size())
	}
	got, err := list.Bytes(0)
	if err != nil || len(got) != 0 {
		t.Fatalf("empty blob = %v, %v", got, err)
	}
	reopened, err := NewBinaryList(NewByteArrayStreamFixed(stream.ToBytes()), 4)
	if err != nil {
		t.Fatal(err)
	}
	got, err = reopened.Bytes(0)
	if err != nil || len(got) != 0 {
		t.Fatalf("reopened empty blob = %v, %v", got, err)
	}
}

func TestBinaryListInitializeAndGuess(t *testing.T) {
	stream := NewAutoSizeStream(16)
	list, err := NewBinaryList(stream, 5)
	if err != nil {
		t.Fatal(err)
	}
	if err := list.Initialize(); err != nil {
		t.Fatal(err)
	}
	if want := int64(list.SchemaSize() + list.IndexSize()); stream.Size() != want {
		t.Fatalf("initialized size = %d, want %d", stream.Size(), want)
	}
	guessed, err := GuessListCapacity(stream)
	if err != nil {
		t.Fatal(err)
	}
	if guessed != 5 {
		t.Fatalf("guessed empty capacity = %d", guessed)
	}
	if _, err := list.AddReader(strings.NewReader("x")); err != nil {
		t.Fatal(err)
	}
	guessed, err = GuessListCapacity(stream)
	if err != nil {
		t.Fatal(err)
	}
	if guessed != 5 {
		t.Fatalf("guessed populated capacity = %d", guessed)
	}
}

func TestBinaryListAddReader(t *testing.T) {
	list, err := NewBinaryList(NewAutoSizeStream(16), 4)
	if err != nil {
		t.Fatal(err)
	}
	index, err := list.AddReader(strings.NewReader("streamed blob"))
	if err != nil || index != 0 {
		t.Fatalf("AddReader = %d, %v", index, err)
	}
	got, err := list.Bytes(0)
	if err != nil || string(got) != "streamed blob" {
		t.Fatalf("blob = %q, %v", got, err)
	}
	index, err = list.AddReader(strings.NewReader(""))
	if err != nil || index != 1 {
		t.Fatalf("empty AddReader = %d, %v", index, err)
	}
	got, err = list.Bytes(1)
	if err != nil || len(got) != 0 {
		t.Fatalf("empty blob = %v, %v", got, err)
	}
}

func TestBinaryListCapacity(t *testing.T) {
	list, err := NewBinaryList(NewAutoSizeStream(8), 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := list.Add([]byte("a")); err != nil {
		t.Fatal(err)
	}
	if _, err := list.Add([]byte("b")); !errors.Is(err, ErrOutOfBounds) {
		t.Fatalf("full list error = %v", err)
	}
}

func TestBinaryListEmptyAndBinaryData(t *testing.T) {
	list, err := NewBinaryList(NewAutoSizeStream(8), 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := list.Add(nil); err != nil {
		t.Fatal(err)
	}
	payload := []byte{0, 1, 0, 255, 0}
	if _, err := list.Add(payload); err != nil {
		t.Fatal(err)
	}
	got, err := list.Bytes(0)
	if err != nil || len(got) != 0 {
		t.Fatalf("empty = %v, %v", got, err)
	}
	got, err = list.Bytes(1)
	if err != nil || !bytes.Equal(got, payload) {
		t.Fatalf("binary = %v, %v", got, err)
	}
}
