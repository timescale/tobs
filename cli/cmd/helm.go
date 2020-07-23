package cmd

import (
	"github.com/spf13/cobra"
)

// helmCmd represents the helm command
var helmCmd = &cobra.Command{
	Use:   "helm",
	Short: "Subcommand for Helm operations",
}

func init() {
	rootCmd.AddCommand(helmCmd)
}
