package prometheus

import (
	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
)

// prometheusCmd represents the prometheus command
var prometheusCmd = &cobra.Command{
	Use:   "prometheus",
	Short: "Subcommand for Prometheus operations",
}

func init() {
	cmd.RootCmd.AddCommand(prometheusCmd)
}
