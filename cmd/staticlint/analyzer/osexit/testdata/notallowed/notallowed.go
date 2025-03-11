package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("os exit test data")
	os.Exit(0) // want "os.Exit calls not allowed"
}

func osExitAllowed() {
	fmt.Println("os exit test data")
	os.Exit(0)
}
