package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"unicode/utf8"

	"github.com/magedq1/BinaryTableData/tinygo/binarytable"
)

// BlobInfo describes one entry in a binary list.
type BlobInfo struct {
	Index int    `json:"index"`
	Size  int64  `json:"size"`
	Name  string `json:"name,omitempty"`
}

// ListInfo describes an open binary list.
type ListInfo struct {
	Path       string     `json:"path"`
	ReadOnly   bool       `json:"readOnly"`
	Compressed bool       `json:"compressed"`
	Entries    int        `json:"entries"`
	MaxEntries int        `json:"maxEntries"`
	SchemaSize int        `json:"schemaSize"`
	IndexSize  int        `json:"indexSize"`
	FileSize   int64      `json:"fileSize"`
	Blobs      []BlobInfo `json:"blobs"`
}

// BlobData contains a blob plus preview material.
type BlobData struct {
	Index     int    `json:"index"`
	Size      int64  `json:"size"`
	MIME      string `json:"mime"`
	IsText    bool   `json:"isText"`
	Text      string `json:"text"`
	Hex       string `json:"hex"`
	Base64    string `json:"base64"`
	Truncated bool   `json:"truncated"`
}

type listSession struct {
	mu         sync.Mutex
	path       string
	max        int
	stream     *binarytable.FileStream
	source     io.Closer
	list       *binarytable.BinaryList
	readOnly   bool
	compressed bool
}

func (s *listSession) close() error {
	if s.stream != nil {
		err := s.stream.Close()
		s.stream = nil
		return err
	}
	if s.source != nil {
		err := s.source.Close()
		s.source = nil
		return err
	}
	return nil
}

// CreateListRequest describes a new binary list file.
type CreateListRequest struct {
	Path       string `json:"path"`
	MaxEntries int    `json:"maxEntries"`
	Overwrite  bool   `json:"overwrite"`
}

// CreateList creates an empty binary list with an explicit capacity and opens
// it. The reserved prefix lets later opens recover maxEntries even when the
// list has no entries.
func (a *App) CreateList(request CreateListRequest) (*ListInfo, error) {
	key, err := normalizePath(request.Path)
	if err != nil {
		return nil, err
	}
	if request.MaxEntries <= 0 {
		return nil, fmt.Errorf("max entries must be > 0")
	}
	if !request.Overwrite {
		if _, err := os.Stat(key); err == nil {
			return nil, fmt.Errorf("%s already exists", key)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(key), 0o755); err != nil {
		return nil, err
	}

	a.mu.Lock()
	if existing := a.tables[key]; existing != nil {
		delete(a.tables, key)
		existing.mu.Lock()
		_ = existing.close()
		existing.mu.Unlock()
	}
	if existing := a.lists[key]; existing != nil {
		delete(a.lists, key)
		existing.mu.Lock()
		_ = existing.close()
		existing.mu.Unlock()
	}
	a.mu.Unlock()

	stream, err := binarytable.CreateFileStream(key, 0, binarytable.ReadWrite)
	if err != nil {
		return nil, err
	}
	list, err := binarytable.NewBinaryList(stream, request.MaxEntries)
	if err != nil {
		_ = stream.Close()
		_ = os.Remove(key)
		return nil, err
	}
	if err := list.Initialize(); err != nil {
		_ = stream.Close()
		_ = os.Remove(key)
		return nil, err
	}
	if err := stream.Sync(); err != nil {
		_ = stream.Close()
		_ = os.Remove(key)
		return nil, err
	}
	session := &listSession{path: key, max: request.MaxEntries, stream: stream, list: list}
	a.mu.Lock()
	a.lists[key] = session
	a.mu.Unlock()
	return session.info()
}

// OpenList opens or creates a binary list. maxListSize <= 0 uses the capacity
// recovered from the file, or 1000 for a brand-new empty file.
func (a *App) OpenList(path string, maxListSize int) (*ListInfo, error) {
	key, err := normalizePath(path)
	if err != nil {
		return nil, err
	}
	explicitMax := maxListSize > 0
	a.mu.Lock()
	existing := a.lists[key]
	if existing != nil && !existing.compressed && (maxListSize <= 0 || existing.max == maxListSize) {
		a.mu.Unlock()
		return existing.info()
	}
	if existing != nil {
		delete(a.lists, key)
	}
	a.mu.Unlock()
	if existing != nil {
		existing.mu.Lock()
		_ = existing.close()
		existing.mu.Unlock()
	}

	stream, err := binarytable.OpenOrCreateFileStream(key)
	if err != nil {
		return nil, err
	}
	if maxListSize <= 0 {
		maxListSize, err = binarytable.GuessListCapacity(stream)
		if err != nil {
			_ = stream.Close()
			return nil, err
		}
		if maxListSize <= 0 {
			maxListSize = 1000
		}
	}
	list, err := binarytable.NewBinaryList(stream, maxListSize)
	if err != nil {
		_ = stream.Close()
		return nil, err
	}
	if explicitMax && list.Size() == 0 && stream.Size() < int64(list.SchemaSize()+list.IndexSize()) {
		if err := list.Initialize(); err != nil {
			_ = stream.Close()
			return nil, err
		}
		if err := stream.Sync(); err != nil {
			_ = stream.Close()
			return nil, err
		}
	}
	session := &listSession{path: key, max: maxListSize, stream: stream, list: list}
	a.mu.Lock()
	a.lists[key] = session
	a.mu.Unlock()
	return session.info()
}

// OpenCompressedList reads a binary list through a
// DeflateBinaryStreamReader equivalent. Add, replace, and delete are disabled.
func (a *App) OpenCompressedList(path string, maxListSize int) (*ListInfo, error) {
	key, err := normalizePath(path)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	existing := a.lists[key]
	if existing != nil && existing.compressed && (maxListSize <= 0 || existing.max == maxListSize) {
		a.mu.Unlock()
		return existing.info()
	}
	if existing != nil {
		delete(a.lists, key)
	}
	a.mu.Unlock()
	if existing != nil {
		existing.mu.Lock()
		_ = existing.close()
		existing.mu.Unlock()
	}

	source, err := binarytable.OpenFileStream(key, binarytable.ReadOnly)
	if err != nil {
		return nil, err
	}
	reader, err := binarytable.NewDeflateReader(source, 8)
	if err != nil {
		_ = source.Close()
		return nil, err
	}
	if maxListSize <= 0 {
		maxListSize, err = binarytable.GuessListCapacity(reader)
		if err != nil {
			_ = source.Close()
			return nil, err
		}
		if maxListSize <= 0 {
			maxListSize = 1000
		}
	}
	list, err := binarytable.NewBinaryList(binarytable.AsStream(reader), maxListSize)
	if err != nil {
		_ = source.Close()
		return nil, err
	}
	session := &listSession{
		path: key, max: maxListSize, source: source, list: list,
		readOnly: true, compressed: true,
	}
	a.mu.Lock()
	a.lists[key] = session
	a.mu.Unlock()
	return session.info()
}

// CloseList closes one open list.
func (a *App) CloseList(path string) error {
	key, err := normalizePath(path)
	if err != nil {
		return err
	}
	a.mu.Lock()
	session := a.lists[key]
	delete(a.lists, key)
	a.mu.Unlock()
	if session == nil {
		return nil
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	return session.close()
}

func (a *App) getList(path string) (*listSession, error) {
	key, err := normalizePath(path)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	session := a.lists[key]
	a.mu.Unlock()
	if session == nil {
		return nil, fmt.Errorf("list is not open: %s", key)
	}
	return session, nil
}

func (s *listSession) info() (*ListInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.infoLocked()
}

func (s *listSession) infoLocked() (*ListInfo, error) {
	stat, err := os.Stat(s.path)
	if err != nil {
		return nil, err
	}
	blobs := make([]BlobInfo, 0, s.list.Size())
	for i := 0; i < s.list.Size(); i++ {
		reader, err := s.list.Reader(i)
		if err != nil {
			return nil, err
		}
		blobs = append(blobs, BlobInfo{Index: i, Size: reader.Size()})
	}
	return &ListInfo{
		Path:       s.path,
		ReadOnly:   s.readOnly,
		Compressed: s.compressed,
		Entries:    s.list.Size(),
		MaxEntries: s.list.MaxSize(),
		SchemaSize: s.list.SchemaSize(),
		IndexSize:  s.list.IndexSize(),
		FileSize:   stat.Size(),
		Blobs:      blobs,
	}, nil
}

// AddBlobFromPath embeds a file into the list.
func (a *App) AddBlobFromPath(path string, sourcePath string) (*BlobInfo, error) {
	session, err := a.getList(path)
	if err != nil {
		return nil, err
	}
	source, err := os.Open(sourcePath)
	if err != nil {
		return nil, err
	}
	defer source.Close()
	stat, err := source.Stat()
	if err != nil {
		return nil, err
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.readOnly {
		return nil, binarytable.ErrReadOnly
	}
	index, err := session.list.AddReader(source)
	if err != nil {
		return nil, err
	}
	return &BlobInfo{Index: index, Size: stat.Size(), Name: filepath.Base(sourcePath)}, nil
}

// AddBlobFromBase64 embeds bridge-supplied bytes into the list.
func (a *App) AddBlobFromBase64(path, name, dataBase64 string) (*BlobInfo, error) {
	session, err := a.getList(path)
	if err != nil {
		return nil, err
	}
	data, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 data: %w", err)
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.readOnly {
		return nil, binarytable.ErrReadOnly
	}
	index, err := session.list.Add(data)
	if err != nil {
		return nil, err
	}
	return &BlobInfo{Index: index, Size: int64(len(data)), Name: name}, nil
}

// GetBlob reads a blob and returns preview material.
func (a *App) GetBlob(path string, index, previewLimit int) (*BlobData, error) {
	session, err := a.getList(path)
	if err != nil {
		return nil, err
	}
	if previewLimit <= 0 {
		previewLimit = 256 * 1024
	}
	if previewLimit > 1024*1024 {
		previewLimit = 1024 * 1024
	}
	session.mu.Lock()
	data, err := session.list.Bytes(index)
	session.mu.Unlock()
	if err != nil {
		return nil, err
	}
	mime := http.DetectContentType(data)
	if len(data) == 0 {
		mime = "application/octet-stream"
	}
	textual := utf8.Valid(data) && isTextData(data)
	result := &BlobData{
		Index: index,
		Size:  int64(len(data)),
		MIME:  mime,
	}
	limit := previewLimit
	if limit > len(data) {
		limit = len(data)
	}
	if textual {
		result.IsText = true
		result.Text = string(data[:limit])
	}
	hexLimit := 256
	if hexLimit > limit {
		hexLimit = limit
	}
	result.Hex = hex.EncodeToString(data[:hexLimit])
	if len(data) <= 8*1024*1024 {
		result.Base64 = base64.StdEncoding.EncodeToString(data)
	}
	result.Truncated = limit < len(data)
	return result, nil
}

func isTextData(data []byte) bool {
	for _, b := range data {
		if b == '\n' || b == '\r' || b == '\t' {
			continue
		}
		if b < 0x20 || b == 0x7f {
			return false
		}
	}
	return true
}

// ExportBlob writes one blob to a file.
func (a *App) ExportBlob(path string, index int, outputPath string) error {
	session, err := a.getList(path)
	if err != nil {
		return err
	}
	session.mu.Lock()
	data, err := session.list.Bytes(index)
	session.mu.Unlock()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0o666)
}

// DeleteBlob removes one blob by rebuilding the list atomically.
func (a *App) DeleteBlob(path string, index int) (*ListInfo, error) {
	session, err := a.getList(path)
	if err != nil {
		return nil, err
	}
	return session.rewrite(index, nil, true)
}

// ReplaceBlobFromPath replaces one blob with a file.
func (a *App) ReplaceBlobFromPath(path string, index int, sourcePath string) (*ListInfo, error) {
	source, err := os.Open(sourcePath)
	if err != nil {
		return nil, err
	}
	defer source.Close()
	session, err := a.getList(path)
	if err != nil {
		return nil, err
	}
	return session.rewrite(index, source, false)
}

func (s *listSession) reopenLocked() {
	stream, err := binarytable.OpenOrCreateFileStream(s.path)
	if err != nil {
		return
	}
	list, err := binarytable.NewBinaryList(stream, s.max)
	if err != nil {
		_ = stream.Close()
		return
	}
	s.stream = stream
	s.list = list
}

func (s *listSession) rewrite(index int, replacement io.Reader, skip bool) (*ListInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.readOnly || s.stream == nil {
		return nil, binarytable.ErrReadOnly
	}
	if index < 0 || index >= s.list.Size() {
		return nil, binarytable.ErrOutOfBounds
	}

	temp, err := os.CreateTemp(filepath.Dir(s.path), ".binarytable-list-*")
	if err != nil {
		return nil, err
	}
	tempPath := temp.Name()
	if err := temp.Close(); err != nil {
		_ = os.Remove(tempPath)
		return nil, err
	}
	_ = os.Remove(tempPath)
	out, err := binarytable.CreateFileStream(tempPath, 0, binarytable.ReadWrite)
	if err != nil {
		return nil, err
	}
	newList, err := binarytable.NewBinaryList(out, s.max)
	if err != nil {
		_ = out.Close()
		_ = os.Remove(tempPath)
		return nil, err
	}
	for i := 0; i < s.list.Size(); i++ {
		if i == index {
			if !skip {
				if replacement == nil {
					_ = out.Close()
					_ = os.Remove(tempPath)
					return nil, fmt.Errorf("replacement reader is nil")
				}
				if _, err := newList.AddReader(replacement); err != nil {
					_ = out.Close()
					_ = os.Remove(tempPath)
					return nil, err
				}
			}
			continue
		}
		reader, err := s.list.Reader(i)
		if err != nil {
			_ = out.Close()
			_ = os.Remove(tempPath)
			return nil, err
		}
		if _, err := newList.AddReader(binarytable.NewIOReader(reader)); err != nil {
			_ = out.Close()
			_ = os.Remove(tempPath)
			return nil, err
		}
	}
	if err := out.Sync(); err != nil {
		_ = out.Close()
		_ = os.Remove(tempPath)
		return nil, err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tempPath)
		return nil, err
	}

	if err := s.stream.Close(); err != nil {
		_ = os.Remove(tempPath)
		return nil, err
	}
	s.stream = nil

	backup := s.path + ".bak"
	_ = os.Remove(backup)
	if err := os.Rename(s.path, backup); err != nil {
		_ = os.Remove(tempPath)
		s.reopenLocked()
		return nil, err
	}
	if err := os.Rename(tempPath, s.path); err != nil {
		_ = os.Rename(backup, s.path)
		_ = os.Remove(tempPath)
		s.reopenLocked()
		return nil, err
	}
	_ = os.Remove(backup)
	stream, err := binarytable.OpenOrCreateFileStream(s.path)
	if err != nil {
		return nil, err
	}
	list, err := binarytable.NewBinaryList(stream, s.max)
	if err != nil {
		_ = stream.Close()
		return nil, err
	}
	s.stream = stream
	s.list = list
	return s.infoLocked()
}
