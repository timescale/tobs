package grafana

import (
	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
)

// grafanaCmd represents the grafana command
var grafanaCmd = &cobra.Command{
	Use:   "grafana",
	Short: "Subcommand for Grafana operations",
}

func init() {
	cmd.RootCmd.AddCommand(grafanaCmd)
}
