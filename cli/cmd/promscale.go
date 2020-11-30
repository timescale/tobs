package cmd

import (
"github.com/spf13/cobra"
)

// promscaleCmd represents the promscale command
var promscaleCmd = &cobra.Command{
	Use:   "promscale",
	Short: "Subcommand for Promscale operations",
}

func init() {
	rootCmd.AddCommand(promscaleCmd)
}

