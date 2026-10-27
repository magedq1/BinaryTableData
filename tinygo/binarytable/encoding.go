package binarytable

import (
	"fmt"
	"math"
)

// EncodeInt32 encodes v in big-endian order.
func EncodeInt32(v int32) []byte {
	return AppendInt32(make([]byte, 0, 4), v)
}

// EncodeInt64 encodes v in big-endian order.
func EncodeInt64(v int64) []byte {
	return AppendInt64(make([]byte, 0, 8), v)
}

// EncodeFloat64 encodes v in IEEE-754 big-endian order.
func EncodeFloat64(v float64) []byte {
	return AppendFloat64(make([]byte, 0, 8), v)
}

// AppendInt32 appends v in big-endian order to dst.
func AppendInt32(dst []byte, v int32) []byte {
	u := uint32(v)
	return append(dst, byte(u>>24), byte(u>>16), byte(u>>8), byte(u))
}

// AppendInt64 appends v in big-endian order to dst.
func AppendInt64(dst []byte, v int64) []byte {
	u := uint64(v)
	return append(dst,
		byte(u>>56), byte(u>>48), byte(u>>40), byte(u>>32),
		byte(u>>24), byte(u>>16), byte(u>>8), byte(u),
	)
}

// AppendFloat64 appends v in IEEE-754 big-endian order to dst.
func AppendFloat64(dst []byte, v float64) []byte {
	return AppendInt64(dst, int64(math.Float64bits(v)))
}

// DecodeInt32 decodes a big-endian int32 from b.
func DecodeInt32(b []byte) (int32, error) {
	if len(b) < 4 {
		return 0, fmt.Errorf("%w: int32 needs 4 bytes, got %d", ErrCorrupt, len(b))
	}
	return int32(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])), nil
}

// DecodeInt64 decodes a big-endian int64 from b.
func DecodeInt64(b []byte) (int64, error) {
	if len(b) < 8 {
		return 0, fmt.Errorf("%w: int64 needs 8 bytes, got %d", ErrCorrupt, len(b))
	}
	u := uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
	return int64(u), nil
}

// DecodeFloat64 decodes a big-endian IEEE-754 float64 from b.
func DecodeFloat64(b []byte) (float64, error) {
	u, err := DecodeInt64(b)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(uint64(u)), nil
}

// putInt32 writes v into dst, which must be at least 4 bytes long.
func putInt32(dst []byte, v int32) {
	u := uint32(v)
	dst[0], dst[1], dst[2], dst[3] = byte(u>>24), byte(u>>16), byte(u>>8), byte(u)
}

// putInt64 writes v into dst, which must be at least 8 bytes long.
func putInt64(dst []byte, v int64) {
	u := uint64(v)
	dst[0], dst[1], dst[2], dst[3] = byte(u>>56), byte(u>>48), byte(u>>40), byte(u>>32)
	dst[4], dst[5], dst[6], dst[7] = byte(u>>24), byte(u>>16), byte(u>>8), byte(u)
}
