package cmd

import (
	"github.com/spf13/cobra"
)

// grafanaCmd represents the grafana command
var grafanaCmd = &cobra.Command{
	Use:   "grafana",
	Short: "Subcommand for Grafana operations",
}

func init() {
	rootCmd.AddCommand(grafanaCmd)
}
