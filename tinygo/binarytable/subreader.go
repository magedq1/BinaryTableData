package binarytable

import (
	"fmt"
	"io"
)

// SubReader exposes a contiguous window of another reader.
type SubReader struct {
	source Reader
	start  int64
	length int64
}

// NewSubReader returns a read-only view of source from start (inclusive) to end
// (exclusive). The end may be omitted, in which case the source length is used.
//
// It returns an error when start is negative or end is smaller than start.
func NewSubReader(source Reader, start int64, end ...int64) (*SubReader, error) {
	if source == nil {
		return nil, ErrNoReader
	}
	limit := source.Size()
	if len(end) > 0 {
		limit = end[0]
	}
	if start < 0 {
		return nil, fmt.Errorf("%w: start must be >= 0", ErrOutOfBounds)
	}
	if limit < start {
		return nil, fmt.Errorf("%w: end (%d) is smaller than start (%d)", ErrOutOfBounds, limit, start)
	}
	return &SubReader{source: source, start: start, length: limit - start}, nil
}

// Size implements Reader.
func (r *SubReader) Size() int64 { return r.length }

// ReadAt implements Reader.
func (r *SubReader) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, ErrOutOfBounds
	}
	if off >= r.length {
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	want := len(p)
	if limit := r.length - off; int64(want) > limit {
		want = int(limit)
	}
	n, err := r.source.ReadAt(p[:want], r.start+off)
	if n < len(p) && err == nil {
		err = io.EOF
	}
	return n, err
}

// Source returns the wrapped reader.
func (r *SubReader) Source() Reader { return r.source }

// Start returns the absolute start offset of the window.
func (r *SubReader) Start() int64 { return r.start }

// End returns the absolute end offset of the window.
func (r *SubReader) End() int64 { return r.start + r.length }

// SubReader returns a nested window relative to this window.
func (r *SubReader) SubReader(start, end int64) (*SubReader, error) {
	if start < 0 || end > r.length || end < start {
		return nil, ErrOutOfBounds
	}
	return &SubReader{source: r.source, start: r.start + start, length: end - start}, nil
}

// JoinedReader concatenates several readers into one logical stream.
type JoinedReader struct {
	readers []Reader
	ends    []int64
	size    int64
}

// NewJoinedReader concatenates readers in the supplied order.
func NewJoinedReader(readers ...Reader) *JoinedReader {
	j := &JoinedReader{readers: make([]Reader, 0, len(readers))}
	var total int64
	for _, r := range readers {
		if r == nil {
			continue
		}
		total += r.Size()
		j.readers = append(j.readers, r)
		j.ends = append(j.ends, total)
	}
	j.size = total
	return j
}

// Size implements Reader.
func (j *JoinedReader) Size() int64 { return j.size }

// ReadAt implements Reader.
func (j *JoinedReader) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, ErrOutOfBounds
	}
	if off >= j.size {
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	readers := len(j.readers)
	if readers == 0 {
		return 0, io.EOF
	}

	// Find the first reader containing off.
	index := 0
	for index < readers && off >= j.ends[index] {
		index++
	}
	if index >= readers {
		return 0, io.EOF
	}
	start := int64(0)
	if index > 0 {
		start = j.ends[index-1]
	}
	pos := off - start
	total := 0
	for index < readers && total < len(p) {
		want := len(p) - total
		if limit := j.readers[index].Size() - pos; int64(want) > limit {
			want = int(limit)
		}
		if want <= 0 {
			index++
			pos = 0
			continue
		}
		n, err := j.readers[index].ReadAt(p[total:total+want], pos)
		if n > 0 {
			total += n
			pos += int64(n)
		}
		if err != nil && err != io.EOF {
			return total, err
		}
		if n < want {
			return total, io.EOF
		}
		index++
		pos = 0
	}
	if total < len(p) {
		return total, io.EOF
	}
	return total, nil
}
