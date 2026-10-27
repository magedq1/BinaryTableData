package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/magedq1/BinaryTableData/tinygo/binarytable"
)

func runTable(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		tableHelp(stdout)
		return 0
	}
	switch args[0] {
	case "create":
		return tableCreate(args[1:], stdout, stderr)
	case "info":
		return tableInfo(args[1:], stdout, stderr)
	case "schema":
		return tableSchema(args[1:], stdout, stderr)
	case "get":
		return tableGet(args[1:], stdout, stderr)
	case "set":
		return tableSet(args[1:], stdin, stdout, stderr)
	case "dump":
		return tableDump(args[1:], stdout, stderr)
	case "search":
		return tableSearch(args[1:], stdout, stderr)
	default:
		return usageError(stderr, "unknown table command %q", args[0])
	}
}

func tableHelp(w io.Writer) {
	fmt.Fprint(w, `Usage: binarytable table <command> [flags]

Commands:
  create  Create a new fixed-capacity table file
  info    Show file, schema, and capacity information
  schema  Print the raw schema text
  get     Read one cell
  set     Write one cell
  dump    Print every column of a row
  search  Find the first row matching a value

Column specifications use name:type[:length]:
  int, long, double, string[:64], raw:length

Examples:
  binarytable table create -file users.db -rows 1000 \
      -col _id:int -col name:string:64 -col score:double

  binarytable table set -file users.db -row 0 -col name -value Alice
  binarytable table get -file users.db -row 0 -col name
  binarytable table dump -file users.db -row 0
  binarytable table search -file users.db -col name -value Alice
`)
}

func tableCreate(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("table create", stderr)
	file := fs.String("file", "", "output table file")
	rows := fs.Int("rows", -1, "maximum number of rows")
	force := fs.Bool("force", false, "overwrite the file if it exists")
	var columns columnList
	fs.Var(&columns, "col", "column spec name:type[:length], repeatable")
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), `Usage: binarytable table create -file PATH -rows N -col SPEC [-col SPEC ...] [-force]

Column specs:
  _id:int
  name:string:64
  score:double
  payload:raw:1024
`)
		fs.PrintDefaults()
	}
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *file == "" {
		return usageError(stderr, "table create needs -file")
	}
	if *rows < 0 {
		return usageError(stderr, "table create needs -rows >= 0")
	}
	if len(columns) == 0 {
		return usageError(stderr, "table create needs at least one -col")
	}
	cols, err := parseColumnSpecs(columns)
	if err != nil {
		return commandError(stderr, err)
	}
	header, err := binarytable.NewHeader(*rows, cols)
	if err != nil {
		return commandError(stderr, err)
	}
	if !*force {
		exists, err := fileExists(*file)
		if err != nil {
			return commandError(stderr, err)
		}
		if exists {
			return commandError(stderr, fmt.Errorf("%s already exists; use -force to overwrite", *file))
		}
	}
	table, err := binarytable.CreateFile(header, *file)
	if err != nil {
		return commandError(stderr, err)
	}
	if err := table.Close(); err != nil {
		return commandError(stderr, err)
	}
	fmt.Fprintf(stdout, "created %s: rows=%d row_size=%d schema_size=%d capacity=%d\n",
		*file, header.MaxRows(), header.MaxRowLength(), header.SchemaSize(), header.Capacity())
	return 0
}

func tableInfo(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("table info", stderr)
	file := fs.String("file", "", "table file")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *file == "" {
		return usageError(stderr, "table info needs -file")
	}
	table, err := binarytable.OpenFileReadOnly(*file)
	if err != nil {
		return commandError(stderr, err)
	}
	defer table.Close()
	info, err := os.Stat(*file)
	if err != nil {
		return commandError(stderr, err)
	}
	header := table.Header
	fmt.Fprintf(stdout, "File: %s\nSize: %d bytes\nRows: %d\nRow size: %d\nSchema size: %d\nCapacity: %d\nColumns: %d\n",
		*file, info.Size(), header.MaxRows(), header.MaxRowLength(), header.SchemaSize(), header.Capacity(), header.ColumnCount())
	for _, column := range header.Columns() {
		index, _ := header.ColumnIndex(column.Name)
		fmt.Fprintf(stdout, "  %-3d %-24s %-7s %d\n", index, column.Name, column.Type, column.Length)
	}
	return 0
}

func tableSchema(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("table schema", stderr)
	file := fs.String("file", "", "table file")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *file == "" {
		return usageError(stderr, "table schema needs -file")
	}
	table, err := binarytable.OpenFileReadOnly(*file)
	if err != nil {
		return commandError(stderr, err)
	}
	defer table.Close()
	_, err = stdout.Write(table.Header.SchemaBytes())
	if err != nil {
		return commandError(stderr, err)
	}
	return 0
}

func tableGet(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("table get", stderr)
	file := fs.String("file", "", "table file")
	row := fs.Int("row", -1, "row index")
	column := fs.String("col", "", "column name or zero-based index")
	hexMode := fs.Bool("hex", false, "print raw column bytes as hexadecimal")
	out := fs.String("out", "", "write raw bytes to this file, or - for stdout")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *file == "" || *row < 0 || *column == "" {
		return usageError(stderr, "table get needs -file, -row >= 0, and -col")
	}
	if *hexMode && *out != "" {
		return usageError(stderr, "table get: -hex and -out cannot be combined")
	}
	table, err := binarytable.OpenFileReadOnly(*file)
	if err != nil {
		return commandError(stderr, err)
	}
	defer table.Close()
	info, err := resolveColumn(table.Header, *column)
	if err != nil {
		return commandError(stderr, err)
	}
	if err := printCell(table, *row, info, *hexMode, *out, stdout); err != nil {
		return commandError(stderr, err)
	}
	return 0
}

func tableSet(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("table set", stderr)
	file := fs.String("file", "", "table file")
	row := fs.Int("row", -1, "row index")
	column := fs.String("col", "", "column name or zero-based index")
	value := fs.String("value", "", "typed value (string, int, long, double, or raw text)")
	valueHex := fs.String("value-hex", "", "raw bytes as hexadecimal")
	valueFile := fs.String("value-file", "", "read raw bytes from this file, or - for stdin")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *file == "" || *row < 0 || *column == "" {
		return usageError(stderr, "table set needs -file, -row >= 0, and -col")
	}
	visited := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) { visited[f.Name] = true })
	opts := valueOptions{
		value:    *value,
		valueSet: visited["value"],
		hex:      *valueHex,
		hexSet:   visited["value-hex"],
		file:     *valueFile,
		fileSet:  visited["value-file"],
	}
	if err := opts.requireOne(); err != nil {
		return usageError(stderr, "table set: %v", err)
	}
	table, err := binarytable.OpenFile(*file)
	if err != nil {
		return commandError(stderr, err)
	}
	defer table.Close()
	info, err := resolveColumn(table.Header, *column)
	if err != nil {
		return commandError(stderr, err)
	}
	if err := setCell(table, *row, info, opts, stdin); err != nil {
		return commandError(stderr, err)
	}
	fmt.Fprintf(stdout, "row=%d column=%s set\n", *row, info.Column.Name)
	return 0
}

func tableDump(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("table dump", stderr)
	file := fs.String("file", "", "table file")
	row := fs.Int("row", -1, "row index")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *file == "" || *row < 0 {
		return usageError(stderr, "table dump needs -file and -row >= 0")
	}
	table, err := binarytable.OpenFileReadOnly(*file)
	if err != nil {
		return commandError(stderr, err)
	}
	defer table.Close()
	for _, column := range table.Header.Columns() {
		info, _ := table.Header.ColumnInfoByName(column.Name)
		value, err := formatCell(table, *row, info)
		if err != nil {
			return commandError(stderr, err)
		}
		fmt.Fprintf(stdout, "%d\t%s\t%s\t%s\n", info.Index, info.Column.Name, info.Column.Type, value)
	}
	return 0
}

func tableSearch(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("table search", stderr)
	file := fs.String("file", "", "table file")
	column := fs.String("col", "", "column name or zero-based index")
	value := fs.String("value", "", "value to search for")
	start := fs.Int("start", 0, "first row to inspect")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *file == "" || *column == "" {
		return usageError(stderr, "table search needs -file, -col, and -value")
	}
	if *start < 0 {
		return usageError(stderr, "table search: -start must be >= 0")
	}
	table, err := binarytable.OpenFileReadOnly(*file)
	if err != nil {
		return commandError(stderr, err)
	}
	defer table.Close()
	info, err := resolveColumn(table.Header, *column)
	if err != nil {
		return commandError(stderr, err)
	}
	var row int
	switch info.Column.Type {
	case binarytable.TypeInt:
		v, err := strconv.ParseInt(*value, 10, 32)
		if err != nil {
			return commandError(stderr, fmt.Errorf("invalid INT value %q", *value))
		}
		row, err = table.IndexOfInt(int32(v), info.Index, *start)
		if err != nil {
			return commandError(stderr, err)
		}
	case binarytable.TypeLong:
		v, err := strconv.ParseInt(*value, 10, 64)
		if err != nil {
			return commandError(stderr, fmt.Errorf("invalid LONG value %q", *value))
		}
		row, err = table.IndexOfLong(v, info.Index, *start)
		if err != nil {
			return commandError(stderr, err)
		}
	case binarytable.TypeDouble:
		v, err := strconv.ParseFloat(*value, 64)
		if err != nil {
			return commandError(stderr, fmt.Errorf("invalid DOUBLE value %q", *value))
		}
		row, err = table.IndexOfDouble(v, info.Index, *start)
		if err != nil {
			return commandError(stderr, err)
		}
	case binarytable.TypeString:
		row, err = table.IndexOfString(*value, info.Index, *start)
		if err != nil {
			return commandError(stderr, err)
		}
	default:
		return commandError(stderr, fmt.Errorf("search is not supported for %s columns", info.Column.Type))
	}
	fmt.Fprintln(stdout, row)
	if row < 0 {
		return 1
	}
	return 0
}
