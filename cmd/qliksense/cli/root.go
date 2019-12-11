package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/deislabs/porter/pkg/porter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func RootCmd() *cobra.Command {
	p := porter.New()
	cmd := &cobra.Command{
		Use:   "qliksense",
		Short: "qliksense cli tool",
		Long: `qliksense cli tool provides a wrapper around the porter api as well as
		provides addition functionality`,
		SilenceUsage: true,
	}

	cobra.OnInitialize(initConfig)

	cmd.AddCommand(installPorterMixin(p))

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	return cmd
}

func InitAndExecute() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.SetEnvPrefix("QLIKSENSE")
	viper.AutomaticEnv()
}
