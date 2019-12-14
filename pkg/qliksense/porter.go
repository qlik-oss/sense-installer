package qliksense

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CallPorter ...
func (p *Qliksense) CallPorter(args []string) error {
	var (
		cmd     *exec.Cmd
		err     error
		outText string
		stdout  bytes.Buffer
	)
	cmd = exec.Command(p.porterExe, strings.Join(args, " "))
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return err
	}
	outText = stdout.String()
	outText = strings.ReplaceAll(outText, "porter", "qliksense porter")
	fmt.Print(outText)
	return nil
}
