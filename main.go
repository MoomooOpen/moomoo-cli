package main

import (
	"os"

	"github.com/MoomooOpen/moomoo-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
