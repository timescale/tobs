package cmd

import (
	"github.com/spf13/cobra"
)

// grafanaCmd represents the grafana command
var promlensCmd = &cobra.Command{
	Use:   "promlens",
	Short: "Subcommand for Grafana operations",
}

func init() {
	rootCmd.AddCommand(promlensCmd)
}
