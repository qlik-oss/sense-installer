package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	if tmpDir, err := ioutil.TempDir("", ""); err != nil {
		panic(err)
	} else {
		fmt.Print(tmpDir)
	}
}
