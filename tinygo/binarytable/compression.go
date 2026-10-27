package binarytable

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
)

// DeflateData compresses data with zlib, which is the format produced by
// java.util.zip.Deflater and therefore by the Kotlin implementation.
func DeflateData(data []byte) ([]byte, error) {
	return DeflateDataRange(data, 0, len(data))
}

// DeflateDataRange compresses data[start:start+length].
func DeflateDataRange(data []byte, start, length int) ([]byte, error) {
	if start < 0 || length < 0 || start > len(data)-length {
		return nil, ErrOutOfBounds
	}
	var out bytes.Buffer
	writer := zlib.NewWriter(&out)
	if _, err := writer.Write(data[start : start+length]); err != nil {
		_ = writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// InflateData decompresses zlib data.
func InflateData(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%w: zlib: %v", ErrCorrupt, err)
	}
	out, err := io.ReadAll(reader)
	closeErr := reader.Close()
	if err != nil {
		return nil, fmt.Errorf("%w: zlib: %v", ErrCorrupt, err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("%w: zlib: %v", ErrCorrupt, closeErr)
	}
	return out, nil
}
