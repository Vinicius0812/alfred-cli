package main

import (
	"os"

	"github.com/Vinicius0812/alfred-cli/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
