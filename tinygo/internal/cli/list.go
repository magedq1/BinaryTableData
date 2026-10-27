package cli

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/magedq1/BinaryTableData/tinygo/binarytable"
)

func runList(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		listHelp(stdout)
		return 0
	}
	switch args[0] {
	case "create":
		return listCreate(args[1:], stdout, stderr)
	case "info":
		return listInfo(args[1:], stdout, stderr)
	case "add":
		return listAdd(args[1:], stdin, stdout, stderr)
	case "get":
		return listGet(args[1:], stdout, stderr)
	case "dump":
		return listDump(args[1:], stdout, stderr)
	default:
		return usageError(stderr, "unknown list command %q", args[0])
	}
}

func listHelp(w io.Writer) {
	fmt.Fprint(w, `Usage: binarytable list <command> [flags]

Commands:
  create  Create an empty list with an explicit capacity
  info    Show list size and capacity
  add     Append a blob from a file (or - for stdin)
  get     Read a blob by index
  dump    Print the index and size of every blob

The -max flag is the list capacity used when the file was created. For existing
files, omit -max (or pass 0) to recover the capacity from the reserved prefix or
the first data offset. A brand-new list defaults to 1000 when -max is omitted.
Index entries are 32-bit, so a list is limited to about 2 GiB.

Examples:
  binarytable list create -file blobs.bin -max 5000
  binarytable list info -file blobs.bin
  binarytable list add -file blobs.bin -in avatar.png
  binarytable list get -file blobs.bin -index 0 -out avatar.png
  binarytable list dump -file blobs.bin
`)
}

func listCreate(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("list create", stderr)
	file := fs.String("file", "", "list file")
	max := fs.Int("max", 1000, "maximum number of blobs")
	force := fs.Bool("force", false, "overwrite an existing file")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *file == "" || *max <= 0 {
		return usageError(stderr, "list create needs -file and -max > 0")
	}
	if !*force {
		if _, err := os.Stat(*file); err == nil {
			return commandError(stderr, fmt.Errorf("%s already exists; use -force to overwrite", *file))
		} else if !os.IsNotExist(err) {
			return commandError(stderr, err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(*file), 0o755); err != nil {
		return commandError(stderr, err)
	}
	stream, err := binarytable.CreateFileStream(*file, 0, binarytable.ReadWrite)
	if err != nil {
		return commandError(stderr, err)
	}
	list, err := binarytable.NewBinaryList(stream, *max)
	if err != nil {
		_ = stream.Close()
		return commandError(stderr, err)
	}
	if err := list.Initialize(); err != nil {
		_ = stream.Close()
		return commandError(stderr, err)
	}
	if err := stream.Sync(); err != nil {
		_ = stream.Close()
		return commandError(stderr, err)
	}
	if err := stream.Close(); err != nil {
		return commandError(stderr, err)
	}
	fmt.Fprintf(stdout, "created %s: max_entries=%d prefix=%d bytes\n", *file, *max, list.SchemaSize()+list.IndexSize())
	return 0
}

func listInfo(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("list info", stderr)
	file := fs.String("file", "", "list file")
	max := fs.Int("max", 0, "list capacity; 0 recovers it from the file")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *file == "" {
		return usageError(stderr, "list info needs -file")
	}
	stream, err := binarytable.OpenFileStream(*file, binarytable.ReadOnly)
	if err != nil {
		return commandError(stderr, err)
	}
	defer stream.Close()
	resolved, err := resolveListMax(stream, *max)
	if err != nil {
		return commandError(stderr, err)
	}
	list, err := binarytable.NewBinaryList(stream, resolved)
	if err != nil {
		return commandError(stderr, err)
	}
	info, err := os.Stat(*file)
	if err != nil {
		return commandError(stderr, err)
	}
	fmt.Fprintf(stdout, "File: %s\nSize: %d bytes\nEntries: %d\nMax entries: %d\nSchema size: %d\nIndex size: %d\n",
		*file, info.Size(), list.Size(), list.MaxSize(), list.SchemaSize(), list.IndexSize())
	return 0
}

func listAdd(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("list add", stderr)
	file := fs.String("file", "", "list file")
	in := fs.String("in", "", "blob file, or - for stdin")
	max := fs.Int("max", 0, "list capacity; 0 recovers it from the file")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *file == "" || *in == "" {
		return usageError(stderr, "list add needs -file and -in")
	}
	data, err := readInput(*in, stdin)
	if err != nil {
		return commandError(stderr, err)
	}
	stream, err := binarytable.OpenOrCreateFileStream(*file)
	if err != nil {
		return commandError(stderr, err)
	}
	defer stream.Close()
	resolved, err := resolveListMax(stream, *max)
	if err != nil {
		return commandError(stderr, err)
	}
	list, err := binarytable.NewBinaryList(stream, resolved)
	if err != nil {
		return commandError(stderr, err)
	}
	index, err := list.Add(data)
	if err != nil {
		return commandError(stderr, err)
	}
	fmt.Fprintf(stdout, "index=%d size=%d\n", index, len(data))
	return 0
}

func listGet(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("list get", stderr)
	file := fs.String("file", "", "list file")
	index := fs.Int("index", -1, "blob index")
	max := fs.Int("max", 0, "list capacity; 0 recovers it from the file")
	hexMode := fs.Bool("hex", false, "print the blob as hexadecimal")
	out := fs.String("out", "", "write the blob to this file, or - for stdout")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *file == "" || *index < 0 {
		return usageError(stderr, "list get needs -file and -index >= 0")
	}
	if *hexMode && *out != "" {
		return usageError(stderr, "list get: -hex and -out cannot be combined")
	}
	stream, err := binarytable.OpenFileStream(*file, binarytable.ReadOnly)
	if err != nil {
		return commandError(stderr, err)
	}
	defer stream.Close()
	resolved, err := resolveListMax(stream, *max)
	if err != nil {
		return commandError(stderr, err)
	}
	list, err := binarytable.NewBinaryList(stream, resolved)
	if err != nil {
		return commandError(stderr, err)
	}
	data, err := list.Bytes(*index)
	if err != nil {
		return commandError(stderr, err)
	}
	if *hexMode {
		if _, err := fmt.Fprintln(stdout, hex.EncodeToString(data)); err != nil {
			return commandError(stderr, err)
		}
		return 0
	}
	target := *out
	if target == "" {
		target = "-"
	}
	if err := writeOutput(target, stdout, data); err != nil {
		return commandError(stderr, err)
	}
	return 0
}

func listDump(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("list dump", stderr)
	file := fs.String("file", "", "list file")
	max := fs.Int("max", 0, "list capacity; 0 recovers it from the file")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *file == "" {
		return usageError(stderr, "list dump needs -file")
	}
	stream, err := binarytable.OpenFileStream(*file, binarytable.ReadOnly)
	if err != nil {
		return commandError(stderr, err)
	}
	defer stream.Close()
	resolved, err := resolveListMax(stream, *max)
	if err != nil {
		return commandError(stderr, err)
	}
	list, err := binarytable.NewBinaryList(stream, resolved)
	if err != nil {
		return commandError(stderr, err)
	}
	for i := 0; i < list.Size(); i++ {
		reader, err := list.Reader(i)
		if err != nil {
			return commandError(stderr, err)
		}
		fmt.Fprintf(stdout, "%d\t%d\n", i, reader.Size())
	}
	return 0
}

func resolveListMax(reader binarytable.Reader, requested int) (int, error) {
	if requested > 0 {
		return requested, nil
	}
	guessed, err := binarytable.GuessListCapacity(reader)
	if err != nil {
		return 0, err
	}
	if guessed <= 0 {
		return 1000, nil
	}
	return guessed, nil
}
