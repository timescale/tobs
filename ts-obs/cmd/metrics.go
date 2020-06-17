package cmd

import (
	"github.com/spf13/cobra"
)

// metricsCmd represents the metrics command
var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Subcommand for metrics operations",
}

func init() {
	rootCmd.AddCommand(metricsCmd)
}
