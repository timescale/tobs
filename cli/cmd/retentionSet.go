package cmd

import (
	"context"
	"fmt"

	"github.com/timescale/tobs/cli/pkg/pgconn"

	"github.com/spf13/cobra"
)

// retentionSetCmd represents the metrics retention set command
var retentionSetCmd = &cobra.Command{
	Use:   "set <metric> <days>",
	Short: "Sets data retention period in days for a specific metric",
	Args:  cobra.ExactArgs(2),
	RunE:  retentionSet,
}

func init() {
	retentionCmd.AddCommand(retentionSetCmd)
}

func retentionSet(cmd *cobra.Command, args []string) error {
	var err error

	metric := args[0]
	retention_period := args[1]

	pool, err := pgconn.OpenConnectionToDB(namespace, name, user, dbname, FORWARD_PORT_TSDB)
	if err != nil {
		return fmt.Errorf("could not set retention period for %v: %w", metric, err)
	}
	defer pool.Close()

	fmt.Printf("Setting retention period for %v to %v days\n", metric, retention_period)
	_, err = pool.Exec(context.Background(), "SELECT prom_api.set_metric_retention_period($1, INTERVAL '1 day' * $2)", metric, retention_period)
	if err != nil {
		return fmt.Errorf("could not set retention period for %v: %w", metric, err)
	}

	return nil
}
