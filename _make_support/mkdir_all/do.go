package main

import (
	"os"
)

func main() {
	if err := os.MkdirAll(os.Args[1], os.ModePerm); err != nil {
		panic(err)
	}
}
