package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/magedq1/BinaryTableData/tinygo/binarytable"
)

func runDeflate(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		deflateHelp(stdout)
		return 0
	}
	switch args[0] {
	case "pack", "write":
		return deflatePack(args[1:], stdin, stdout, stderr)
	case "unpack", "read":
		return deflateUnpack(args[1:], stdin, stdout, stderr)
	case "info":
		return deflateInfo(args[1:], stdin, stdout, stderr)
	default:
		return usageError(stderr, "unknown deflate command %q", args[0])
	}
}

func deflateHelp(w io.Writer) {
	fmt.Fprint(w, `Usage: binarytable deflate <command> [flags]

Commands:
  pack    Compress a file using the Kotlin-compatible block format
  write   Alias for pack
  unpack  Read a block-deflate file (DeflateBinaryStreamReader equivalent)
  read    Alias for unpack
  info    Print the block-deflate header and sizes

The block format reserves 512 index bytes and is compatible with
DeflateBinaryStreamReader/Writer in the Kotlin library.

Examples:
  binarytable deflate pack -in data.bin -out data.deflate -block 8192
  binarytable deflate unpack -in data.deflate -out data.bin -cache 8
  binarytable deflate info -in data.deflate
  binarytable deflate read -in data.deflate -out - | sha256sum
`)
}

func deflatePack(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("deflate pack", stderr)
	in := fs.String("in", "", "input file, or - for stdin")
	out := fs.String("out", "", "output file, or - for stdout")
	block := fs.Int("block", binarytable.DefaultDeflateBlockSize, "block size in bytes")
	header := fs.Int("header", binarytable.DefaultDeflateHeaderSize, "reserved caller header size")
	force := fs.Bool("force", false, "overwrite the output file")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *in == "" || *out == "" {
		return usageError(stderr, "deflate pack needs -in and -out")
	}
	if err := rejectSameFile(*in, *out); err != nil {
		return commandError(stderr, err)
	}
	source, closeSource, err := openReader(*in, stdin)
	if err != nil {
		return commandError(stderr, err)
	}
	if closeSource != nil {
		defer closeSource()
	}

	status := stdout
	if *out == "-" {
		status = stderr
	}
	if *out == "-" {
		output := binarytable.NewAutoSizeStream(4096)
		if err := binarytable.Deflate(source, output, *block, *header); err != nil {
			return commandError(stderr, err)
		}
		if err := writeOutput("-", stdout, output.ToBytes()); err != nil {
			return commandError(stderr, err)
		}
		fmt.Fprintf(status, "packed %d bytes into %d bytes\n", source.Size(), output.Size())
		return 0
	}
	if !*force {
		if exists, err := fileExists(*out); err != nil {
			return commandError(stderr, err)
		} else if exists {
			return commandError(stderr, fmt.Errorf("%s already exists; use -force to overwrite", *out))
		}
	}
	output, err := binarytable.CreateFileStream(*out, 0, binarytable.ReadWrite)
	if err != nil {
		return commandError(stderr, err)
	}
	defer output.Close()
	if err := binarytable.Deflate(source, output, *block, *header); err != nil {
		return commandError(stderr, err)
	}
	if err := output.Sync(); err != nil {
		return commandError(stderr, err)
	}
	fmt.Fprintf(status, "packed %d bytes into %d bytes\n", source.Size(), output.Size())
	return 0
}

func deflateUnpack(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("deflate unpack", stderr)
	in := fs.String("in", "", "input file, or - for stdin")
	out := fs.String("out", "", "output file, or - for stdout")
	cache := fs.Int("cache", 8, "number of decompressed blocks to cache")
	force := fs.Bool("force", false, "overwrite the output file")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *in == "" || *out == "" {
		return usageError(stderr, "deflate unpack needs -in and -out")
	}
	if err := rejectSameFile(*in, *out); err != nil {
		return commandError(stderr, err)
	}
	source, closeSource, err := openReader(*in, stdin)
	if err != nil {
		return commandError(stderr, err)
	}
	if closeSource != nil {
		defer closeSource()
	}
	reader, err := binarytable.NewDeflateReader(source, *cache)
	if err != nil {
		return commandError(stderr, err)
	}
	status := stdout
	if *out == "-" {
		status = stderr
	}
	if *out == "-" {
		data, err := binarytable.ReadAll(reader)
		if err != nil {
			return commandError(stderr, err)
		}
		if err := writeOutput("-", stdout, data); err != nil {
			return commandError(stderr, err)
		}
		fmt.Fprintf(status, "unpacked %d bytes\n", reader.Size())
		return 0
	}
	if !*force {
		if exists, err := fileExists(*out); err != nil {
			return commandError(stderr, err)
		} else if exists {
			return commandError(stderr, fmt.Errorf("%s already exists; use -force to overwrite", *out))
		}
	}
	output, err := binarytable.CreateFileStream(*out, reader.Size(), binarytable.ReadWrite)
	if err != nil {
		return commandError(stderr, err)
	}
	defer output.Close()
	if err := binarytable.CopyAt(output, 0, reader, 0, reader.Size()); err != nil {
		return commandError(stderr, err)
	}
	if err := output.Sync(); err != nil {
		return commandError(stderr, err)
	}
	fmt.Fprintf(status, "unpacked %d bytes\n", reader.Size())
	return 0
}

func deflateInfo(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("deflate info", stderr)
	in := fs.String("in", "", "input file, or - for stdin")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *in == "" {
		return usageError(stderr, "deflate info needs -in")
	}
	source, closeSource, err := openReader(*in, stdin)
	if err != nil {
		return commandError(stderr, err)
	}
	if closeSource != nil {
		defer closeSource()
	}
	header, err := binarytable.ReadDeflateHeader(source)
	if err != nil {
		return commandError(stderr, err)
	}
	ratio := 0.0
	if header.OriginalLength > 0 {
		ratio = float64(source.Size()) / float64(header.OriginalLength)
	}
	fmt.Fprintf(stdout, "File: %s\nCompressed size: %d bytes\nOriginal size: %d bytes\nRatio: %.3f\nBlock size: %d bytes\nCaller header: %d bytes\nBlocks: %d / %d\n",
		*in, source.Size(), header.OriginalLength, ratio, header.BlockSize, header.HeaderSize, header.Blocks, header.MaxBlocks)
	return 0
}

func runZlib(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(stdout, `Usage: binarytable zlib <command> [flags]

Commands:
  compress    Compress bytes with zlib
  decompress  Decompress zlib bytes

Both commands accept -in and -out, where - means stdin or stdout.

Examples:
  binarytable zlib compress -in data.bin -out data.zlib
  binarytable zlib decompress -in data.zlib -out data.bin
`)
		return 0
	}
	switch args[0] {
	case "compress":
		return zlibCompress(args[1:], stdin, stdout, stderr)
	case "decompress":
		return zlibDecompress(args[1:], stdin, stdout, stderr)
	default:
		return usageError(stderr, "unknown zlib command %q", args[0])
	}
}

func zlibCompress(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("zlib compress", stderr)
	in := fs.String("in", "", "input file, or - for stdin")
	out := fs.String("out", "", "output file, or - for stdout")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *in == "" || *out == "" {
		return usageError(stderr, "zlib compress needs -in and -out")
	}
	data, err := readInput(*in, stdin)
	if err != nil {
		return commandError(stderr, err)
	}
	compressed, err := binarytable.DeflateData(data)
	if err != nil {
		return commandError(stderr, err)
	}
	if err := writeOutput(*out, stdout, compressed); err != nil {
		return commandError(stderr, err)
	}
	status := stdout
	if *out == "-" {
		status = stderr
	}
	fmt.Fprintf(status, "compressed %d bytes into %d bytes\n", len(data), len(compressed))
	return 0
}

func zlibDecompress(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := newFlagSet("zlib decompress", stderr)
	in := fs.String("in", "", "input file, or - for stdin")
	out := fs.String("out", "", "output file, or - for stdout")
	if code, ok := parseFlags(fs, args); !ok {
		return code
	}
	if *in == "" || *out == "" {
		return usageError(stderr, "zlib decompress needs -in and -out")
	}
	data, err := readInput(*in, stdin)
	if err != nil {
		return commandError(stderr, err)
	}
	plain, err := binarytable.InflateData(data)
	if err != nil {
		return commandError(stderr, err)
	}
	if err := writeOutput(*out, stdout, plain); err != nil {
		return commandError(stderr, err)
	}
	status := stdout
	if *out == "-" {
		status = stderr
	}
	fmt.Fprintf(status, "decompressed %d bytes into %d bytes\n", len(data), len(plain))
	return 0
}

func openReader(path string, stdin io.Reader) (binarytable.Reader, func() error, error) {
	if path == "-" {
		data, err := io.ReadAll(stdin)
		if err != nil {
			return nil, nil, err
		}
		return binarytable.NewByteArrayReader(data), nil, nil
	}
	stream, err := binarytable.OpenFileStream(path, binarytable.ReadOnly)
	if err != nil {
		return nil, nil, err
	}
	return stream, stream.Close, nil
}

func rejectSameFile(in, out string) error {
	if in == "-" || out == "-" {
		return nil
	}
	if filepath.Clean(in) == filepath.Clean(out) {
		return fmt.Errorf("input and output must be different files")
	}
	if inInfo, err := os.Stat(in); err == nil {
		if outInfo, err := os.Stat(out); err == nil && os.SameFile(inInfo, outInfo) {
			return fmt.Errorf("input and output refer to the same file")
		}
	}
	return nil
}
