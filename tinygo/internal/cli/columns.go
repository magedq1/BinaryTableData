package cli

import (
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/magedq1/BinaryTableData/tinygo/binarytable"
)

type columnList []string

func (c *columnList) String() string { return strings.Join(*c, ",") }

func (c *columnList) Set(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("empty column specification")
	}
	*c = append(*c, value)
	return nil
}

// parseColumnSpec parses "name:type" or "name:type:length".
func parseColumnSpec(spec string) (binarytable.Column, error) {
	parts := strings.Split(spec, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return binarytable.Column{}, fmt.Errorf("invalid column %q: use name:type[:length]", spec)
	}
	name := parts[0]
	if name == "" {
		return binarytable.Column{}, fmt.Errorf("invalid column %q: empty name", spec)
	}
	if strings.ContainsAny(name, "/\n\r") {
		return binarytable.Column{}, fmt.Errorf("invalid column name %q", name)
	}
	kind := strings.ToLower(parts[1])
	length := 0
	if len(parts) == 3 {
		v, err := strconv.Atoi(parts[2])
		if err != nil || v <= 0 {
			return binarytable.Column{}, fmt.Errorf("invalid length in column %q", spec)
		}
		length = v
	}
	switch kind {
	case "int", "int32", "i32":
		if length != 0 && length != 4 {
			return binarytable.Column{}, fmt.Errorf("INT column %q must use length 4", name)
		}
		return binarytable.IntColumn(name), nil
	case "long", "int64", "i64":
		if length != 0 && length != 8 {
			return binarytable.Column{}, fmt.Errorf("LONG column %q must use length 8", name)
		}
		return binarytable.LongColumn(name), nil
	case "double", "float64", "f64":
		if length != 0 && length != 8 {
			return binarytable.Column{}, fmt.Errorf("DOUBLE column %q must use length 8", name)
		}
		return binarytable.DoubleColumn(name), nil
	case "string", "str", "text":
		if length == 0 {
			length = 64
		}
		return binarytable.StringColumn(name, length), nil
	case "raw", "bytes", "blob":
		if length == 0 {
			return binarytable.Column{}, fmt.Errorf("RAW column %q needs an explicit length, for example %s:raw:1024", name, name)
		}
		return binarytable.RawColumn(name, length), nil
	default:
		return binarytable.Column{}, fmt.Errorf("unknown column type %q in %q", parts[1], spec)
	}
}

func parseColumnSpecs(specs []string) ([]binarytable.Column, error) {
	columns := make([]binarytable.Column, 0, len(specs))
	for _, spec := range specs {
		column, err := parseColumnSpec(spec)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	return columns, nil
}

func resolveColumn(header *binarytable.Header, spec string) (binarytable.ColumnInfo, error) {
	if index, err := strconv.Atoi(spec); err == nil {
		info, ok := header.ColumnInfo(index)
		if !ok {
			return binarytable.ColumnInfo{}, fmt.Errorf("column index %d is out of range", index)
		}
		return info, nil
	}
	info, ok := header.ColumnInfoByName(spec)
	if !ok {
		return binarytable.ColumnInfo{}, fmt.Errorf("unknown column %q", spec)
	}
	return info, nil
}

type valueOptions struct {
	value    string
	valueSet bool
	hex      string
	hexSet   bool
	file     string
	fileSet  bool
}

func (o valueOptions) count() int {
	n := 0
	if o.valueSet {
		n++
	}
	if o.hexSet {
		n++
	}
	if o.fileSet {
		n++
	}
	return n
}

func (o valueOptions) rawBytes(stdin io.Reader) ([]byte, error) {
	if o.hexSet {
		text := strings.ReplaceAll(strings.TrimSpace(o.hex), " ", "")
		text = strings.TrimPrefix(text, "0x")
		return hex.DecodeString(text)
	}
	if o.fileSet {
		return readInput(o.file, stdin)
	}
	return nil, fmt.Errorf("no raw value source")
}

func (o valueOptions) requireOne() error {
	if o.count() != 1 {
		return fmt.Errorf("specify exactly one of -value, -value-hex, or -value-file")
	}
	return nil
}

func setCell(table *binarytable.Table, row int, info binarytable.ColumnInfo, opts valueOptions, stdin io.Reader) error {
	if err := opts.requireOne(); err != nil {
		return err
	}
	if opts.hexSet || opts.fileSet {
		data, err := opts.rawBytes(stdin)
		if err != nil {
			return err
		}
		return table.SetValue(row, info.Index, data)
	}
	switch info.Column.Type {
	case binarytable.TypeInt:
		v, err := strconv.ParseInt(opts.value, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid INT value %q", opts.value)
		}
		return table.SetInt(row, info.Index, int32(v))
	case binarytable.TypeLong:
		v, err := strconv.ParseInt(opts.value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid LONG value %q", opts.value)
		}
		return table.SetLong(row, info.Index, v)
	case binarytable.TypeDouble:
		v, err := strconv.ParseFloat(opts.value, 64)
		if err != nil {
			return fmt.Errorf("invalid DOUBLE value %q", opts.value)
		}
		return table.SetDouble(row, info.Index, v)
	case binarytable.TypeString:
		return table.SetString(row, info.Index, opts.value)
	case binarytable.TypeRaw:
		return table.SetValue(row, info.Index, []byte(opts.value))
	default:
		return fmt.Errorf("unsupported column type %s", info.Column.Type)
	}
}

func formatCell(table *binarytable.Table, row int, info binarytable.ColumnInfo) (string, error) {
	switch info.Column.Type {
	case binarytable.TypeInt:
		v, err := table.GetInt(row, info.Index)
		if err != nil {
			return "", err
		}
		return strconv.FormatInt(int64(v), 10), nil
	case binarytable.TypeLong:
		v, err := table.GetLong(row, info.Index)
		if err != nil {
			return "", err
		}
		return strconv.FormatInt(v, 10), nil
	case binarytable.TypeDouble:
		v, err := table.GetDouble(row, info.Index)
		if err != nil {
			return "", err
		}
		return strconv.FormatFloat(v, 'g', -1, 64), nil
	case binarytable.TypeString:
		v, err := table.GetString(row, info.Index)
		if err != nil {
			return "", err
		}
		return strconv.Quote(v), nil
	case binarytable.TypeRaw:
		v, err := table.GetValue(row, info.Index)
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(v), nil
	default:
		return "", fmt.Errorf("unsupported column type %s", info.Column.Type)
	}
}

func printCell(table *binarytable.Table, row int, info binarytable.ColumnInfo, hexMode bool, out string, stdout io.Writer) error {
	if hexMode || info.Column.Type == binarytable.TypeRaw && out == "" {
		raw, err := table.GetValue(row, info.Index)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, hex.EncodeToString(raw))
		return err
	}
	if out != "" {
		raw, err := table.GetValue(row, info.Index)
		if err != nil {
			return err
		}
		return writeOutput(out, stdout, raw)
	}
	switch info.Column.Type {
	case binarytable.TypeInt:
		v, err := table.GetInt(row, info.Index)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, v)
		return err
	case binarytable.TypeLong:
		v, err := table.GetLong(row, info.Index)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, v)
		return err
	case binarytable.TypeDouble:
		v, err := table.GetDouble(row, info.Index)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, strconv.FormatFloat(v, 'g', -1, 64))
		return err
	case binarytable.TypeString:
		v, err := table.GetString(row, info.Index)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, v)
		return err
	default:
		return fmt.Errorf("unsupported column type %s", info.Column.Type)
	}
}
