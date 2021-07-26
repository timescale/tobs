package superuser

import (
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
	"github.com/timescale/tobs/cli/cmd/timescaledb"
)

// Cmd represents the timescaledb superuser command
var superuserCmd = &cobra.Command{
	Use:   "superuser",
	Short: "Subcommand for TimescaleDB super-user operations",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := root.RootCmd.PersistentPreRunE(cmd, args)
		if err != nil {
			return fmt.Errorf("could not read global flag: %w", err)
		}

		return nil
	},
}

func init() {
	timescaledb.Cmd.AddCommand(superuserCmd)
}
