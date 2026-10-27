package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/magedq1/BinaryTableData/tinygo/binarytable"
)

// DeflateInfo describes a block-deflate file.
type DeflateInfo struct {
	Path           string  `json:"path"`
	FileSize       int64   `json:"fileSize"`
	BlockSize      int     `json:"blockSize"`
	HeaderSize     int     `json:"headerSize"`
	MaxBlocks      int     `json:"maxBlocks"`
	Blocks         int     `json:"blocks"`
	OriginalLength int64   `json:"originalLength"`
	OutputSize     int64   `json:"outputSize"`
	Ratio          float64 `json:"ratio"`
}

// InspectDeflate reads the header of a block-deflate file.
func (a *App) InspectDeflate(path string) (*DeflateInfo, error) {
	key, err := normalizePath(path)
	if err != nil {
		return nil, err
	}
	return inspectDeflateFile(key)
}

func inspectDeflateFile(path string) (*DeflateInfo, error) {
	stream, err := binarytable.OpenFileStream(path, binarytable.ReadOnly)
	if err != nil {
		return nil, err
	}
	defer stream.Close()
	header, err := binarytable.ReadDeflateHeader(stream)
	if err != nil {
		return nil, err
	}
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	info := &DeflateInfo{
		Path:           path,
		FileSize:       stat.Size(),
		BlockSize:      header.BlockSize,
		HeaderSize:     header.HeaderSize,
		MaxBlocks:      header.MaxBlocks,
		Blocks:         header.Blocks,
		OriginalLength: header.OriginalLength,
		OutputSize:     stat.Size(),
	}
	if header.OriginalLength > 0 {
		info.Ratio = float64(stat.Size()) / float64(header.OriginalLength)
	}
	return info, nil
}

// PackDeflate compresses a file with the Kotlin-compatible block format.
func (a *App) PackDeflate(inputPath, outputPath string, blockSize, headerSize int, overwrite bool) (*DeflateInfo, error) {
	in, err := normalizePath(inputPath)
	if err != nil {
		return nil, err
	}
	out, err := normalizePath(outputPath)
	if err != nil {
		return nil, err
	}
	if in == out {
		return nil, fmt.Errorf("input and output must be different files")
	}
	if blockSize <= 0 {
		return nil, fmt.Errorf("block size must be > 0")
	}
	if headerSize < 0 {
		return nil, fmt.Errorf("header size must be >= 0")
	}
	if !overwrite {
		if _, err := os.Stat(out); err == nil {
			return nil, fmt.Errorf("%s already exists", out)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return nil, err
	}
	source, err := binarytable.OpenFileStream(in, binarytable.ReadOnly)
	if err != nil {
		return nil, err
	}
	defer source.Close()
	output, err := binarytable.CreateFileStream(out, 0, binarytable.ReadWrite)
	if err != nil {
		return nil, err
	}
	if err := binarytable.Deflate(source, output, blockSize, headerSize); err != nil {
		_ = output.Close()
		_ = os.Remove(out)
		return nil, err
	}
	if err := output.Sync(); err != nil {
		_ = output.Close()
		_ = os.Remove(out)
		return nil, err
	}
	if err := output.Close(); err != nil {
		_ = os.Remove(out)
		return nil, err
	}
	return inspectDeflateFile(out)
}

// UnpackDeflate decompresses a block-deflate file.
func (a *App) UnpackDeflate(inputPath, outputPath string, cacheMax int, overwrite bool) (*DeflateInfo, error) {
	in, err := normalizePath(inputPath)
	if err != nil {
		return nil, err
	}
	out, err := normalizePath(outputPath)
	if err != nil {
		return nil, err
	}
	if in == out {
		return nil, fmt.Errorf("input and output must be different files")
	}
	if cacheMax < 0 {
		cacheMax = 0
	}
	if cacheMax > 64 {
		cacheMax = 64
	}
	if !overwrite {
		if _, err := os.Stat(out); err == nil {
			return nil, fmt.Errorf("%s already exists", out)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return nil, err
	}
	source, err := binarytable.OpenFileStream(in, binarytable.ReadOnly)
	if err != nil {
		return nil, err
	}
	defer source.Close()
	reader, err := binarytable.NewDeflateReader(source, cacheMax)
	if err != nil {
		return nil, err
	}
	output, err := binarytable.CreateFileStream(out, reader.Size(), binarytable.ReadWrite)
	if err != nil {
		return nil, err
	}
	if err := binarytable.CopyAt(output, 0, reader, 0, reader.Size()); err != nil {
		_ = output.Close()
		_ = os.Remove(out)
		return nil, err
	}
	if err := output.Sync(); err != nil {
		_ = output.Close()
		_ = os.Remove(out)
		return nil, err
	}
	if err := output.Close(); err != nil {
		_ = os.Remove(out)
		return nil, err
	}
	info, err := inspectDeflateFile(in)
	if err != nil {
		return nil, err
	}
	info.OutputSize = reader.Size()
	return info, nil
}
