package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rancher/k3d/v3/cmd/cluster"
	"github.com/rancher/k3d/v3/cmd/image"
	"github.com/rancher/k3d/v3/cmd/kubeconfig"
	"github.com/rancher/k3d/v3/cmd/node"
	"github.com/rancher/k3d/v3/pkg/runtimes"
	"github.com/rancher/k3d/v3/version"
	log "github.com/sirupsen/logrus"
)

// RootFlags describes a struct that holds flags that can be set on root level of the command
type K3dFlags struct {
	debugLogging bool
	version      bool
}

var flags = K3dFlags{}

// var cfgFile string

func getK3dCmd() *cobra.Command {
	// kedCmd represents the base command when called without any subcommands
	var k3dCmd = &cobra.Command{
		Use:   "k3d",
		Short: "https://k3d.io/ -> Run k3s in Docker!",
		Long: `https://k3d.io/
k3d is a wrapper CLI that helps you to easily create k3s clusters inside docker.
Nodes of a k3d cluster are docker containers running a k3s image.
All Nodes of a k3d cluster are part of the same docker network.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GOOOOOOO %T", runtimes.SelectedRuntime)
			if flags.version {
				printVersion()
			} else {
				if err := cmd.Usage(); err != nil {
					log.Fatalln(err)
				}
			}
		},
	}
	// add subcommands
	k3dCmd.AddCommand(cluster.NewCmdCluster())
	k3dCmd.AddCommand(kubeconfig.NewCmdKubeconfig())
	k3dCmd.AddCommand(node.NewCmdNode())
	k3dCmd.AddCommand(image.NewCmdImage())
	return k3dCmd

}

func printVersion() {
	fmt.Printf("k3d version %s\n", version.GetVersion())
	fmt.Printf("k3s version %s (default)\n", version.K3sVersion)
}

func initRuntime() {
	runtime, err := runtimes.GetRuntime("docker")
	if err != nil {
		log.Fatalln(err)
	}
	runtimes.SelectedRuntime = runtime
	log.Debugf("Selected runtime is '%T'", runtimes.SelectedRuntime)
}
