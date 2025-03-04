package allowed

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("os exit test data")
	os.Exit(0)
}

func osExitAllowed() {
	fmt.Println("os exit test data")
	os.Exit(0)
}
