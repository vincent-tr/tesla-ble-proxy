package cmd

import (
	"tesla-ble-proxy/pkg/web"

	"github.com/spf13/cobra"
)

var address string
var debug bool

func init() {

	cmd := &cobra.Command{
		Use:   "web",
		Short: "Starts the web server",
		Run: func(_ *cobra.Command, _ []string) {

			web.Serve(address, debug)
		},
	}

	cmd.Flags().StringVarP(&address, "address", "a", ":80", "Address for the webserver to listen on")
	cmd.Flags().BoolVarP(&debug, "debug", "", false, "Web server debug mode")

	rootCmd.AddCommand(cmd)
}
