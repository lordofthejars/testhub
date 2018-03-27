package main

import (
	"os"

	"github.com/lordofthejars/testhub/hub"
	"github.com/spf13/cobra"
)

var configuration *hub.Config

var RootCmd = &cobra.Command{
	Use:   "testhub",
	Short: "Interact with Test Hub",
}

func main() {
	var repositoryPath string

	var port int
	var configPath string

	var cmdStart = &cobra.Command{
		Use:   "start [options]",
		Short: "Start Test Hub server",
		Long:  `start is used to start Test Hub server to collect test data`,
		Run: func(cmd *cobra.Command, args []string) {
			configuration, error := hub.NewConfig(configPath)
			if error != nil {
				hub.Error("Fatal Error while reading configuration: %s", error.Error())
				os.Exit(-1)
			}

			// Update with content provided by CLI flags
			if port != 0 {
				configuration.Port = port
			}

			if len(repositoryPath) > 0 {
				configuration.Repository.Path = repositoryPath
			}

			hub.StartServer(configuration)
		},
	}

	cmdStart.Flags().IntVarP(&port, "port", "p", 0, "port to start Test Hub server")
	cmdStart.Flags().StringVarP(&configPath, "config", "c", "", "configuration file for Test Hub")
	cmdStart.Flags().StringVar(&repositoryPath, "repository.path", "", "Configures Test Hub to use disk repository to given path")

	RootCmd.AddCommand(cmdStart)

	if err := RootCmd.Execute(); err != nil {
		hub.Error(err.Error())
		os.Exit(1)
	}
}
