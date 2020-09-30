package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rancher/k3d/v3/cmd/cluster"
	"github.com/rancher/k3d/v3/cmd/image"
	"github.com/rancher/k3d/v3/cmd/kubeconfig"
	"github.com/rancher/k3d/v3/cmd/node"
	cliutil "github.com/rancher/k3d/v3/cmd/util"
	"github.com/rancher/k3d/v3/pkg/runtimes"
	"github.com/rancher/k3d/v3/version"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

// RootFlags describes a struct that holds flags that can be set on root level of the command
type K3dFlags struct {
	debugLogging bool
	version      bool
}

var flags = K3dFlags{}

// var cfgFile string

// kedCmd represents the base command when called without any subcommands
var k3dCmd = &cobra.Command{
	Use:   "k3d",
	Short: "https://k3d.io/ -> Run k3s in Docker!",
	Long: `https://k3d.io/
k3d is a wrapper CLI that helps you to easily create k3s clusters inside docker.
Nodes of a k3d cluster are docker containers running a k3s image.
All Nodes of a k3d cluster are part of the same docker network.`,
	Run: func(cmd *cobra.Command, args []string) {
		if flags.version {
			printVersion()
		} else {
			if err := cmd.Usage(); err != nil {
				log.Fatalln(err)
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the k3dCmd.
func Execute() {
	if len(os.Args) > 1 {
		parts := os.Args[1:]
		// Check if it's a built-in command, else try to execute it as a plugin
		if _, _, err := k3dCmd.Find(parts); err != nil {
			pluginFound, err := cliutil.HandlePlugin(context.Background(), parts)
			if err != nil {
				log.Errorf("Failed to execute plugin '%+v'", parts)
				log.Fatalln(err)
			} else if pluginFound {
				os.Exit(0)
			}
		}
	}
	if err := k3dCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func init() {
	cobra.OnInitialize(initLogging, initRuntime)

	// add persistent flags (present to all subcommands)
	// k3dCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.k3d/config.yaml)")
	k3dCmd.PersistentFlags().BoolVar(&flags.debugLogging, "verbose", false, "Enable verbose output (debug logging)")

	// add local flags
	k3dCmd.Flags().BoolVar(&flags.version, "version", false, "Show k3d and default k3s version")

	// add subcommands
	k3dCmd.AddCommand(cluster.NewCmdCluster())
	k3dCmd.AddCommand(kubeconfig.NewCmdKubeconfig())
	k3dCmd.AddCommand(node.NewCmdNode())
	k3dCmd.AddCommand(image.NewCmdImage())

	k3dCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show k3d and default k3s version",
		Long:  "Show k3d and default k3s version",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
		},
	})
}

// initLogging initializes the logger
func initLogging() {
	if flags.debugLogging {
		log.SetLevel(log.DebugLevel)
	} else {
		switch logLevel := strings.ToUpper(os.Getenv("LOG_LEVEL")); logLevel {
		case "DEBUG":
			log.SetLevel(log.DebugLevel)
		case "WARN":
			log.SetLevel(log.WarnLevel)
		case "ERROR":
			log.SetLevel(log.ErrorLevel)
		default:
			log.SetLevel(log.InfoLevel)
		}
	}
	log.SetOutput(ioutil.Discard)
	log.AddHook(&writer.Hook{
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
	})
	log.AddHook(&writer.Hook{
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.InfoLevel,
			log.DebugLevel,
		},
	})
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
}

func initRuntime() {
	runtime, err := runtimes.GetRuntime("docker")
	if err != nil {
		log.Fatalln(err)
	}
	runtimes.SelectedRuntime = runtime
	log.Debugf("Selected runtime is '%T'", runtimes.SelectedRuntime)
}

func printVersion() {
	fmt.Printf("k3d version %s\n", version.GetVersion())
	fmt.Printf("k3s version %s (default)\n", version.K3sVersion)
}
