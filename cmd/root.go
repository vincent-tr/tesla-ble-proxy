package cmd

import (
	"os"
	"tesla-ble-proxy/pkg/web"

	"github.com/spf13/cobra"
)

var address string
var debug bool

var rootCmd = &cobra.Command{
	Use:   "tesla-ble-proxy",
	Short: "tesla-ble-proxy - Tesla BLE Rest API (proxy)",
	Run:   run,
}

func init() {
	rootCmd.Flags().StringVarP(&address, "address", "a", ":80", "Address for the webserver to listen on")
	rootCmd.Flags().BoolVarP(&debug, "debug", "", false, "Web server debug mode")
}

func run(_ *cobra.Command, _ []string) {

	web.Serve(address, debug)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
