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
	var cert string
	var key string
	var authUsersPath string
	var authSecret string

	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Start Test Hub server",
		Long:  `start is used to start Test Hub server to collect test data`,
		Run: func(cmd *cobra.Command, args []string) {
			configuration, error := hub.NewConfig(configPath)
			if error != nil {
				hub.Error("Fatal Error while reading configuration: %s", error.Error())
				os.Exit(-1)
			}

			// Override configuration with content provided by CLI flags
			if port != 0 {
				configuration.Port = port
			}

			if len(repositoryPath) > 0 {
				configuration.Repository.Path = repositoryPath
			}

			if len(cert) > 0 {
				configuration.Cert = cert
			}

			if len(key) > 0 {
				configuration.Key = key
			}

			if len(authUsersPath) > 0 {
				configuration.Authentication.UsersPath = authUsersPath
			}

			if len(authSecret) > 0 {
				configuration.Authentication.Secret = authSecret
			}

			hub.StartServer(configuration)
		},
	}

	cmdStart.Flags().IntVarP(&port, "port", "p", 0, "port to start Test Hub server")
	cmdStart.Flags().StringVarP(&configPath, "config", "c", "", "configuration file for Test Hub")
	cmdStart.Flags().StringVar(&repositoryPath, "repository.path", "", "configures Test Hub to use disk repository to given path")
	cmdStart.Flags().StringVar(&cert, "cert", "", "configures location of certificate file to use in https")
	cmdStart.Flags().StringVar(&key, "key", "", "configures location of key file to use in https")
	cmdStart.Flags().StringVarP(&authUsersPath, "authentication.userspath", "u", "", "configures the location of users file")
	cmdStart.Flags().StringVarP(&authSecret, "authentication.secret", "s", "", "configures the secret to create JWT token")

	RootCmd.AddCommand(cmdStart)

	if err := RootCmd.Execute(); err != nil {
		hub.Error(err.Error())
		os.Exit(1)
	}
}
