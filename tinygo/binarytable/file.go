package binarytable

import (
	"io"
	"os"
	"sync"
)

// FileMode selects how a FileStream is opened.
type FileMode int

const (
	// ReadOnly opens the file for reading only.
	ReadOnly FileMode = iota
	// ReadWrite opens the file for reading and writing.
	ReadWrite
)

// FileStream is a random-access stream backed by an *os.File.
//
// The Kotlin implementation uses memory-mapped files. Go/TinyGo does not offer
// a portable memory-mapping API, so this port uses positional reads and writes
// (ReadAt/WriteAt), which are available on hosted TinyGo targets such as
// linux, wasm/wasi, and windows.
//
// FileStream is safe for concurrent use by multiple goroutines.
type FileStream struct {
	mu   sync.RWMutex
	file *os.File
	mode FileMode
	size int64
	name string
}

// OpenFileStream opens an existing file.
func OpenFileStream(path string, mode FileMode) (*FileStream, error) {
	flag := os.O_RDONLY
	if mode == ReadWrite {
		flag = os.O_RDWR
	}
	f, err := os.OpenFile(path, flag, 0)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	return &FileStream{file: f, mode: mode, size: info.Size(), name: path}, nil
}

// OpenOrCreateFileStream opens a file for reading and writing, creating an
// empty file when it does not exist. The file is not truncated.
func OpenOrCreateFileStream(path string) (*FileStream, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o666)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	return &FileStream{file: f, mode: ReadWrite, size: info.Size(), name: path}, nil
}

// CreateFileStream creates (or truncates) a file with exactly size bytes.
//
// Creating a table file with CreateFileStream mirrors the Kotlin
// BinaryStreamImplMapped(file, schemaSize + maxRows*rowSize) constructor: the
// file has a fixed capacity and can be addressed at any valid row.
func CreateFileStream(path string, size int64, mode FileMode) (*FileStream, error) {
	if size < 0 {
		return nil, ErrOutOfBounds
	}
	if mode == ReadOnly {
		return nil, ErrReadOnly
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return nil, err
	}
	if err := f.Truncate(size); err != nil {
		_ = f.Close()
		return nil, err
	}
	return &FileStream{file: f, mode: mode, size: size, name: path}, nil
}

// Name returns the file path.
func (s *FileStream) Name() string { return s.name }

// Size implements Reader.
func (s *FileStream) Size() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.size
}

// ReadAt implements Reader.
func (s *FileStream) ReadAt(p []byte, off int64) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if off < 0 {
		return 0, ErrOutOfBounds
	}
	return s.file.ReadAt(p, off)
}

// WriteAt implements Writer.
func (s *FileStream) WriteAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.mode == ReadOnly {
		return 0, ErrReadOnly
	}
	if off < 0 {
		return 0, ErrOutOfBounds
	}
	n, err := s.file.WriteAt(p, off)
	if end := off + int64(n); end > s.size {
		s.size = end
	}
	if err == nil && n < len(p) {
		err = io.ErrShortWrite
	}
	return n, err
}

// Grow extends the file to at least size bytes. It cannot shrink the file.
func (s *FileStream) Grow(size int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.mode == ReadOnly {
		return ErrReadOnly
	}
	if size < 0 {
		return ErrOutOfBounds
	}
	if size <= s.size {
		return nil
	}
	if err := s.file.Truncate(size); err != nil {
		return err
	}
	s.size = size
	return nil
}

// Sync flushes file contents to stable storage.
func (s *FileStream) Sync() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.file.Sync()
}

// Close closes the file.
func (s *FileStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.file.Close()
}
