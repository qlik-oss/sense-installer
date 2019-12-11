package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Install() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "return version",
		Long:  `...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("installing")
			return nil
		},
	}

	viper.BindPFlags(cmd.Flags())

	return cmd
}
