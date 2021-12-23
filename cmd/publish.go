package cmd

import (
	"os"

	"github.com/rubiojr/tavern"
	"github.com/spf13/cobra"
)

var url *string

const defaultURL = "https://pub.rbel.co"

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish Charm FS files to a Tavern server",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		if *url == defaultURL && os.Getenv("TAVERN_SERVER_URL") != "" {
			*url = os.Getenv("TAVERN_SERVER_URL")
		}

		cfg := tavern.DefaultConfig()
		pc, err := tavern.NewClientWithConfig(cfg)
		if err != nil {
			panic(err)
		}

		return pc.Publish(args[0])
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
	url = publishCmd.Flags().StringP("server-url", "s", defaultURL, "Tavern server URL")
}
