package binarytable

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"sync"
)

// DefaultBinaryListSchemaSize is the number of reserved bytes at the start of a
// BinaryList stream. It matches the Kotlin implementation.
const DefaultBinaryListSchemaSize = 100

// BinaryList stores variable-length blobs in a random-access stream.
//
// On-disk layout (all integers big-endian):
//
//	offset 0                      int32  number of stored blobs
//	offset 100                    int32  data start offsets, 4 bytes each
//	offset 100 + indexSize        blob payload area
//
// The index area is reserved for maxListSize entries, so data always starts at
// max(stream size, 100 + maxListSize*4). Because offsets are 32-bit, the total
// stream size is limited to roughly 2 GiB.
type BinaryList struct {
	mu         sync.Mutex
	stream     Stream
	schemaSize int
	indexSize  int
	maxSize    int
	realSize   int
}

// NewBinaryList opens or initializes a list. maxListSize is the maximum number
// of blobs that can be stored; a value <= 0 selects 1000.
func NewBinaryList(stream Stream, maxListSize int) (*BinaryList, error) {
	if stream == nil {
		return nil, fmt.Errorf("binarytable: nil stream")
	}
	if maxListSize <= 0 {
		maxListSize = 1000
	}
	if int64(maxListSize) > (math.MaxInt32-DefaultBinaryListSchemaSize)/4 {
		return nil, fmt.Errorf("%w: BinaryList index is too large", ErrOutOfBounds)
	}
	indexSize := maxListSize * 4
	size := 0
	if stream.Size() >= 4 {
		b, err := ReadBytesAt(stream, 0, 4)
		if err != nil {
			return nil, err
		}
		v, err := DecodeInt32(b)
		if err != nil {
			return nil, err
		}
		if v > 0 {
			size = int(v)
		}
	}
	if size > maxListSize {
		return nil, fmt.Errorf("%w: list contains %d entries but capacity is %d", ErrCorrupt, size, maxListSize)
	}
	return &BinaryList{
		stream:     stream,
		schemaSize: DefaultBinaryListSchemaSize,
		indexSize:  indexSize,
		maxSize:    maxListSize,
		realSize:   size,
	}, nil
}

// Size returns the number of stored blobs.
func (l *BinaryList) Size() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.realSize
}

// MaxSize returns the configured maximum number of blobs.
func (l *BinaryList) MaxSize() int { return l.maxSize }

// SchemaSize returns the fixed 100-byte prefix length.
func (l *BinaryList) SchemaSize() int { return l.schemaSize }

// IndexSize returns the size of the reserved index area.
func (l *BinaryList) IndexSize() int { return l.indexSize }

// Initialize reserves the fixed prefix and index area for a new list. It is
// idempotent and does not change existing entries. The stream must support
// growing (the built-in memory and file streams do).
func (l *BinaryList) Initialize() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.stream.Size() < 4 {
		if err := WriteAtFull(l.stream, 0, EncodeInt32(int32(l.realSize))); err != nil {
			return err
		}
	}
	minimum := int64(l.schemaSize + l.indexSize)
	if l.stream.Size() >= minimum {
		return nil
	}
	if grower, ok := l.stream.(Grower); ok {
		return grower.Grow(minimum)
	}
	return WriteAtFull(l.stream, minimum-1, []byte{0})
}

// GuessListCapacity recovers maxListSize from the data start offset or from
// the reserved prefix size. The Kotlin format does not store the capacity, so
// this is a heuristic; callers normally fall back to 1000 when it returns 0.
func GuessListCapacity(reader Reader) (int, error) {
	if reader == nil {
		return 0, ErrNoReader
	}
	size := reader.Size()
	if size < 4 {
		return 0, nil
	}
	n := 104
	if size < int64(n) {
		n = int(size)
	}
	prefix, err := ReadBytesAt(reader, 0, n)
	if err != nil {
		return 0, err
	}
	count, err := DecodeInt32(prefix[:4])
	if err != nil {
		return 0, err
	}
	if count > 0 && n >= 104 {
		first, err := DecodeInt32(prefix[100:104])
		if err != nil {
			return 0, err
		}
		if first > DefaultBinaryListSchemaSize {
			indexBytes := int64(first) - int64(DefaultBinaryListSchemaSize)
			if indexBytes%4 == 0 && indexBytes/4 <= int64(maxInt()) {
				return int(indexBytes / 4), nil
			}
		}
	}
	if count == 0 && size >= DefaultBinaryListSchemaSize {
		indexBytes := size - int64(DefaultBinaryListSchemaSize)
		if indexBytes%4 == 0 && indexBytes/4 <= int64(maxInt()) {
			return int(indexBytes / 4), nil
		}
	}
	return 0, nil
}

// Add appends a blob and returns its index.
func (l *BinaryList) Add(data []byte) (int, error) {
	return l.AddReader(bytes.NewReader(data))
}

// AddReader streams a blob from reader into the list without loading the whole
// blob into memory. It returns the new index.
func (l *BinaryList) AddReader(reader io.Reader) (int, error) {
	if reader == nil {
		return 0, fmt.Errorf("binarytable: nil blob reader")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.addReaderLocked(reader)
}

// AddFunc builds a blob with write and appends it. It is the Go equivalent of
// the Kotlin add { outputStream -> ... } overload.
func (l *BinaryList) AddFunc(write func(io.Writer) error) (int, error) {
	if write == nil {
		return 0, fmt.Errorf("binarytable: nil writer function")
	}
	var buf bytes.Buffer
	if err := write(&buf); err != nil {
		return 0, err
	}
	return l.Add(buf.Bytes())
}

func (l *BinaryList) addReaderLocked(reader io.Reader) (int, error) {
	if l.realSize >= l.maxSize {
		return 0, fmt.Errorf("%w: list is full (%d entries)", ErrOutOfBounds, l.maxSize)
	}
	start := l.stream.Size()
	if minimum := int64(l.schemaSize + l.indexSize); start < minimum {
		start = minimum
	}
	if start > math.MaxInt32 {
		return 0, fmt.Errorf("%w: 32-bit BinaryList offset overflow", ErrOutOfBounds)
	}
	written, err := io.Copy(NewIOWriter(l.stream, start), reader)
	if err != nil {
		return 0, err
	}
	if written > math.MaxInt32-start {
		return 0, fmt.Errorf("%w: blob would exceed the 32-bit BinaryList limit", ErrOutOfBounds)
	}
	// A zero-length payload does not extend a stream. Reserve the data start
	// position explicitly so the index never points past the stream end.
	if l.stream.Size() < start {
		grower, ok := l.stream.(Grower)
		if !ok {
			return 0, fmt.Errorf("binarytable: stream does not support empty blobs")
		}
		if err := grower.Grow(start); err != nil {
			return 0, err
		}
	}
	indexPos := int64(l.schemaSize + l.realSize*4)
	if err := WriteAtFull(l.stream, indexPos, EncodeInt32(int32(start))); err != nil {
		return 0, err
	}
	l.realSize++
	if err := WriteAtFull(l.stream, 0, EncodeInt32(int32(l.realSize))); err != nil {
		l.realSize--
		return 0, err
	}
	return l.realSize - 1, nil
}

// Reader returns a read-only view of one blob.
func (l *BinaryList) Reader(index int) (*SubReader, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if index < 0 || index >= l.realSize {
		return nil, fmt.Errorf("%w: index %d, size %d", ErrOutOfBounds, index, l.realSize)
	}
	startBytes, err := ReadBytesAt(l.stream, int64(l.schemaSize+index*4), 4)
	if err != nil {
		return nil, err
	}
	start, err := DecodeInt32(startBytes)
	if err != nil {
		return nil, err
	}
	if start <= 0 {
		return nil, fmt.Errorf("%w: invalid start position %d for index %d", ErrCorrupt, start, index)
	}
	end := l.stream.Size()
	if index+1 < l.realSize {
		nextBytes, err := ReadBytesAt(l.stream, int64(l.schemaSize+(index+1)*4), 4)
		if err != nil {
			return nil, err
		}
		next, err := DecodeInt32(nextBytes)
		if err != nil {
			return nil, err
		}
		if next > 0 {
			end = int64(next)
		}
	}
	if end < int64(start) {
		return nil, fmt.Errorf("%w: end position %d is smaller than start %d", ErrCorrupt, end, start)
	}
	return NewSubReader(l.stream, int64(start), end)
}

// Open returns one blob as a sequential io.Reader.
func (l *BinaryList) Open(index int) (io.Reader, error) {
	reader, err := l.Reader(index)
	if err != nil {
		return nil, err
	}
	return NewIOReader(reader), nil
}

// Bytes returns a copy of one blob.
func (l *BinaryList) Bytes(index int) ([]byte, error) {
	reader, err := l.Reader(index)
	if err != nil {
		return nil, err
	}
	return ReadAll(reader)
}

// ToBytes returns a copy of the complete underlying stream.
func (l *BinaryList) ToBytes() ([]byte, error) {
	return ReadAll(l.stream)
}
