package cmd

import (
	"context"
	"fmt"

	"github.com/timescale/tobs/cli/pkg/pgconn"

	"github.com/spf13/cobra"
)

// retentionSetDefaultCmd represents the retention set-default command
var retentionSetDefaultCmd = &cobra.Command{
	Use:   "set-default <days>",
	Short: "Sets default data retention period in days",
	Args:  cobra.ExactArgs(1),
	RunE:  retentionSetDefault,
}

func init() {
	retentionCmd.AddCommand(retentionSetDefaultCmd)
}

func retentionSetDefault(cmd *cobra.Command, args []string) error {
	var err error

	retention_period := args[0]

	pool, err := pgconn.OpenConnectionToDB(namespace, name, user, dbname, FORWARD_PORT_TSDB)
	if err != nil {
		return fmt.Errorf("could not set default retention period: %w", err)
	}
	defer pool.Close()

	fmt.Printf("Setting default retention period to %v days\n", retention_period)
	_, err = pool.Exec(context.Background(), "SELECT prom_api.set_default_retention_period(INTERVAL '1 day' * $1)", retention_period)
	if err != nil {
		return fmt.Errorf("could not set default retention period: %w", err)
	}

	return nil
}
