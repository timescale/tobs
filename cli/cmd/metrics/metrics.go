package metrics

import (
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
)

// metricsCmd represents the metrics command
var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Subcommand for metrics operations",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := root.RootCmd.PersistentPreRunE(cmd, args)
		if err != nil {
			return fmt.Errorf("could not read global flag: %w", err)
		}

		return nil
	},
}

func init() {
	root.RootCmd.AddCommand(metricsCmd)
}
