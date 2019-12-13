package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func porter(q *qlikSenseCmd) *cobra.Command {
	var (
		cobCmd  *cobra.Command
		cmd     *exec.Cmd
		err     error
		outText string
		stdout  bytes.Buffer
	)
	cobCmd = &cobra.Command{
		Use:   "porter",
		Short: "Execute a porter command",
		RunE: func(cobCmd *cobra.Command, args []string) error {
			cmd = exec.Command(q.porterExe, strings.Join(args, " "))
			cmd.Stdout = &stdout
			cmd.Stderr = os.Stderr
			if err = cmd.Run(); err != nil {
				return err
			}
			outText = stdout.String()
			outText = strings.ReplaceAll(outText, "porter", "qliksense porter")
			fmt.Print(outText)
			return nil
		},
	}
	return cobCmd
}
