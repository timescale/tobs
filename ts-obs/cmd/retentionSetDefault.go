package cmd

import (
	"context"
	"fmt"

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
	retentionSetDefaultCmd.Flags().StringP("user", "U", "postgres", "database user name")
	retentionSetDefaultCmd.Flags().StringP("dbname", "d", "postgres", "database name to connect to")
}

func retentionSetDefault(cmd *cobra.Command, args []string) error {
	var err error

	retention_period := args[0]

	var user string
	user, err = cmd.Flags().GetString("user")
	if err != nil {
		return fmt.Errorf("could not set default retention period: %w", err)
	}

	var dbname string
	dbname, err = cmd.Flags().GetString("dbname")
	if err != nil {
		return fmt.Errorf("could not set default retention period: %w", err)
	}

	pool, err := OpenConnectionToDB(namespace, name, user, dbname, FORWARD_PORT_TSDB)
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
