package cmd

import (
	"os"

	"github.com/rubiojr/tavern/client"
	"github.com/spf13/cobra"
)

var url, charmHost *string
var charmHTTPPort, charmSSHPort *int

const defaultURL = "https://pub.rbel.co"

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish Charm FS files to a Tavern server",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		if *url == defaultURL && os.Getenv("TAVERN_SERVER_URL") != "" {
			*url = os.Getenv("TAVERN_SERVER_URL")
		}

		cfg := client.DefaultConfig()
		cfg.ServerURL = *url
		cfg.CharmServerHost = *charmHost
		cfg.CharmServerHTTPPort = *charmHTTPPort
		pc, err := client.NewClientWithConfig(cfg)
		if err != nil {
			return err
		}

		return pc.Publish(args[0])
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
	url = publishCmd.Flags().StringP("server-url", "s", defaultURL, "Tavern server URL")
	charmHost = publishCmd.Flags().StringP("charm-server-host", "", "cloud.charm.sh", "Charm server URL")
	charmHTTPPort = publishCmd.Flags().IntP("charm-server-http-port", "", 35354, "Charm server URL")
	charmSSHPort = publishCmd.Flags().IntP("charm-server-ssh-port", "", 35353, "Charm server URL")
}
