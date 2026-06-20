package main

import (
	"os"

	"github.com/yottayoshida/intent-diff/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
