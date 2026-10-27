package binarytable

import (
	"io"
)

// Reader is a random-access, read-only byte source.
//
// ReadAt follows the io.ReaderAt contract: it reads len(p) bytes starting at
// off, returns the number of bytes read, and returns io.EOF when fewer bytes
// than requested are available at the end of the stream.
type Reader interface {
	ReadAt(p []byte, off int64) (n int, err error)
	Size() int64
}

// Writer is a random-access byte sink.
//
// WriteAt follows the io.WriterAt contract. Implementations must not retain p.
type Writer interface {
	WriteAt(p []byte, off int64) (n int, err error)
}

// Stream is a random-access read/write byte source.
type Stream interface {
	Reader
	Writer
}

// Grower is implemented by streams that can extend their logical size without
// writing payload bytes. BinaryList uses it to reserve the data start position
// for empty blobs, matching the Kotlin growable stream behaviour.
type Grower interface {
	Grow(size int64) error
}

// readOnlyStream adapts a Reader to the Stream interface. Writes return
// ErrReadOnly.
type readOnlyStream struct {
	Reader
}

// AsStream adapts a Reader to a Stream whose write methods return ErrReadOnly.
func AsStream(r Reader) Stream {
	if s, ok := r.(Stream); ok {
		return s
	}
	return readOnlyStream{Reader: r}
}

// WriteAt implements Writer.
func (s readOnlyStream) WriteAt([]byte, int64) (int, error) {
	return 0, ErrReadOnly
}

// ReadAtFull fills p from r at off. It returns io.ErrUnexpectedEOF when the
// stream ends before p is full.
func ReadAtFull(r Reader, off int64, p []byte) error {
	if off < 0 {
		return ErrOutOfBounds
	}
	done := 0
	for done < len(p) {
		n, err := r.ReadAt(p[done:], off+int64(done))
		if n > 0 {
			done += n
		}
		if err != nil {
			if err == io.EOF {
				if done == len(p) {
					return nil
				}
				return io.ErrUnexpectedEOF
			}
			return err
		}
		if n == 0 {
			return io.ErrUnexpectedEOF
		}
	}
	return nil
}

// ReadBytesAt allocates n bytes and fills them from r at off.
func ReadBytesAt(r Reader, off int64, n int) ([]byte, error) {
	if n < 0 {
		return nil, ErrOutOfBounds
	}
	b := make([]byte, n)
	if err := ReadAtFull(r, off, b); err != nil {
		return nil, err
	}
	return b, nil
}

// ReadAll reads the complete stream.
func ReadAll(r Reader) ([]byte, error) {
	size := r.Size()
	if size < 0 {
		return nil, ErrCorrupt
	}
	if size > int64(maxInt()) {
		return nil, ErrOutOfBounds
	}
	return ReadBytesAt(r, 0, int(size))
}

// WriteAtFull writes all of p to w at off.
func WriteAtFull(w Writer, off int64, p []byte) error {
	if off < 0 {
		return ErrOutOfBounds
	}
	done := 0
	for done < len(p) {
		n, err := w.WriteAt(p[done:], off+int64(done))
		if n > 0 {
			done += n
		}
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
	}
	return nil
}

// CopyAt copies n bytes from src at srcOff to dst at dstOff.
func CopyAt(dst Writer, dstOff int64, src Reader, srcOff, n int64) error {
	if n < 0 {
		return ErrOutOfBounds
	}
	buf := make([]byte, 32*1024)
	for n > 0 {
		chunk := int64(len(buf))
		if chunk > n {
			chunk = n
		}
		if err := ReadAtFull(src, srcOff, buf[:chunk]); err != nil {
			return err
		}
		if err := WriteAtFull(dst, dstOff, buf[:chunk]); err != nil {
			return err
		}
		srcOff += chunk
		dstOff += chunk
		n -= chunk
	}
	return nil
}

func maxInt() int {
	return int(^uint(0) >> 1)
}
