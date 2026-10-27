package binarytable

import (
	"fmt"
	"io"
	"math"
)

// Deflate stream header offsets. They match the Kotlin
// DeflateBinaryStreamReader companion object.
const (
	// DeflatePosBlockSize is the absolute offset of the int32 block size.
	DeflatePosBlockSize = 100
	// DeflatePosHeaderSize is the absolute offset of the int32 caller header
	// size. 512 bytes are always reserved after it for the block index.
	DeflatePosHeaderSize = 104
	// DeflatePosMaxBlocks is the absolute offset of the int32 maximum number
	// of blocks.
	DeflatePosMaxBlocks = 108
	// DeflatePosBlocksCount is the absolute offset of the int32 number of used
	// blocks.
	DeflatePosBlocksCount = 112
	// DeflatePosOriginalLength is the absolute offset of the int64 original
	// (uncompressed) length.
	DeflatePosOriginalLength = 116

	// DefaultDeflateBlockSize is the Kotlin default block size.
	DefaultDeflateBlockSize = 8192
	// DefaultDeflateHeaderSize is the Kotlin default user header size.
	DefaultDeflateHeaderSize = 0
	// DeflateIndexReserve is the fixed gap between the caller header and the
	// block index.
	DeflateIndexReserve = 512
)

// Deflate compresses source into output using the block format of the Kotlin
// library.
//
// The output starts 512 bytes after headerSize. Bytes [0,headerSize) are left
// untouched for caller data. Each compressed block has a 16-byte index entry
// (start and end offset, both int64 relative to the block area).
func Deflate(source Reader, output Writer, blockSize, headerSize int) error {
	if source == nil || output == nil {
		return fmt.Errorf("binarytable: nil source or output")
	}
	if blockSize <= 0 || blockSize > math.MaxInt32 {
		return fmt.Errorf("%w: invalid block size %d", ErrOutOfBounds, blockSize)
	}
	if headerSize < 0 || headerSize > math.MaxInt32 {
		return fmt.Errorf("%w: invalid header size %d", ErrOutOfBounds, headerSize)
	}
	length := source.Size()
	if length < 0 {
		return ErrCorrupt
	}
	maxBlocks64 := length/int64(blockSize) + 1
	if maxBlocks64 > int64(maxInt()) {
		return fmt.Errorf("%w: too many deflate blocks", ErrOutOfBounds)
	}
	maxBlocks := int(maxBlocks64)
	usedBlocks := maxBlocks
	if int64(maxBlocks)*int64(blockSize) == length {
		usedBlocks--
	}
	if usedBlocks < 0 {
		usedBlocks = 1
	}
	if maxBlocks > math.MaxInt32 || usedBlocks > math.MaxInt32 {
		return fmt.Errorf("%w: too many deflate blocks", ErrOutOfBounds)
	}
	if maxBlocks > maxInt()/16 {
		return fmt.Errorf("%w: deflate index is too large", ErrOutOfBounds)
	}

	if err := WriteAtFull(output, DeflatePosBlockSize, EncodeInt32(int32(blockSize))); err != nil {
		return err
	}
	if err := WriteAtFull(output, DeflatePosHeaderSize, EncodeInt32(int32(headerSize))); err != nil {
		return err
	}
	if err := WriteAtFull(output, DeflatePosMaxBlocks, EncodeInt32(int32(maxBlocks))); err != nil {
		return err
	}
	if err := WriteAtFull(output, DeflatePosBlocksCount, EncodeInt32(int32(usedBlocks))); err != nil {
		return err
	}
	if err := WriteAtFull(output, DeflatePosOriginalLength, EncodeInt64(length)); err != nil {
		return err
	}

	indexStart := int64(headerSize) + DeflateIndexReserve
	blockStart := indexStart + int64(maxBlocks*16)
	nextCompressed := blockStart
	buf := make([]byte, blockSize)
	for i := 0; i < usedBlocks; i++ {
		inputPos := int64(i) * int64(blockSize)
		readSize := blockSize
		if remaining := length - inputPos; remaining < int64(readSize) {
			readSize = int(remaining)
		}
		if readSize > 0 {
			if err := ReadAtFull(source, inputPos, buf[:readSize]); err != nil {
				return err
			}
		}
		compressed, err := DeflateData(buf[:readSize])
		if err != nil {
			return err
		}
		if err := WriteAtFull(output, nextCompressed, compressed); err != nil {
			return err
		}
		relativeStart := nextCompressed - blockStart
		relativeEnd := relativeStart + int64(len(compressed))
		indexPos := indexStart + int64(i*16)
		if err := WriteAtFull(output, indexPos, EncodeInt64(relativeStart)); err != nil {
			return err
		}
		if err := WriteAtFull(output, indexPos+8, EncodeInt64(relativeEnd)); err != nil {
			return err
		}
		nextCompressed += int64(len(compressed))
	}
	return nil
}

// DeflateHeader is the metadata stored at the beginning of a block Deflate
// stream. It is the Go equivalent of the constants exposed by
// DeflateBinaryStreamReader in Kotlin.
type DeflateHeader struct {
	BlockSize      int
	HeaderSize     int
	MaxBlocks      int
	Blocks         int
	OriginalLength int64
}

// ReadDeflateHeader reads and validates the block-deflate header.
func ReadDeflateHeader(source Reader) (DeflateHeader, error) {
	if source == nil {
		return DeflateHeader{}, ErrNoReader
	}
	if source.Size() < int64(DeflatePosOriginalLength+8) {
		return DeflateHeader{}, fmt.Errorf("%w: deflate header is truncated", ErrCorrupt)
	}
	raw, err := ReadBytesAt(source, DeflatePosBlockSize, 24)
	if err != nil {
		return DeflateHeader{}, err
	}
	blockSize, err := DecodeInt32(raw[0:4])
	if err != nil {
		return DeflateHeader{}, err
	}
	headerSize, err := DecodeInt32(raw[4:8])
	if err != nil {
		return DeflateHeader{}, err
	}
	maxBlocks, err := DecodeInt32(raw[8:12])
	if err != nil {
		return DeflateHeader{}, err
	}
	blocks, err := DecodeInt32(raw[12:16])
	if err != nil {
		return DeflateHeader{}, err
	}
	originalLength, err := DecodeInt64(raw[16:24])
	if err != nil {
		return DeflateHeader{}, err
	}
	if blockSize <= 0 || headerSize < 0 || maxBlocks < 0 || blocks < 0 || blocks > maxBlocks || originalLength < 0 {
		return DeflateHeader{}, fmt.Errorf("%w: invalid deflate header", ErrCorrupt)
	}
	return DeflateHeader{
		BlockSize:      int(blockSize),
		HeaderSize:     int(headerSize),
		MaxBlocks:      int(maxBlocks),
		Blocks:         int(blocks),
		OriginalLength: originalLength,
	}, nil
}

// DeflateReader presents the uncompressed content of a deflate block stream.
type DeflateReader struct {
	source         Reader
	blockSize      int
	headerSize     int
	maxBlocks      int
	blocksCount    int
	originalLength int64
	indexStart     int64
	blocksStart    int64
	cache          *Cache[int32, *deflateBlock]
}

type deflateBlock struct {
	index int32
	data  []byte
}

// NewDeflateReader reads the deflate header and returns a reader of the
// original (uncompressed) data. cacheMax is the number of decompressed blocks
// kept in memory; values <= 0 disable the cache.
func NewDeflateReader(source Reader, cacheMax int) (*DeflateReader, error) {
	header, err := ReadDeflateHeader(source)
	if err != nil {
		return nil, err
	}
	indexStart := int64(header.HeaderSize) + DeflateIndexReserve
	if int64(header.MaxBlocks) > (math.MaxInt64-indexStart)/16 {
		return nil, fmt.Errorf("%w: deflate index overflows int64", ErrCorrupt)
	}
	blocksStart := indexStart + int64(header.MaxBlocks)*16
	if blocksStart > source.Size() {
		return nil, fmt.Errorf("%w: deflate block area starts beyond the stream", ErrCorrupt)
	}
	if cacheMax < 0 {
		cacheMax = 0
	}
	return &DeflateReader{
		source:         source,
		blockSize:      header.BlockSize,
		headerSize:     header.HeaderSize,
		maxBlocks:      header.MaxBlocks,
		blocksCount:    header.Blocks,
		originalLength: header.OriginalLength,
		indexStart:     indexStart,
		blocksStart:    blocksStart,
		cache:          NewCache[int32, *deflateBlock](cacheMax),
	}, nil
}

// Size returns the original uncompressed length.
func (d *DeflateReader) Size() int64 { return d.originalLength }

// ReadAt implements Reader.
func (d *DeflateReader) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, ErrOutOfBounds
	}
	if len(p) == 0 {
		return 0, nil
	}
	if off >= d.originalLength {
		return 0, io.EOF
	}
	want := len(p)
	if remaining := d.originalLength - off; int64(want) > remaining {
		want = int(remaining)
	}
	total := 0
	for total < want {
		position := off + int64(total)
		index := int32(position / int64(d.blockSize))
		block, err := d.block(index)
		if err != nil {
			return total, err
		}
		local := position - int64(index)*int64(d.blockSize)
		if local < 0 || local >= int64(len(block.data)) {
			return total, fmt.Errorf("%w: deflate block %d does not cover position %d", ErrCorrupt, index, position)
		}
		n := copy(p[total:want], block.data[local:])
		if n == 0 {
			return total, fmt.Errorf("%w: empty deflate block %d", ErrCorrupt, index)
		}
		total += n
	}
	if want < len(p) {
		return total, io.EOF
	}
	return total, nil
}

func (d *DeflateReader) block(index int32) (*deflateBlock, error) {
	if index < 0 || int(index) >= d.blocksCount {
		return nil, fmt.Errorf("%w: deflate block index %d", ErrOutOfBounds, index)
	}
	if block, ok := d.cache.Get(index); ok {
		return block, nil
	}
	entryPos := d.indexStart + int64(index)*16
	entry, err := ReadBytesAt(d.source, entryPos, 16)
	if err != nil {
		return nil, err
	}
	relativeStart, err := DecodeInt64(entry[:8])
	if err != nil {
		return nil, err
	}
	relativeEnd, err := DecodeInt64(entry[8:])
	if err != nil {
		return nil, err
	}
	if relativeStart < 0 || relativeEnd < relativeStart {
		return nil, fmt.Errorf("%w: invalid deflate index for block %d", ErrCorrupt, index)
	}
	start := d.blocksStart + relativeStart
	end := d.blocksStart + relativeEnd
	if start < d.blocksStart || end < start || end > d.source.Size() {
		return nil, fmt.Errorf("%w: deflate block %d points outside the stream", ErrCorrupt, index)
	}
	compressedLen := end - start
	if compressedLen > int64(maxInt()) {
		return nil, ErrOutOfBounds
	}
	compressed, err := ReadBytesAt(d.source, start, int(compressedLen))
	if err != nil {
		return nil, err
	}
	inflated, err := InflateData(compressed)
	if err != nil {
		return nil, err
	}
	expected := int64(d.blockSize)
	if remaining := d.originalLength - int64(index)*int64(d.blockSize); remaining < expected {
		expected = remaining
	}
	if expected < 0 || int64(len(inflated)) != expected {
		return nil, fmt.Errorf("%w: deflate block %d has length %d, expected %d", ErrCorrupt, index, len(inflated), expected)
	}
	block := &deflateBlock{index: index, data: inflated}
	d.cache.Put(index, block)
	return block, nil
}

// InflateStream reads and decompresses an entire deflate stream.
func InflateStream(source Reader, cacheMax int) ([]byte, error) {
	reader, err := NewDeflateReader(source, cacheMax)
	if err != nil {
		return nil, err
	}
	return ReadAll(reader)
}
