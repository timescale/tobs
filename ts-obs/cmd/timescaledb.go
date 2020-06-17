package cmd

import (
	"github.com/spf13/cobra"
)

// timescaledbCmd represents the timescaledb command
var timescaledbCmd = &cobra.Command{
	Use:   "timescaledb",
	Short: "Subcommand for TimescaleDB/PostgreSQL operations",
}

func init() {
	rootCmd.AddCommand(timescaledbCmd)
}
