package cmd

import (
	"context"

	"github.com/rubiojr/tavern"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the Tavern server",
	Run: func(cmd *cobra.Command, args []string) {
		s := tavern.NewServerWithConfig(&tavern.Config{UploadsPath: *path, Addr: *addr, CharmServerURL: *charmServerURL})
		s.Serve(context.Background())
	},
}

var path *string
var addr *string
var charmServerURL *string

func init() {
	rootCmd.AddCommand(serveCmd)
	path = serveCmd.Flags().StringP("path", "p", tavern.ServerDefaultUploadsPath, "Path where the files will be uploaded/served")
	addr = serveCmd.Flags().StringP("address", "a", tavern.ServerDefaultAddr, "Listening address")
	charmServerURL = serveCmd.Flags().StringP("charm-server-url", "", tavern.ServerDefaultCharmServerURL, "Charm server URL address")
}
