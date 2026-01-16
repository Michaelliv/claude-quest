//go:build !debug

package main

import (
	"fmt"
	"os"
)

func runStudio() {
	fmt.Println("Studio mode is not available in release builds.")
	fmt.Println("Build with: go build -tags debug -o cq . && ./cq studio")
	os.Exit(1)
}
