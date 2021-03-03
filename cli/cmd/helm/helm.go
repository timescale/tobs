package helm

import (
	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
)

// helmCmd represents the helm command
var helmCmd = &cobra.Command{
	Use:   "helm",
	Short: "Subcommand for Helm operations",
}

func init() {
	cmd.RootCmd.AddCommand(helmCmd)
}
