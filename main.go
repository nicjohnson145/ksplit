package main

import (
	"fmt"
	"os"

	"github.com/nicjohnson145/ksplit/cmd"
)

// Version info set by goreleaser
var (
	version = "development"
	date    = "unknown"
)

func main() {
	if err := cmd.Root().Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
