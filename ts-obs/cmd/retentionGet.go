package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// retentionGetCmd represents the metrics retention get command
var retentionGetCmd = &cobra.Command{
	Use:   "get <metric>",
	Short: "Gets data retention period in days for a specific metric",
	Args:  cobra.ExactArgs(1),
	RunE:  retentionGet,
}

func init() {
	retentionCmd.AddCommand(retentionGetCmd)
	retentionGetCmd.Flags().StringP("user", "U", "postgres", "database user name")
	retentionGetCmd.Flags().StringP("dbname", "d", "postgres", "database name to connect to")
}

func retentionGet(cmd *cobra.Command, args []string) error {
	var err error

	metric := args[0]

	var user string
	user, err = cmd.Flags().GetString("user")
	if err != nil {
		return fmt.Errorf("could not get retention period for %v: %w", metric, err)
	}

	var dbname string
	dbname, err = cmd.Flags().GetString("dbname")
	if err != nil {
		return fmt.Errorf("could not get retention period for %v: %w", metric, err)
	}

	var name string
	name, err = cmd.Flags().GetString("name")
	if err != nil {
		return fmt.Errorf("could not get retention period for %v: %w", metric, err)
	}

	var namespace string
	namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return fmt.Errorf("could not get retention period for %v: %w", metric, err)
	}

	pool, err := OpenConnectionToDB(namespace, name, user, dbname, FORWARD_PORT_TSDB)
	if err != nil {
		return fmt.Errorf("could not get retention period for %v: %w", metric, err)
	}
	defer pool.Close()

	fmt.Printf("Getting retention period for %v\n", metric)
	var secs int
	err = pool.QueryRow(context.Background(), "SELECT EXTRACT(epoch FROM _prom_catalog.get_metric_retention_period($1))", metric).Scan(&secs)
	if err != nil {
		return fmt.Errorf("could not get retention period for %v: %w", metric, err)
	}

	retention := time.Duration(secs) * time.Second
	fmt.Println(retention)

	return nil
}
