// Command binarytable is the command-line interface for the Go/TinyGo port of
// BinaryTableData.
package main

import (
	"os"

	"github.com/magedq1/BinaryTableData/tinygo/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
