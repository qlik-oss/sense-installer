package main

import (
	"os"
)

func main() {
	if err := os.RemoveAll(os.Args[1]); err != nil {
		panic(err)
	}
}
