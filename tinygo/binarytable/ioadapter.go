package binarytable

import "io"

// NewIOReader adapts a random-access Reader to a sequential io.Reader.
//
// The adapter starts at position zero. It is the Go equivalent of
// BinaryStreamReader.asInputStream() in the Kotlin library.
func NewIOReader(source Reader) io.Reader {
	return &ioReader{source: source}
}

type ioReader struct {
	source   Reader
	position int64
}

func (r *ioReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.position >= r.source.Size() {
		return 0, io.EOF
	}
	n, err := r.source.ReadAt(p, r.position)
	if n > 0 {
		r.position += int64(n)
	}
	if n == 0 && err == nil {
		return 0, io.EOF
	}
	return n, err
}

// NewIOWriter adapts a random-access Writer to a sequential io.Writer starting
// at start. It is the Go equivalent of asOutputStream(startPos).
func NewIOWriter(target Writer, start int64) io.Writer {
	if start < 0 {
		start = 0
	}
	return &ioWriter{target: target, position: start}
}

type ioWriter struct {
	target   Writer
	position int64
}

func (w *ioWriter) Write(p []byte) (int, error) {
	n, err := w.target.WriteAt(p, w.position)
	if n > 0 {
		w.position += int64(n)
	}
	if err == nil && n < len(p) {
		err = io.ErrShortWrite
	}
	return n, err
}
