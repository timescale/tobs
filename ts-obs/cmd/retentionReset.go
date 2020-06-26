package cmd

import (
	"context"
	"fmt"

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
	retentionResetCmd.Flags().StringP("user", "U", "postgres", "database user name")
	retentionResetCmd.Flags().StringP("dbname", "d", "postgres", "database name to connect to")
}

func retentionReset(cmd *cobra.Command, args []string) error {
	var err error

	metric := args[0]

	var user string
	user, err = cmd.Flags().GetString("user")
	if err != nil {
		return fmt.Errorf("could not reset retention period for %v: %w", metric, err)
	}

	var dbname string
	dbname, err = cmd.Flags().GetString("dbname")
	if err != nil {
		return fmt.Errorf("could not reset retention period for %v: %w", metric, err)
	}

	var name string
	name, err = cmd.Flags().GetString("name")
	if err != nil {
		return fmt.Errorf("could not reset retention period for %v: %w", metric, err)
	}

	var namespace string
	namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return fmt.Errorf("could not reset retention period for %v: %w", metric, err)
	}

	pool, err := OpenConnectionToDB(namespace, name, user, dbname, FORWARD_PORT_TSDB)
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
