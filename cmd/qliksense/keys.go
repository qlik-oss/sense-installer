package main

import (
	"fmt"

	"github.com/qlik-oss/sense-installer/pkg/qliksense"
	"github.com/spf13/cobra"
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "keys for qliksense",
}

func keysRotateCmd(q *qliksense.Qliksense) *cobra.Command {
	c := &cobra.Command{
		Use:   "rotate",
		Short: "Rotate qliksense application keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("deleting stored application keys")
			if err := q.DeleteKeysClusterBackup(); err != nil {
				return err
			}
			fmt.Println("next install will rotate all qliksense keys")
			return nil
		},
	}
	return c
}
