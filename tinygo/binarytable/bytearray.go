package binarytable

import (
	"fmt"
	"io"
	"sync"
)

// ByteArrayReader is a read-only stream backed by a byte slice.
type ByteArrayReader struct {
	data []byte
}

// NewByteArrayReader returns a reader over data. The slice is not copied, so
// the caller must not modify it while the reader is in use.
func NewByteArrayReader(data []byte) *ByteArrayReader {
	return &ByteArrayReader{data: data}
}

// Size returns the number of bytes in the stream.
func (r *ByteArrayReader) Size() int64 { return int64(len(r.data)) }

// ReadAt implements Reader.
func (r *ByteArrayReader) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, ErrOutOfBounds
	}
	if off >= int64(len(r.data)) {
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n := copy(p, r.data[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

// Bytes returns the underlying slice. It is exposed for zero-copy integrations
// and must be treated as read-only.
func (r *ByteArrayReader) Bytes() []byte { return r.data }

// ByteArrayStream is an in-memory random-access stream over a fixed-size byte
// slice. It mirrors the Kotlin ByteArrayStream: writes that reach the end of the
// backing slice are truncated, and Size reports the highest written position
// when shrinkToWritten is true.
type ByteArrayStream struct {
	mu              sync.RWMutex
	data            []byte
	length          int64
	shrinkToWritten bool
}

// NewByteArrayStream wraps data in a stream whose Size is the highest position
// written so far. This matches the default Kotlin constructor.
func NewByteArrayStream(data []byte) *ByteArrayStream {
	return &ByteArrayStream{data: data, shrinkToWritten: true}
}

// NewByteArrayStreamFixed wraps data in a stream whose Size is always
// len(data), regardless of how much has been written. This matches the Kotlin
// constructor with shrinkToWrotePos = false.
func NewByteArrayStreamFixed(data []byte) *ByteArrayStream {
	return &ByteArrayStream{data: data}
}

// NewByteArrayStreamCapacity allocates a zero-filled in-memory stream of size
// bytes.
func NewByteArrayStreamCapacity(size int) *ByteArrayStream {
	if size < 0 {
		size = 0
	}
	return NewByteArrayStream(make([]byte, size))
}

// Size implements Reader.
func (s *ByteArrayStream) Size() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.shrinkToWritten {
		return s.length
	}
	return int64(len(s.data))
}

// ReadAt implements Reader. Reads are limited to the current Size.
func (s *ByteArrayStream) ReadAt(p []byte, off int64) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	size := int64(len(s.data))
	if s.shrinkToWritten {
		size = s.length
	}
	if off < 0 {
		return 0, ErrOutOfBounds
	}
	if off >= size {
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n := copy(p, s.data[off:size])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

// WriteAt implements Writer.
//
// A write starting past the end of the backing slice returns ErrOutOfBounds.
// A write that crosses the end is truncated and returns io.ErrShortWrite.
// This is the Go equivalent of the Kotlin implementation, which silently
// clips such writes.
func (s *ByteArrayStream) WriteAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if off < 0 {
		return 0, ErrOutOfBounds
	}
	if off > int64(len(s.data)) {
		return 0, ErrOutOfBounds
	}
	available := int64(len(s.data)) - off
	n := len(p)
	short := false
	if int64(n) > available {
		n = int(available)
		short = true
	}
	copy(s.data[off:off+int64(n)], p[:n])
	if end := off + int64(n); end > s.length {
		s.length = end
	}
	if short {
		return n, io.ErrShortWrite
	}
	return n, nil
}

// Grow extends the logical size to at least size bytes. It cannot grow beyond
// the backing slice.
func (s *ByteArrayStream) Grow(size int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if size < 0 {
		return ErrOutOfBounds
	}
	if size > int64(len(s.data)) {
		return fmt.Errorf("%w: cannot grow fixed byte slice to %d bytes", ErrOutOfBounds, size)
	}
	if size > s.length {
		s.length = size
	}
	return nil
}

// ToBytes returns a copy of the written bytes.
func (s *ByteArrayStream) ToBytes() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	size := int64(len(s.data))
	if s.shrinkToWritten {
		size = s.length
	}
	if size < 0 || size > int64(maxInt()) {
		return nil
	}
	out := make([]byte, int(size))
	copy(out, s.data[:size])
	return out
}

// Backing returns the underlying slice. It must be treated as read-only.
func (s *ByteArrayStream) Backing() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

// AutoSizeStream is an in-memory stream that grows in fixed-size blocks. It
// mirrors the Kotlin ByteArrayStreamAutoSize / ByteArrayStreamWriterAutoSize.
type AutoSizeStream struct {
	mu        sync.RWMutex
	blockSize int
	blocks    [][]byte
	length    int64
}

// NewAutoSizeStream returns a growable stream. A blockSize <= 0 selects the
// default of 1024 bytes, matching the Kotlin implementation.
func NewAutoSizeStream(blockSize int) *AutoSizeStream {
	if blockSize <= 0 {
		blockSize = 1024
	}
	return &AutoSizeStream{blockSize: blockSize}
}

// BlockSize returns the configured block size.
func (s *AutoSizeStream) BlockSize() int { return s.blockSize }

// Size implements Reader.
func (s *AutoSizeStream) Size() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.length
}

// ReadAt implements Reader.
func (s *AutoSizeStream) ReadAt(p []byte, off int64) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if off < 0 {
		return 0, ErrOutOfBounds
	}
	if off >= s.length {
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	want := len(p)
	if limit := s.length - off; int64(want) > limit {
		want = int(limit)
	}
	remaining := want
	dst := 0
	blockIndex := int(off / int64(s.blockSize))
	blockOffset := int(off % int64(s.blockSize))
	for remaining > 0 && blockIndex < len(s.blocks) {
		n := s.blockSize - blockOffset
		if n > remaining {
			n = remaining
		}
		copy(p[dst:dst+n], s.blocks[blockIndex][blockOffset:blockOffset+n])
		dst += n
		remaining -= n
		blockIndex++
		blockOffset = 0
	}
	n := want - remaining
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

// WriteAt implements Writer. The stream grows as needed.
func (s *AutoSizeStream) WriteAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if off < 0 {
		return 0, ErrOutOfBounds
	}
	if len(p) == 0 {
		return 0, nil
	}
	end := off + int64(len(p))
	if end < off {
		return 0, ErrOutOfBounds
	}
	lastBlock := int((end - 1) / int64(s.blockSize))
	for len(s.blocks) <= lastBlock {
		s.blocks = append(s.blocks, make([]byte, s.blockSize))
	}

	remaining := len(p)
	src := 0
	blockIndex := int(off / int64(s.blockSize))
	blockOffset := int(off % int64(s.blockSize))
	for remaining > 0 {
		n := s.blockSize - blockOffset
		if n > remaining {
			n = remaining
		}
		copy(s.blocks[blockIndex][blockOffset:blockOffset+n], p[src:src+n])
		src += n
		remaining -= n
		blockIndex++
		blockOffset = 0
	}
	if end > s.length {
		s.length = end
	}
	return len(p), nil
}

// Grow extends the logical size to at least size bytes without changing the
// payload. Missing bytes read as zero.
func (s *AutoSizeStream) Grow(size int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if size < 0 {
		return ErrOutOfBounds
	}
	if size == 0 {
		return nil
	}
	lastBlock := int((size - 1) / int64(s.blockSize))
	for len(s.blocks) <= lastBlock {
		s.blocks = append(s.blocks, make([]byte, s.blockSize))
	}
	if size > s.length {
		s.length = size
	}
	return nil
}

// ToBytes returns a copy of the written data.
func (s *AutoSizeStream) ToBytes() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.length < 0 || s.length > int64(maxInt()) {
		return nil
	}
	out := make([]byte, int(s.length))
	dst := 0
	for _, block := range s.blocks {
		if dst >= len(out) {
			break
		}
		n := len(block)
		if n > len(out)-dst {
			n = len(out) - dst
		}
		copy(out[dst:dst+n], block[:n])
		dst += n
	}
	return out
}
