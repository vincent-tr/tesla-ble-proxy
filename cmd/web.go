package cmd

import (
	"tesla-ble-proxy/pkg/web"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "web",
		Short: "Starts the web server",
		Run: func(_ *cobra.Command, _ []string) {

			web.Serve()
		},
	})
}
