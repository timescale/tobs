package cmd

import (
	"context"
	"fmt"

	"github.com/timescale/tobs/cli/pkg/pgconn"

	"github.com/spf13/cobra"
)

// retentionResetCmd represents the retention reset command
var retentionResetCmd = &cobra.Command{
	Use:   "reset <metric>",
	Short: "Resets data retention period to default for a specific metric",
	Args:  cobra.ExactArgs(1),
	RunE:  retentionReset,
}

func init() {
	retentionCmd.AddCommand(retentionResetCmd)
}

func retentionReset(cmd *cobra.Command, args []string) error {
	var err error

	metric := args[0]

	pool, err := pgconn.OpenConnectionToDB(namespace, name, user, dbname, FORWARD_PORT_TSDB)
	if err != nil {
		return fmt.Errorf("could not reset retention period for %v: %w", metric, err)
	}
	defer pool.Close()

	fmt.Printf("Resetting retention period for %v back to default\n", metric)
	_, err = pool.Exec(context.Background(), "SELECT prom_api.reset_metric_retention_period($1)", metric)
	if err != nil {
		return fmt.Errorf("could not reset retention period for %v: %w", metric, err)
	}

	return nil

}
