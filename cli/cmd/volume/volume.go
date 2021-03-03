package volume

import (
	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
)

// volumeCmd represents the volume command
var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Subcommand for Volume operations",
}

func init() {
	cmd.RootCmd.AddCommand(volumeCmd)
}
