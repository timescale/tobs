package cmd

import (
	"context"
	"fmt"

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
	retentionSetCmd.Flags().StringP("user", "U", "postgres", "database user name")
	retentionSetCmd.Flags().StringP("dbname", "d", "postgres", "database name to connect to")
}

func retentionSet(cmd *cobra.Command, args []string) error {
	var err error

	metric := args[0]
	retention_period := args[1]

	var user string
	user, err = cmd.Flags().GetString("user")
	if err != nil {
		return fmt.Errorf("could not set retention period for %v: %w", metric, err)
	}

	var dbname string
	dbname, err = cmd.Flags().GetString("dbname")
	if err != nil {
		return fmt.Errorf("could not set retention period for %v: %w", metric, err)
	}

	var name string
	name, err = cmd.Flags().GetString("name")
	if err != nil {
		return fmt.Errorf("could not set retention period for %v: %w", metric, err)
	}

	var namespace string
	namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return fmt.Errorf("could not set retention period for %v: %w", metric, err)
	}

	pool, err := OpenConnectionToDB(namespace, name, user, dbname, FORWARD_PORT_TSDB)
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
