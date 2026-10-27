// Package binarytable is a TinyGo-friendly Go port of the BinaryTableData
// Kotlin library.
//
// The package stores rows in a fixed-size schema. Every numeric value is
// encoded in big-endian order, exactly like the Kotlin implementation, so the
// files produced by either implementation can be read by the other.
//
// The original Kotlin library relies on kotlinx.serialization and reflection for
// automatic object mapping. TinyGo deliberately supports only a small part of
// the reflect package, therefore this port is schema-first: declare columns
// explicitly, then read and write values through typed helpers. See the
// README and the docs directory for examples.
package binarytable

import "errors"

var (
	// ErrReadOnly is returned when a write is attempted on a stream or table
	// that has no writer.
	ErrReadOnly = errors.New("binarytable: stream is read-only")

	// ErrOutOfBounds is returned when an index, offset, or length is outside
	// the valid range.
	ErrOutOfBounds = errors.New("binarytable: index out of bounds")

	// ErrOverSize is returned when a value does not fit in its fixed-size
	// column.
	ErrOverSize = errors.New("binarytable: value is larger than the column")

	// ErrCorrupt is returned when persisted data does not follow the
	// BinaryTableData format.
	ErrCorrupt = errors.New("binarytable: corrupt data")

	// ErrNoReader is returned when a table was created without a reader.
	ErrNoReader = errors.New("binarytable: table has no reader")
)
