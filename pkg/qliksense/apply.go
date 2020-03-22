package qliksense

import (
	"io"
)

func (q *Qliksense) ApplyCRFromReader(r io.Reader) error {
	if err := q.LoadCr(r); err != nil {
		return err
	}
	opts := &InstallCommandOptions{}
	if err := q.InstallQK8s("", opts, true); err != nil {
		return err
	}
	return nil
}
