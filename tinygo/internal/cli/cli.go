// Package cli implements the binarytable command-line interface.
//
// The package is separate from cmd/binarytable so it can be tested without
// launching a process.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

// Version is the CLI version reported by "binarytable version".
const Version = "0.1.0"

// Run executes the CLI with the supplied arguments (excluding the program
// name). It returns the process exit code: 0 for success, 1 for runtime
// errors, and 2 for usage errors.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stdout)
		return 0
	}
	switch args[0] {
	case "help", "-h", "--help":
		usage(stdout)
		return 0
	case "version", "--version":
		fmt.Fprintf(stdout, "binarytable %s\n", Version)
		return 0
	case "table":
		return runTable(args[1:], stdin, stdout, stderr)
	case "list":
		return runList(args[1:], stdin, stdout, stderr)
	case "deflate":
		return runDeflate(args[1:], stdin, stdout, stderr)
	case "zlib":
		return runZlib(args[1:], stdin, stdout, stderr)
	case "aes":
		return runAES(args[1:], stdin, stdout, stderr)
	case "rsa":
		return runRSA(args[1:], stdin, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "binarytable: unknown command %q\n\n", args[0])
		usage(stderr)
		return 2
	}
}

func usage(w io.Writer) {
	fmt.Fprint(w, `binarytable - structured binary table data

Usage:
  binarytable <command> [flags]

Commands:
  table    Create, inspect, read, write, and search fixed-schema tables
  list     Store and retrieve variable-length blobs
  deflate  Pack and unpack the block deflate format
  zlib     Compress and decompress raw zlib data
  aes      AES-256-CBC encryption and decryption
  rsa      RSA keys, encryption, and SHA-256 signatures
  version  Print the CLI version
  help     Show this help

Examples:
  binarytable table create -file users.db -rows 1000 -col _id:int -col name:string:64
  binarytable table set -file users.db -row 0 -col name -value Alice
  binarytable table get -file users.db -row 0 -col name
  binarytable list add -file blobs.bin -in avatar.png
  binarytable deflate pack -in data.bin -out data.bin.deflate

Run "binarytable <command> help" for command-specific help.
`)
}

func newFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: binarytable %s [flags]\n", name)
		fs.PrintDefaults()
	}
	return fs
}

func parseFlags(fs *flag.FlagSet, args []string) (code int, ok bool) {
	err := fs.Parse(args)
	if err == nil {
		if fs.NArg() > 0 {
			fmt.Fprintf(fs.Output(), "binarytable: unexpected arguments: %v\n", fs.Args())
			return 2, false
		}
		return 0, true
	}
	if errors.Is(err, flag.ErrHelp) {
		return 0, false
	}
	return 2, false
}

func commandError(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "binarytable: %v\n", err)
	return 1
}

func usageError(stderr io.Writer, format string, args ...any) int {
	fmt.Fprintf(stderr, "binarytable: "+format+"\n", args...)
	return 2
}

// readInput reads path, or stdin when path is "-".
func readInput(path string, stdin io.Reader) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(stdin)
	}
	return os.ReadFile(path)
}

// writeOutput writes data to path, or stdout when path is "-".
func writeOutput(path string, stdout io.Writer, data []byte) error {
	if path == "-" {
		for len(data) > 0 {
			n, err := stdout.Write(data)
			if n > 0 {
				data = data[n:]
			}
			if err != nil {
				return err
			}
			if n == 0 {
				return io.ErrShortWrite
			}
		}
		return nil
	}
	return os.WriteFile(path, data, 0o666)
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
