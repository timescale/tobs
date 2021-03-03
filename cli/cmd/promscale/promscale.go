package promscale

import (
	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
)

// promscaleCmd represents the promscale command
var promscaleCmd = &cobra.Command{
	Use:   "promscale",
	Short: "Subcommand for Promscale operations",
}

func init() {
	cmd.RootCmd.AddCommand(promscaleCmd)
}
