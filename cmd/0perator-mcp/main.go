package main

import (
	"fmt"
	"os"

	"github.com/akulkarni/0perator/internal/cli"
)

func main() {
	if err := cli.MCPStart(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
