package binarytable

import (
	"fmt"
	"strconv"
	"strings"
)

// ColumnType is the logical type of a column.
type ColumnType uint8

const (
	// TypeRaw is an opaque fixed-size byte column.
	TypeRaw ColumnType = iota
	// TypeInt is a signed 32-bit integer stored in 4 bytes.
	TypeInt
	// TypeLong is a signed 64-bit integer stored in 8 bytes.
	TypeLong
	// TypeDouble is a 64-bit IEEE-754 float stored in 8 bytes.
	TypeDouble
	// TypeString is a UTF-8 string stored in a fixed-size byte column.
	TypeString
)

// String returns the wire name used in the table schema.
func (t ColumnType) String() string {
	switch t {
	case TypeRaw:
		return "RAW"
	case TypeInt:
		return "INT"
	case TypeLong:
		return "LONG"
	case TypeDouble:
		return "DOUBLE"
	case TypeString:
		return "STRING"
	default:
		return "UNKNOWN"
	}
}

func parseColumnType(s string) (ColumnType, error) {
	switch s {
	case "RAW":
		return TypeRaw, nil
	case "INT":
		return TypeInt, nil
	case "LONG":
		return TypeLong, nil
	case "DOUBLE":
		return TypeDouble, nil
	case "STRING":
		return TypeString, nil
	default:
		return TypeRaw, fmt.Errorf("%w: unknown column type %q", ErrCorrupt, s)
	}
}

// Column is one fixed-size field in a table schema.
type Column struct {
	Name   string
	Length int
	Type   ColumnType
}

// RawColumn creates a RAW column of length bytes.
func RawColumn(name string, length int) Column {
	return Column{Name: name, Length: length, Type: TypeRaw}
}

// IntColumn creates a 4-byte INT column.
func IntColumn(name string) Column {
	return Column{Name: name, Length: 4, Type: TypeInt}
}

// LongColumn creates an 8-byte LONG column.
func LongColumn(name string) Column {
	return Column{Name: name, Length: 8, Type: TypeLong}
}

// DoubleColumn creates an 8-byte DOUBLE column.
func DoubleColumn(name string) Column {
	return Column{Name: name, Length: 8, Type: TypeDouble}
}

// StringColumn creates a fixed-size STRING column. Use a length that can hold
// the longest UTF-8 representation of the values you intend to store.
func StringColumn(name string, length int) Column {
	return Column{Name: name, Length: length, Type: TypeString}
}

// encodeColumns serializes the schema exactly like the Kotlin implementation:
//
//	name/length/TYPE\n
func encodeColumns(columns []Column) []byte {
	var b strings.Builder
	for _, c := range columns {
		b.WriteString(c.Name)
		b.WriteByte('/')
		b.WriteString(strconv.Itoa(c.Length))
		b.WriteByte('/')
		b.WriteString(c.Type.String())
		b.WriteByte('\n')
	}
	return []byte(b.String())
}

func parseColumns(schema []byte) ([]Column, error) {
	text := string(schema)
	if text == "" {
		return nil, nil
	}
	// Accept LF, CRLF, and CR line endings, like Kotlin's readLines().
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	columns := make([]Column, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "/", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("%w: invalid schema line %q", ErrCorrupt, line)
		}
		length, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("%w: invalid length in schema line %q", ErrCorrupt, line)
		}
		columnType, err := parseColumnType(parts[2])
		if err != nil {
			return nil, err
		}
		columns = append(columns, Column{Name: parts[0], Length: length, Type: columnType})
	}
	return columns, nil
}
