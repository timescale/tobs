package promlens

import (
	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
)

// promlensCmd represents the promlens command
var promlensCmd = &cobra.Command{
	Use:   "promlens",
	Short: "Subcommand for Promlens operations",
}

func init() {
	cmd.RootCmd.AddCommand(promlensCmd)
}
