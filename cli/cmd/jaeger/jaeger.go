package jaeger

import (
	"github.com/spf13/cobra"
	"github.com/timescale/tobs/cli/cmd"
)

// jaegerCmd represents the jaeger command
var jaegerCmd = &cobra.Command{
	Use:   "jaeger",
	Short: "Subcommand for Jaeger operations",
}

func init() {
	cmd.RootCmd.AddCommand(jaegerCmd)
}
