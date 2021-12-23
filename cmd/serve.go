package cmd

import (
	"context"

	"github.com/rubiojr/tavern/server"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the Tavern server",
	Run: func(cmd *cobra.Command, args []string) {
		s := server.NewServerWithConfig(&server.Config{UploadsPath: *path, Addr: *addr, CharmServerURL: *charmServerURL})
		s.Serve(context.Background())
	},
}

var path *string
var addr *string
var charmServerURL *string

func init() {
	rootCmd.AddCommand(serveCmd)
	path = serveCmd.Flags().StringP("path", "p", server.ServerDefaultUploadsPath, "Path where the files will be uploaded/served")
	addr = serveCmd.Flags().StringP("address", "a", server.ServerDefaultAddr, "Listening address")
	charmServerURL = serveCmd.Flags().StringP("charm-server-url", "", server.ServerDefaultCharmServerURL, "Charm server URL address")
}
