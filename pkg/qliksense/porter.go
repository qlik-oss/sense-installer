package qliksense

import (
	"io"
	"fmt"
	"bufio"
	"os"
	"os/exec"
)
// ProcessLine ...
type ProcessLine func(string) *string
// CallPorter ...
func (p *Qliksense) CallPorter(args []string, processor ProcessLine) (string,error) {
	var (
		outText string
		cmd     *exec.Cmd
		err     error
		output io.ReadCloser
		scanner                             *bufio.Scanner
		done chan struct{}
	)
	cmd = exec.Command(p.porterExe,args[:]...)
	if output,err = cmd.StdoutPipe(); err !=nil {
		return "",err
	}
	cmd.Stderr = os.Stderr
	
	done = make(chan struct{})
	scanner = bufio.NewScanner(output)
	go func() {
		for scanner.Scan() {
			var text string
			var newText *string
			text = scanner.Text()
			if processor != nil {
				newText = processor(text)
				if newText != nil {
				  outText = outText + fmt.Sprintln(*newText)
				}
			} else {
			  outText = outText + fmt.Sprintln(text)
			}
		}
		done <- struct{}{}
	}()
	if err = cmd.Start(); err != nil {
		return "",err
	}
	<-done
	if err = cmd.Wait(); err != nil {
		return "",err
	}
	if err = scanner.Err(); err != nil {
		return "",err
	}

	return outText,nil
}
