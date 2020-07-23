package cmd

import (
	"github.com/spf13/cobra"
)

// prometheusCmd represents the prometheus command
var prometheusCmd = &cobra.Command{
	Use:   "prometheus",
	Short: "Subcommand for Prometheus operations",
}

func init() {
	rootCmd.AddCommand(prometheusCmd)
}
