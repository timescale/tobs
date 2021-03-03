package metrics

import (
	"fmt"

	"github.com/spf13/cobra"
	root "github.com/timescale/tobs/cli/cmd"
)

var user string
var dbname string

// metricsCmd represents the metrics command
var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Subcommand for metrics operations",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error

		err = root.RootCmd.PersistentPreRunE(cmd, args)
		if err != nil {
			return fmt.Errorf("could not read global flag: %w", err)
		}

		user, err = cmd.Flags().GetString("user")
		if err != nil {
			return fmt.Errorf("could not read flag: %w", err)
		}

		dbname, err = cmd.Flags().GetString("dbname")
		if err != nil {
			return fmt.Errorf("could not read flag: %w", err)
		}

		return nil
	},
}

func init() {
	root.RootCmd.AddCommand(metricsCmd)
	metricsCmd.PersistentFlags().StringP("user", "U", "", "database user name")
	metricsCmd.PersistentFlags().StringP("dbname", "d", "postgres", "database name to connect to")
}
