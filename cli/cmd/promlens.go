package cmd

import (
	"github.com/spf13/cobra"
)

// promlensCmd represents the promlens command
var promlensCmd = &cobra.Command{
	Use:   "promlens",
	Short: "Subcommand for Promlens operations",
}

func init() {
	rootCmd.AddCommand(promlensCmd)
}
