package main

import (
	"mycheck/multichecker"
)

func main() {
	if err := multichecker.Run(); err != nil {
		panic(err)
	}
}
