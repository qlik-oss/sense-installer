package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func main() {
	initAndExecute()
}

func init() {
	// initialize runtime for k3d
	fmt.Println("GOT INITIALIZE")
	cobra.OnInitialize(initLogging, initRuntime)
}
